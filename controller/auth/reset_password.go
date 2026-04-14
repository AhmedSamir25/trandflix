package auth

import (
	cryptorand "crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"trendflix/database"
	"trendflix/models"
	"trendflix/utils"
)

type resetPasswordEmailRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

func ResetPasswordRequest(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "If the email exists, a reset code has been sent",
	}

	if database.DbConn == nil {
		log.Println("database connection is not initialized")
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	var request resetPasswordEmailRequest
	if err := c.BodyParser(&request); err != nil {
		log.Printf("Error parsing request body: %v", err)
		context["statusText"] = "bad"
		context["msg"] = "Invalid request body"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	if request.Email == "" {
		context["statusText"] = "bad"
		context["msg"] = "Email is required"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	var user models.User
	result := database.DbConn.Where("email = ?", request.Email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusOK).JSON(context)
		}
		log.Println("Error querying database:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	resetCode, err := generateResetCode()
	if err != nil {
		log.Println("Error generating reset code:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error generating reset code"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	expiresAt := time.Now().Add(resetTokenDuration())
	err = database.DbConn.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.ResetToken{}).Error; err != nil {
			return err
		}

		resetToken := models.ResetToken{
			UserID:    user.ID,
			Code:      resetCode,
			ExpiresAt: expiresAt,
		}

		return tx.Create(&resetToken).Error
	})
	if err != nil {
		log.Println("Error saving reset token:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error saving reset token"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	if err := utils.SendEmail(user.Email, "Password Reset Code", resetCode); err != nil {
		log.Println("Error sending email:", err)
		if deleteErr := database.DbConn.Where("user_id = ? AND code = ?", user.ID, resetCode).Delete(&models.ResetToken{}).Error; deleteErr != nil {
			log.Println("Error deleting reset token after email failure:", deleteErr)
		}
		context["statusText"] = "bad"
		context["msg"] = "Error sending email"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	return c.Status(fiber.StatusOK).JSON(context)
}

func ResetPassword(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Password reset successful",
	}

	if database.DbConn == nil {
		log.Println("database connection is not initialized")
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	var request resetPasswordRequest
	if err := c.BodyParser(&request); err != nil {
		log.Printf("Error parsing request body: %v", err)
		context["statusText"] = "bad"
		context["msg"] = "Invalid request"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	request.Code = strings.TrimSpace(request.Code)
	request.NewPassword = strings.TrimSpace(request.NewPassword)

	if request.Email == "" || request.Code == "" || request.NewPassword == "" {
		context["statusText"] = "bad"
		context["msg"] = "Email, code and new password are required"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	var user models.User
	result := database.DbConn.Where("email = ?", request.Email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "Invalid or expired reset code"
			return c.Status(fiber.StatusBadRequest).JSON(context)
		}
		log.Println("Error querying database:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	var resetToken models.ResetToken
	result = database.DbConn.Where("user_id = ? AND code = ? AND expires_at > ?", user.ID, request.Code, time.Now()).First(&resetToken)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "Invalid or expired reset code"
			return c.Status(fiber.StatusBadRequest).JSON(context)
		}
		log.Println("Error querying reset token:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error hashing password:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error hashing password"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	err = database.DbConn.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
			return err
		}

		return tx.Where("user_id = ?", user.ID).Delete(&models.ResetToken{}).Error
	})
	if err != nil {
		log.Println("Error resetting password:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error updating password"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	return c.Status(fiber.StatusOK).JSON(context)
}

func generateResetCode() (string, error) {
	max := big.NewInt(1000000)
	value, err := cryptorand.Int(cryptorand.Reader, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", value.Int64()), nil
}

func resetTokenDuration() time.Duration {
	value := strings.TrimSpace(os.Getenv("RESET_TOKEN_EXPIRE_MINUTES"))
	if value == "" {
		return 15 * time.Minute
	}

	minutes, err := strconv.Atoi(value)
	if err != nil || minutes <= 0 {
		return 15 * time.Minute
	}

	return time.Duration(minutes) * time.Minute
}

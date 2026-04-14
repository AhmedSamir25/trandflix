package middleware

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"trendflix/database"
	"trendflix/models"
)

func Authenticate(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "bad",
		"msg":        "Unauthorized",
	}

	if database.DbConn == nil {
		log.Println("database connection is not initialized")
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	authorizationHeader := strings.TrimSpace(c.Get("Authorization"))
	if authorizationHeader == "" || !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(context)
	}

	tokenString := strings.TrimSpace(strings.TrimPrefix(authorizationHeader, "Bearer "))
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(context)
	}

	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		log.Println("JWT_SECRET is not configured")
		context["msg"] = "Authentication error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(context)
	}

	subject, ok := claims["sub"].(string)
	if !ok || strings.TrimSpace(subject) == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(context)
	}

	userID, err := strconv.ParseUint(subject, 10, 64)
	if err != nil || userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(context)
	}

	var user models.User
	result := database.DbConn.First(&user, userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusUnauthorized).JSON(context)
		}

		log.Println("Error querying authenticated user:", result.Error)
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	c.Locals("currentUser", user)
	return c.Next()
}

func RequireAdmin(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "bad",
		"msg":        "Forbidden",
	}

	userValue := c.Locals("currentUser")
	user, ok := userValue.(models.User)
	if !ok {
		context["msg"] = "Unauthorized"
		return c.Status(fiber.StatusUnauthorized).JSON(context)
	}

	if strings.TrimSpace(user.Role) != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(context)
	}

	return c.Next()
}

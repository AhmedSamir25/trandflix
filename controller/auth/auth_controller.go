package auth

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"trendflix/database"
	"trendflix/models"
)

type createUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Avatar   string `json:"avatar"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func CreateUser(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Create User",
	}

	if database.DbConn == nil {
		log.Println("database connection is not initialized")
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	var request createUserRequest

	if err := c.BodyParser(&request); err != nil {
		log.Printf("Error parsing request body: %v", err)
		context["statusText"] = "bad"
		context["msg"] = "Invalid request"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	request.Name = strings.TrimSpace(request.Name)
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	request.Password = strings.TrimSpace(request.Password)
	request.Avatar = strings.TrimSpace(request.Avatar)

	if request.Name == "" || request.Email == "" || request.Password == "" {
		context["statusText"] = "bad"
		context["msg"] = "Name, email and password are required"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	var existingUser models.User
	result := database.DbConn.Where("email = ?", request.Email).First(&existingUser)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Println("Error querying database:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}
	if result.RowsAffected > 0 {
		context["statusText"] = "bad"
		context["msg"] = "Email already exists"
		return c.Status(fiber.StatusConflict).JSON(context)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error hashing password:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error hashing password"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	record := models.User{
		Name:     request.Name,
		Email:    request.Email,
		Password: string(hashedPassword),
		Avatar:   request.Avatar,
		Role:     "user",
	}

	result = database.DbConn.Create(&record)
	if result.Error != nil {
		log.Println("Error in saving data:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Error in saving user"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	token, err := generateAuthToken(record)
	if err != nil {
		log.Println("Error generating token:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error generating token"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["id"] = record.ID
	context["token"] = token
	context["msg"] = "User created successfully"
	return c.Status(fiber.StatusCreated).JSON(context)
}

func LoginUser(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Login Successful",
	}

	if database.DbConn == nil {
		log.Println("database connection is not initialized")
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	var request loginRequest
	if err := c.BodyParser(&request); err != nil {
		log.Printf("Error parsing request body: %v", err)
		context["statusText"] = "bad"
		context["msg"] = "Invalid request"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	request.Password = strings.TrimSpace(request.Password)

	if request.Email == "" || request.Password == "" {
		context["statusText"] = "bad"
		context["msg"] = "Email and password are required"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	var user models.User
	result := database.DbConn.Where("email = ?", request.Email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "Invalid email or password"
			return c.Status(fiber.StatusUnauthorized).JSON(context)
		}
		log.Println("Error querying database:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		context["statusText"] = "bad"
		context["msg"] = "Invalid email or password"
		return c.Status(fiber.StatusUnauthorized).JSON(context)
	}
	if err != nil {
		log.Println("Error comparing password:", err)
		context["statusText"] = "bad"
		context["msg"] = "Authentication error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	token, err := generateAuthToken(user)
	if err != nil {
		log.Println("Error generating token:", err)
		context["statusText"] = "bad"
		context["msg"] = "Authentication error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["id"] = user.ID
	context["role"] = user.Role
	context["token"] = token
	context["msg"] = "Login Successful"
	return c.Status(fiber.StatusOK).JSON(context)
}

func generateAuthToken(user models.User) (string, error) {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		return "", errors.New("JWT_SECRET is not configured")
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  strconv.FormatUint(uint64(user.ID), 10),
		"role": user.Role,
		"iat":  now.Unix(),
		"exp":  now.Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

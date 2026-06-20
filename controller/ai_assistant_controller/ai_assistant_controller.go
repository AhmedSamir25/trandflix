package aiassistantcontroller

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"

	"trendflix/database"
	"trendflix/models"
	"trendflix/services"
)

type recommendRequest struct {
	UserMessage string   `json:"user_message"`
	Type        string   `json:"type"`
	Categories  []string `json:"categories"`
	Mood        string   `json:"mood"`
	Language    string   `json:"language"`
	Limit       int      `json:"limit"`
}

func Recommend(c *fiber.Ctx) error {
	contextMap := fiber.Map{
		"statusText": "Ok",
		"msg":        "Recommendations generated successfully",
	}

	if database.DbConn == nil {
		log.Println("database connection is not initialized")
		contextMap["statusText"] = "bad"
		contextMap["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(contextMap)
	}

	var request recommendRequest
	if err := c.BodyParser(&request); err != nil {
		contextMap["statusText"] = "bad"
		contextMap["msg"] = "Invalid request"
		return c.Status(fiber.StatusBadRequest).JSON(contextMap)
	}

	message := strings.TrimSpace(request.UserMessage)
	if message == "" {
		contextMap["statusText"] = "bad"
		contextMap["msg"] = "user_message is required"
		return c.Status(fiber.StatusBadRequest).JSON(contextMap)
	}

	if len([]rune(message)) > 800 {
		contextMap["statusText"] = "bad"
		contextMap["msg"] = "user_message is too long"
		return c.Status(fiber.StatusBadRequest).JSON(contextMap)
	}

	user, _ := currentUserFromContext(c)

	repo := &services.GormAIRepository{DB: database.DbConn}
	client := services.NewOpenRouterRecommenderWithSettings(services.LoadAISettingsConfig(database.DbConn))

	service := services.NewAIRecommendationService(repo, client)

	serviceRequest := services.AIRecommendationRequest{
		UserMessage: message,
		Type:        request.Type,
		Categories:  request.Categories,
		Mood:        request.Mood,
		Language:    request.Language,
		Limit:       request.Limit,
	}

	ctx, cancel := context.WithTimeout(c.Context(), services.AIRequestTimeout)
	defer cancel()

	result, err := service.Recommend(ctx, serviceRequest)
	if err != nil {
		log.Println("AI assistant recommend error:", err)
		contextMap["statusText"] = "bad"
		if errors.Is(err, services.ErrAIModelUnavailablePublic) {
			contextMap["msg"] = "AI assistant is unavailable right now"
			return c.Status(fiber.StatusServiceUnavailable).JSON(contextMap)
		}
		contextMap["msg"] = "Unable to generate recommendations right now"
		return c.Status(fiber.StatusServiceUnavailable).JSON(contextMap)
	}

	contextMap["success"] = true
	contextMap["data"] = result
	_ = user
	return c.Status(fiber.StatusOK).JSON(contextMap)
}

func currentUserFromContext(c *fiber.Ctx) (models.User, bool) {
	userValue := c.Locals("currentUser")
	user, ok := userValue.(models.User)
	return user, ok
}

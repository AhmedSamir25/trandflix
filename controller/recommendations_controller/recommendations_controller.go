package recommendationscontroller

import (
	"context"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"

	"trendflix/database"
	"trendflix/models"
	"trendflix/services"
)

func GetRecommendationsForYou(c *fiber.Ctx) error {
	response := fiber.Map{
		"statusText": "Ok",
		"msg":        "Recommendations fetched successfully",
	}

	if database.DbConn == nil {
		log.Println("database connection is not initialized")
		response["statusText"] = "bad"
		response["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}

	user, err := currentUserFromContext(c, response)
	if err != nil {
		return err
	}

	limit := normalizeLimit(c.QueryInt("limit", services.DefaultLimit))
	language := strings.TrimSpace(c.Query("language"))

	repo := &services.GormRepository{DB: database.DbConn}
	client := services.NewOpenRouterRecommenderWithSettings(services.LoadAISettingsConfig(database.DbConn))
	service := services.NewAIForYouService(repo, client)

	ctx, cancel := context.WithTimeout(c.Context(), services.AIForYouRequestTimeout)
	defer cancel()

	recommendations, err := service.ForUser(ctx, user.ID, language, limit)
	if err != nil {
		log.Println("Error building recommendations:", err)
		response["statusText"] = "bad"
		response["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}

	response["success"] = true
	response["data"] = recommendations
	return c.Status(fiber.StatusOK).JSON(response)
}

func currentUserFromContext(c *fiber.Ctx, response fiber.Map) (models.User, error) {
	userValue := c.Locals("currentUser")
	user, ok := userValue.(models.User)
	if !ok {
		response["statusText"] = "bad"
		response["msg"] = "Unauthorized"
		return models.User{}, c.Status(fiber.StatusUnauthorized).JSON(response)
	}

	return user, nil
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return services.DefaultLimit
	}
	if limit > services.MaxLimit {
		return services.MaxLimit
	}
	return limit
}

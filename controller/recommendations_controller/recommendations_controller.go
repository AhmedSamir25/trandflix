package recommendationscontroller

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"trendflix/database"
	"trendflix/models"
	"trendflix/services"
)

func GetRecommendationsForYou(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Recommendations fetched successfully",
	}

	if database.DbConn == nil {
		log.Println("database connection is not initialized")
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	user, err := currentUserFromContext(c, context)
	if err != nil {
		return err
	}

	limit := normalizeLimit(c.QueryInt("limit", services.DefaultLimit))

	repo := &services.GormRepository{DB: database.DbConn}
	service := services.NewRecommendationService(repo)

	recommendations, err := service.ForUser(user.ID, limit)
	if err != nil {
		log.Println("Error building recommendations:", err)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["success"] = true
	context["data"] = recommendations
	return c.Status(fiber.StatusOK).JSON(context)
}

func currentUserFromContext(c *fiber.Ctx, context fiber.Map) (models.User, error) {
	userValue := c.Locals("currentUser")
	user, ok := userValue.(models.User)
	if !ok {
		context["statusText"] = "bad"
		context["msg"] = "Unauthorized"
		return models.User{}, c.Status(fiber.StatusUnauthorized).JSON(context)
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

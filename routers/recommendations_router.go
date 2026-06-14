package routers

import (
	"github.com/gofiber/fiber/v2"

	recommendationscontroller "trendflix/controller/recommendations_controller"
	"trendflix/middleware"
)

func RegisterRecommendationRoutes(app *fiber.App) {
	recommendations := app.Group("/api/recommendations", middleware.Authenticate)
	recommendations.Get("/for-you", recommendationscontroller.GetRecommendationsForYou)
}

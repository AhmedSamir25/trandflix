package routers

import (
	"github.com/gofiber/fiber/v2"

	aiassistantcontroller "trendflix/controller/ai_assistant_controller"
	"trendflix/middleware"
)

func RegisterAIAssistantRoutes(app *fiber.App) {
	assistant := app.Group("/api/ai-assistant", middleware.Authenticate)
	assistant.Post("/recommend", aiassistantcontroller.Recommend)
}

package routers

import (
	"github.com/gofiber/fiber/v2"

	chatcontroller "trendflix/controller/chat_controller"
	"trendflix/middleware"
)

func RegisterChatRoutes(app *fiber.App) {
	chat := app.Group("/chat", middleware.Authenticate)
	chat.Post("/trendflix", chatcontroller.Reply)
}

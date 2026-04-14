package routers

import (
	"github.com/gofiber/fiber/v2"

	authcontroller "trendflix/controller/auth"
)

func RegisterAuthRoutes(app *fiber.App) {
	authGroup := app.Group("/auth")

	authGroup.Post("/register", authcontroller.CreateUser)
	authGroup.Post("/login", authcontroller.LoginUser)
	authGroup.Post("/reset-password/request", authcontroller.ResetPasswordRequest)
	authGroup.Post("/reset-password", authcontroller.ResetPassword)
}

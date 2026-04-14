package main

import (
	"os"

	"github.com/gofiber/fiber/v2"

	"trendflix/database"
	"trendflix/routers"
)

func main() {
	database.ConnDB()

	app := fiber.New()

	routers.RegisterAuthRoutes(app)
	routers.RegisterCategoryRoutes(app)
	routers.RegisterItemRoutes(app)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	if err := app.Listen(":" + port); err != nil {
		panic(err)
	}
}

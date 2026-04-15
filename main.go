package main

import (
	"os"

	"github.com/gofiber/fiber/v2"

	"trendflix/database"
	"trendflix/routers"
)

func main() {
	database.ConnDB()
	database.Migrate()

	app := fiber.New()
	app.Static("/upload", "./upload")

	routers.RegisterAuthRoutes(app)
	routers.RegisterCategoryRoutes(app)
	routers.RegisterFavoriteRoutes(app)
	routers.RegisterItemRoutes(app)
	routers.RegisterReviewRoutes(app)
	routers.RegisterUploadRoutes(app)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	if err := app.Listen(":" + port); err != nil {
		panic(err)
	}
}

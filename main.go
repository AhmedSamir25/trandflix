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
	database.SeedAdmin()
	database.SeedCategories()

	app := fiber.New()
	app.Static("/upload", "./upload")

	routers.RegisterAuthRoutes(app)
	routers.RegisterChatRoutes(app)
	routers.RegisterCategoryRoutes(app)
	routers.RegisterFavoriteRoutes(app)
	routers.RegisterItemRoutes(app)
	routers.RegisterReviewRoutes(app)
	routers.RegisterUploadRoutes(app)
	routers.RegisterViewRoutes(app)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "4000"
	}

	if err := app.Listen(":" + port); err != nil {
		panic(err)
	}
}

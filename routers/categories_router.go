package routers

import (
	"github.com/gofiber/fiber/v2"

	categoriescontroller "trendflix/controller/categories_controller"
	"trendflix/middleware"
)

func RegisterCategoryRoutes(app *fiber.App) {
	categories := app.Group("/categories")
	categories.Get("", categoriescontroller.GetCategories)

	adminCategories := app.Group("/categories", middleware.Authenticate, middleware.RequireAdmin)
	adminCategories.Post("", categoriescontroller.CreateCategory)
	adminCategories.Put("/:id", categoriescontroller.UpdateCategory)
	adminCategories.Delete("/:id", categoriescontroller.DeleteCategory)
}

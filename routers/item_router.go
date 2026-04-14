package routers

import (
	"github.com/gofiber/fiber/v2"

	itemcontroller "trendflix/controller/item_controller"
	"trendflix/middleware"
)

func RegisterItemRoutes(app *fiber.App) {
	items := app.Group("/items")
	items.Get("", itemcontroller.GetItems)
	items.Get("/:id", itemcontroller.GetItemByID)

	adminItems := app.Group("/items", middleware.Authenticate, middleware.RequireAdmin)
	adminItems.Post("", itemcontroller.CreateItem)
	adminItems.Put("/:id", itemcontroller.UpdateItem)
	adminItems.Delete("/:id", itemcontroller.DeleteItem)
}

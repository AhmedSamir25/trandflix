package routers

import (
	"github.com/gofiber/fiber/v2"

	listscontroller "trendflix/controller/lists_controller"
	"trendflix/middleware"
)

func RegisterListRoutes(app *fiber.App) {
	lists := app.Group("/lists", middleware.Authenticate)
	lists.Get("", listscontroller.GetLists)
	lists.Post("", listscontroller.CreateList)
	lists.Get("/:list_id", listscontroller.GetListItems)
	lists.Post("/:list_id/items/:item_id", listscontroller.AddItemToList)
	lists.Delete("/:list_id/items/:item_id", listscontroller.RemoveItemFromList)
	lists.Delete("/:list_id", listscontroller.DeleteList)
}

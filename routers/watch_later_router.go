package routers

import (
	"github.com/gofiber/fiber/v2"

	watchlatercontroller "trendflix/controller/watch_later_controller"
	"trendflix/middleware"
)

func RegisterWatchLaterRoutes(app *fiber.App) {
	watchLater := app.Group("/watch-later", middleware.Authenticate)
	watchLater.Get("", watchlatercontroller.GetWatchLater)
	watchLater.Post("/:item_id", watchlatercontroller.AddWatchLater)
	watchLater.Delete("/:item_id", watchlatercontroller.RemoveWatchLater)
}

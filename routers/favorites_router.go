package routers

import (
	"github.com/gofiber/fiber/v2"

	favoritescontroller "trendflix/controller/favorites_controller"
	"trendflix/middleware"
)

func RegisterFavoriteRoutes(app *fiber.App) {
	favorites := app.Group("/favorites", middleware.Authenticate)
	favorites.Get("", favoritescontroller.GetFavorites)
	favorites.Post("/:item_id", favoritescontroller.AddFavorite)
	favorites.Delete("/:item_id", favoritescontroller.RemoveFavorite)
}

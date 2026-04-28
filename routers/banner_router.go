package routers

import (
	"github.com/gofiber/fiber/v2"

	bannercontroller "trendflix/controller/banner_controller"
	"trendflix/middleware"
)

func RegisterBannerRoutes(app *fiber.App) {
	app.Get("/banners", bannercontroller.GetActiveBanners)

	admin := app.Group("/banners", middleware.Authenticate, middleware.RequireAdmin)
	admin.Get("/all", bannercontroller.GetAllBanners)
	admin.Post("/", bannercontroller.CreateBanner)
	admin.Put("/:id", bannercontroller.UpdateBanner)
	admin.Delete("/:id", bannercontroller.DeleteBanner)
}

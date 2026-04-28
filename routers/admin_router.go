package routers

import (
	"github.com/gofiber/fiber/v2"

	admincontroller "trendflix/controller/admin_controller"
	"trendflix/middleware"
)

func RegisterAdminRoutes(app *fiber.App) {
	admin := app.Group("/admin", middleware.Authenticate, middleware.RequireAdmin)
	admin.Get("/stats", admincontroller.GetStats)
	admin.Get("/stats/overview", admincontroller.GetOverviewStats)
	admin.Get("/stats/types", admincontroller.GetTypeStats)
	admin.Get("/stats/categories", admincontroller.GetCategoryStats)
	admin.Get("/stats/users", admincontroller.GetUserStats)
}

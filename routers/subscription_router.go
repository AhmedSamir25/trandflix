package routers

import (
	"github.com/gofiber/fiber/v2"

	subscriptioncontroller "trendflix/controller/subscription_controller"
	"trendflix/middleware"
)

func RegisterSubscriptionRoutes(app *fiber.App) {
	sub := app.Group("/subscription", middleware.Authenticate)

	sub.Get("/plan", subscriptioncontroller.GetActivePlan)
	sub.Get("/me", subscriptioncontroller.GetMySubscription)
	sub.Post("/checkout", subscriptioncontroller.Checkout)
	sub.Post("/renew", subscriptioncontroller.Renew)
	sub.Post("/cancel", subscriptioncontroller.Cancel)
}

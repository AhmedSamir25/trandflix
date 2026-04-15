package routers

import (
	"github.com/gofiber/fiber/v2"

	reviewscontroller "trendflix/controller/reviews_controller"
	"trendflix/middleware"
)

func RegisterReviewRoutes(app *fiber.App) {
	reviews := app.Group("/reviews")
	reviews.Get("/item/:item_id", reviewscontroller.GetReviewsByItem)

	authReviews := app.Group("/reviews", middleware.Authenticate)
	authReviews.Post("", reviewscontroller.CreateReview)
	authReviews.Put("/:id", reviewscontroller.UpdateReview)
	authReviews.Delete("/:id", reviewscontroller.DeleteReview)
}

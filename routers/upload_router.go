package routers

import (
	"github.com/gofiber/fiber/v2"

	uploadcontroller "trendflix/controller/upload_controller"
	"trendflix/middleware"
)

func RegisterUploadRoutes(app *fiber.App) {
	upload := app.Group("/upload")
	upload.Post("/avatar", uploadcontroller.UploadAvatar)

	adminUpload := app.Group("/upload", middleware.Authenticate, middleware.RequireAdmin)
	adminUpload.Post("/item-image", uploadcontroller.UploadItemImage)
}

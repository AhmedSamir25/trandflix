package routers

import (
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

func RegisterViewRoutes(app *fiber.App) {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	viewDir := filepath.Join(cwd, "view")

	// Frontend entry + static assets
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile(filepath.Join(viewDir, "index.html"))
	})

	// Backward-compatible route (older JS used /pages/auth.html)
	// Must be registered before the /pages static handler.
	app.Get("/pages/auth.html", func(c *fiber.Ctx) error {
		return c.Redirect("/pages/auth/auth.html", fiber.StatusTemporaryRedirect)
	})

	app.Get("/detail/:id", func(c *fiber.Ctx) error {
		return c.SendFile(filepath.Join(viewDir, "pages", "detail.html"))
	})

	app.Static("/assets", filepath.Join(viewDir, "assets"))
	app.Static("/pages", filepath.Join(viewDir, "pages"))
}


package httpx

import (
	"github.com/gofiber/fiber/v2"
	"trendflix/models"
)

func CurrentUser(c *fiber.Ctx) (models.User, bool) {
	value := c.Locals("currentUser")
	user, ok := value.(models.User)
	return user, ok
}

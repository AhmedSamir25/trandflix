package httpx

import "github.com/gofiber/fiber/v2"

func Success(c *fiber.Ctx, status int, message string, data fiber.Map) error {
	if data == nil {
		data = fiber.Map{}
	}
	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

func SuccessData(c *fiber.Ctx, status int, message string, key string, value any) error {
	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    fiber.Map{key: value},
	})
}

func Error(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"message": message,
		"errors":  fiber.Map{},
	})
}

func ValidationError(c *fiber.Ctx, message string, errors fiber.Map) error {
	if errors == nil {
		errors = fiber.Map{}
	}
	return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
		"success": false,
		"message": message,
		"errors":  errors,
	})
}

package subscriptionplancontroller

import (
	"errors"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"trendflix/database"
	"trendflix/models"
)

type updatePlanRequest struct {
	Name              string  `json:"name"`
	Price             float64 `json:"price"`
	Currency          string  `json:"currency"`
	BillingPeriodDays uint    `json:"billing_period_days"`
	IsActive          bool    `json:"is_active"`
}

func GetPlan(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Subscription plan fetched successfully",
	}

	if database.DbConn == nil {
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	if err := requireAdmin(c, context); err != nil {
		return err
	}

	var plan models.SubscriptionPlan
	result := database.DbConn.First(&plan)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "No subscription plan found"
			return c.Status(fiber.StatusNotFound).JSON(context)
		}
		log.Println("Error querying subscription plan:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["plan"] = plan
	return c.Status(fiber.StatusOK).JSON(context)
}

func UpdatePlan(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Subscription plan updated successfully",
	}

	if database.DbConn == nil {
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	if err := requireAdmin(c, context); err != nil {
		return err
	}

	var request updatePlanRequest
	if err := c.BodyParser(&request); err != nil {
		context["statusText"] = "bad"
		context["msg"] = "Invalid request"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	request.Name = strings.TrimSpace(request.Name)
	request.Currency = strings.ToUpper(strings.TrimSpace(request.Currency))

	if request.Name == "" {
		context["statusText"] = "bad"
		context["msg"] = "Plan name is required"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	if request.Price <= 0 {
		context["statusText"] = "bad"
		context["msg"] = "Price must be greater than zero"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	if request.Currency == "" {
		context["statusText"] = "bad"
		context["msg"] = "Currency is required"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	if request.BillingPeriodDays == 0 {
		request.BillingPeriodDays = 30
	}

	var plan models.SubscriptionPlan
	result := database.DbConn.First(&plan)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "No subscription plan found"
			return c.Status(fiber.StatusNotFound).JSON(context)
		}
		log.Println("Error querying subscription plan:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	plan.Name = request.Name
	plan.Price = request.Price
	plan.Currency = request.Currency
	plan.BillingPeriodDays = request.BillingPeriodDays
	plan.IsActive = request.IsActive

	if err := database.DbConn.Save(&plan).Error; err != nil {
		log.Println("Error updating subscription plan:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error updating subscription plan"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["plan"] = plan
	return c.Status(fiber.StatusOK).JSON(context)
}

func requireAdmin(c *fiber.Ctx, context fiber.Map) error {
	userValue := c.Locals("currentUser")
	user, ok := userValue.(models.User)
	if !ok {
		context["statusText"] = "bad"
		context["msg"] = "Unauthorized"
		return c.Status(fiber.StatusUnauthorized).JSON(context)
	}

	if strings.TrimSpace(user.Role) != "admin" {
		context["statusText"] = "bad"
		context["msg"] = "Forbidden"
		return c.Status(fiber.StatusForbidden).JSON(context)
	}

	return nil
}

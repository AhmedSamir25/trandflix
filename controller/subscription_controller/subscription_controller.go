package subscriptioncontroller

import (
	"errors"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"trendflix/database"
	"trendflix/models"
)

type checkoutRequest struct {
	CardName   string `json:"card_name"`
	CardNumber string `json:"card_number"`
	CardExpiry string `json:"card_expiry"`
	CardCVC    string `json:"card_cvc"`
}

func GetActivePlan(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Plan fetched successfully",
	}

	if database.DbConn == nil {
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	var plan models.SubscriptionPlan
	result := database.DbConn.Where("is_active = ?", true).First(&plan)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "No active subscription plan found"
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

func GetMySubscription(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Subscription fetched successfully",
	}

	if database.DbConn == nil {
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	user, err := currentUserFromContext(c, context)
	if err != nil {
		return err
	}

	var subscription models.Subscription
	result := database.DbConn.Preload("Plan").Where("user_id = ?", user.ID).Order("id DESC").First(&subscription)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["subscription"] = nil
			context["is_active"] = false
			return c.Status(fiber.StatusOK).JSON(context)
		}
		log.Println("Error querying subscription:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["subscription"] = subscription
	context["is_active"] = subscription.IsActive()
	return c.Status(fiber.StatusOK).JSON(context)
}

func Checkout(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Subscription created successfully",
	}

	if database.DbConn == nil {
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	user, err := currentUserFromContext(c, context)
	if err != nil {
		return err
	}

	var plan models.SubscriptionPlan
	result := database.DbConn.Where("is_active = ?", true).First(&plan)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "No active subscription plan found"
			return c.Status(fiber.StatusNotFound).JSON(context)
		}
		log.Println("Error querying subscription plan:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	var request checkoutRequest
	if err := c.BodyParser(&request); err != nil {
		context["statusText"] = "bad"
		context["msg"] = "Invalid request"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	if msg, ok := validateMockPayment(request); !ok {
		context["statusText"] = "bad"
		context["msg"] = msg
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	var existing models.Subscription
	checkResult := database.DbConn.Where("user_id = ? AND status = ?", user.ID, models.SubscriptionStatusActive).
		Where("ends_at > ?", time.Now()).
		First(&existing)
	if checkResult.Error == nil {
		context["statusText"] = "bad"
		context["msg"] = "You already have an active subscription"
		return c.Status(fiber.StatusConflict).JSON(context)
	}

	tx := database.DbConn.Begin()
	if tx.Error != nil {
		log.Println("Error starting transaction:", tx.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	now := time.Now()
	subscription := models.Subscription{
		UserID:   user.ID,
		PlanID:   plan.ID,
		Status:   models.SubscriptionStatusActive,
		StartsAt: now,
		EndsAt:   now.AddDate(0, 0, int(plan.BillingPeriodDays)),
	}

	if err := tx.Create(&subscription).Error; err != nil {
		tx.Rollback()
		log.Println("Error saving subscription:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error saving subscription"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	normalized := normalizeCardNumber(request.CardNumber)
	last4 := ""
	if len(normalized) >= 4 {
		last4 = normalized[len(normalized)-4:]
	}

	payment := models.Payment{
		UserID:         user.ID,
		SubscriptionID: &subscription.ID,
		PlanID:         plan.ID,
		Amount:         plan.Price,
		Currency:       plan.Currency,
		Status:         models.PaymentStatusPaid,
		MockCardLast4:  last4,
	}

	if err := tx.Create(&payment).Error; err != nil {
		tx.Rollback()
		log.Println("Error saving payment:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error saving payment"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	if err := tx.Commit().Error; err != nil {
		log.Println("Error committing transaction:", err)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	subscription.Plan = plan
	context["subscription"] = subscription
	context["payment_id"] = payment.ID
	return c.Status(fiber.StatusCreated).JSON(context)
}

func Renew(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Subscription renewed successfully",
	}

	if database.DbConn == nil {
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	user, err := currentUserFromContext(c, context)
	if err != nil {
		return err
	}

	var plan models.SubscriptionPlan
	result := database.DbConn.Where("is_active = ?", true).First(&plan)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "No active subscription plan found"
			return c.Status(fiber.StatusNotFound).JSON(context)
		}
		log.Println("Error querying subscription plan:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	var request checkoutRequest
	if err := c.BodyParser(&request); err != nil {
		context["statusText"] = "bad"
		context["msg"] = "Invalid request"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	if msg, ok := validateMockPayment(request); !ok {
		context["statusText"] = "bad"
		context["msg"] = msg
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	var subscription models.Subscription
	result = database.DbConn.Where("user_id = ?", user.ID).Order("id DESC").First(&subscription)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "No existing subscription found to renew"
			return c.Status(fiber.StatusNotFound).JSON(context)
		}
		log.Println("Error querying subscription:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	now := time.Now()
	newStartsAt := now
	if subscription.Status == models.SubscriptionStatusActive && subscription.EndsAt.After(now) {
		newStartsAt = subscription.EndsAt
	}

	tx := database.DbConn.Begin()
	if tx.Error != nil {
		log.Println("Error starting transaction:", tx.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	subscription.Status = models.SubscriptionStatusActive
	subscription.PlanID = plan.ID
	subscription.StartsAt = newStartsAt
	subscription.EndsAt = newStartsAt.AddDate(0, 0, int(plan.BillingPeriodDays))

	if err := tx.Save(&subscription).Error; err != nil {
		tx.Rollback()
		log.Println("Error renewing subscription:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error renewing subscription"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	normalized := normalizeCardNumber(request.CardNumber)
	last4 := ""
	if len(normalized) >= 4 {
		last4 = normalized[len(normalized)-4:]
	}

	payment := models.Payment{
		UserID:         user.ID,
		SubscriptionID: &subscription.ID,
		PlanID:         plan.ID,
		Amount:         plan.Price,
		Currency:       plan.Currency,
		Status:         models.PaymentStatusPaid,
		MockCardLast4:  last4,
	}

	if err := tx.Create(&payment).Error; err != nil {
		tx.Rollback()
		log.Println("Error saving payment:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error saving payment"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	if err := tx.Commit().Error; err != nil {
		log.Println("Error committing transaction:", err)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	subscription.Plan = plan
	context["subscription"] = subscription
	context["payment_id"] = payment.ID
	return c.Status(fiber.StatusOK).JSON(context)
}

func Cancel(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Subscription cancelled successfully",
	}

	if database.DbConn == nil {
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	user, err := currentUserFromContext(c, context)
	if err != nil {
		return err
	}

	var subscription models.Subscription
	result := database.DbConn.Where("user_id = ? AND status = ?", user.ID, models.SubscriptionStatusActive).
		First(&subscription)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "No active subscription found"
			return c.Status(fiber.StatusNotFound).JSON(context)
		}
		log.Println("Error querying subscription:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	subscription.Status = models.SubscriptionStatusCancelled

	if err := database.DbConn.Save(&subscription).Error; err != nil {
		log.Println("Error cancelling subscription:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error cancelling subscription"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["subscription"] = subscription
	return c.Status(fiber.StatusOK).JSON(context)
}

func currentUserFromContext(c *fiber.Ctx, context fiber.Map) (models.User, error) {
	userValue := c.Locals("currentUser")
	user, ok := userValue.(models.User)
	if !ok {
		context["statusText"] = "bad"
		context["msg"] = "Unauthorized"
		return models.User{}, c.Status(fiber.StatusUnauthorized).JSON(context)
	}

	return user, nil
}

func normalizeCardNumber(card string) string {
	re := regexp.MustCompile(`[\s-]`)
	return re.ReplaceAllString(card, "")
}

func validateMockPayment(request checkoutRequest) (string, bool) {
	request.CardName = strings.TrimSpace(request.CardName)
	request.CardExpiry = strings.TrimSpace(request.CardExpiry)
	request.CardCVC = strings.TrimSpace(request.CardCVC)

	if request.CardName == "" {
		return "Cardholder name is required", false
	}

	normalized := normalizeCardNumber(request.CardNumber)
	if len(normalized) < 12 || len(normalized) > 19 {
		return "Card number must be between 12 and 19 digits", false
	}

	if _, err := strconv.ParseInt(normalized, 10, 64); err != nil {
		return "Card number must contain only digits", false
	}

	if request.CardExpiry == "" {
		return "Card expiry date is required", false
	}

	expPattern := regexp.MustCompile(`^\d{2}/\d{2}$`)
	if !expPattern.MatchString(request.CardExpiry) {
		return "Expiry date must be in MM/YY format", false
	}

	if len(request.CardCVC) < 3 || len(request.CardCVC) > 4 {
		return "CVC must be 3 or 4 digits", false
	}

	if _, err := strconv.Atoi(request.CardCVC); err != nil {
		return "CVC must contain only digits", false
	}

	return "", true
}

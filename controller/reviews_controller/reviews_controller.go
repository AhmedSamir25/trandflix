package reviewscontroller

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"trendflix/database"
	"trendflix/models"
)

type createReviewRequest struct {
	ItemID  uint   `json:"item_id"`
	Rating  uint   `json:"rating"`
	Comment string `json:"comment"`
}

type updateReviewRequest struct {
	Rating  uint   `json:"rating"`
	Comment string `json:"comment"`
}

func GetReviewsByItem(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Reviews fetched successfully",
	}

	if database.DbConn == nil {
		log.Println("database connection is not initialized")
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	itemID, err := parseItemID(c, context)
	if err != nil {
		return err
	}

	if err := ensureItemExists(itemID, context, c); err != nil {
		return err
	}

	var reviews []models.Review
	result := database.DbConn.Where("item_id = ?", itemID).Order("created_at DESC").Find(&reviews)
	if result.Error != nil {
		log.Println("Error querying reviews:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["reviews"] = reviews
	return c.Status(fiber.StatusOK).JSON(context)
}

func CreateReview(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Review created successfully",
	}

	if database.DbConn == nil {
		log.Println("database connection is not initialized")
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	user, err := currentUserFromContext(c, context)
	if err != nil {
		return err
	}

	var request createReviewRequest
	if err := c.BodyParser(&request); err != nil {
		log.Printf("Error parsing request body: %v", err)
		context["statusText"] = "bad"
		context["msg"] = "Invalid request"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	review, msg, statusCode := buildReviewFromCreateRequest(request)
	if statusCode != 0 {
		context["statusText"] = "bad"
		context["msg"] = msg
		return c.Status(statusCode).JSON(context)
	}

	if err := ensureItemExists(review.ItemID, context, c); err != nil {
		return err
	}

	var existingReview models.Review
	result := database.DbConn.Where("user_id = ? AND item_id = ?", user.ID, review.ItemID).First(&existingReview)
	if result.Error == nil {
		context["statusText"] = "bad"
		context["msg"] = "You already reviewed this item"
		return c.Status(fiber.StatusConflict).JSON(context)
	}
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Println("Error querying review:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	review.UserID = user.ID

	if err := database.DbConn.Create(&review).Error; err != nil {
		log.Println("Error saving review:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error saving review"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["id"] = review.ID
	context["review"] = review
	return c.Status(fiber.StatusCreated).JSON(context)
}

func UpdateReview(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Review updated successfully",
	}

	if database.DbConn == nil {
		log.Println("database connection is not initialized")
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	user, err := currentUserFromContext(c, context)
	if err != nil {
		return err
	}

	reviewID, err := parseReviewID(c, context)
	if err != nil {
		return err
	}

	var request updateReviewRequest
	if err := c.BodyParser(&request); err != nil {
		log.Printf("Error parsing request body: %v", err)
		context["statusText"] = "bad"
		context["msg"] = "Invalid request"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	rating, comment, msg, statusCode := normalizeReviewInput(request.Rating, request.Comment)
	if statusCode != 0 {
		context["statusText"] = "bad"
		context["msg"] = msg
		return c.Status(statusCode).JSON(context)
	}

	var review models.Review
	result := database.DbConn.Where("id = ? AND user_id = ?", reviewID, user.ID).First(&review)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "Review not found"
			return c.Status(fiber.StatusNotFound).JSON(context)
		}

		log.Println("Error querying review:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	review.Rating = rating
	review.Comment = comment

	if err := database.DbConn.Save(&review).Error; err != nil {
		log.Println("Error updating review:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error updating review"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["review"] = review
	return c.Status(fiber.StatusOK).JSON(context)
}

func DeleteReview(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Review deleted successfully",
	}

	if database.DbConn == nil {
		log.Println("database connection is not initialized")
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	user, err := currentUserFromContext(c, context)
	if err != nil {
		return err
	}

	reviewID, err := parseReviewID(c, context)
	if err != nil {
		return err
	}

	var review models.Review
	result := database.DbConn.Where("id = ? AND user_id = ?", reviewID, user.ID).First(&review)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "Review not found"
			return c.Status(fiber.StatusNotFound).JSON(context)
		}

		log.Println("Error querying review:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	if err := database.DbConn.Delete(&review).Error; err != nil {
		log.Println("Error deleting review:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error deleting review"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

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

func parseItemID(c *fiber.Ctx, context fiber.Map) (uint, error) {
	itemID, err := strconv.ParseUint(strings.TrimSpace(c.Params("item_id")), 10, 64)
	if err != nil || itemID == 0 {
		context["statusText"] = "bad"
		context["msg"] = "Invalid item id"
		return 0, c.Status(fiber.StatusBadRequest).JSON(context)
	}

	return uint(itemID), nil
}

func parseReviewID(c *fiber.Ctx, context fiber.Map) (uint, error) {
	reviewID, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil || reviewID == 0 {
		context["statusText"] = "bad"
		context["msg"] = "Invalid review id"
		return 0, c.Status(fiber.StatusBadRequest).JSON(context)
	}

	return uint(reviewID), nil
}

func ensureItemExists(itemID uint, context fiber.Map, c *fiber.Ctx) error {
	var item models.Item
	result := database.DbConn.Select("id").First(&item, itemID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "Item not found"
			return c.Status(fiber.StatusNotFound).JSON(context)
		}

		log.Println("Error querying item:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	return nil
}

func buildReviewFromCreateRequest(request createReviewRequest) (models.Review, string, int) {
	rating, comment, msg, statusCode := normalizeReviewInput(request.Rating, request.Comment)
	if statusCode != 0 {
		return models.Review{}, msg, statusCode
	}

	if request.ItemID == 0 {
		return models.Review{}, "Item id is required", fiber.StatusBadRequest
	}

	return models.Review{
		ItemID:  request.ItemID,
		Rating:  rating,
		Comment: comment,
	}, "", 0
}

func normalizeReviewInput(rating uint, comment string) (uint, string, string, int) {
	comment = strings.TrimSpace(comment)

	if rating < 1 || rating > 5 {
		return 0, "", "Rating must be between 1 and 5", fiber.StatusBadRequest
	}

	return rating, comment, "", 0
}

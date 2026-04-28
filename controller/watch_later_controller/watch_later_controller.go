package watchlatercontroller

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

func GetWatchLater(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Watch later items fetched successfully",
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

	var items []models.Item
	result := database.DbConn.
		Model(&models.Item{}).
		Joins("JOIN watch_later ON watch_later.item_id = items.id").
		Where("watch_later.user_id = ?", user.ID).
		Preload("Categories").
		Order("watch_later.created_at DESC").
		Find(&items)
	if result.Error != nil {
		log.Println("Error fetching watch later:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["items"] = items
	return c.Status(fiber.StatusOK).JSON(context)
}

func AddWatchLater(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Item added to watch later successfully",
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

	itemID, err := parseItemID(c, context)
	if err != nil {
		return err
	}

	if err := ensureItemExists(itemID, context, c); err != nil {
		return err
	}

	var wl models.WatchLater
	result := database.DbConn.Where("user_id = ? AND item_id = ?", user.ID, itemID).First(&wl)
	if result.Error == nil {
		context["statusText"] = "bad"
		context["msg"] = "Item already in watch later"
		return c.Status(fiber.StatusConflict).JSON(context)
	}
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Println("Error querying watch later:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	wl = models.WatchLater{
		UserID: user.ID,
		ItemID: itemID,
	}

	if err := database.DbConn.Create(&wl).Error; err != nil {
		log.Println("Error saving watch later:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error saving watch later"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["id"] = wl.ID
	context["watchLater"] = wl
	return c.Status(fiber.StatusCreated).JSON(context)
}

func RemoveWatchLater(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Item removed from watch later successfully",
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

	itemID, err := parseItemID(c, context)
	if err != nil {
		return err
	}

	var wl models.WatchLater
	result := database.DbConn.Where("user_id = ? AND item_id = ?", user.ID, itemID).First(&wl)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "Watch later item not found"
			return c.Status(fiber.StatusNotFound).JSON(context)
		}

		log.Println("Error querying watch later:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	if err := database.DbConn.Delete(&wl).Error; err != nil {
		log.Println("Error deleting watch later:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error deleting watch later"
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

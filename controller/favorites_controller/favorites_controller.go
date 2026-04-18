package favoritescontroller

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

func GetFavorites(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Favorites fetched successfully",
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
		Joins("JOIN favorites ON favorites.item_id = items.id").
		Where("favorites.user_id = ?", user.ID).
		Preload("Categories").
		Order("favorites.created_at DESC").
		Find(&items)
	if result.Error != nil {
		log.Println("Error fetching favorites:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["items"] = items
	return c.Status(fiber.StatusOK).JSON(context)
}

func AddFavorite(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Favorite added successfully",
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

	var favorite models.Favorite
	result := database.DbConn.Where("user_id = ? AND item_id = ?", user.ID, itemID).First(&favorite)
	if result.Error == nil {
		context["statusText"] = "bad"
		context["msg"] = "Item already in favorites"
		return c.Status(fiber.StatusConflict).JSON(context)
	}
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Println("Error querying favorite:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	favorite = models.Favorite{
		UserID: user.ID,
		ItemID: itemID,
	}

	if err := database.DbConn.Create(&favorite).Error; err != nil {
		log.Println("Error saving favorite:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error saving favorite"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["id"] = favorite.ID
	context["favorite"] = favorite
	return c.Status(fiber.StatusCreated).JSON(context)
}

func RemoveFavorite(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Favorite removed successfully",
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

	var favorite models.Favorite
	result := database.DbConn.Where("user_id = ? AND item_id = ?", user.ID, itemID).First(&favorite)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "Favorite not found"
			return c.Status(fiber.StatusNotFound).JSON(context)
		}

		log.Println("Error querying favorite:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	if err := database.DbConn.Delete(&favorite).Error; err != nil {
		log.Println("Error deleting favorite:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error deleting favorite"
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

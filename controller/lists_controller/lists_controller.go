package listscontroller

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

func GetLists(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Lists fetched successfully",
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

	var lists []models.UserList
	result := database.DbConn.
		Where("user_id = ?", user.ID).
		Order("created_at DESC").
		Find(&lists)
	if result.Error != nil {
		log.Println("Error fetching lists:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["lists"] = lists
	return c.Status(fiber.StatusOK).JSON(context)
}

func CreateList(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "List created successfully",
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

	var req struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Error parsing request body: %v", err)
		context["statusText"] = "bad"
		context["msg"] = "Invalid request"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		context["statusText"] = "bad"
		context["msg"] = "List name is required"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	list := models.UserList{
		UserID: user.ID,
		Name:   name,
	}

	if err := database.DbConn.Create(&list).Error; err != nil {
		log.Println("Error saving list:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error saving list"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["id"] = list.ID
	context["list"] = list
	return c.Status(fiber.StatusCreated).JSON(context)
}

func DeleteList(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "List deleted successfully",
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

	listID, err := parseListID(c, context)
	if err != nil {
		return err
	}

	tx := database.DbConn.Begin()
	if tx.Error != nil {
		log.Println("Error starting transaction:", tx.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	var list models.UserList
	result := tx.Where("id = ? AND user_id = ?", listID, user.ID).First(&list)
	if result.Error != nil {
		tx.Rollback()
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "List not found"
			return c.Status(fiber.StatusNotFound).JSON(context)
		}

		log.Println("Error querying list:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	if err := tx.Where("user_list_id = ?", list.ID).Delete(&models.UserListItem{}).Error; err != nil {
		tx.Rollback()
		log.Println("Error deleting list items:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error deleting list items"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	if err := tx.Delete(&list).Error; err != nil {
		tx.Rollback()
		log.Println("Error deleting list:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error deleting list"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	if err := tx.Commit().Error; err != nil {
		log.Println("Error committing transaction:", err)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	return c.Status(fiber.StatusOK).JSON(context)
}

func GetListItems(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "List items fetched successfully",
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

	listID, err := parseListID(c, context)
	if err != nil {
		return err
	}

	var list models.UserList
	result := database.DbConn.Where("id = ? AND user_id = ?", listID, user.ID).First(&list)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "List not found"
			return c.Status(fiber.StatusNotFound).JSON(context)
		}

		log.Println("Error querying list:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	var items []models.Item
	result = database.DbConn.
		Model(&models.Item{}).
		Joins("JOIN user_list_items ON user_list_items.item_id = items.id").
		Where("user_list_items.user_list_id = ?", list.ID).
		Preload("Categories").
		Order("user_list_items.created_at DESC").
		Find(&items)
	if result.Error != nil {
		log.Println("Error fetching list items:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["list"] = list
	context["items"] = items
	return c.Status(fiber.StatusOK).JSON(context)
}

func AddItemToList(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Item added to list successfully",
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

	listID, err := parseListID(c, context)
	if err != nil {
		return err
	}

	var list models.UserList
	result := database.DbConn.Where("id = ? AND user_id = ?", listID, user.ID).First(&list)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "List not found"
			return c.Status(fiber.StatusNotFound).JSON(context)
		}

		log.Println("Error querying list:", result.Error)
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

	var item models.UserListItem
	result = database.DbConn.Where("user_list_id = ? AND item_id = ?", list.ID, itemID).First(&item)
	if result.Error == nil {
		context["statusText"] = "bad"
		context["msg"] = "Item already in list"
		return c.Status(fiber.StatusConflict).JSON(context)
	}
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Println("Error querying list item:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	item = models.UserListItem{
		UserListID: list.ID,
		ItemID:     itemID,
	}

	if err := database.DbConn.Create(&item).Error; err != nil {
		log.Println("Error saving list item:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error saving list item"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["id"] = item.ID
	context["item"] = item
	return c.Status(fiber.StatusCreated).JSON(context)
}

func RemoveItemFromList(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Item removed from list successfully",
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

	listID, err := parseListID(c, context)
	if err != nil {
		return err
	}

	var list models.UserList
	result := database.DbConn.Where("id = ? AND user_id = ?", listID, user.ID).First(&list)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "List not found"
			return c.Status(fiber.StatusNotFound).JSON(context)
		}

		log.Println("Error querying list:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	itemID, err := parseItemID(c, context)
	if err != nil {
		return err
	}

	var item models.UserListItem
	result = database.DbConn.Where("user_list_id = ? AND item_id = ?", list.ID, itemID).First(&item)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "Item not found in list"
			return c.Status(fiber.StatusNotFound).JSON(context)
		}

		log.Println("Error querying list item:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	if err := database.DbConn.Delete(&item).Error; err != nil {
		log.Println("Error deleting list item:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error deleting list item"
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

func parseListID(c *fiber.Ctx, context fiber.Map) (uint, error) {
	listID, err := strconv.ParseUint(strings.TrimSpace(c.Params("list_id")), 10, 64)
	if err != nil || listID == 0 {
		context["statusText"] = "bad"
		context["msg"] = "Invalid list id"
		return 0, c.Status(fiber.StatusBadRequest).JSON(context)
	}

	return uint(listID), nil
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

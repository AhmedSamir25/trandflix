package itemcontroller

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"trendflix/database"
	"trendflix/models"
)

type itemRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Type        string  `json:"type"`
	CoverImage  string  `json:"cover_image"`
	ReleaseDate string  `json:"release_date"`
	Author      *string `json:"author"`
	Director    *string `json:"director"`
	Developer   *string `json:"developer"`
	Duration    *uint   `json:"duration"`
	PagesCount  *uint   `json:"pages_count"`
	Platform    *string `json:"platform"`
	Rating      float64 `json:"rating"`
	CategoryIDs []uint  `json:"category_ids"`
}

func GetItems(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Items fetched successfully",
	}

	if database.DbConn == nil {
		log.Println("database connection is not initialized")
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	var items []models.Item
	result := database.DbConn.Preload("Categories").Order("id DESC").Find(&items)
	if result.Error != nil {
		log.Println("Error fetching items:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["items"] = items
	return c.Status(fiber.StatusOK).JSON(context)
}

func GetItemByID(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Item fetched successfully",
	}

	if database.DbConn == nil {
		log.Println("database connection is not initialized")
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	id, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil || id == 0 {
		context["statusText"] = "bad"
		context["msg"] = "Invalid item id"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	var item models.Item
	result := database.DbConn.Preload("Categories").First(&item, id)
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

	context["item"] = item
	return c.Status(fiber.StatusOK).JSON(context)
}

func CreateItem(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Item created successfully",
	}

	if database.DbConn == nil {
		log.Println("database connection is not initialized")
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	if err := requireAdminAccess(c, context); err != nil {
		return err
	}

	var request itemRequest
	if err := c.BodyParser(&request); err != nil {
		log.Printf("Error parsing request body: %v", err)
		context["statusText"] = "bad"
		context["msg"] = "Invalid request"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	record, msg, statusCode := buildItemFromRequest(request)
	if statusCode != 0 {
		context["statusText"] = "bad"
		context["msg"] = msg
		return c.Status(statusCode).JSON(context)
	}

	categories, msg, statusCode := loadCategoriesByIDs(request.CategoryIDs)
	if statusCode != 0 {
		context["statusText"] = "bad"
		context["msg"] = msg
		return c.Status(statusCode).JSON(context)
	}

	tx := database.DbConn.Begin()
	if tx.Error != nil {
		log.Println("Error starting transaction:", tx.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		log.Println("Error saving item:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error saving item"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	if err := replaceItemCategories(tx, &record, categories); err != nil {
		tx.Rollback()
		log.Println("Error saving item categories:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error saving item categories"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	if err := tx.Commit().Error; err != nil {
		log.Println("Error committing transaction:", err)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	record.Categories = categories

	context["id"] = record.ID
	context["item"] = record
	return c.Status(fiber.StatusCreated).JSON(context)
}

func UpdateItem(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Item updated successfully",
	}

	if database.DbConn == nil {
		log.Println("database connection is not initialized")
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	if err := requireAdminAccess(c, context); err != nil {
		return err
	}

	id, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil || id == 0 {
		context["statusText"] = "bad"
		context["msg"] = "Invalid item id"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	var request itemRequest
	if err := c.BodyParser(&request); err != nil {
		log.Printf("Error parsing request body: %v", err)
		context["statusText"] = "bad"
		context["msg"] = "Invalid request"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	updatedItem, msg, statusCode := buildItemFromRequest(request)
	if statusCode != 0 {
		context["statusText"] = "bad"
		context["msg"] = msg
		return c.Status(statusCode).JSON(context)
	}

	categories, msg, statusCode := loadCategoriesByIDs(request.CategoryIDs)
	if statusCode != 0 {
		context["statusText"] = "bad"
		context["msg"] = msg
		return c.Status(statusCode).JSON(context)
	}

	tx := database.DbConn.Begin()
	if tx.Error != nil {
		log.Println("Error starting transaction:", tx.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	var item models.Item
	result := tx.Preload("Categories").First(&item, id)
	if result.Error != nil {
		tx.Rollback()
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

	item.Title = updatedItem.Title
	item.Description = updatedItem.Description
	item.Type = updatedItem.Type
	item.CoverImage = updatedItem.CoverImage
	item.ReleaseDate = updatedItem.ReleaseDate
	item.Author = updatedItem.Author
	item.Director = updatedItem.Director
	item.Developer = updatedItem.Developer
	item.Duration = updatedItem.Duration
	item.PagesCount = updatedItem.PagesCount
	item.Platform = updatedItem.Platform
	item.Rating = updatedItem.Rating

	if err := tx.Save(&item).Error; err != nil {
		tx.Rollback()
		log.Println("Error updating item:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error updating item"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	if err := replaceItemCategories(tx, &item, categories); err != nil {
		tx.Rollback()
		log.Println("Error updating item categories:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error updating item categories"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	if err := tx.Commit().Error; err != nil {
		log.Println("Error committing transaction:", err)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	item.Categories = categories

	context["item"] = item
	return c.Status(fiber.StatusOK).JSON(context)
}

func DeleteItem(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Item deleted successfully",
	}

	if database.DbConn == nil {
		log.Println("database connection is not initialized")
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	if err := requireAdminAccess(c, context); err != nil {
		return err
	}

	id, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil || id == 0 {
		context["statusText"] = "bad"
		context["msg"] = "Invalid item id"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	tx := database.DbConn.Begin()
	if tx.Error != nil {
		log.Println("Error starting transaction:", tx.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	var item models.Item
	result := tx.First(&item, id)
	if result.Error != nil {
		tx.Rollback()
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

	if err := tx.Where("item_id = ?", item.ID).Delete(&models.CategoryItem{}).Error; err != nil {
		tx.Rollback()
		log.Println("Error deleting item categories:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error deleting item categories"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	if err := tx.Delete(&item).Error; err != nil {
		tx.Rollback()
		log.Println("Error deleting item:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error deleting item"
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

func requireAdminAccess(c *fiber.Ctx, context fiber.Map) error {
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

func buildItemFromRequest(request itemRequest) (models.Item, string, int) {
	request.Title = strings.TrimSpace(request.Title)
	request.Description = strings.TrimSpace(request.Description)
	request.Type = strings.ToLower(strings.TrimSpace(request.Type))
	request.CoverImage = strings.TrimSpace(request.CoverImage)
	request.ReleaseDate = strings.TrimSpace(request.ReleaseDate)
	request.Author = normalizeOptionalString(request.Author)
	request.Director = normalizeOptionalString(request.Director)
	request.Developer = normalizeOptionalString(request.Developer)
	request.Platform = normalizeOptionalString(request.Platform)

	if request.Title == "" || request.Type == "" || request.ReleaseDate == "" {
		return models.Item{}, "Title, type and release_date are required", fiber.StatusBadRequest
	}

	if request.Type != "book" && request.Type != "movie" && request.Type != "game" {
		return models.Item{}, "Type must be book, movie or game", fiber.StatusBadRequest
	}

	releaseDate, err := time.Parse("2006-01-02", request.ReleaseDate)
	if err != nil {
		return models.Item{}, "release_date must be in YYYY-MM-DD format", fiber.StatusBadRequest
	}

	item := models.Item{
		Title:       request.Title,
		Description: request.Description,
		Type:        request.Type,
		CoverImage:  request.CoverImage,
		ReleaseDate: releaseDate,
		Author:      request.Author,
		Director:    request.Director,
		Developer:   request.Developer,
		Duration:    request.Duration,
		PagesCount:  request.PagesCount,
		Platform:    request.Platform,
		Rating:      request.Rating,
	}

	switch item.Type {
	case "book":
		item.Director = nil
		item.Developer = nil
		item.Duration = nil
		item.Platform = nil
	case "movie":
		item.Author = nil
		item.Developer = nil
		item.PagesCount = nil
		item.Platform = nil
	case "game":
		item.Author = nil
		item.Director = nil
		item.Duration = nil
		item.PagesCount = nil
	}

	return item, "", 0
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func loadCategoriesByIDs(categoryIDs []uint) ([]models.Category, string, int) {
	if len(categoryIDs) == 0 {
		return []models.Category{}, "", 0
	}

	uniqueIDs := make([]uint, 0, len(categoryIDs))
	seen := make(map[uint]struct{}, len(categoryIDs))
	for _, categoryID := range categoryIDs {
		if categoryID == 0 {
			return nil, "Invalid category id", fiber.StatusBadRequest
		}

		if _, exists := seen[categoryID]; exists {
			continue
		}

		seen[categoryID] = struct{}{}
		uniqueIDs = append(uniqueIDs, categoryID)
	}

	var categories []models.Category
	result := database.DbConn.Where("id IN ?", uniqueIDs).Find(&categories)
	if result.Error != nil {
		log.Println("Error querying categories:", result.Error)
		return nil, "Database error", fiber.StatusInternalServerError
	}

	categoryByID := make(map[uint]models.Category, len(categories))
	for _, category := range categories {
		categoryByID[category.ID] = category
	}

	orderedCategories := make([]models.Category, 0, len(uniqueIDs))
	for _, categoryID := range uniqueIDs {
		category, exists := categoryByID[categoryID]
		if !exists {
			return nil, "One or more categories were not found", fiber.StatusBadRequest
		}

		orderedCategories = append(orderedCategories, category)
	}

	return orderedCategories, "", 0
}

func replaceItemCategories(tx *gorm.DB, item *models.Item, categories []models.Category) error {
	association := tx.Model(item).Association("Categories")
	if association.Error != nil {
		return association.Error
	}

	if len(categories) == 0 {
		return association.Clear()
	}

	return association.Replace(categories)
}

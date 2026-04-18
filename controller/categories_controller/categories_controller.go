package categoriescontroller

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

type categoryRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func GetCategories(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Categories fetched successfully",
	}

	if database.DbConn == nil {
		log.Println("database connection is not initialized")
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	var categories []models.Category
	result := database.DbConn.Order("name ASC").Find(&categories)
	if result.Error != nil {
		log.Println("Error fetching categories:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["categories"] = categories
	return c.Status(fiber.StatusOK).JSON(context)
}

func CreateCategory(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Category created successfully",
	}

	if database.DbConn == nil {
		log.Println("database connection is not initialized")
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	var request categoryRequest
	if err := c.BodyParser(&request); err != nil {
		log.Printf("Error parsing request body: %v", err)
		context["statusText"] = "bad"
		context["msg"] = "Invalid request"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	request.Name = strings.TrimSpace(request.Name)
	request.Slug = strings.TrimSpace(request.Slug)

	if request.Name == "" || request.Slug == "" {
		context["statusText"] = "bad"
		context["msg"] = "Name and slug are required"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	var existingCategory models.Category
	result := database.DbConn.Where("slug = ?", request.Slug).First(&existingCategory)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Println("Error querying database:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}
	if result.RowsAffected > 0 {
		context["statusText"] = "bad"
		context["msg"] = "Slug already exists"
		return c.Status(fiber.StatusConflict).JSON(context)
	}

	record := models.Category{
		Name: request.Name,
		Slug: request.Slug,
	}

	result = database.DbConn.Create(&record)
	if result.Error != nil {
		log.Println("Error saving category:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Error saving category"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["id"] = record.ID
	context["category"] = record
	return c.Status(fiber.StatusCreated).JSON(context)
}

func UpdateCategory(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Category updated successfully",
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
		context["msg"] = "Invalid category id"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	var request categoryRequest
	if err := c.BodyParser(&request); err != nil {
		log.Printf("Error parsing request body: %v", err)
		context["statusText"] = "bad"
		context["msg"] = "Invalid request"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	request.Name = strings.TrimSpace(request.Name)
	request.Slug = strings.TrimSpace(request.Slug)

	if request.Name == "" || request.Slug == "" {
		context["statusText"] = "bad"
		context["msg"] = "Name and slug are required"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	var category models.Category
	result := database.DbConn.First(&category, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "Category not found"
			return c.Status(fiber.StatusNotFound).JSON(context)
		}

		log.Println("Error querying category:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	var duplicateCategory models.Category
	result = database.DbConn.Where("slug = ? AND id <> ?", request.Slug, id).First(&duplicateCategory)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Println("Error checking category slug:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}
	if result.RowsAffected > 0 {
		context["statusText"] = "bad"
		context["msg"] = "Slug already exists"
		return c.Status(fiber.StatusConflict).JSON(context)
	}

	category.Name = request.Name
	category.Slug = request.Slug

	result = database.DbConn.Save(&category)
	if result.Error != nil {
		log.Println("Error updating category:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Error updating category"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["category"] = category
	return c.Status(fiber.StatusOK).JSON(context)
}

func DeleteCategory(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Category deleted successfully",
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
		context["msg"] = "Invalid category id"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	var category models.Category
	result := database.DbConn.First(&category, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "Category not found"
			return c.Status(fiber.StatusNotFound).JSON(context)
		}

		log.Println("Error querying category:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	result = database.DbConn.Delete(&category)
	if result.Error != nil {
		log.Println("Error deleting category:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Error deleting category"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	return c.Status(fiber.StatusOK).JSON(context)
}

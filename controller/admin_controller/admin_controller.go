package admincontroller

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"trendflix/database"
	"trendflix/models"
)

type typeCount struct {
	Type  string `json:"type" gorm:"column:type"`
	Count int64  `json:"count" gorm:"column:count"`
}

type categoryCount struct {
	ID        uint   `json:"id" gorm:"column:id"`
	Name      string `json:"name" gorm:"column:name"`
	Slug      string `json:"slug" gorm:"column:slug"`
	ItemCount int64  `json:"item_count" gorm:"column:item_count"`
}

type roleCount struct {
	Role  string `json:"role" gorm:"column:role"`
	Count int64  `json:"count" gorm:"column:count"`
}

func ensureDatabase(context fiber.Map) error {
	if database.DbConn == nil {
		log.Println("database connection is not initialized")
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return errors.New("database is not initialized")
	}

	return nil
}

func failWithDatabaseError(c *fiber.Ctx, context fiber.Map, logMessage string, err error) error {
	log.Println(logMessage, err)
	context["statusText"] = "bad"
	context["msg"] = "Database error"
	return c.Status(fiber.StatusInternalServerError).JSON(context)
}

func fetchOverviewStats() (fiber.Map, error) {
	var totalItems int64
	var totalUsers int64
	var totalCategories int64

	if err := database.DbConn.Model(&models.Item{}).Count(&totalItems).Error; err != nil {
		return nil, err
	}

	if err := database.DbConn.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		return nil, err
	}

	if err := database.DbConn.Model(&models.Category{}).Count(&totalCategories).Error; err != nil {
		return nil, err
	}

	var averageRating float64
	if err := database.DbConn.
		Model(&models.Item{}).
		Select("COALESCE(AVG(NULLIF(rating, 0)), 0)").
		Scan(&averageRating).Error; err != nil {
		return nil, err
	}

	var latestItem models.Item
	latestResult := database.DbConn.Order("id DESC").First(&latestItem)
	if latestResult.Error != nil && !errors.Is(latestResult.Error, gorm.ErrRecordNotFound) {
		return nil, latestResult.Error
	}

	overview := fiber.Map{
		"total_items":      totalItems,
		"total_users":      totalUsers,
		"total_categories": totalCategories,
		"average_rating":   averageRating,
		"latest_item":      latestItem,
	}

	return overview, nil
}

func fetchTypeCounts() ([]typeCount, error) {
	var typeCounts []typeCount
	if err := database.DbConn.
		Model(&models.Item{}).
		Select("type, COUNT(*) AS count").
		Group("type").
		Scan(&typeCounts).Error; err != nil {
		return nil, err
	}

	return typeCounts, nil
}

func fetchCategoryCounts() ([]categoryCount, error) {
	var categoryCounts []categoryCount
	if err := database.DbConn.
		Table("categories").
		Select("categories.id, categories.name, categories.slug, COUNT(category_item.item_id) AS item_count").
		Joins("LEFT JOIN category_item ON category_item.category_id = categories.id").
		Group("categories.id, categories.name, categories.slug").
		Order("categories.name ASC").
		Scan(&categoryCounts).Error; err != nil {
		return nil, err
	}

	return categoryCounts, nil
}

func fetchUserStats() (fiber.Map, error) {
	var totalUsers int64
	if err := database.DbConn.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		return nil, err
	}

	var roleCounts []roleCount
	if err := database.DbConn.
		Model(&models.User{}).
		Select("role, COUNT(*) AS count").
		Group("role").
		Scan(&roleCounts).Error; err != nil {
		return nil, err
	}

	var recentUsers []models.User
	if err := database.DbConn.Order("id DESC").Limit(5).Find(&recentUsers).Error; err != nil {
		return nil, err
	}

	return fiber.Map{
		"total_users":  totalUsers,
		"role_counts":  roleCounts,
		"recent_users": recentUsers,
	}, nil
}

func fetchRecentItems() ([]models.Item, error) {
	var recentItems []models.Item
	if err := database.DbConn.Preload("Categories").Order("id DESC").Limit(5).Find(&recentItems).Error; err != nil {
		return nil, err
	}

	return recentItems, nil
}

func GetOverviewStats(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Overview stats fetched successfully",
	}

	if err := ensureDatabase(context); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	overview, err := fetchOverviewStats()
	if err != nil {
		return failWithDatabaseError(c, context, "Error fetching overview stats:", err)
	}

	context["overview"] = overview
	return c.Status(fiber.StatusOK).JSON(context)
}

func GetTypeStats(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Type stats fetched successfully",
	}

	if err := ensureDatabase(context); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	typeCounts, err := fetchTypeCounts()
	if err != nil {
		return failWithDatabaseError(c, context, "Error counting items by type:", err)
	}

	context["type_counts"] = typeCounts
	return c.Status(fiber.StatusOK).JSON(context)
}

func GetCategoryStats(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Category stats fetched successfully",
	}

	if err := ensureDatabase(context); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	categoryCounts, err := fetchCategoryCounts()
	if err != nil {
		return failWithDatabaseError(c, context, "Error counting items by category:", err)
	}

	context["category_counts"] = categoryCounts
	return c.Status(fiber.StatusOK).JSON(context)
}

func GetUserStats(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "User stats fetched successfully",
	}

	if err := ensureDatabase(context); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	userStats, err := fetchUserStats()
	if err != nil {
		return failWithDatabaseError(c, context, "Error fetching user stats:", err)
	}

	context["user_stats"] = userStats
	return c.Status(fiber.StatusOK).JSON(context)
}

func GetStats(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Admin stats fetched successfully",
	}

	if err := ensureDatabase(context); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	overview, err := fetchOverviewStats()
	if err != nil {
		return failWithDatabaseError(c, context, "Error fetching overview stats:", err)
	}

	typeCounts, err := fetchTypeCounts()
	if err != nil {
		return failWithDatabaseError(c, context, "Error counting items by type:", err)
	}

	categoryCounts, err := fetchCategoryCounts()
	if err != nil {
		return failWithDatabaseError(c, context, "Error counting items by category:", err)
	}

	userStats, err := fetchUserStats()
	if err != nil {
		return failWithDatabaseError(c, context, "Error fetching user stats:", err)
	}

	recentItems, err := fetchRecentItems()
	if err != nil {
		return failWithDatabaseError(c, context, "Error fetching recent items:", err)
	}

	context["stats"] = fiber.Map{
		"total_items":      overview["total_items"],
		"total_users":      overview["total_users"],
		"total_categories": overview["total_categories"],
		"average_rating":   overview["average_rating"],
		"latest_item":      overview["latest_item"],
		"type_counts":      typeCounts,
		"category_counts":  categoryCounts,
		"user_stats":       userStats,
		"recent_items":     recentItems,
	}

	return c.Status(fiber.StatusOK).JSON(context)
}

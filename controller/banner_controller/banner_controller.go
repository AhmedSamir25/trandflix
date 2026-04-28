package bannercontroller

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

type bannerRequest struct {
	Title     string `json:"title"`
	Subtitle  string `json:"subtitle"`
	ImageURL  string `json:"image_url"`
	LinkURL   string `json:"link_url"`
	IsActive  *bool  `json:"is_active"`
	SortOrder int    `json:"sort_order"`
}

func GetActiveBanners(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Banners fetched successfully",
	}

	if database.DbConn == nil {
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	var banners []models.Banner
	result := database.DbConn.Where("is_active = ?", true).Order("sort_order ASC, id ASC").Find(&banners)
	if result.Error != nil {
		log.Println("Error fetching banners:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["banners"] = banners
	return c.Status(fiber.StatusOK).JSON(context)
}

func GetAllBanners(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Banners fetched successfully",
	}

	if database.DbConn == nil {
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	var banners []models.Banner
	result := database.DbConn.Order("sort_order ASC, id ASC").Find(&banners)
	if result.Error != nil {
		log.Println("Error fetching banners:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["banners"] = banners
	return c.Status(fiber.StatusOK).JSON(context)
}

func CreateBanner(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Banner created successfully",
	}

	if database.DbConn == nil {
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	var request bannerRequest
	if err := c.BodyParser(&request); err != nil {
		context["statusText"] = "bad"
		context["msg"] = "Invalid request"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	banner, msg, statusCode := buildBannerFromRequest(request, true)
	if statusCode != 0 {
		context["statusText"] = "bad"
		context["msg"] = msg
		return c.Status(statusCode).JSON(context)
	}

	if err := database.DbConn.Create(&banner).Error; err != nil {
		log.Println("Error saving banner:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error saving banner"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["id"] = banner.ID
	context["banner"] = banner
	return c.Status(fiber.StatusCreated).JSON(context)
}

func UpdateBanner(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Banner updated successfully",
	}

	if database.DbConn == nil {
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	id, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil || id == 0 {
		context["statusText"] = "bad"
		context["msg"] = "Invalid banner id"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	var request bannerRequest
	if err := c.BodyParser(&request); err != nil {
		context["statusText"] = "bad"
		context["msg"] = "Invalid request"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	var banner models.Banner
	result := database.DbConn.First(&banner, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "Banner not found"
			return c.Status(fiber.StatusNotFound).JSON(context)
		}
		log.Println("Error querying banner:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	updated, msg, statusCode := buildBannerFromRequest(request, false)
	if statusCode != 0 {
		context["statusText"] = "bad"
		context["msg"] = msg
		return c.Status(statusCode).JSON(context)
	}

	banner.Title = updated.Title
	banner.Subtitle = updated.Subtitle
	banner.ImageURL = updated.ImageURL
	banner.LinkURL = updated.LinkURL
	banner.IsActive = updated.IsActive
	banner.SortOrder = updated.SortOrder

	if err := database.DbConn.Save(&banner).Error; err != nil {
		log.Println("Error updating banner:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error updating banner"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["banner"] = banner
	return c.Status(fiber.StatusOK).JSON(context)
}

func DeleteBanner(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Banner deleted successfully",
	}

	if database.DbConn == nil {
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	id, err := strconv.ParseUint(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil || id == 0 {
		context["statusText"] = "bad"
		context["msg"] = "Invalid banner id"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	var banner models.Banner
	result := database.DbConn.First(&banner, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "Banner not found"
			return c.Status(fiber.StatusNotFound).JSON(context)
		}
		log.Println("Error querying banner:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	if err := database.DbConn.Delete(&banner).Error; err != nil {
		log.Println("Error deleting banner:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error deleting banner"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	return c.Status(fiber.StatusOK).JSON(context)
}

func buildBannerFromRequest(request bannerRequest, isNew bool) (models.Banner, string, int) {
	request.Title = strings.TrimSpace(request.Title)
	request.Subtitle = strings.TrimSpace(request.Subtitle)
	request.ImageURL = strings.TrimSpace(request.ImageURL)
	request.LinkURL = strings.TrimSpace(request.LinkURL)

	if request.Title == "" {
		return models.Banner{}, "Title is required", fiber.StatusBadRequest
	}
	if request.ImageURL == "" {
		return models.Banner{}, "Image URL is required", fiber.StatusBadRequest
	}

	isActive := true
	if request.IsActive != nil {
		isActive = *request.IsActive
	}

	return models.Banner{
		Title:     request.Title,
		Subtitle:  request.Subtitle,
		ImageURL:  request.ImageURL,
		LinkURL:   request.LinkURL,
		IsActive:  isActive,
		SortOrder: request.SortOrder,
	}, "", 0
}

package database

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"trendflix/models"
)

const (
	defaultAdminName     = "Admin"
	defaultAdminEmail    = "admin@trendflix.local"
	defaultAdminPassword = "admin123456"
)

var defaultCategories = []models.Category{
	{Name: "Action", NameAr: "أكشن", Slug: "action"},
	{Name: "Adventure", NameAr: "مغامرة", Slug: "adventure"},
	{Name: "Animation", NameAr: "رسوم متحركة", Slug: "animation"},
	{Name: "Biography", NameAr: "سيرة ذاتية", Slug: "biography"},
	{Name: "Comedy", NameAr: "كوميديا", Slug: "comedy"},
	{Name: "Crime", NameAr: "جريمة", Slug: "crime"},
	{Name: "Documentary", NameAr: "وثائقي", Slug: "documentary"},
	{Name: "Drama", NameAr: "دراما", Slug: "drama"},
	{Name: "Family", NameAr: "عائلي", Slug: "family"},
	{Name: "Fantasy", NameAr: "خيال", Slug: "fantasy"},
	{Name: "History", NameAr: "تاريخي", Slug: "history"},
	{Name: "Horror", NameAr: "رعب", Slug: "horror"},
	{Name: "Music", NameAr: "موسيقي", Slug: "music"},
	{Name: "Mystery", NameAr: "غموض", Slug: "mystery"},
	{Name: "Musical", NameAr: "موسيقي غنائي", Slug: "musical"},
	{Name: "Psychological", NameAr: "نفسي", Slug: "psychological"},
	{Name: "Romance", NameAr: "رومانسي", Slug: "romance"},
	{Name: "Sci-Fi", NameAr: "خيال علمي", Slug: "sci-fi"},
	{Name: "Sport", NameAr: "رياضة", Slug: "sport"},
	{Name: "Superhero", NameAr: "أبطال خارقون", Slug: "superhero"},
	{Name: "Suspense", NameAr: "تشويق", Slug: "suspense"},
	{Name: "Thriller", NameAr: "إثارة", Slug: "thriller"},
	{Name: "War", NameAr: "حرب", Slug: "war"},
	{Name: "Western", NameAr: "غربي", Slug: "western"},
}

var defaultBanners = []models.Banner{
	{
		Title:     "Discover Entertainment",
		Subtitle:  "Movies, TV shows, games, and books picked for you every time you open TrendFlix.",
		ImageURL:  "/assets/images/default-banner.svg",
		LinkURL:   "",
		IsActive:  true,
		SortOrder: 0,
	},
}

func SeedAdmin() {
	if DbConn == nil {
		panic("database is not connected")
	}

	adminName := strings.TrimSpace(os.Getenv("ADMIN_NAME"))
	if adminName == "" {
		adminName = defaultAdminName
	}

	adminEmail := strings.ToLower(strings.TrimSpace(os.Getenv("ADMIN_EMAIL")))
	if adminEmail == "" {
		adminEmail = defaultAdminEmail
	}

	adminPassword := strings.TrimSpace(os.Getenv("ADMIN_PASSWORD"))
	if adminPassword == "" {
		adminPassword = defaultAdminPassword
	}

	adminAvatar := strings.TrimSpace(os.Getenv("ADMIN_AVATAR"))

	var existingUser models.User
	result := DbConn.Where("email = ?", adminEmail).First(&existingUser)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		panic(fmt.Sprintf("admin seed query failed: %v", result.Error))
	}

	if result.RowsAffected > 0 {
		if strings.TrimSpace(existingUser.Role) == "admin" {
			log.Printf("admin seed: admin user already exists for %s", adminEmail)
			return
		}

		if err := DbConn.Model(&existingUser).Updates(map[string]interface{}{
			"name":   adminName,
			"avatar": adminAvatar,
			"role":   "admin",
		}).Error; err != nil {
			panic(fmt.Sprintf("admin seed update failed: %v", err))
		}

		log.Printf("admin seed: promoted existing user %s to admin", adminEmail)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Sprintf("admin seed password hash failed: %v", err))
	}

	admin := models.User{
		Name:     adminName,
		Email:    adminEmail,
		Password: string(hashedPassword),
		Avatar:   adminAvatar,
		Role:     "admin",
	}

	if err := DbConn.Create(&admin).Error; err != nil {
		panic(fmt.Sprintf("admin seed create failed: %v", err))
	}

	log.Printf("admin seed: created admin user %s", adminEmail)
}

func SeedCategories() {
	if DbConn == nil {
		panic("database is not connected")
	}

	for _, category := range defaultCategories {
		var existingCategory models.Category
		result := DbConn.Where("slug = ?", category.Slug).First(&existingCategory)
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			panic(fmt.Sprintf("category seed query failed for %s: %v", category.Slug, result.Error))
		}

		if result.RowsAffected > 0 {
			if existingCategory.Name == category.Name && existingCategory.NameAr == category.NameAr {
				continue
			}

			if err := DbConn.Model(&existingCategory).Updates(map[string]interface{}{
				"name":    category.Name,
				"name_ar": category.NameAr,
			}).Error; err != nil {
				panic(fmt.Sprintf("category seed update failed for %s: %v", category.Slug, err))
			}

			continue
		}

		if err := DbConn.Create(&category).Error; err != nil {
			panic(fmt.Sprintf("category seed create failed for %s: %v", category.Slug, err))
		}
	}

	log.Printf("category seed: ensured %d default categories", len(defaultCategories))
}

func SeedBanners() {
	if DbConn == nil {
		panic("database is not connected")
	}

	var count int64
	if err := DbConn.Model(&models.Banner{}).Count(&count).Error; err != nil {
		panic(fmt.Sprintf("banner seed count failed: %v", err))
	}

	if count > 0 {
		log.Printf("banner seed: skipped because %d banners already exist", count)
		return
	}

	if err := DbConn.Create(&defaultBanners).Error; err != nil {
		panic(fmt.Sprintf("banner seed create failed: %v", err))
	}

	log.Printf("banner seed: created %d default banners", len(defaultBanners))
}

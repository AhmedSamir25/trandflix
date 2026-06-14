package services

import (
	"strings"

	"gorm.io/gorm"

	"trendflix/models"
)

type GormAIRepository struct {
	DB *gorm.DB
}

func (r *GormAIRepository) SearchItems(keywords []string, itemType string, categorySlugs []string, limit int) ([]models.Item, error) {
	if limit <= 0 {
		limit = AICandidatePoolSize
	}

	query := r.DB.Model(&models.Item{})

	if itemType != "" {
		query = query.Where("items.type = ?", itemType)
	}

	hasCategoryFilter := len(categorySlugs) > 0
	if hasCategoryFilter {
		query = query.
			Joins("JOIN category_item ON category_item.item_id = items.id").
			Joins("JOIN categories ON categories.id = category_item.category_id")
	}

	conditions := make([]string, 0)
	params := make([]interface{}, 0)

	if hasCategoryFilter {
		conditions = append(conditions, "(categories.slug IN ? OR categories.name IN ? OR categories.name_ar IN ?)")
		params = append(params, categorySlugs, categorySlugs, categorySlugs)
	}

	for _, keyword := range keywords {
		like := "%" + keyword + "%"
		conditions = append(conditions, "(items.title LIKE ? OR items.description LIKE ?)")
		params = append(params, like, like)
	}

	if len(conditions) > 0 {
		query = query.Where(strings.Join(conditions, " OR "), params...)
	}

	if hasCategoryFilter {
		query = query.Group("items.id")
	}

	var items []models.Item
	err := query.
		Preload("Categories").
		Order("items.rating DESC, items.id DESC").
		Limit(limit).
		Find(&items).Error
	if err != nil {
		return nil, err
	}

	if len(items) >= limit || (len(keywords) == 0 && !hasCategoryFilter) {
		return items, nil
	}

	if itemType == "" && len(items) < limit {
		topUp, err := r.topUpTopRated(limit-len(items), items, itemType)
		if err != nil {
			return nil, err
		}
		items = append(items, topUp...)
	}

	return items, nil
}

func (r *GormAIRepository) topUpTopRated(count int, existing []models.Item, itemType string) ([]models.Item, error) {
	if count <= 0 {
		return nil, nil
	}

	existingIDs := make([]uint, 0, len(existing))
	for _, item := range existing {
		existingIDs = append(existingIDs, item.ID)
	}

	query := r.DB.Model(&models.Item{})
	if len(existingIDs) > 0 {
		query = query.Where("id NOT IN ?", existingIDs)
	}
	if itemType != "" {
		query = query.Where("type = ?", itemType)
	}

	var items []models.Item
	err := query.
		Preload("Categories").
		Order("rating DESC, id DESC").
		Limit(count).
		Find(&items).Error
	return items, err
}

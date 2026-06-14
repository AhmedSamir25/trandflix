package services

import (
	"trendflix/models"

	"gorm.io/gorm"
)

type GormRepository struct {
	DB *gorm.DB
}

func (r *GormRepository) FetchFavorites(userID uint) ([]models.Item, error) {
	var items []models.Item
	err := r.DB.
		Joins("JOIN favorites ON favorites.item_id = items.id").
		Where("favorites.user_id = ?", userID).
		Preload("Categories").
		Order("favorites.created_at DESC").
		Find(&items).Error
	return items, err
}

func (r *GormRepository) FetchWatchLater(userID uint) ([]models.Item, error) {
	var items []models.Item
	err := r.DB.
		Joins("JOIN watch_later ON watch_later.item_id = items.id").
		Where("watch_later.user_id = ?", userID).
		Preload("Categories").
		Order("watch_later.created_at DESC").
		Find(&items).Error
	return items, err
}

func (r *GormRepository) FetchListItems(userID uint) ([]models.Item, error) {
	var items []models.Item
	err := r.DB.
		Joins("JOIN user_list_items ON user_list_items.item_id = items.id").
		Joins("JOIN user_lists ON user_lists.id = user_list_items.user_list_id").
		Where("user_lists.user_id = ?", userID).
		Preload("Categories").
		Order("user_list_items.created_at DESC").
		Find(&items).Error
	return items, err
}

func (r *GormRepository) FetchReviewedItemIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := r.DB.
		Model(&models.Review{}).
		Where("user_id = ?", userID).
		Pluck("item_id", &ids).Error
	return ids, err
}

func (r *GormRepository) FetchCandidates(categoryIDs, excludedIDs []uint) ([]models.Item, error) {
	if len(categoryIDs) == 0 {
		return nil, nil
	}

	query := r.DB.
		Model(&models.Item{}).
		Joins("JOIN category_item ON category_item.item_id = items.id").
		Where("category_item.category_id IN ?", categoryIDs)

	if len(excludedIDs) > 0 {
		query = query.Where("items.id NOT IN ?", excludedIDs)
	}

	var items []models.Item
	err := query.
		Preload("Categories").
		Group("items.id").
		Find(&items).Error
	return items, err
}

func (r *GormRepository) FetchPopularity(itemIDs []uint) (map[uint]int, error) {
	popularity := make(map[uint]int, len(itemIDs))
	if len(itemIDs) == 0 {
		return popularity, nil
	}

	type countRow struct {
		ItemID uint
		Count  int
	}

	tables := []string{"favorites", "watch_later", "user_list_items", "reviews"}
	for _, table := range tables {
		var rows []countRow
		err := r.DB.
			Table(table).
			Select("item_id, COUNT(*) AS count").
			Where("item_id IN ?", itemIDs).
			Group("item_id").
			Scan(&rows).Error
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			popularity[row.ItemID] += row.Count
		}
	}

	return popularity, nil
}

func (r *GormRepository) FetchTopRated(excludedIDs []uint, limit int) ([]models.Item, error) {
	query := r.DB.Model(&models.Item{})
	if len(excludedIDs) > 0 {
		query = query.Where("id NOT IN ?", excludedIDs)
	}

	var items []models.Item
	err := query.
		Preload("Categories").
		Order("rating DESC, id DESC").
		Limit(limit).
		Find(&items).Error
	return items, err
}

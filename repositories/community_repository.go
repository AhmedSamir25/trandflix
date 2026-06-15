package repositories

import (
	"errors"

	"gorm.io/gorm"
	"trendflix/models"
)

type CommunityRepository struct {
	db *gorm.DB
}

func NewCommunityRepository(db *gorm.DB) *CommunityRepository {
	return &CommunityRepository{db: db}
}

func (r *CommunityRepository) List(search string, limit, offset int) ([]models.Community, int64, error) {
	query := r.db.Model(&models.Community{}).Where("status = ?", models.CommunityStatusApproved)
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var communities []models.Community
	err := query.
		Preload("User").
		Order("members_count DESC, id DESC").
		Limit(limit).Offset(offset).
		Find(&communities).Error
	return communities, total, err
}

func (r *CommunityRepository) Popular(limit int) ([]models.Community, error) {
	var communities []models.Community
	err := r.db.Preload("User").
		Where("status = ?", models.CommunityStatusApproved).
		Order("members_count DESC, id DESC").
		Limit(limit).
		Find(&communities).Error
	return communities, err
}

func (r *CommunityRepository) GetByID(id uint) (*models.Community, error) {
	var community models.Community
	err := r.db.Preload("User").First(&community, id).Error
	return &community, err
}

func (r *CommunityRepository) GetBySlug(slug string) (*models.Community, error) {
	var community models.Community
	err := r.db.Preload("User").
		Where("slug = ? AND status <> ?", slug, models.CommunityStatusRejected).
		First(&community).Error
	return &community, err
}

func (r *CommunityRepository) ListByStatus(status string, limit, offset int) ([]models.Community, int64, error) {
	query := r.db.Model(&models.Community{}).Where("status = ?", status)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var communities []models.Community
	err := query.
		Preload("User").
		Order("id DESC").
		Limit(limit).Offset(offset).
		Find(&communities).Error
	return communities, total, err
}

func (r *CommunityRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Community{}).Count(&count).Error
	return count, err
}

func (r *CommunityRepository) ListAll(search, status, categoryType string, limit, offset int) ([]models.Community, int64, error) {
	query := r.db.Model(&models.Community{})
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", like, like)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if categoryType != "" {
		query = query.Where("category_type = ?", categoryType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var communities []models.Community
	err := query.
		Preload("User").
		Order("id DESC").
		Limit(limit).Offset(offset).
		Find(&communities).Error
	return communities, total, err
}

type CategoryTypeCount struct {
	Type  string `json:"type" gorm:"column:category_type"`
	Count int64  `json:"count" gorm:"column:count"`
}

type StatusCount struct {
	Status string `json:"status" gorm:"column:status"`
	Count  int64  `json:"count" gorm:"column:count"`
}

func (r *CommunityRepository) CountByCategoryType() ([]CategoryTypeCount, error) {
	var counts []CategoryTypeCount
	err := r.db.Model(&models.Community{}).
		Select("category_type, COUNT(*) AS count").
		Group("category_type").
		Scan(&counts).Error
	return counts, err
}

func (r *CommunityRepository) CountGroupByStatus() ([]StatusCount, error) {
	var counts []StatusCount
	err := r.db.Model(&models.Community{}).
		Select("status, COUNT(*) AS count").
		Group("status").
		Scan(&counts).Error
	return counts, err
}

func (r *CommunityRepository) SumMembers() (int64, error) {
	var sum int64
	err := r.db.Model(&models.Community{}).
		Select("COALESCE(SUM(members_count), 0)").
		Scan(&sum).Error
	return sum, err
}

func (r *CommunityRepository) DeleteByID(id uint) error {
	if id == 0 {
		return errors.New("community id is required")
	}
	return r.db.Delete(&models.Community{}, id).Error
}

func (r *CommunityRepository) SetStatus(id uint, status string) error {
	return r.db.Model(&models.Community{}).Where("id = ?", id).Update("status", status).Error
}

func (r *CommunityRepository) CountByStatus(status string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Community{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

func (r *CommunityRepository) SlugExists(slug string, ignoreID uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.Community{}).Where("slug = ?", slug)
	if ignoreID > 0 {
		query = query.Where("id <> ?", ignoreID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *CommunityRepository) Create(community *models.Community) error {
	return r.db.Create(community).Error
}

func (r *CommunityRepository) Update(community *models.Community) error {
	return r.db.Save(community).Error
}

func (r *CommunityRepository) IncrementMembers(communityID uint, by int) error {
	return r.db.Model(&models.Community{}).
		Where("id = ?", communityID).
		UpdateColumn("members_count", gorm.Expr("GREATEST(members_count + ?, 0)", by)).Error
}

func (r *CommunityRepository) IncrementPosts(communityID uint, by int) error {
	return r.db.Model(&models.Community{}).
		Where("id = ?", communityID).
		UpdateColumn("posts_count", gorm.Expr("GREATEST(posts_count + ?, 0)", by)).Error
}

func (r *CommunityRepository) Delete(community *models.Community) error {
	if community.ID == 0 {
		return errors.New("community id is required")
	}
	return r.db.Delete(&models.Community{}, community.ID).Error
}

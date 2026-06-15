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
	query := r.db.Model(&models.Community{})
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
	err := r.db.Preload("User").Where("slug = ?", slug).First(&community).Error
	return &community, err
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

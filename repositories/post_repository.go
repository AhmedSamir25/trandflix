package repositories

import (
	"errors"

	"gorm.io/gorm"
	"trendflix/models"
)

const (
	SortNew = "new"
	SortTop = "top"
	SortHot = "hot"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) ListByCommunity(communityID uint, sort string, limit, offset int) ([]models.Post, int64, error) {
	query := r.db.Model(&models.Post{}).
		Where("community_id = ? AND status = ?", communityID, models.PostStatusPublished)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	dataQuery := r.db.Preload("User").Preload("Community").
		Where("community_id = ? AND status = ?", communityID, models.PostStatusPublished)

	switch sort {
	case SortTop:
		dataQuery = dataQuery.Order("score DESC, id DESC")
	case SortHot:
		dataQuery = dataQuery.Order("(score + comments_count * 2) DESC, id DESC")
	default:
		dataQuery = dataQuery.Order("created_at DESC, id DESC")
	}

	var posts []models.Post
	err := dataQuery.Limit(limit).Offset(offset).Find(&posts).Error
	return posts, total, err
}

func (r *PostRepository) ListByItem(itemType string, itemID uint, limit, offset int) ([]models.Post, int64, error) {
	query := r.db.Model(&models.Post{}).
		Where("related_item_type = ? AND related_item_id = ? AND status = ?", itemType, itemID, models.PostStatusPublished)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var posts []models.Post
	err := query.Preload("User").Preload("Community").
		Order("created_at DESC, id DESC").
		Limit(limit).Offset(offset).
		Find(&posts).Error
	return posts, total, err
}

func (r *PostRepository) GetByID(id uint) (*models.Post, error) {
	var post models.Post
	err := r.db.Preload("User").Preload("Community").First(&post, id).Error
	return &post, err
}

func (r *PostRepository) Create(post *models.Post) error {
	return r.db.Create(post).Error
}

func (r *PostRepository) Update(post *models.Post) error {
	return r.db.Save(post).Error
}

func (r *PostRepository) SoftDelete(post *models.Post) error {
	if post.ID == 0 {
		return errors.New("post id is required")
	}
	return r.db.Model(&models.Post{}).
		Where("id = ?", post.ID).
		Update("status", models.PostStatusDeleted).Error
}

func (r *PostRepository) IncrementComments(postID uint, by int) error {
	return r.db.Model(&models.Post{}).
		Where("id = ?", postID).
		UpdateColumn("comments_count", gorm.Expr("GREATEST(comments_count + ?, 0)", by)).Error
}

func (r *PostRepository) AdjustScore(postID uint, delta int) error {
	return r.db.Model(&models.Post{}).
		Where("id = ?", postID).
		UpdateColumn("score", gorm.Expr("score + ?", delta)).Error
}

func (r *PostRepository) CountRecentByUser(userID uint, since string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Post{}).
		Where("user_id = ? AND created_at >= ?", userID, since).
		Count(&count).Error
	return count, err
}

package repositories

import (
	"errors"

	"gorm.io/gorm"
	"trendflix/models"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) ListByPost(postID uint) ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.
		Where("post_id = ? AND status <> ?", postID, models.CommentStatusDeleted).
		Preload("User").
		Order("score DESC, created_at ASC").
		Find(&comments).Error
	return comments, err
}

func (r *CommentRepository) GetByID(id uint) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.First(&comment, id).Error
	return &comment, err
}

func (r *CommentRepository) Create(comment *models.Comment) error {
	return r.db.Create(comment).Error
}

func (r *CommentRepository) Update(comment *models.Comment) error {
	return r.db.Save(comment).Error
}

func (r *CommentRepository) SoftDelete(comment *models.Comment) error {
	if comment.ID == 0 {
		return errors.New("comment id is required")
	}
	return r.db.Model(&models.Comment{}).
		Where("id = ?", comment.ID).
		Update("status", models.CommentStatusDeleted).Error
}

func (r *CommentRepository) AdjustScore(commentID uint, delta int) error {
	return r.db.Model(&models.Comment{}).
		Where("id = ?", commentID).
		UpdateColumn("score", gorm.Expr("score + ?", delta)).Error
}

func (r *CommentRepository) CountRecentByUser(userID uint, since string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Comment{}).
		Where("user_id = ? AND created_at >= ?", userID, since).
		Count(&count).Error
	return count, err
}

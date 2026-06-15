package repositories

import (
	"gorm.io/gorm"
	"trendflix/models"
)

type VoteRepository struct {
	db *gorm.DB
}

func NewVoteRepository(db *gorm.DB) *VoteRepository {
	return &VoteRepository{db: db}
}

func (r *VoteRepository) GetByUserAndVotable(userID uint, votableID uint, votableType string) (*models.Vote, error) {
	var vote models.Vote
	err := r.db.Where("user_id = ? AND votable_id = ? AND votable_type = ?", userID, votableID, votableType).
		First(&vote).Error
	return &vote, err
}

func (r *VoteRepository) Create(vote *models.Vote) error {
	return r.db.Create(vote).Error
}

func (r *VoteRepository) UpdateVoteType(vote *models.Vote, voteType string) error {
	return r.db.Model(&models.Vote{}).
		Where("id = ?", vote.ID).
		Update("vote_type", voteType).Error
}

func (r *VoteRepository) Delete(vote *models.Vote) error {
	return r.db.Delete(&models.Vote{}, vote.ID).Error
}

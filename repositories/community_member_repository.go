package repositories

import (
	"errors"

	"gorm.io/gorm"
	"trendflix/models"
)

type CommunityMemberRepository struct {
	db *gorm.DB
}

func NewCommunityMemberRepository(db *gorm.DB) *CommunityMemberRepository {
	return &CommunityMemberRepository{db: db}
}

func (r *CommunityMemberRepository) GetByCommunityAndUser(communityID, userID uint) (*models.CommunityMember, error) {
	var member models.CommunityMember
	err := r.db.Where("community_id = ? AND user_id = ?", communityID, userID).First(&member).Error
	return &member, err
}

func (r *CommunityMemberRepository) ListByCommunity(communityID uint, limit, offset int) ([]models.CommunityMember, int64, error) {
	query := r.db.Model(&models.CommunityMember{}).Where("community_id = ? AND status = ?", communityID, models.MemberStatusActive)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var members []models.CommunityMember
	err := query.Preload("User").
		Order("role DESC, joined_at ASC").
		Limit(limit).Offset(offset).
		Find(&members).Error
	return members, total, err
}

func (r *CommunityMemberRepository) Create(member *models.CommunityMember) error {
	return r.db.Create(member).Error
}

func (r *CommunityMemberRepository) Update(member *models.CommunityMember) error {
	return r.db.Save(member).Error
}

func (r *CommunityMemberRepository) Delete(member *models.CommunityMember) error {
	if member.ID == 0 {
		return errors.New("member id is required")
	}
	return r.db.Delete(&models.CommunityMember{}, member.ID).Error
}

func (r *CommunityMemberRepository) CountRecentByUser(userID uint, since string) (int64, error) {
	var count int64
	err := r.db.Model(&models.CommunityMember{}).
		Where("user_id = ? AND created_at >= ?", userID, since).
		Count(&count).Error
	return count, err
}

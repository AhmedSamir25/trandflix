package communities

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"trendflix/database"
	"trendflix/models"
	"trendflix/repositories"
	"trendflix/utils/slugs"
)

var (
	ErrNotFound           = errors.New("community not found")
	ErrSlugTaken          = errors.New("slug is already taken")
	ErrInvalidName        = errors.New("name is required")
	ErrInvalidCategory    = errors.New("invalid category type")
	ErrAlreadyMember      = errors.New("you are already a member of this community")
	ErrNotMember          = errors.New("you are not a member of this community")
	ErrBanned             = errors.New("you are banned from this community")
	ErrRateLimitCommunity = errors.New("you have created too many communities today")
	ErrForbidden          = errors.New("you are not allowed to perform this action")
)

var validCategories = map[string]bool{
	models.CategoryTypeMovies: true,
	models.CategoryTypeSeries: true,
	models.CategoryTypeBooks:  true,
	models.CategoryTypeGames:  true,
	models.CategoryTypeMixed:  true,
}

const maxCommunitiesPerDay = 3

type Service struct {
	communities *repositories.CommunityRepository
	members     *repositories.CommunityMemberRepository
}

func NewService() *Service {
	return &Service{
		communities: repositories.NewCommunityRepository(database.DbConn),
		members:     repositories.NewCommunityMemberRepository(database.DbConn),
	}
}

type ListResult struct {
	Items []models.Community `json:"items"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Pages int                `json:"pages"`
}

type CreateInput struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	AvatarImage  string `json:"avatar_image"`
	CoverImage   string `json:"cover_image"`
	CategoryType string `json:"category_type"`
	Rules        string `json:"rules"`
	IsPrivate    bool   `json:"is_private"`
}

type UpdateInput struct {
	Name         *string `json:"name"`
	Description  *string `json:"description"`
	AvatarImage  *string `json:"avatar_image"`
	CoverImage   *string `json:"cover_image"`
	CategoryType *string `json:"category_type"`
	Rules        *string `json:"rules"`
	IsPrivate    *bool   `json:"is_private"`
}

func (s *Service) List(search string, page, perPage int) (*ListResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	items, total, err := s.communities.List(search, perPage, offset)
	if err != nil {
		return nil, err
	}

	pages := int((total + int64(perPage) - 1) / int64(perPage))
	return &ListResult{Items: items, Total: total, Page: page, Pages: pages}, nil
}

func (s *Service) Popular(limit int) ([]models.Community, error) {
	if limit < 1 || limit > 50 {
		limit = 8
	}
	return s.communities.Popular(limit)
}

func (s *Service) GetBySlug(slug string) (*models.Community, error) {
	community, err := s.communities.GetBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return community, nil
}

func (s *Service) GetByID(id uint) (*models.Community, error) {
	community, err := s.communities.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return community, nil
}

func (s *Service) Create(userID uint, input CreateInput) (*models.Community, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, ErrInvalidName
	}
	category := strings.TrimSpace(input.CategoryType)
	if category == "" || !validCategories[category] {
		return nil, ErrInvalidCategory
	}

	since := time.Now().Add(-24 * time.Hour)
	recent, err := s.members.CountRecentByUser(userID, since.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	if recent >= maxCommunitiesPerDay {
		return nil, ErrRateLimitCommunity
	}

	baseSlug := slugs.Slugify(name)
	slug, err := slugs.Unique(baseSlug, func(candidate string) (bool, error) {
		return s.communities.SlugExists(candidate, 0)
	})
	if err != nil {
		return nil, err
	}

	community := &models.Community{
		Name:         name,
		Slug:         slug,
		Description:  strings.TrimSpace(input.Description),
		AvatarImage:  strings.TrimSpace(input.AvatarImage),
		CoverImage:   strings.TrimSpace(input.CoverImage),
		CategoryType: category,
		CreatedBy:    userID,
		Rules:        strings.TrimSpace(input.Rules),
		IsPrivate:    input.IsPrivate,
		MembersCount: 1,
	}

	txErr := database.DbConn.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(community).Error; err != nil {
			return err
		}
		member := &models.CommunityMember{
			CommunityID: community.ID,
			UserID:      userID,
			Role:        models.MemberRoleAdmin,
			Status:      models.MemberStatusActive,
		}
		return tx.Create(member).Error
	})
	if txErr != nil {
		return nil, txErr
	}

	return s.communities.GetByID(community.ID)
}

func (s *Service) Update(userID uint, id uint, input UpdateInput) (*models.Community, error) {
	community, err := s.communities.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if !s.canManage(community, userID) {
		return nil, ErrForbidden
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, ErrInvalidName
		}
		if name != community.Name {
			newSlug := slugs.Slugify(name)
			if newSlug != community.Slug {
				taken, err := s.communities.SlugExists(newSlug, community.ID)
				if err != nil {
					return nil, err
				}
				if taken {
					return nil, ErrSlugTaken
				}
				community.Slug = newSlug
			}
			community.Name = name
		}
	}
	if input.Description != nil {
		community.Description = strings.TrimSpace(*input.Description)
	}
	if input.AvatarImage != nil {
		community.AvatarImage = strings.TrimSpace(*input.AvatarImage)
	}
	if input.CoverImage != nil {
		community.CoverImage = strings.TrimSpace(*input.CoverImage)
	}
	if input.CategoryType != nil {
		category := strings.TrimSpace(*input.CategoryType)
		if category == "" || !validCategories[category] {
			return nil, ErrInvalidCategory
		}
		community.CategoryType = category
	}
	if input.Rules != nil {
		community.Rules = strings.TrimSpace(*input.Rules)
	}
	if input.IsPrivate != nil {
		community.IsPrivate = *input.IsPrivate
	}

	if err := s.communities.Update(community); err != nil {
		return nil, err
	}
	return community, nil
}

func (s *Service) Delete(userID uint, id uint) error {
	community, err := s.communities.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if !s.canManage(community, userID) {
		return ErrForbidden
	}
	return s.communities.Delete(community)
}

func (s *Service) Join(userID uint, id uint) error {
	if _, err := s.communities.GetByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}

	existing, err := s.members.GetByCommunityAndUser(id, userID)
	if err == nil {
		if existing.Status == models.MemberStatusBanned {
			return ErrBanned
		}
		return ErrAlreadyMember
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	member := &models.CommunityMember{
		CommunityID: id,
		UserID:      userID,
		Role:        models.MemberRoleMember,
		Status:      models.MemberStatusActive,
	}
	if err := s.members.Create(member); err != nil {
		return err
	}
	return s.communities.IncrementMembers(id, 1)
}

func (s *Service) Leave(userID uint, id uint) error {
	member, err := s.members.GetByCommunityAndUser(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotMember
		}
		return err
	}
	if member.Role == models.MemberRoleAdmin {
		return ErrForbidden
	}

	if err := s.members.Delete(member); err != nil {
		return err
	}
	return s.communities.IncrementMembers(id, -1)
}

func (s *Service) Members(communityID uint, page, perPage int) ([]models.CommunityMember, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 30
	}
	offset := (page - 1) * perPage
	return s.members.ListByCommunity(communityID, perPage, offset)
}

func (s *Service) IsMember(communityID, userID uint) (bool, string, error) {
	member, err := s.members.GetByCommunityAndUser(communityID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, "", nil
		}
		return false, "", err
	}
	return member.Status == models.MemberStatusActive, member.Status, nil
}

func (s *Service) canManage(community *models.Community, userID uint) bool {
	if community.CreatedBy == userID {
		return true
	}
	member, err := s.members.GetByCommunityAndUser(community.ID, userID)
	if err != nil {
		return false
	}
	if member.Status != models.MemberStatusActive {
		return false
	}
	return member.Role == models.MemberRoleAdmin || member.Role == models.MemberRoleModerator
}

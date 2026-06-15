package posts

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"trendflix/database"
	"trendflix/models"
	"trendflix/repositories"
	communities "trendflix/services/communities"
)

var (
	ErrNotFound          = errors.New("post not found")
	ErrInvalidTitle      = errors.New("title is required")
	ErrMustJoin          = errors.New("you must join this community before posting")
	ErrBanned            = errors.New("you are banned from this community")
	ErrLocked            = errors.New("this post is locked")
	ErrForbidden         = errors.New("you are not allowed to perform this action")
	ErrRateLimitPost     = errors.New("you are posting too fast, try again later")
	ErrRateLimitComment  = errors.New("you are commenting too fast, try again later")
	ErrCommunityNotFound = errors.New("community not found")
	ErrInvalidPostType   = errors.New("invalid post type")
	ErrInvalidItemType   = errors.New("invalid related item type")
	ErrItemIDRequired    = errors.New("related_item_id is required when related_item_type is set")
)

var validPostTypes = map[string]bool{
	models.PostTypeDiscussion:            true,
	models.PostTypeReview:                true,
	models.PostTypePoll:                  true,
	models.PostTypeRecommendationRequest: true,
}

var validItemTypes = map[string]bool{
	"movie": true, "series": true, "book": true, "game": true,
}

const maxPostsPerHour = 5

type Service struct {
	posts       *repositories.PostRepository
	communities *communities.Service
}

func NewService(communityService *communities.Service) *Service {
	return &Service{
		posts:       repositories.NewPostRepository(database.DbConn),
		communities: communityService,
	}
}

type ListResult struct {
	Items []models.Post `json:"items"`
	Total int64         `json:"total"`
	Page  int           `json:"page"`
	Pages int           `json:"pages"`
}

type CreateInput struct {
	Title           string `json:"title"`
	Body            string `json:"body"`
	PostType        string `json:"post_type"`
	RelatedItemType string `json:"related_item_type"`
	RelatedItemID   *uint  `json:"related_item_id"`
	IsSpoiler       bool   `json:"is_spoiler"`
}

type UpdateInput struct {
	Title     *string `json:"title"`
	Body      *string `json:"body"`
	IsSpoiler *bool   `json:"is_spoiler"`
}

func (s *Service) ListByCommunity(communityID uint, sort string, page, perPage int) (*ListResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	items, total, err := s.posts.ListByCommunity(communityID, sort, perPage, offset)
	if err != nil {
		return nil, err
	}
	pages := int((total + int64(perPage) - 1) / int64(perPage))
	return &ListResult{Items: items, Total: total, Page: page, Pages: pages}, nil
}

func (s *Service) ListByItem(itemType string, itemID uint, page, perPage int) (*ListResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	items, total, err := s.posts.ListByItem(itemType, itemID, perPage, offset)
	if err != nil {
		return nil, err
	}
	pages := int((total + int64(perPage) - 1) / int64(perPage))
	return &ListResult{Items: items, Total: total, Page: page, Pages: pages}, nil
}

func (s *Service) GetByID(id uint) (*models.Post, error) {
	post, err := s.posts.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if post.Status == models.PostStatusDeleted {
		return nil, ErrNotFound
	}
	return post, nil
}

func (s *Service) Create(userID uint, communityID uint, input CreateInput) (*models.Post, error) {
	if _, err := s.communities.GetByID(communityID); err != nil {
		if errors.Is(err, communities.ErrNotFound) {
			return nil, ErrCommunityNotFound
		}
		return nil, err
	}

	active, status, err := s.communities.IsMember(communityID, userID)
	if err != nil {
		return nil, err
	}
	if status == models.MemberStatusBanned {
		return nil, ErrBanned
	}
	if !active {
		return nil, ErrMustJoin
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, ErrInvalidTitle
	}

	postType := strings.TrimSpace(input.PostType)
	if postType == "" {
		postType = models.PostTypeDiscussion
	}
	if !validPostTypes[postType] {
		return nil, ErrInvalidPostType
	}

	relatedItemType := strings.TrimSpace(input.RelatedItemType)
	if relatedItemType != "" {
		if !validItemTypes[relatedItemType] {
			return nil, ErrInvalidItemType
		}
	}

	since := time.Now().Add(-time.Hour).Format(time.RFC3339)
	recent, err := s.posts.CountRecentByUser(userID, since)
	if err != nil {
		return nil, err
	}
	if recent >= maxPostsPerHour {
		return nil, ErrRateLimitPost
	}

	post := &models.Post{
		CommunityID:     communityID,
		UserID:          userID,
		Title:           title,
		Body:            strings.TrimSpace(input.Body),
		PostType:        postType,
		RelatedItemType: relatedItemType,
		RelatedItemID:   input.RelatedItemID,
		IsSpoiler:       input.IsSpoiler,
		Status:          models.PostStatusPublished,
	}

	txErr := database.DbConn.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(post).Error; err != nil {
			return err
		}
		return tx.Model(&models.Community{}).
			Where("id = ?", communityID).
			UpdateColumn("posts_count", gorm.Expr("posts_count + 1")).Error
	})
	if txErr != nil {
		return nil, txErr
	}
	return s.posts.GetByID(post.ID)
}

func (s *Service) Update(userID uint, id uint, input UpdateInput) (*models.Post, error) {
	post, err := s.posts.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if post.Status == models.PostStatusDeleted {
		return nil, ErrNotFound
	}
	if post.UserID != userID {
		return nil, ErrForbidden
	}

	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" {
			return nil, ErrInvalidTitle
		}
		post.Title = title
	}
	if input.Body != nil {
		post.Body = strings.TrimSpace(*input.Body)
	}
	if input.IsSpoiler != nil {
		post.IsSpoiler = *input.IsSpoiler
	}

	if err := s.posts.Update(post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *Service) Delete(userID uint, id uint) error {
	post, err := s.posts.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if post.Status == models.PostStatusDeleted {
		return ErrNotFound
	}
	if post.UserID != userID && !s.canModerate(post.CommunityID, userID) {
		return ErrForbidden
	}

	return database.DbConn.Transaction(func(tx *gorm.DB) error {
		if err := s.posts.SoftDelete(post); err != nil {
			return err
		}
		return tx.Model(&models.Community{}).
			Where("id = ?", post.CommunityID).
			UpdateColumn("posts_count", gorm.Expr("GREATEST(posts_count - 1, 0)")).Error
	})
}

func (s *Service) canModerate(communityID, userID uint) bool {
	memberRepo := repositories.NewCommunityMemberRepository(database.DbConn)
	member, err := memberRepo.GetByCommunityAndUser(communityID, userID)
	if err != nil {
		return false
	}
	if member.Status != models.MemberStatusActive {
		return false
	}
	return member.Role == models.MemberRoleAdmin || member.Role == models.MemberRoleModerator
}

func (s *Service) EnsureCanComment(post *models.Post, userID uint) error {
	if post.IsLocked {
		return ErrLocked
	}
	active, status, err := s.communities.IsMember(post.CommunityID, userID)
	if err != nil {
		return err
	}
	if status == models.MemberStatusBanned {
		return ErrBanned
	}
	if !active {
		return ErrMustJoin
	}
	return nil
}

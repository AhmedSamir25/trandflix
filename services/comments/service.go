package comments

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"trendflix/database"
	"trendflix/models"
	"trendflix/repositories"
)

var (
	ErrNotFound         = errors.New("comment not found")
	ErrInvalidBody      = errors.New("comment body is required")
	ErrParentNotFound   = errors.New("parent comment not found")
	ErrForbidden        = errors.New("you are not allowed to perform this action")
	ErrRateLimitComment = errors.New("you are commenting too fast, try again later")
)

const maxCommentsPerHour = 30

type Service struct {
	comments *repositories.CommentRepository
	posts    *repositories.PostRepository
}

func NewService() *Service {
	return &Service{
		comments: repositories.NewCommentRepository(database.DbConn),
		posts:    repositories.NewPostRepository(database.DbConn),
	}
}

type Node struct {
	models.Comment
	Replies []Node `json:"replies"`
}

type CreateInput struct {
	Body     string `json:"body"`
	ParentID *uint  `json:"parent_id"`
}

type UpdateInput struct {
	Body string `json:"body"`
}

func (s *Service) Tree(postID uint) ([]Node, error) {
	list, err := s.comments.ListByPost(postID)
	if err != nil {
		return nil, err
	}
	return buildTree(list), nil
}

func buildTree(comments []models.Comment) []Node {
	byID := make(map[uint]models.Comment, len(comments))
	children := make(map[uint][]uint, len(comments))
	var rootIDs []uint
	var roots []Node

	for _, c := range comments {
		byID[c.ID] = c
	}

	for _, c := range comments {
		if c.ParentID == nil {
			rootIDs = append(rootIDs, c.ID)
			continue
		}
		if _, ok := byID[*c.ParentID]; !ok {
			rootIDs = append(rootIDs, c.ID)
			continue
		}
		children[*c.ParentID] = append(children[*c.ParentID], c.ID)
	}

	var build func(uint, map[uint]bool) Node
	build = func(id uint, seen map[uint]bool) Node {
		seen[id] = true
		node := Node{Comment: byID[id]}
		for _, childID := range children[id] {
			if seen[childID] {
				continue
			}
			node.Replies = append(node.Replies, build(childID, seen))
		}
		delete(seen, id)
		return node
	}

	for _, id := range rootIDs {
		roots = append(roots, build(id, map[uint]bool{}))
	}
	return roots
}

func (s *Service) Create(userID uint, postID uint, input CreateInput) (*models.Comment, error) {
	body := strings.TrimSpace(input.Body)
	if body == "" {
		return nil, ErrInvalidBody
	}

	since := time.Now().Add(-time.Hour).Format(time.RFC3339)
	recent, err := s.comments.CountRecentByUser(userID, since)
	if err != nil {
		return nil, err
	}
	if recent >= maxCommentsPerHour {
		return nil, ErrRateLimitComment
	}

	if input.ParentID != nil && *input.ParentID != 0 {
		parent, err := s.comments.GetByID(*input.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrParentNotFound
			}
			return nil, err
		}
		if parent.PostID != postID {
			return nil, ErrParentNotFound
		}
	}

	comment := &models.Comment{
		PostID:   postID,
		UserID:   userID,
		ParentID: input.ParentID,
		Body:     body,
		Status:   models.CommentStatusPublished,
	}

	txErr := database.DbConn.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}
		return tx.Model(&models.Post{}).
			Where("id = ?", postID).
			UpdateColumn("comments_count", gorm.Expr("comments_count + 1")).Error
	})
	if txErr != nil {
		return nil, txErr
	}
	return s.comments.GetByID(comment.ID)
}

func (s *Service) Update(userID uint, id uint, input UpdateInput) (*models.Comment, error) {
	comment, err := s.comments.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if comment.Status == models.CommentStatusDeleted {
		return nil, ErrNotFound
	}
	if comment.UserID != userID {
		return nil, ErrForbidden
	}

	body := strings.TrimSpace(input.Body)
	if body == "" {
		return nil, ErrInvalidBody
	}
	comment.Body = body

	if err := s.comments.Update(comment); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *Service) Delete(userID uint, id uint) error {
	comment, err := s.comments.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if comment.Status == models.CommentStatusDeleted {
		return ErrNotFound
	}
	if comment.UserID != userID && !s.canModerate(comment.PostID, userID) {
		return ErrForbidden
	}

	return database.DbConn.Transaction(func(tx *gorm.DB) error {
		if err := s.comments.SoftDelete(comment); err != nil {
			return err
		}
		return tx.Model(&models.Post{}).
			Where("id = ?", comment.PostID).
			UpdateColumn("comments_count", gorm.Expr("GREATEST(comments_count - 1, 0)")).Error
	})
}

func (s *Service) canModerate(postID, userID uint) bool {
	memberRepo := repositories.NewCommunityMemberRepository(database.DbConn)
	post, err := s.posts.GetByID(postID)
	if err != nil {
		return false
	}
	member, err := memberRepo.GetByCommunityAndUser(post.CommunityID, userID)
	if err != nil {
		return false
	}
	if member.Status != models.MemberStatusActive {
		return false
	}
	return member.Role == models.MemberRoleAdmin || member.Role == models.MemberRoleModerator
}

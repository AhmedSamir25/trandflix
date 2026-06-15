package votes

import (
	"errors"
	"strings"

	"gorm.io/gorm"
	"trendflix/database"
	"trendflix/models"
	"trendflix/repositories"
)

var (
	ErrInvalidVoteType = errors.New("vote_type must be 'up' or 'down'")
	ErrNotFound        = errors.New("target not found")
)

type Service struct {
	votes    *repositories.VoteRepository
	posts    *repositories.PostRepository
	comments *repositories.CommentRepository
}

func NewService() *Service {
	return &Service{
		votes:    repositories.NewVoteRepository(database.DbConn),
		posts:    repositories.NewPostRepository(database.DbConn),
		comments: repositories.NewCommentRepository(database.DbConn),
	}
}

type VoteInput struct {
	VoteType string `json:"vote_type"`
}

func (s *Service) VotePost(userID, postID uint, input VoteInput) (*models.Post, error) {
	voteType := strings.TrimSpace(input.VoteType)
	if voteType != models.VoteTypeUp && voteType != models.VoteTypeDown {
		return nil, ErrInvalidVoteType
	}

	post, err := s.posts.GetByID(postID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if post.Status == models.PostStatusDeleted {
		return nil, ErrNotFound
	}

	delta, err := s.applyVote(userID, postID, models.VotableTypePost, voteType)
	if err != nil {
		return nil, err
	}
	if delta != 0 {
		if err := s.posts.AdjustScore(postID, delta); err != nil {
			return nil, err
		}
	}
	return s.posts.GetByID(postID)
}

func (s *Service) VoteComment(userID, commentID uint, input VoteInput) (*models.Comment, error) {
	voteType := strings.TrimSpace(input.VoteType)
	if voteType != models.VoteTypeUp && voteType != models.VoteTypeDown {
		return nil, ErrInvalidVoteType
	}

	comment, err := s.comments.GetByID(commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if comment.Status == models.CommentStatusDeleted {
		return nil, ErrNotFound
	}

	delta, err := s.applyVote(userID, commentID, models.VotableTypeComment, voteType)
	if err != nil {
		return nil, err
	}
	if delta != 0 {
		if err := s.comments.AdjustScore(commentID, delta); err != nil {
			return nil, err
		}
	}
	return s.comments.GetByID(commentID)
}

// applyVote returns the score delta to apply to the target.
//   +1: new upvote
//   -1: new downvote
//   +2: down -> up
//   -2: up -> down
//    0: same vote toggled off (removed)
func (s *Service) applyVote(userID, votableID uint, votableType, voteType string) (int, error) {
	existing, err := s.votes.GetByUserAndVotable(userID, votableID, votableType)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		vote := &models.Vote{
			UserID:      userID,
			VotableID:   votableID,
			VotableType: votableType,
			VoteType:    voteType,
		}
		if err := s.votes.Create(vote); err != nil {
			return 0, err
		}
		if voteType == models.VoteTypeUp {
			return 1, nil
		}
		return -1, nil
	}

	if existing.VoteType == voteType {
		if err := s.votes.Delete(existing); err != nil {
			return 0, err
		}
		if voteType == models.VoteTypeUp {
			return -1, nil
		}
		return 1, nil
	}

	if err := s.votes.UpdateVoteType(existing, voteType); err != nil {
		return 0, err
	}
	if voteType == models.VoteTypeUp {
		return 2, nil
	}
	return -2, nil
}

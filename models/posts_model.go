package models

import "time"

const (
	PostTypeDiscussion           = "discussion"
	PostTypeReview               = "review"
	PostTypePoll                 = "poll"
	PostTypeRecommendationRequest = "recommendation_request"

	PostStatusPublished = "published"
	PostStatusHidden    = "hidden"
	PostStatusDeleted   = "deleted"
)

type Post struct {
	ID              uint       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CommunityID     uint       `gorm:"column:community_id;not null;index" json:"community_id"`
	Community       Community  `gorm:"foreignKey:CommunityID;references:ID" json:"community,omitempty"`
	UserID          uint       `gorm:"column:user_id;not null;index" json:"user_id"`
	User            User       `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Title           string     `gorm:"column:title;not null" json:"title"`
	Body            string     `gorm:"column:body;type:text" json:"body"`
	PostType        string     `gorm:"column:post_type;type:varchar(50);not null;default:discussion" json:"post_type"`
	RelatedItemID   *uint      `gorm:"column:related_item_id;index" json:"related_item_id"`
	RelatedItemType string     `gorm:"column:related_item_type;type:varchar(50);index" json:"related_item_type"`
	IsSpoiler       bool       `gorm:"column:is_spoiler;default:false" json:"is_spoiler"`
	IsPinned        bool       `gorm:"column:is_pinned;default:false" json:"is_pinned"`
	IsLocked        bool       `gorm:"column:is_locked;default:false" json:"is_locked"`
	Score           int        `gorm:"column:score;default:0" json:"score"`
	CommentsCount   int        `gorm:"column:comments_count;default:0" json:"comments_count"`
	Status          string     `gorm:"column:status;type:varchar(50);not null;default:published;index" json:"status"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime;index" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Post) TableName() string {
	return "posts"
}

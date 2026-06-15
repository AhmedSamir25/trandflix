package models

import "time"

const (
	CategoryTypeMovies = "movies"
	CategoryTypeSeries = "series"
	CategoryTypeBooks  = "books"
	CategoryTypeGames  = "games"
	CategoryTypeMixed  = "mixed"

	CommunityStatusPending  = "pending"
	CommunityStatusApproved = "approved"
	CommunityStatusRejected = "rejected"
)

type Community struct {
	ID           uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name         string    `gorm:"column:name;type:varchar(191);not null" json:"name"`
	Slug         string    `gorm:"column:slug;type:varchar(191);not null;uniqueIndex" json:"slug"`
	Description  string    `gorm:"column:description;type:text" json:"description"`
	AvatarImage  string    `gorm:"column:avatar_image" json:"avatar_image"`
	CoverImage   string    `gorm:"column:cover_image" json:"cover_image"`
	CategoryType string    `gorm:"column:category_type;type:varchar(191);not null;index" json:"category_type"`
	CreatedBy    uint      `gorm:"column:created_by;not null;index" json:"created_by"`
	User         User      `gorm:"foreignKey:CreatedBy;references:ID" json:"user,omitempty"`
	Rules        string    `gorm:"column:rules;type:text" json:"rules"`
	IsPrivate    bool      `gorm:"column:is_private;default:false" json:"is_private"`
	Status       string    `gorm:"column:status;type:varchar(50);not null;default:approved;index" json:"status"`
	MembersCount int       `gorm:"column:members_count;default:0" json:"members_count"`
	PostsCount   int       `gorm:"column:posts_count;default:0" json:"posts_count"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Community) TableName() string {
	return "communities"
}

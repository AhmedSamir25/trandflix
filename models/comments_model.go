package models

import "time"

const (
	CommentStatusPublished = "published"
	CommentStatusHidden    = "hidden"
	CommentStatusDeleted   = "deleted"
)

type Comment struct {
	ID        uint       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PostID    uint       `gorm:"column:post_id;not null;index" json:"post_id"`
	Post      Post       `gorm:"foreignKey:PostID;references:ID" json:"post,omitempty"`
	UserID    uint       `gorm:"column:user_id;not null;index" json:"user_id"`
	User      User       `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	ParentID  *uint      `gorm:"column:parent_id;index" json:"parent_id"`
	Parent    *Comment   `gorm:"foreignKey:ParentID;references:ID" json:"parent,omitempty"`
	Replies   []Comment  `gorm:"foreignKey:ParentID;references:ID" json:"replies,omitempty"`
	Body      string     `gorm:"column:body;type:text;not null" json:"body"`
	Score     int        `gorm:"column:score;default:0" json:"score"`
	Status    string     `gorm:"column:status;type:varchar(50);not null;default:published;index" json:"status"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Comment) TableName() string {
	return "comments"
}

package models

import "time"

const (
	VotableTypePost    = "post"
	VotableTypeComment = "comment"

	VoteTypeUp   = "up"
	VoteTypeDown = "down"
)

type Vote struct {
	ID          uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID      uint      `gorm:"column:user_id;not null;index;uniqueIndex:idx_votes_user_votable" json:"user_id"`
	User        User      `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	VotableID   uint      `gorm:"column:votable_id;not null;uniqueIndex:idx_votes_user_votable" json:"votable_id"`
	VotableType string    `gorm:"column:votable_type;type:varchar(50);not null;uniqueIndex:idx_votes_user_votable" json:"votable_type"`
	VoteType    string    `gorm:"column:vote_type;type:varchar(10);not null" json:"vote_type"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Vote) TableName() string {
	return "votes"
}

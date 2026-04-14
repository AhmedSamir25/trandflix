package models

import "time"

type ResetToken struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"column:user_id;not null;index;uniqueIndex:idx_reset_tokens_user_code" json:"user_id"`
	Code      string    `gorm:"column:code;type:char(6);not null;uniqueIndex:idx_reset_tokens_user_code" json:"-"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null;index" json:"expires_at"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (ResetToken) TableName() string {
	return "reset_tokens"
}

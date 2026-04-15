package models

import "time"

type Review struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"column:user_id;not null;index;uniqueIndex:idx_reviews_user_item" json:"user_id"`
	ItemID    uint      `gorm:"column:item_id;not null;index;uniqueIndex:idx_reviews_user_item" json:"item_id"`
	Rating    uint      `gorm:"column:rating;type:tinyint unsigned;not null" json:"rating"`
	Comment   string    `gorm:"column:comment;type:text" json:"comment"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Review) TableName() string {
	return "reviews"
}

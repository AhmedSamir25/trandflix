package models

import "time"

type Favorite struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"column:user_id;not null" json:"user_id"`
	ItemID    uint      `gorm:"column:item_id;not null" json:"item_id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (Favorite) TableName() string {
	return "favorites"
}

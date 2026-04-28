package models

import "time"

type WatchLater struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"column:user_id;not null;uniqueIndex:idx_watch_later_user_item" json:"user_id"`
	ItemID    uint      `gorm:"column:item_id;not null;uniqueIndex:idx_watch_later_user_item" json:"item_id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (WatchLater) TableName() string {
	return "watch_later"
}

package models

import "time"

type UserListItem struct {
	ID         uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserListID uint      `gorm:"column:user_list_id;not null;uniqueIndex:idx_user_list_item_list_item" json:"user_list_id"`
	ItemID     uint      `gorm:"column:item_id;not null;uniqueIndex:idx_user_list_item_list_item" json:"item_id"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (UserListItem) TableName() string {
	return "user_list_items"
}

package models

import "time"

type UserList struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"column:user_id;not null;index" json:"user_id"`
	Name      string    `gorm:"column:name;type:varchar(255);not null" json:"name"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (UserList) TableName() string {
	return "user_lists"
}

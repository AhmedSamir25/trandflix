package models

import "time"

type User struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;not null" json:"name"`
	Email     string    `gorm:"column:email;type:varchar(191);not null;uniqueIndex" json:"email"`
	Password  string    `gorm:"column:password;not null" json:"-"`
	Avatar    string    `gorm:"column:avatar" json:"avatar"`
	Role      string    `gorm:"column:role;not null;default:user" json:"role"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

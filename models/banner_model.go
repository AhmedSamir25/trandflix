package models

import "time"

type Banner struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Title     string    `gorm:"column:title;not null" json:"title"`
	Subtitle  string    `gorm:"column:subtitle;type:text" json:"subtitle"`
	ImageURL  string    `gorm:"column:image_url;not null" json:"image_url"`
	LinkURL   string    `gorm:"column:link_url" json:"link_url"`
	IsActive  bool      `gorm:"column:is_active;default:true" json:"is_active"`
	SortOrder int       `gorm:"column:sort_order;default:0" json:"sort_order"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Banner) TableName() string {
	return "banners"
}

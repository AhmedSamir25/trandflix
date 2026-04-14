package models

import "time"

type Item struct {
	ID          uint       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Title       string     `gorm:"column:title;not null" json:"title"`
	Description string     `gorm:"column:description;type:text" json:"description"`
	Type        string     `gorm:"column:type;not null" json:"type"`
	CoverImage  string     `gorm:"column:cover_image" json:"cover_image"`
	ReleaseDate time.Time  `gorm:"column:release_date;type:date" json:"release_date"`
	Author      *string    `gorm:"column:author" json:"author"`
	Director    *string    `gorm:"column:director" json:"director"`
	Developer   *string    `gorm:"column:developer" json:"developer"`
	Duration    *uint      `gorm:"column:duration" json:"duration"`
	PagesCount  *uint      `gorm:"column:pages_count" json:"pages_count"`
	Platform    *string    `gorm:"column:platform" json:"platform"`
	Rating      float64    `gorm:"column:rating" json:"rating"`
	Categories  []Category `gorm:"many2many:category_item;joinForeignKey:ItemID;joinReferences:CategoryID" json:"categories,omitempty"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Item) TableName() string {
	return "items"
}

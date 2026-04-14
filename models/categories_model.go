package models

import "time"

type Category struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;not null" json:"name"`
	Slug      string    `gorm:"column:slug;not null" json:"slug"`
	Items     []Item    `gorm:"many2many:category_item;joinForeignKey:CategoryID;joinReferences:ItemID" json:"items,omitempty"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Category) TableName() string {
	return "categories"
}

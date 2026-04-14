package models

type CategoryItem struct {
	ID         uint `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ItemID     uint `gorm:"column:item_id;not null;index;uniqueIndex:idx_category_item_item_category" json:"item_id"`
	CategoryID uint `gorm:"column:category_id;not null;index;uniqueIndex:idx_category_item_item_category" json:"category_id"`
}

func (CategoryItem) TableName() string {
	return "category_item"
}

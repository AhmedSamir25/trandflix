package models

import "time"

type SubscriptionPlan struct {
	ID               uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name             string    `gorm:"column:name;not null" json:"name"`
	Price            float64   `gorm:"column:price;not null" json:"price"`
	Currency         string    `gorm:"column:currency;not null;default:USD" json:"currency"`
	BillingPeriodDays uint     `gorm:"column:billing_period_days;not null;default:30" json:"billing_period_days"`
	IsActive         bool      `gorm:"column:is_active;not null;default:true" json:"is_active"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (SubscriptionPlan) TableName() string {
	return "subscription_plans"
}

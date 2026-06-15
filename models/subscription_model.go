package models

import "time"

const (
	SubscriptionStatusActive    = "active"
	SubscriptionStatusCancelled = "cancelled"
	SubscriptionStatusExpired   = "expired"
)

type Subscription struct {
	ID        uint       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    uint       `gorm:"column:user_id;not null;index" json:"user_id"`
	PlanID    uint       `gorm:"column:plan_id;not null" json:"plan_id"`
	Status    string     `gorm:"column:status;not null" json:"status"`
	StartsAt  time.Time  `gorm:"column:starts_at;not null" json:"starts_at"`
	EndsAt    time.Time  `gorm:"column:ends_at;not null" json:"ends_at"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	Plan SubscriptionPlan `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}

func (s Subscription) IsActive() bool {
	if s.Status != SubscriptionStatusActive {
		return false
	}
	now := time.Now()
	return now.After(s.StartsAt) && now.Before(s.EndsAt)
}

package models

import "time"

const (
	PaymentStatusPaid   = "paid"
	PaymentStatusFailed = "failed"
)

type Payment struct {
	ID             uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID         uint      `gorm:"column:user_id;not null;index" json:"user_id"`
	SubscriptionID *uint     `gorm:"column:subscription_id" json:"subscription_id"`
	PlanID         uint      `gorm:"column:plan_id;not null" json:"plan_id"`
	Amount         float64   `gorm:"column:amount;not null" json:"amount"`
	Currency       string    `gorm:"column:currency;not null;default:USD" json:"currency"`
	Status         string    `gorm:"column:status;not null" json:"status"`
	MockCardLast4  string    `gorm:"column:mock_card_last4;size:4" json:"mock_card_last4"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (Payment) TableName() string {
	return "payments"
}

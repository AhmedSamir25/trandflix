package models

import "time"

const (
	AIProviderOpenRouter       = "openrouter"
	AIProviderOpenAICompatible = "openai_compatible"
)

// AISetting stores the AI provider configuration as a single row (ID = 1).
// Values saved from the admin dashboard take priority over the .env fallbacks
// baked into the recommender.
type AISetting struct {
	ID                      uint      `gorm:"column:id;primaryKey" json:"id"`
	Provider                string    `gorm:"column:provider;not null;default:openai_compatible" json:"provider"`
	OpenRouterAPIKey        string    `gorm:"column:openrouter_api_key" json:"openrouter_api_key"`
	OpenRouterModel         string    `gorm:"column:openrouter_model" json:"openrouter_model"`
	OpenAICompatibleAPIKey  string    `gorm:"column:openai_compatible_api_key" json:"openai_compatible_api_key"`
	OpenAICompatibleBaseURL string    `gorm:"column:openai_compatible_base_url" json:"openai_compatible_base_url"`
	OpenAICompatibleModel   string    `gorm:"column:openai_compatible_model" json:"openai_compatible_model"`
	CreatedAt               time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt               time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (AISetting) TableName() string {
	return "ai_settings"
}

package aisettingscontroller

import (
	"errors"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"trendflix/database"
	"trendflix/models"
)

type updateAISettingsRequest struct {
	Provider                string `json:"provider"`
	OpenRouterAPIKey        string `json:"openrouter_api_key"`
	OpenRouterModel         string `json:"openrouter_model"`
	OpenAICompatibleAPIKey  string `json:"openai_compatible_api_key"`
	OpenAICompatibleBaseURL string `json:"openai_compatible_base_url"`
	OpenAICompatibleModel   string `json:"openai_compatible_model"`
}

const aiKeyMaskPlaceholder = "••••"

func maskAPIKey(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 4 {
		return aiKeyMaskPlaceholder
	}
	return aiKeyMaskPlaceholder + value[len(value)-4:]
}

// GetAISettings returns the AI configuration. API keys are masked so the real
// secret is never exposed to the browser.
func GetAISettings(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "AI settings fetched successfully",
	}

	if database.DbConn == nil {
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	var settings models.AISetting
	result := database.DbConn.First(&settings)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "No AI settings found"
			return c.Status(fiber.StatusNotFound).JSON(context)
		}
		log.Println("Error querying AI settings:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["settings"] = fiber.Map{
		"provider":                  settings.Provider,
		"openrouter_api_key":        maskAPIKey(settings.OpenRouterAPIKey),
		"openrouter_api_key_set":    strings.TrimSpace(settings.OpenRouterAPIKey) != "",
		"openrouter_model":          settings.OpenRouterModel,
		"openai_compatible_api_key": maskAPIKey(settings.OpenAICompatibleAPIKey),
		"openai_compatible_api_key_set": strings.TrimSpace(settings.OpenAICompatibleAPIKey) != "",
		"openai_compatible_base_url":    settings.OpenAICompatibleBaseURL,
		"openai_compatible_model":       settings.OpenAICompatibleModel,
	}
	return c.Status(fiber.StatusOK).JSON(context)
}

// UpdateAISettings updates the AI configuration. API keys are only overwritten
// when a non-empty, non-masked value is submitted so the masked display in the
// browser cannot clobber the stored secret.
func UpdateAISettings(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "AI settings updated successfully",
	}

	if database.DbConn == nil {
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	var request updateAISettingsRequest
	if err := c.BodyParser(&request); err != nil {
		context["statusText"] = "bad"
		context["msg"] = "Invalid request"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	provider := strings.ToLower(strings.TrimSpace(request.Provider))
	if provider != models.AIProviderOpenRouter && provider != models.AIProviderOpenAICompatible {
		context["statusText"] = "bad"
		context["msg"] = "Provider must be openrouter or openai_compatible"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	var settings models.AISetting
	result := database.DbConn.First(&settings)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			context["statusText"] = "bad"
			context["msg"] = "No AI settings found"
			return c.Status(fiber.StatusNotFound).JSON(context)
		}
		log.Println("Error querying AI settings:", result.Error)
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	settings.Provider = provider
	settings.OpenRouterModel = strings.TrimSpace(request.OpenRouterModel)
	settings.OpenAICompatibleBaseURL = strings.TrimSpace(request.OpenAICompatibleBaseURL)
	settings.OpenAICompatibleModel = strings.TrimSpace(request.OpenAICompatibleModel)

	if incoming := strings.TrimSpace(request.OpenRouterAPIKey); incoming != "" && incoming != maskAPIKey(settings.OpenRouterAPIKey) && !strings.HasPrefix(incoming, aiKeyMaskPlaceholder) {
		settings.OpenRouterAPIKey = incoming
	}
	if incoming := strings.TrimSpace(request.OpenAICompatibleAPIKey); incoming != "" && incoming != maskAPIKey(settings.OpenAICompatibleAPIKey) && !strings.HasPrefix(incoming, aiKeyMaskPlaceholder) {
		settings.OpenAICompatibleAPIKey = incoming
	}

	if err := database.DbConn.Save(&settings).Error; err != nil {
		log.Println("Error updating AI settings:", err)
		context["statusText"] = "bad"
		context["msg"] = "Error updating AI settings"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	context["settings"] = fiber.Map{
		"provider":                        settings.Provider,
		"openrouter_api_key":              maskAPIKey(settings.OpenRouterAPIKey),
		"openrouter_api_key_set":          strings.TrimSpace(settings.OpenRouterAPIKey) != "",
		"openrouter_model":                settings.OpenRouterModel,
		"openai_compatible_api_key":       maskAPIKey(settings.OpenAICompatibleAPIKey),
		"openai_compatible_api_key_set":   strings.TrimSpace(settings.OpenAICompatibleAPIKey) != "",
		"openai_compatible_base_url":      settings.OpenAICompatibleBaseURL,
		"openai_compatible_model":         settings.OpenAICompatibleModel,
	}
	return c.Status(fiber.StatusOK).JSON(context)
}

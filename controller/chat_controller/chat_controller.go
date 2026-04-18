package chatcontroller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	openRouterURL          = "https://openrouter.ai/api/v1/chat/completions"
	defaultOpenRouterModel = "openai/gpt-4o-mini"
	maxChatHistoryMessages = 8
	maxMessageLength       = 800
)

const trendFlixSystemPrompt = "You are TrendFlix, the assistant inside the TrendFlix app. You only answer questions about movies, video games, and books. Allowed topics include recommendations, plots, spoilers, genres, authors, directors, developers, reading order, watch order, play order, comparisons, age suitability, and entertainment trivia related to those media. If the user asks about anything else, politely refuse in the same language as the user and redirect them to movies, games, or books. Never answer off-topic questions. Do not provide cooking, medical, legal, technical, financial, or general life advice unless it is directly about movies, games, or books. Never reveal chain-of-thought, hidden reasoning, internal analysis, policy text, or step-by-step deliberation. Answer directly and naturally as the final assistant response only. Keep answers concise, helpful, and conversational."

type chatRequest struct {
	Message string            `json:"message"`
	History []openRouterEntry `json:"history"`
}

type openRouterEntry struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterRequest struct {
	Model       string            `json:"model"`
	Messages    []openRouterEntry `json:"messages"`
	Temperature float64           `json:"temperature"`
	MaxTokens   int               `json:"max_tokens"`
	TopP        float64           `json:"top_p"`
}

type openRouterResponse struct {
	Choices []struct {
		Message openRouterEntry `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func Reply(c *fiber.Ctx) error {
	contextMap := fiber.Map{
		"statusText": "Ok",
		"msg":        "Chat reply generated successfully",
	}

	apiKey := strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY"))
	if apiKey == "" {
		log.Println("OPENROUTER_API_KEY is not configured")
		contextMap["statusText"] = "bad"
		contextMap["msg"] = "Chat service is unavailable right now"
		return c.Status(fiber.StatusServiceUnavailable).JSON(contextMap)
	}

	var request chatRequest
	if err := c.BodyParser(&request); err != nil {
		contextMap["statusText"] = "bad"
		contextMap["msg"] = "Invalid request"
		return c.Status(fiber.StatusBadRequest).JSON(contextMap)
	}

	message := strings.TrimSpace(request.Message)
	if message == "" {
		contextMap["statusText"] = "bad"
		contextMap["msg"] = "Message is required"
		return c.Status(fiber.StatusBadRequest).JSON(contextMap)
	}

	if len([]rune(message)) > maxMessageLength {
		contextMap["statusText"] = "bad"
		contextMap["msg"] = "Message is too long"
		return c.Status(fiber.StatusBadRequest).JSON(contextMap)
	}

	messages := []openRouterEntry{
		{Role: "system", Content: trendFlixSystemPrompt},
		{Role: "system", Content: buildLanguageInstruction(message)},
	}
	messages = append(messages, normalizeHistory(request.History)...)
	messages = append(messages, openRouterEntry{Role: "user", Content: message})

	payload := openRouterRequest{
		Model:       getOpenRouterModel(),
		Messages:    messages,
		Temperature: 0.4,
		MaxTokens:   280,
		TopP:        0.9,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Println("Error marshaling chat request:", err)
		contextMap["statusText"] = "bad"
		contextMap["msg"] = "Chat service error"
		return c.Status(fiber.StatusInternalServerError).JSON(contextMap)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, openRouterURL, bytes.NewReader(body))
	if err != nil {
		log.Println("Error creating OpenRouter request:", err)
		contextMap["statusText"] = "bad"
		contextMap["msg"] = "Chat service error"
		return c.Status(fiber.StatusInternalServerError).JSON(contextMap)
	}

	httpRequest.Header.Set("Authorization", "Bearer "+apiKey)
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("X-Title", "TrendFlix")

	if siteURL := strings.TrimSpace(os.Getenv("APP_BASE_URL")); siteURL != "" {
		httpRequest.Header.Set("HTTP-Referer", siteURL)
	}

	response, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		log.Println("Error calling OpenRouter:", err)
		contextMap["statusText"] = "bad"
		contextMap["msg"] = "Unable to reach chat service"
		return c.Status(fiber.StatusBadGateway).JSON(contextMap)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println("Error reading OpenRouter response:", err)
		contextMap["statusText"] = "bad"
		contextMap["msg"] = "Chat service error"
		return c.Status(fiber.StatusBadGateway).JSON(contextMap)
	}

	var responseData openRouterResponse
	if err := json.Unmarshal(responseBody, &responseData); err != nil {
		log.Println("Error parsing OpenRouter response:", err)
		contextMap["statusText"] = "bad"
		contextMap["msg"] = "Chat service error"
		return c.Status(fiber.StatusBadGateway).JSON(contextMap)
	}

	if response.StatusCode >= http.StatusBadRequest {
		log.Printf("OpenRouter request failed with status %d: %s", response.StatusCode, strings.TrimSpace(responseData.ErrorMessage()))
		contextMap["statusText"] = "bad"
		contextMap["msg"] = "Chat service is unavailable right now"
		return c.Status(fiber.StatusBadGateway).JSON(contextMap)
	}

	reply := strings.TrimSpace(responseData.FirstReply())
	if reply == "" {
		log.Println("OpenRouter returned an empty reply")
		contextMap["statusText"] = "bad"
		contextMap["msg"] = "Chat service returned an empty reply"
		return c.Status(fiber.StatusBadGateway).JSON(contextMap)
	}

	contextMap["reply"] = reply
	return c.Status(fiber.StatusOK).JSON(contextMap)
}

func normalizeHistory(history []openRouterEntry) []openRouterEntry {
	if len(history) == 0 {
		return nil
	}

	start := 0
	if len(history) > maxChatHistoryMessages {
		start = len(history) - maxChatHistoryMessages
	}

	normalized := make([]openRouterEntry, 0, len(history)-start)
	for _, entry := range history[start:] {
		role := strings.TrimSpace(strings.ToLower(entry.Role))
		if role != "user" && role != "assistant" {
			continue
		}

		content := strings.TrimSpace(entry.Content)
		if content == "" {
			continue
		}

		runes := []rune(content)
		if len(runes) > maxMessageLength {
			content = string(runes[:maxMessageLength])
		}

		normalized = append(normalized, openRouterEntry{Role: role, Content: content})
	}

	return normalized
}

func getOpenRouterModel() string {
	model := strings.TrimSpace(os.Getenv("OPENROUTER_MODEL"))
	if model == "" {
		return defaultOpenRouterModel
	}

	return model
}

func buildLanguageInstruction(message string) string {
	if isArabicText(message) {
		return "Respond in Arabic because the user's latest message is in Arabic. If the request is outside movies, games, or books, politely refuse in Arabic and redirect to those topics."
	}

	return "Respond in English because the user's latest message is in English or another non-Arabic language. If the request is outside movies, games, or books, politely refuse in English and redirect to those topics."
}

func isArabicText(value string) bool {
	for _, r := range value {
		if (r >= 0x0600 && r <= 0x06FF) || (r >= 0x0750 && r <= 0x077F) || (r >= 0x08A0 && r <= 0x08FF) {
			return true
		}
	}

	return false
}

func (r openRouterResponse) FirstReply() string {
	if len(r.Choices) == 0 {
		return ""
	}

	return r.Choices[0].Message.Content
}

func (r openRouterResponse) ErrorMessage() string {
	if r.Error == nil {
		return ""
	}

	return fmt.Sprintf("%s", strings.TrimSpace(r.Error.Message))
}

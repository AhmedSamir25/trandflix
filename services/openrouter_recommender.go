package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
)

type OpenRouterRecommender struct {
	Provider string
	APIKey   string
	Model    string
	URL      string
	Client   *http.Client
}

func NewOpenRouterRecommender() *OpenRouterRecommender {
	provider := getAIProvider()

	return &OpenRouterRecommender{
		Provider: provider,
		APIKey:   getAIAPIKey(provider),
		Model:    getAIModel(provider),
		URL:      getAIURL(provider),
		Client:   &http.Client{Timeout: AIRequestTimeout},
	}
}

func (r *OpenRouterRecommender) Available() bool {
	return strings.TrimSpace(r.APIKey) != ""
}

type openRouterAIRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
	TopP        float64       `json:"top_p"`
}

type openRouterAIResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (r *OpenRouterRecommender) Recommend(ctx context.Context, messages []ChatMessage) (string, error) {
	if !r.Available() {
		return "", errAIModelUnavailable
	}

	payload := openRouterAIRequest{
		Model:       r.Model,
		Messages:    messages,
		Temperature: 0.4,
		MaxTokens:   900,
		TopP:        0.9,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	if ctx == nil {
		ctx = context.Background()
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, r.URL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	request.Header.Set("Authorization", "Bearer "+r.APIKey)
	request.Header.Set("Content-Type", "application/json")
	if r.Provider == aiProviderOpenRouter {
		request.Header.Set("X-Title", "TrendFlix")
	}

	if r.Provider == aiProviderOpenRouter && strings.TrimSpace(os.Getenv("APP_BASE_URL")) != "" {
		siteURL := strings.TrimSpace(os.Getenv("APP_BASE_URL"))
		request.Header.Set("HTTP-Referer", siteURL)
	}

	client := r.Client
	if client == nil {
		client = &http.Client{Timeout: AIRequestTimeout}
	}

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	var parsed openRouterAIResponse
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return "", err
	}

	if response.StatusCode >= http.StatusBadRequest {
		msg := strings.TrimSpace(parsed.firstErrorMessage())
		if msg == "" {
			msg = http.StatusText(response.StatusCode)
		}
		return "", errors.New(msg)
	}

	reply := strings.TrimSpace(parsed.firstReply())
	if reply == "" {
		return "", errors.New("ai service returned an empty reply")
	}

	return reply, nil
}

func (r openRouterAIResponse) firstReply() string {
	if len(r.Choices) == 0 {
		return ""
	}
	return r.Choices[0].Message.Content
}

func (r openRouterAIResponse) firstErrorMessage() string {
	if r.Error == nil {
		return ""
	}
	return strings.TrimSpace(r.Error.Message)
}

func getAIProvider() string {
	provider := strings.ToLower(strings.TrimSpace(os.Getenv("AI_PROVIDER")))
	switch provider {
	case "", aiProviderOpenRouter:
		return aiProviderOpenRouter
	case aiProviderOpenAICompatible, "openai-compatible", "openai", "compatible":
		return aiProviderOpenAICompatible
	default:
		return aiProviderOpenRouter
	}
}

func getAIAPIKey(provider string) string {
	if provider == aiProviderOpenAICompatible {
		for _, key := range []string{"OPENAI_COMPATIBLE_API_KEY", "OPENAI_API_KEY", "AI_API_KEY"} {
			if value := strings.TrimSpace(os.Getenv(key)); value != "" {
				return value
			}
		}
		return ""
	}

	if value := strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY")); value != "" {
		return value
	}
	return strings.TrimSpace(os.Getenv("AI_API_KEY"))
}

func getAIURL(provider string) string {
	if provider == aiProviderOpenAICompatible {
		for _, key := range []string{"OPENAI_COMPATIBLE_BASE_URL", "OPENAI_BASE_URL", "AI_BASE_URL"} {
			if value := normalizeAIChatCompletionsURL(os.Getenv(key)); value != "" {
				return value
			}
		}
		return aiDefaultOpenAICompatibleURL
	}

	return aiOpenRouterURL
}

func getAIModel(provider string) string {
	if provider == aiProviderOpenAICompatible {
		for _, key := range []string{"OPENAI_COMPATIBLE_MODEL", "OPENAI_MODEL", "AI_MODEL"} {
			if value := strings.TrimSpace(os.Getenv(key)); value != "" {
				return value
			}
		}
		return aiDefaultOpenAICompatibleModel
	}

	if model := strings.TrimSpace(os.Getenv("OPENROUTER_MODEL")); model != "" {
		return model
	}
	return aiDefaultOpenRouterModel
}

func normalizeAIChatCompletionsURL(value string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(value), "/")
	if baseURL == "" {
		return ""
	}
	if strings.HasSuffix(baseURL, "/chat/completions") {
		return baseURL
	}
	return baseURL + "/chat/completions"
}

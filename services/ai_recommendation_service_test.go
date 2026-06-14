package services

import (
	"context"
	"errors"
	"testing"

	"trendflix/models"
)

type fakeAIRepository struct {
	items []models.Item
	err   error
}

func (f *fakeAIRepository) SearchItems(keywords []string, itemType string, categorySlugs []string, limit int) ([]models.Item, error) {
	if f.err != nil {
		return nil, f.err
	}

	result := make([]models.Item, 0, len(f.items))
	for _, item := range f.items {
		if itemType != "" && item.Type != itemType {
			continue
		}
		result = append(result, item)
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

type fakeLLMRecommender struct {
	reply string
	err   error
}

func (f *fakeLLMRecommender) Recommend(ctx context.Context, messages []ChatMessage) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.reply, nil
}

func aiItem(id uint, title, itemType string) models.Item {
	return models.Item{
		ID:          id,
		Title:       title,
		Type:        itemType,
		Description: "Description for " + title,
		Rating:      8,
		CoverImage:  "https://example.com/" + title + ".jpg",
		Categories:  []models.Category{{ID: id, Name: "Action"}},
	}
}

func TestDatabaseRecommendationHasValidIDAndActions(t *testing.T) {
	repo := &fakeAIRepository{items: []models.Item{aiItem(15, "Inception", "movie")}}
	client := &fakeLLMRecommender{
		reply: `{"reply":"Here you go.","recommendations":[{"title":"Inception","type":"movie","description":"smart mystery","reason":"matches your request","ref_id":15}]}`,
	}

	service := NewAIRecommendationService(repo, client)
	result, err := service.Recommend(context.Background(), AIRecommendationRequest{
		UserMessage: "I want a smart mystery movie",
		Type:        "movie",
		Limit:       5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Recommendations) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(result.Recommendations))
	}

	rec := result.Recommendations[0]
	if rec.Source != AIRecommendationSourceDatabase {
		t.Errorf("expected source database, got %s", rec.Source)
	}
	if !rec.IsAvailableInApp {
		t.Error("database item should be available in app")
	}
	if rec.ItemID == nil || *rec.ItemID != 15 {
		t.Errorf("expected item id 15, got %v", rec.ItemID)
	}
	if !rec.CanAddToFavorites || !rec.CanAddToWatchLater {
		t.Error("database item should allow favorites and watch later")
	}
}

func TestExternalRecommendationNeverHasIDOrActions(t *testing.T) {
	repo := &fakeAIRepository{items: []models.Item{aiItem(15, "Inception", "movie")}}
	client := &fakeLLMRecommender{
		reply: `{"reply":"Here you go.","recommendations":[
			{"title":"Inception","type":"movie","description":"","reason":"","ref_id":15},
			{"title":"Interstellar","type":"movie","description":"space","reason":"external","ref_id":null}
		]}`,
	}

	service := NewAIRecommendationService(repo, client)
	result, err := service.Recommend(context.Background(), AIRecommendationRequest{
		UserMessage: "smart sci-fi",
		Limit:       5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var external *AIRecommendation
	for i := range result.Recommendations {
		if result.Recommendations[i].Source == AIRecommendationSourceExternal {
			external = &result.Recommendations[i]
		}
	}
	if external == nil {
		t.Fatal("expected at least one external recommendation")
	}

	if external.ItemID != nil {
		t.Errorf("external item must have nil id, got %v", external.ItemID)
	}
	if external.IsAvailableInApp {
		t.Error("external item must not be available in app")
	}
	if external.CanAddToFavorites || external.CanAddToWatchLater {
		t.Error("external item must not allow favorites or watch later")
	}
}

func TestFabricatedRefIDIsTreatedAsExternal(t *testing.T) {
	repo := &fakeAIRepository{items: []models.Item{aiItem(15, "Inception", "movie")}}
	client := &fakeLLMRecommender{
		reply: `{"reply":"ok","recommendations":[
			{"title":"Fake Item","type":"movie","description":"","reason":"","ref_id":9999}
		]}`,
	}

	service := NewAIRecommendationService(repo, client)
	result, err := service.Recommend(context.Background(), AIRecommendationRequest{
		UserMessage: "something",
		Limit:       5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Recommendations) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(result.Recommendations))
	}
	rec := result.Recommendations[0]
	if rec.Source != AIRecommendationSourceExternal {
		t.Errorf("fabricated ref_id should become external, got %s", rec.Source)
	}
	if rec.ItemID != nil {
		t.Errorf("external item must have nil id, got %v", rec.ItemID)
	}
}

func TestSeriesAliasMapsToTvShow(t *testing.T) {
	if normalizeAIType("series") != models.ItemTypeTVShow {
		t.Error("series should map to tv_show")
	}
	if normalizeAIType("tv") != models.ItemTypeTVShow {
		t.Error("tv should map to tv_show")
	}
	if normalizeAIType("all") != "" {
		t.Error("all should map to empty filter")
	}
	if normalizeAIType("movie") != "movie" {
		t.Error("movie should map to movie")
	}
}

func TestFallbackWhenAIUnavailableButDatabaseMatchesExist(t *testing.T) {
	repo := &fakeAIRepository{items: []models.Item{
		aiItem(1, "Matrix", "movie"),
		aiItem(2, "Dune", "movie"),
	}}
	client := &fakeLLMRecommender{err: errors.New("network down")}

	service := NewAIRecommendationService(repo, client)
	result, err := service.Recommend(context.Background(), AIRecommendationRequest{
		UserMessage: "sci-fi movie",
		Limit:       5,
	})
	if err != nil {
		t.Fatalf("expected fallback result, got error: %v", err)
	}
	if len(result.Recommendations) != 2 {
		t.Fatalf("expected 2 fallback recommendations, got %d", len(result.Recommendations))
	}
	for _, rec := range result.Recommendations {
		if rec.Source != AIRecommendationSourceDatabase {
			t.Error("fallback recommendations should come from the database")
		}
		if rec.Reason == "" {
			t.Error("fallback recommendations should include a reason")
		}
	}
}

func TestUnavailableWhenNoDatabaseAndNoAI(t *testing.T) {
	repo := &fakeAIRepository{items: nil}
	client := &fakeLLMRecommender{err: errors.New("network down")}

	service := NewAIRecommendationService(repo, client)
	_, err := service.Recommend(context.Background(), AIRecommendationRequest{
		UserMessage: "anything",
		Limit:       5,
	})
	if !errors.Is(err, ErrAIModelUnavailablePublic) {
		t.Fatalf("expected ErrAIModelUnavailablePublic, got %v", err)
	}
}

func TestAILimitIsClamped(t *testing.T) {
	if normalizeAILimit(0) != AIDefaultLimit {
		t.Error("limit <= 0 should fall back to default")
	}
	if normalizeAILimit(1000) != AIMaxLimit {
		t.Error("limit above max should be clamped")
	}
	if normalizeAILimit(7) != 7 {
		t.Error("valid limit should be returned unchanged")
	}
}

func TestLanguageDetectionFromArabicMessage(t *testing.T) {
	if normalizeAILanguage("", "أريد فيلماً") != "Arabic" {
		t.Error("Arabic message should resolve to Arabic language")
	}
	if normalizeAILanguage("", "I want a movie") != "English" {
		t.Error("English message should resolve to English language")
	}
	if normalizeAILanguage("ar", "hello") != "Arabic" {
		t.Error("explicit arabic language should be respected")
	}
}

func TestParseAIResponseStripsCodeFences(t *testing.T) {
	raw := "```json\n{\"reply\":\"hi\",\"recommendations\":[]}\n```"
	parsed, err := parseAIResponse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.Reply != "hi" {
		t.Errorf("expected reply 'hi', got %q", parsed.Reply)
	}
}

func TestOpenRouterProviderIsDefault(t *testing.T) {
	t.Setenv("AI_PROVIDER", "")
	t.Setenv("OPENROUTER_API_KEY", "router-key")
	t.Setenv("OPENROUTER_MODEL", "deepseek/deepseek-chat")

	client := NewOpenRouterRecommender()
	if client.Provider != aiProviderOpenRouter {
		t.Fatalf("expected openrouter provider, got %q", client.Provider)
	}
	if client.APIKey != "router-key" {
		t.Fatalf("expected OpenRouter key, got %q", client.APIKey)
	}
	if client.Model != "deepseek/deepseek-chat" {
		t.Fatalf("expected OpenRouter model, got %q", client.Model)
	}
	if client.URL != aiOpenRouterURL {
		t.Fatalf("expected OpenRouter URL, got %q", client.URL)
	}
}

func TestOpenAICompatibleProviderUsesCompatibleEnv(t *testing.T) {
	t.Setenv("AI_PROVIDER", "openai_compatible")
	t.Setenv("OPENAI_COMPATIBLE_API_KEY", "compatible-key")
	t.Setenv("OPENAI_COMPATIBLE_BASE_URL", "https://api.deepseek.com/v1")
	t.Setenv("OPENAI_COMPATIBLE_MODEL", "deepseek-chat")

	client := NewOpenRouterRecommender()
	if client.Provider != aiProviderOpenAICompatible {
		t.Fatalf("expected openai compatible provider, got %q", client.Provider)
	}
	if client.APIKey != "compatible-key" {
		t.Fatalf("expected compatible key, got %q", client.APIKey)
	}
	if client.Model != "deepseek-chat" {
		t.Fatalf("expected compatible model, got %q", client.Model)
	}
	expectedURL := "https://api.deepseek.com/v1/chat/completions"
	if client.URL != expectedURL {
		t.Fatalf("expected compatible URL %q, got %q", expectedURL, client.URL)
	}
}

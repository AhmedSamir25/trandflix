package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"trendflix/models"
)

// capturingLLM records the messages handed to the recommender so tests can
// assert that the user's saved items reach the AI prompt.
type capturingLLM struct {
	reply    string
	err      error
	messages []ChatMessage
}

func (c *capturingLLM) Recommend(_ context.Context, messages []ChatMessage) (string, error) {
	c.messages = messages
	if c.err != nil {
		return "", c.err
	}
	return c.reply, nil
}

func forYouItem(id uint, title, itemType string, rating float64, categories ...models.Category) models.Item {
	return models.Item{
		ID:          id,
		Title:       title,
		Type:        itemType,
		Description: "Description for " + title,
		Rating:      rating,
		CoverImage:  "https://example.com/" + title + ".jpg",
		Categories:  categories,
	}
}

func TestAIForYouReturnsAIPickedCandidates(t *testing.T) {
	action := cat(1, "Action")

	repo := &fakeRepository{
		favorites: []models.Item{
			forYouItem(101, "Fav Action", "movie", 8, action),
		},
		candidates: []models.Item{
			forYouItem(201, "Action One", "movie", 9, action),
			forYouItem(202, "Action Two", "movie", 7, action),
		},
		popularity: map[uint]int{},
	}
	client := &fakeLLMRecommender{
		reply: `{"reply":"Based on your taste","recommendations":[{"title":"Action One","type":"movie","description":"","reason":"matches your love of action","ref_id":201}]}`,
	}

	service := NewAIForYouService(repo, client)
	recs, err := service.ForUser(context.Background(), 1, "en", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recs) == 0 {
		t.Fatal("expected recommendations")
	}
	// The AI's pick must come first; remaining slots are filled by the fallback.
	if recs[0].ID != 201 {
		t.Fatalf("expected AI-picked item 201 first, got %d", recs[0].ID)
	}
	if !strings.Contains(strings.ToLower(recs[0].Reason), "action") {
		t.Errorf("expected AI reason to mention taste, got %q", recs[0].Reason)
	}
}

func TestAIForYouExcludesAlreadySavedItems(t *testing.T) {
	action := cat(1, "Action")

	repo := &fakeRepository{
		favorites: []models.Item{
			forYouItem(101, "Fav Action", "movie", 8, action),
		},
		candidates: []models.Item{
			forYouItem(201, "Action Candidate", "movie", 9, action),
		},
		popularity: map[uint]int{},
	}
	// The model is tricked into returning the favorite, which must be ignored
	// because it is not part of the candidate pool.
	client := &fakeLLMRecommender{
		reply: `{"reply":"ok","recommendations":[{"title":"Fav Action","type":"movie","reason":"x","ref_id":101},{"title":"Action Candidate","type":"movie","reason":"y","ref_id":201}]}`,
	}

	service := NewAIForYouService(repo, client)
	recs, err := service.ForUser(context.Background(), 1, "en", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, rec := range recs {
		if rec.ID == 101 {
			t.Fatalf("favorite item %d must never be recommended", rec.ID)
		}
	}
	if len(recs) == 0 {
		t.Fatal("expected the valid candidate to be recommended")
	}
	if recs[0].ID != 201 {
		t.Fatalf("expected candidate 201, got %d", recs[0].ID)
	}
}

func TestAIForYouFallsBackWhenAIUnavailable(t *testing.T) {
	action := cat(1, "Action")

	repo := &fakeRepository{
		favorites: []models.Item{
			forYouItem(101, "Fav Action", "movie", 8, action),
		},
		candidates: []models.Item{
			forYouItem(201, "Action Candidate", "movie", 9, action),
		},
		popularity: map[uint]int{},
	}
	client := &fakeLLMRecommender{err: errors.New("network down")}

	service := NewAIForYouService(repo, client)
	recs, err := service.ForUser(context.Background(), 1, "en", 5)
	if err != nil {
		t.Fatalf("expected fallback recommendations, got error: %v", err)
	}
	if len(recs) == 0 {
		t.Fatal("expected fallback recommendations when AI is unavailable")
	}
	if recs[0].ID != 201 {
		t.Fatalf("expected fallback to pick candidate 201, got %d", recs[0].ID)
	}
}

func TestAIForYouFallsBackOnInvalidJSON(t *testing.T) {
	action := cat(1, "Action")

	repo := &fakeRepository{
		favorites: []models.Item{
			forYouItem(101, "Fav Action", "movie", 8, action),
		},
		candidates: []models.Item{
			forYouItem(201, "Action Candidate", "movie", 9, action),
		},
		popularity: map[uint]int{},
	}
	client := &fakeLLMRecommender{reply: "this is not json at all"}

	service := NewAIForYouService(repo, client)
	recs, err := service.ForUser(context.Background(), 1, "en", 5)
	if err != nil {
		t.Fatalf("expected graceful fallback, got error: %v", err)
	}
	if len(recs) == 0 {
		t.Fatal("expected fallback recommendations on invalid AI output")
	}
}

func TestAIForYouEmptyWhenNoFavoritesOrWatchLater(t *testing.T) {
	repo := &fakeRepository{
		topRated: []models.Item{
			forYouItem(301, "Top Movie", "movie", 9.5, cat(1, "Action")),
			forYouItem(302, "Top Book", "book", 9.0, cat(2, "Drama")),
		},
		popularity: map[uint]int{},
	}
	client := &capturingLLM{
		reply: `{"reply":"hi","recommendations":[{"title":"Top Movie","type":"movie","reason":"x","ref_id":301}]}`,
	}

	service := NewAIForYouService(repo, client)
	recs, err := service.ForUser(context.Background(), 1, "en", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recs) != 0 {
		t.Fatalf("expected no recommendations without favorites/watch later, got %d", len(recs))
	}
	// The AI must never be consulted when there is no taste signal.
	if client.messages != nil {
		t.Fatal("AI must be skipped when the user has no favorites or watch later")
	}
}

func TestAIForYouShowsRecommendationsWithOnlyWatchLater(t *testing.T) {
	action := cat(1, "Action")

	repo := &fakeRepository{
		watchLater: []models.Item{
			forYouItem(102, "Watch Later Action", "movie", 7, action),
		},
		candidates: []models.Item{
			forYouItem(201, "Action Candidate", "movie", 9, action),
		},
		popularity: map[uint]int{},
	}
	client := &fakeLLMRecommender{
		reply: `{"reply":"ok","recommendations":[{"title":"Action Candidate","type":"movie","reason":"matches your watch later","ref_id":201}]}`,
	}

	service := NewAIForYouService(repo, client)
	recs, err := service.ForUser(context.Background(), 1, "en", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recs) == 0 {
		t.Fatal("expected recommendations when the user has watch later items")
	}
	if recs[0].ID != 201 {
		t.Fatalf("expected candidate 201, got %d", recs[0].ID)
	}
}

func TestAIForYouPromptIncludesSavedItems(t *testing.T) {
	action := cat(1, "Action")
	comedy := cat(2, "Comedy")

	repo := &fakeRepository{
		favorites: []models.Item{
			forYouItem(101, "Fav Action Movie", "movie", 8, action),
		},
		watchLater: []models.Item{
			forYouItem(102, "Watch Later Comedy", "movie", 7, comedy),
		},
		listItems: []models.Item{
			forYouItem(103, "Listed Game", "game", 8, action),
		},
		candidates: []models.Item{
			forYouItem(201, "Action Candidate", "movie", 9, action),
		},
		popularity: map[uint]int{},
	}
	client := &capturingLLM{
		reply: `{"reply":"ok","recommendations":[]}`,
	}

	service := NewAIForYouService(repo, client)
	_, err := service.ForUser(context.Background(), 1, "en", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.messages) < 1 {
		t.Fatal("expected the AI model to receive messages")
	}

	systemPrompt := client.messages[0].Content
	for _, needle := range []string{"Fav Action Movie", "Watch Later Comedy", "Listed Game", "Action Candidate"} {
		if !strings.Contains(systemPrompt, needle) {
			t.Errorf("expected AI prompt to include %q", needle)
		}
	}
}

func TestAIForYouFabricatedRefIDIsDropped(t *testing.T) {
	action := cat(1, "Action")

	repo := &fakeRepository{
		favorites: []models.Item{
			forYouItem(101, "Fav Action", "movie", 8, action),
		},
		candidates: []models.Item{
			forYouItem(201, "Action Candidate", "movie", 9, action),
		},
		popularity: map[uint]int{},
	}
	client := &fakeLLMRecommender{
		reply: `{"reply":"ok","recommendations":[{"title":"Ghost","type":"movie","reason":"x","ref_id":9999}]}`,
	}

	service := NewAIForYouService(repo, client)
	recs, err := service.ForUser(context.Background(), 1, "en", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, rec := range recs {
		if rec.ID == 9999 {
			t.Fatal("fabricated ref_id must never appear in recommendations")
		}
	}
}

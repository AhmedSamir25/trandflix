package services

import (
	"testing"

	"trendflix/models"
)

type fakeRepository struct {
	favorites   []models.Item
	watchLater  []models.Item
	listItems   []models.Item
	reviewedIDs []uint
	candidates  []models.Item
	popularity  map[uint]int
	topRated    []models.Item
}

func (f *fakeRepository) FetchFavorites(uint) ([]models.Item, error) {
	return f.favorites, nil
}

func (f *fakeRepository) FetchWatchLater(uint) ([]models.Item, error) {
	return f.watchLater, nil
}

func (f *fakeRepository) FetchListItems(uint) ([]models.Item, error) {
	return f.listItems, nil
}

func (f *fakeRepository) FetchReviewedItemIDs(uint) ([]uint, error) {
	return f.reviewedIDs, nil
}

func (f *fakeRepository) FetchCandidates(categoryIDs, excludedIDs []uint) ([]models.Item, error) {
	categorySet := map[uint]struct{}{}
	for _, id := range categoryIDs {
		categorySet[id] = struct{}{}
	}
	excludedSet := map[uint]struct{}{}
	for _, id := range excludedIDs {
		excludedSet[id] = struct{}{}
	}

	result := make([]models.Item, 0, len(f.candidates))
	for _, candidate := range f.candidates {
		if _, excluded := excludedSet[candidate.ID]; excluded {
			continue
		}
		matchesCategory := false
		for _, category := range candidate.Categories {
			if _, ok := categorySet[category.ID]; ok {
				matchesCategory = true
				break
			}
		}
		if !matchesCategory {
			continue
		}
		result = append(result, candidate)
	}
	return result, nil
}

func (f *fakeRepository) FetchPopularity([]uint) (map[uint]int, error) {
	return f.popularity, nil
}

func (f *fakeRepository) FetchTopRated(excludedIDs []uint, limit int) ([]models.Item, error) {
	excludedSet := map[uint]struct{}{}
	for _, id := range excludedIDs {
		excludedSet[id] = struct{}{}
	}

	result := make([]models.Item, 0, limit)
	for _, item := range f.topRated {
		if _, excluded := excludedSet[item.ID]; excluded {
			continue
		}
		result = append(result, item)
		if len(result) >= limit {
			break
		}
	}
	return result, nil
}

func cat(id uint, name string) models.Category {
	return models.Category{ID: id, Name: name}
}

func item(id uint, title, itemType string, rating float64, categories ...models.Category) models.Item {
	return models.Item{
		ID:         id,
		Title:      title,
		Type:       itemType,
		Rating:     rating,
		Categories: categories,
	}
}

func TestRecommendationsRankedByCategoryInterest(t *testing.T) {
	action := cat(1, "Action")
	comedy := cat(2, "Comedy")

	repo := &fakeRepository{
		favorites: []models.Item{
			item(101, "Fav Action A", "movie", 8, action),
			item(102, "Fav Action B", "movie", 7, action),
		},
		watchLater: []models.Item{
			item(103, "Watch Later Action", "movie", 7, action),
		},
		listItems: []models.Item{
			item(104, "List Comedy", "movie", 6, comedy),
		},
		candidates: []models.Item{
			item(201, "Action Candidate", "movie", 9, action),
			item(202, "Comedy Candidate", "movie", 9, comedy),
			item(203, "Action Comedy Candidate", "movie", 6, action, comedy),
		},
		popularity: map[uint]int{},
	}

	service := NewRecommendationService(repo)
	recommendations, err := service.ForUser(1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(recommendations) == 0 {
		t.Fatal("expected recommendations, got none")
	}

	if recommendations[0].ID != 201 {
		t.Fatalf("expected top recommendation to be Action candidate (201), got %d", recommendations[0].ID)
	}

	if recommendations[0].Reason == "" {
		t.Fatal("expected recommendation to include a reason")
	}

	for _, recommendation := range recommendations {
		if recommendation.ID == 101 || recommendation.ID == 102 ||
			recommendation.ID == 103 || recommendation.ID == 104 {
			t.Fatalf("interacted item %d should not be recommended", recommendation.ID)
		}
	}
}

func TestAlreadyInteractedAndReviewedItemsExcluded(t *testing.T) {
	action := cat(1, "Action")

	repo := &fakeRepository{
		favorites: []models.Item{
			item(101, "Fav Action", "movie", 8, action),
		},
		reviewedIDs: []uint{202},
		candidates: []models.Item{
			item(201, "New Action", "movie", 9, action),
			item(202, "Reviewed Action", "movie", 9, action),
		},
		popularity: map[uint]int{},
	}

	service := NewRecommendationService(repo)
	recommendations, err := service.ForUser(1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ids := map[uint]struct{}{}
	for _, recommendation := range recommendations {
		ids[recommendation.ID] = struct{}{}
	}

	if _, ok := ids[101]; ok {
		t.Error("favorite item should be excluded from recommendations")
	}
	if _, ok := ids[202]; ok {
		t.Error("reviewed item should be excluded from recommendations")
	}
	if _, ok := ids[201]; !ok {
		t.Error("new action item should be recommended")
	}
}

func TestNewUserReceivesFallbackRecommendations(t *testing.T) {
	action := cat(1, "Action")
	drama := cat(3, "Drama")

	repo := &fakeRepository{
		topRated: []models.Item{
			item(301, "Top Movie", "movie", 9.5, action),
			item(302, "Top Book", "book", 9.0, drama),
			item(303, "Top Game", "game", 8.5, action),
		},
		popularity: map[uint]int{},
	}

	service := NewRecommendationService(repo)
	recommendations, err := service.ForUser(1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(recommendations) == 0 {
		t.Fatal("expected fallback recommendations for new user, got none")
	}

	for _, recommendation := range recommendations {
		if recommendation.Reason == "" {
			t.Error("fallback recommendations should include a reason")
		}
	}
}

func TestResultsSortedByScoreDescending(t *testing.T) {
	action := cat(1, "Action")

	repo := &fakeRepository{
		favorites: []models.Item{
			item(101, "Fav Action", "movie", 8, action),
		},
		candidates: []models.Item{
			item(201, "Low Rated Action", "movie", 3, action),
			item(202, "High Rated Action", "movie", 9, action),
			item(203, "Mid Rated Action", "movie", 6, action),
		},
		popularity: map[uint]int{},
	}

	service := NewRecommendationService(repo)
	recommendations, err := service.ForUser(1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(recommendations) < 2 {
		t.Fatal("expected at least two recommendations to verify sorting")
	}

	for i := 1; i < len(recommendations); i++ {
		if recommendations[i].Score > recommendations[i-1].Score {
			t.Fatalf(
				"recommendations not sorted by score descending: %d before %d at index %d",
				recommendations[i-1].Score,
				recommendations[i].Score,
				i,
			)
		}
	}
}

func TestLimitIsClamped(t *testing.T) {
	if normalizeLimit(0) != DefaultLimit {
		t.Error("limit <= 0 should fall back to DefaultLimit")
	}
	if normalizeLimit(1000) != MaxLimit {
		t.Error("limit above MaxLimit should be clamped to MaxLimit")
	}
	if normalizeLimit(5) != 5 {
		t.Error("valid limit should be returned unchanged")
	}
}

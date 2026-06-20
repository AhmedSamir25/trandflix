package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"trendflix/models"
)

const (
	// AIForYouCandidatePool is the maximum number of catalog items offered to
	// the AI model when building personalized recommendations.
	AIForYouCandidatePool = 30

	// AIForYouProfileSample caps how many saved items per source are described
	// to the AI model so the prompt stays compact.
	AIForYouProfileSample = 12

	// AIForYouRequestTimeout caps how long the For You endpoint waits on the AI
	// model before falling back to the instant rule-based recommender.
	AIForYouRequestTimeout = 12 * time.Second
)

// AIForYouService builds personalized "For You" recommendations by feeding the
// user's favorites, watch later, and custom lists to the AI model and asking
// it to pick from catalog candidates.
type AIForYouService struct {
	repo     Repository
	client   LLMRecommender
	fallback *RecommendationService
}

// NewAIForYouService wires the AI for-you service with a rule-based fallback
// that is used when the AI model is unavailable or returns unusable output.
func NewAIForYouService(repo Repository, client LLMRecommender) *AIForYouService {
	return &AIForYouService{
		repo:     repo,
		client:   client,
		fallback: NewRecommendationService(repo),
	}
}

// ForUser returns AI-driven recommendations derived from the user's saved
// items. When the user has no history, or the AI model is unavailable/fails,
// it transparently falls back to the existing rule-based recommender.
func (s *AIForYouService) ForUser(ctx context.Context, userID uint, language string, limit int) ([]Recommendation, error) {
	limit = normalizeLimit(limit)
	language = normalizeAILanguage(language, "")

	favorites, err := s.repo.FetchFavorites(userID)
	if err != nil {
		return nil, err
	}
	watchLater, err := s.repo.FetchWatchLater(userID)
	if err != nil {
		return nil, err
	}

	// Without favorites or watch later there is no taste signal to personalize
	// from, so the For You section is hidden instead of showing generic items.
	if len(favorites) == 0 && len(watchLater) == 0 {
		return []Recommendation{}, nil
	}

	listItems, err := s.repo.FetchListItems(userID)
	if err != nil {
		return nil, err
	}
	reviewedIDs, err := s.repo.FetchReviewedItemIDs(userID)
	if err != nil {
		return nil, err
	}

	profile := buildInterestProfile(favorites, watchLater, listItems)

	excluded := buildExclusionSet(favorites, watchLater, listItems, reviewedIDs)

	candidates, err := s.repo.FetchCandidates(profile.categoryIDs(), excluded.slice())
	if err != nil {
		return nil, err
	}

	if len(candidates) < AIForYouCandidatePool {
		topUp, err := s.repo.FetchTopRated(excluded.slice(), AIForYouCandidatePool)
		if err != nil {
			return nil, err
		}
		candidates = mergeUniqueItems(candidates, topUp)
	}
	candidates = firstItems(candidates, AIForYouCandidatePool)

	if len(candidates) == 0 {
		return s.fallback.ForUser(userID, limit)
	}

	if s.client == nil {
		return s.fallback.ForUser(userID, limit)
	}

	candidateByID := make(map[uint]models.Item, len(candidates))
	for _, item := range candidates {
		candidateByID[item.ID] = item
	}

	messages := buildForYouAIMessages(favorites, watchLater, listItems, candidates, language, limit)

	rawReply, err := s.client.Recommend(ctx, messages)
	if err != nil {
		return s.fallback.ForUser(userID, limit)
	}

	parsed, parseErr := parseAIResponse(rawReply)
	if parseErr != nil {
		return s.fallback.ForUser(userID, limit)
	}

	recommendations := mergeForYouRecommendations(parsed.Recommendations, candidateByID, limit)

	if len(recommendations) < limit {
		recommendations = s.topUpFromFallback(userID, limit, recommendations)
	}

	return recommendations, nil
}

func (s *AIForYouService) topUpFromFallback(userID uint, limit int, current []Recommendation) []Recommendation {
	fallback, err := s.fallback.ForUser(userID, limit)
	if err != nil || len(fallback) == 0 {
		return current
	}

	existing := make(map[uint]struct{}, len(current))
	for _, rec := range current {
		existing[rec.ID] = struct{}{}
	}

	for _, rec := range fallback {
		if len(current) >= limit {
			break
		}
		if _, ok := existing[rec.ID]; ok {
			continue
		}
		existing[rec.ID] = struct{}{}
		current = append(current, rec)
	}
	return current
}

// mergeForYouRecommendations converts the AI's selected ref_ids into catalog
// recommendations. Only valid ref_ids that exist in the candidate pool are
// kept; fabricated ids and external-only suggestions are dropped so the For
// You row always links to real in-app items.
func mergeForYouRecommendations(drafts []aiRecommendationDraft, candidateByID map[uint]models.Item, limit int) []Recommendation {
	seen := make(map[uint]struct{}, len(drafts))
	recommendations := make([]Recommendation, 0, len(drafts))

	for _, draft := range drafts {
		if draft.RefID == nil {
			continue
		}
		item, ok := candidateByID[*draft.RefID]
		if !ok {
			continue
		}
		if _, dup := seen[item.ID]; dup {
			continue
		}
		seen[item.ID] = struct{}{}

		reason := strings.TrimSpace(draft.Reason)
		if reason == "" {
			reason = "Recommended for you based on your taste"
		}

		recommendations = append(recommendations, Recommendation{
			ID:         item.ID,
			Title:      item.Title,
			Type:       item.Type,
			Categories: categoryNames(item.Categories),
			Reason:     reason,
			CoverImage: item.CoverImage,
			Rating:     item.Rating,
		})

		if len(recommendations) >= limit {
			break
		}
	}
	return recommendations
}

// buildForYouAIMessages assembles the chat payload describing the user's
// taste (favorites / watch later / lists) and the candidate catalog items the
// model is allowed to recommend from.
func buildForYouAIMessages(favorites, watchLater, listItems, candidates []models.Item, language string, limit int) []ChatMessage {
	var builder strings.Builder
	builder.WriteString("You are TrendFlix, the AI entertainment assistant. ")
	builder.WriteString("Your task is to personalize recommendations based on what the user already enjoys on TrendFlix.\n\n")
	builder.WriteString("Rules:\n")
	builder.WriteString("- Recommend ONLY from the candidate catalog items listed below, referencing them by ref_id.\n")
	builder.WriteString("- Never invent a ref_id and never suggest items that are not in the candidate list.\n")
	builder.WriteString("- Infer the user's taste from their favorites, watch later, and custom lists, then pick the best matches.\n")
	builder.WriteString("- Every recommendation must include a short, specific reason that references the user's taste.\n")
	builder.WriteString(fmt.Sprintf("- Return at most %d recommendations.\n", limit))
	builder.WriteString("- Respond ONLY with valid minified JSON in exactly this shape: ")
	builder.WriteString(`{"reply":"short friendly reply","recommendations":[{"title":"","type":"movie|tv_show|game|book","description":"","reason":"","ref_id":0,"image":null}]}`)
	builder.WriteString(".\n")
	builder.WriteString("- Do not include any text outside the JSON object.\n")
	builder.WriteString(fmt.Sprintf("- Reply in %s.\n\n", language))

	builder.WriteString("User's favorites:\n")
	writeProfileItemList(&builder, favorites)
	builder.WriteString("\nUser's watch later:\n")
	writeProfileItemList(&builder, watchLater)
	builder.WriteString("\nUser's custom lists:\n")
	writeProfileItemList(&builder, listItems)

	builder.WriteString("\nCandidate catalog items (pick from these only):\n")
	for _, item := range candidates {
		builder.WriteString(fmt.Sprintf(
			"- ref_id:%d | title:%s | type:%s | rating:%.1f | year:%s | categories:%s | description:%s\n",
			item.ID,
			escapeCSV(item.Title),
			item.Type,
			item.Rating,
			formatReleaseYear(item.ReleaseDate),
			escapeCSV(strings.Join(sortedCategoryNames(item.Categories), ", ")),
			escapeCSV(truncate(item.Description, 140)),
		))
	}

	return []ChatMessage{
		{Role: "system", Content: builder.String()},
		{Role: "user", Content: "Recommend items from the catalog that best match my taste."},
	}
}

func writeProfileItemList(builder *strings.Builder, items []models.Item) {
	if len(items) == 0 {
		builder.WriteString("- none\n")
		return
	}

	for i, item := range items {
		if i >= AIForYouProfileSample {
			builder.WriteString(fmt.Sprintf("- (and %d more)\n", len(items)-AIForYouProfileSample))
			break
		}
		builder.WriteString(fmt.Sprintf(
			"- title:%s | type:%s | rating:%.1f | categories:%s\n",
			escapeCSV(item.Title),
			item.Type,
			item.Rating,
			escapeCSV(strings.Join(sortedCategoryNames(item.Categories), ", ")),
		))
	}
}

func mergeUniqueItems(existing, extra []models.Item) []models.Item {
	seen := make(map[uint]struct{}, len(existing))
	merged := make([]models.Item, 0, len(existing)+len(extra))

	for _, item := range existing {
		seen[item.ID] = struct{}{}
		merged = append(merged, item)
	}
	for _, item := range extra {
		if _, ok := seen[item.ID]; ok {
			continue
		}
		seen[item.ID] = struct{}{}
		merged = append(merged, item)
	}
	return merged
}

func firstItems(items []models.Item, limit int) []models.Item {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

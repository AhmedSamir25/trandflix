package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"trendflix/models"
)

const (
	AIRecommendationSourceDatabase = "database"
	AIRecommendationSourceExternal = "external"

	AITypeAll    = "all"
	AITypeSeries = "series"

	AIDefaultLimit = 10
	AIMaxLimit     = 20

	AICandidatePoolSize = 30
	AIRequestTimeout    = 40 * time.Second

	aiProviderOpenRouter           = "openrouter"
	aiProviderOpenAICompatible     = "openai_compatible"
	aiOpenRouterURL                = "https://openrouter.ai/api/v1/chat/completions"
	aiDefaultOpenRouterModel       = "openai/gpt-4o-mini"
	aiDefaultOpenAICompatibleURL   = "https://api.openai.com/v1/chat/completions"
	aiDefaultOpenAICompatibleModel = "gpt-4o-mini"
)

var (
	errAIModelUnavailable = errors.New("ai service is unavailable")

	// ErrAIModelUnavailablePublic is returned when no database matches exist and
	// the AI model could not produce recommendations.
	ErrAIModelUnavailablePublic = errors.New("ai assistant is unavailable")

	englishStopwords = map[string]struct{}{
		"the": {}, "a": {}, "an": {}, "and": {}, "or": {}, "for": {}, "of": {}, "to": {},
		"in": {}, "on": {}, "with": {}, "i": {}, "want": {}, "me": {}, "please": {},
		"is": {}, "are": {}, "was": {}, "were": {}, "be": {}, "something": {}, "some": {},
		"any": {}, "good": {}, "movie": {}, "movies": {}, "series": {}, "show": {},
		"shows": {}, "tv": {}, "game": {}, "games": {}, "book": {}, "books": {},
		"recommend": {}, "suggest": {}, "recommendation": {}, "recommendations": {},
		"about": {}, "like": {}, "similar": {}, "that": {}, "this": {}, "it": {},
		"very": {}, "really": {}, "you": {}, "my": {}, "we": {}, "they": {},
	}

	aiKeywordSplitRegex = regexp.MustCompile(`[\s,;.!?()']+`)
)

type AIRecommendation struct {
	Title              string   `json:"title"`
	Type               string   `json:"type"`
	Description        string   `json:"description"`
	Reason             string   `json:"reason"`
	Source             string   `json:"source"`
	IsAvailableInApp   bool     `json:"is_available_in_app"`
	ItemID             *uint    `json:"item_id"`
	CanAddToFavorites  bool     `json:"can_add_to_favorites"`
	CanAddToWatchLater bool     `json:"can_add_to_watch_later"`
	Image              *string  `json:"image"`
	Categories         []string `json:"categories,omitempty"`
	Rating             float64  `json:"rating,omitempty"`
	ReleaseYear        *int     `json:"release_year,omitempty"`
}

type AIRecommendationResult struct {
	Reply           string             `json:"reply"`
	Recommendations []AIRecommendation `json:"recommendations"`
}

type AIRecommendationRequest struct {
	UserMessage string
	Type        string
	Categories  []string
	Mood        string
	Language    string
	Limit       int
}

type AIRepository interface {
	SearchItems(keywords []string, itemType string, categorySlugs []string, limit int) ([]models.Item, error)
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLMRecommender interface {
	Recommend(ctx context.Context, messages []ChatMessage) (string, error)
}

type AIRecommendationService struct {
	repo   AIRepository
	client LLMRecommender
}

func NewAIRecommendationService(repo AIRepository, client LLMRecommender) *AIRecommendationService {
	return &AIRecommendationService{repo: repo, client: client}
}

func (s *AIRecommendationService) Recommend(ctx context.Context, request AIRecommendationRequest) (*AIRecommendationResult, error) {
	limit := normalizeAILimit(request.Limit)
	itemType := normalizeAIType(request.Type)
	language := normalizeAILanguage(request.Language, request.UserMessage)

	keywords := extractKeywords(request.UserMessage)
	candidates, err := s.repo.SearchItems(keywords, itemType, request.Categories, AICandidatePoolSize)
	if err != nil {
		return nil, err
	}

	candidateByID := make(map[uint]models.Item, len(candidates))
	for _, item := range candidates {
		candidateByID[item.ID] = item
	}

	if s.client == nil {
		return s.buildFallbackResult(ctx, candidateByID, candidates, request, limit, language, errAIModelUnavailable)
	}

	messages := buildAIMessages(request, candidates, language, limit)

	rawReply, err := s.client.Recommend(ctx, messages)
	if err != nil {
		return s.buildFallbackResult(ctx, candidateByID, candidates, request, limit, language, err)
	}

	parsed, parseErr := parseAIResponse(rawReply)
	if parseErr != nil {
		return s.buildFallbackResult(ctx, candidateByID, candidates, request, limit, language, parseErr)
	}

	recommendations := s.mergeRecommendations(parsed.Recommendations, candidateByID, limit)

	reply := strings.TrimSpace(parsed.Reply)
	if reply == "" {
		reply = defaultReply(recommendations, language)
	}

	return &AIRecommendationResult{
		Reply:           reply,
		Recommendations: recommendations,
	}, nil
}

func (s *AIRecommendationService) buildFallbackResult(
	_ context.Context,
	_ map[uint]models.Item,
	candidates []models.Item,
	request AIRecommendationRequest,
	limit int,
	language string,
	cause error,
) (*AIRecommendationResult, error) {
	if len(candidates) == 0 {
		if cause != nil {
			return nil, fmt.Errorf("%w: %v", ErrAIModelUnavailablePublic, cause)
		}
		return nil, ErrAIModelUnavailablePublic
	}

	recommendations := make([]AIRecommendation, 0, limit)
	for i := 0; i < len(candidates) && len(recommendations) < limit; i++ {
		recommendations = append(recommendations, itemToDatabaseRecommendation(candidates[i], fallbackReason(candidates[i], language)))
	}

	return &AIRecommendationResult{
		Reply:           defaultReply(recommendations, language),
		Recommendations: recommendations,
	}, nil
}

func (s *AIRecommendationService) mergeRecommendations(
	raw []aiRecommendationDraft,
	candidateByID map[uint]models.Item,
	limit int,
) []AIRecommendation {
	seen := make(map[string]struct{}, len(raw))
	merged := make([]AIRecommendation, 0, len(raw))

	for _, draft := range raw {
		title := strings.TrimSpace(draft.Title)
		if title == "" {
			continue
		}
		lowerKey := strings.ToLower(title)
		if _, dup := seen[lowerKey]; dup {
			continue
		}

		if draft.RefID != nil {
			if item, ok := candidateByID[*draft.RefID]; ok {
				seen[lowerKey] = struct{}{}
				merged = append(merged, itemToDatabaseRecommendation(item, draft.Reason))
				continue
			}
		}

		seen[lowerKey] = struct{}{}
		merged = append(merged, draftToExternalRecommendation(draft, title))
	}

	if len(merged) > limit {
		merged = merged[:limit]
	}
	return merged
}

func itemToDatabaseRecommendation(item models.Item, reason string) AIRecommendation {
	recommendation := AIRecommendation{
		Title:              item.Title,
		Type:               item.Type,
		Description:        item.Description,
		Reason:             reason,
		Source:             AIRecommendationSourceDatabase,
		IsAvailableInApp:   true,
		ItemID:             uintPtr(item.ID),
		CanAddToFavorites:  true,
		CanAddToWatchLater: true,
		Rating:             item.Rating,
		Categories:         sortedCategoryNames(item.Categories),
		ReleaseYear:        releaseYear(item.ReleaseDate),
	}

	image := strings.TrimSpace(item.CoverImage)
	recommendation.Image = stringPtrOrNil(image)

	if strings.TrimSpace(reason) == "" {
		recommendation.Reason = "Recommended because it matches your request and is available in TrendFlix."
	}

	return recommendation
}

func draftToExternalRecommendation(draft aiRecommendationDraft, title string) AIRecommendation {
	description := strings.TrimSpace(draft.Description)
	if description == "" {
		description = "An external entertainment suggestion outside the TrendFlix catalog."
	}

	reason := strings.TrimSpace(draft.Reason)
	if reason == "" {
		reason = "An external recommendation related to your request."
	}

	itemType := normalizeAIType(draft.Type)

	var image *string
	if img := strings.TrimSpace(draft.Image); img != "" {
		image = stringPtrOrNil(img)
	}

	return AIRecommendation{
		Title:              title,
		Type:               itemType,
		Description:        description,
		Reason:             reason,
		Source:             AIRecommendationSourceExternal,
		IsAvailableInApp:   false,
		ItemID:             nil,
		CanAddToFavorites:  false,
		CanAddToWatchLater: false,
		Image:              image,
	}
}

func buildAIMessages(request AIRecommendationRequest, candidates []models.Item, language string, limit int) []ChatMessage {
	var builder strings.Builder
	builder.WriteString("You are TrendFlix, the AI Entertainment Assistant inside the TrendFlix app. ")
	builder.WriteString("You recommend movies, TV series, books, and games based on the user's request, taste, mood, and preferences.\n\n")
	builder.WriteString("Rules:\n")
	builder.WriteString("- First prefer titles that already exist inside the TrendFlix database (provided below).\n")
	builder.WriteString("- You may also suggest external titles that are not in the database.\n")
	builder.WriteString("- Every recommendation must include a short, specific reason explaining why it was recommended.\n")
	builder.WriteString("- Only use the field \"ref_id\" to reference a database item, and only with one of the IDs listed below. Never invent a ref_id.\n")
	builder.WriteString("- For external items, leave \"ref_id\" as null.\n")
	builder.WriteString("- Mix database and external items when helpful. Multiple content types are allowed in the same response.\n")
	builder.WriteString(fmt.Sprintf("- Return at most %d recommendations.\n", limit))
	builder.WriteString("- Respond ONLY with valid minified JSON in exactly this shape: ")
	builder.WriteString(`{"reply":"short friendly reply","recommendations":[{"title":"","type":"movie|tv_show|game|book","description":"","reason":"","ref_id":null,"image":null}]}`)
	builder.WriteString(".\n")
	builder.WriteString("- Do not include any text outside the JSON object.\n")
	builder.WriteString(fmt.Sprintf("- Reply in %s.\n", language))

	if len(candidates) == 0 {
		builder.WriteString("\nDatabase items: none.\n")
	} else {
		builder.WriteString("\nDatabase items available in the app:\n")
		for _, item := range candidates {
			builder.WriteString(fmt.Sprintf(
				"- ref_id:%d | title:%s | type:%s | rating:%.1f | year:%s | categories:%s | description:%s\n",
				item.ID,
				escapeCSV(item.Title),
				item.Type,
				item.Rating,
				formatReleaseYear(item.ReleaseDate),
				escapeCSV(strings.Join(sortedCategoryNames(item.Categories), ", ")),
				escapeCSV(truncate(item.Description, 160)),
			))
		}
	}

	messages := []ChatMessage{
		{Role: "system", Content: builder.String()},
	}

	userContent := strings.TrimSpace(request.UserMessage)
	{
		var extras []string
		if v := normalizeAIType(request.Type); v != "" {
			extras = append(extras, "requested_type="+v)
		}
		if len(request.Categories) > 0 {
			extras = append(extras, "categories="+strings.Join(request.Categories, ","))
		}
		if v := strings.TrimSpace(request.Mood); v != "" {
			extras = append(extras, "mood="+v)
		}
		if len(extras) > 0 {
			userContent = userContent + "\n(filters: " + strings.Join(extras, "; ") + ")"
		}
	}

	if userContent == "" {
		userContent = "Recommend something I might enjoy."
	}

	messages = append(messages, ChatMessage{Role: "user", Content: userContent})
	return messages
}

type aiRecommendationDraft struct {
	Title       string `json:"title"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Reason      string `json:"reason"`
	RefID       *uint  `json:"ref_id"`
	Image       string `json:"image"`
}

type aiParsedResponse struct {
	Reply           string                  `json:"reply"`
	Recommendations []aiRecommendationDraft `json:"recommendations"`
}

func parseAIResponse(raw string) (aiParsedResponse, error) {
	trimmed := stripCodeFences(strings.TrimSpace(raw))
	start := strings.IndexByte(trimmed, '{')
	end := strings.LastIndexByte(trimmed, '}')
	if start < 0 || end <= start {
		return aiParsedResponse{}, fmt.Errorf("ai response is not valid JSON")
	}

	var parsed aiParsedResponse
	if err := json.Unmarshal([]byte(trimmed[start:end+1]), &parsed); err != nil {
		return aiParsedResponse{}, fmt.Errorf("ai response could not be parsed: %w", err)
	}

	return parsed, nil
}

func stripCodeFences(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "```") {
		value = strings.TrimPrefix(value, "```")
		if idx := strings.IndexByte(value, '\n'); idx >= 0 {
			value = value[idx+1:]
		}
		value = strings.TrimSuffix(strings.TrimSpace(value), "```")
	}
	return strings.TrimSpace(value)
}

func normalizeAILimit(limit int) int {
	if limit <= 0 {
		return AIDefaultLimit
	}
	if limit > AIMaxLimit {
		return AIMaxLimit
	}
	return limit
}

func normalizeAIType(value string) string {
	t := strings.ToLower(strings.TrimSpace(value))
	switch t {
	case "":
		return ""
	case AITypeAll:
		return ""
	case AITypeSeries, "tv", "show":
		return models.ItemTypeTVShow
	case models.ItemTypeMovie, models.ItemTypeTVShow, models.ItemTypeGame, models.ItemTypeBook:
		return t
	default:
		return ""
	}
}

func normalizeAILanguage(language, message string) string {
	lang := strings.ToLower(strings.TrimSpace(language))
	if lang == "ar" || lang == "arabic" {
		return "Arabic"
	}
	if lang == "en" || lang == "english" {
		return "English"
	}
	if isArabicMessage(message) {
		return "Arabic"
	}
	return "English"
}

func isArabicMessage(value string) bool {
	for _, r := range value {
		if (r >= 0x0600 && r <= 0x06FF) || (r >= 0x0750 && r <= 0x077F) || (r >= 0x08A0 && r <= 0x08FF) {
			return true
		}
	}
	return false
}

func extractKeywords(message string) []string {
	words := aiKeywordSplitRegex.Split(message, -1)
	seen := make(map[string]struct{}, len(words))
	keywords := make([]string, 0, len(words))
	for _, word := range words {
		token := strings.ToLower(strings.Trim(word, "\"'.,()"))
		if len(token) < 3 {
			continue
		}
		if _, stop := englishStopwords[token]; stop {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		keywords = append(keywords, token)
	}
	return keywords
}

func sortedCategoryNames(categories []models.Category) []string {
	names := make([]string, 0, len(categories))
	for _, category := range categories {
		name := strings.TrimSpace(category.Name)
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func releaseYear(date time.Time) *int {
	if date.IsZero() {
		return nil
	}
	year := date.Year()
	if year <= 0 {
		return nil
	}
	return &year
}

func formatReleaseYear(date time.Time) string {
	if date.IsZero() {
		return "unknown"
	}
	year := date.Year()
	if year <= 0 {
		return "unknown"
	}
	return fmt.Sprintf("%d", year)
}

func fallbackReason(item models.Item, language string) string {
	cats := sortedCategoryNames(item.Categories)
	if isArabicLanguage(language) {
		if len(cats) > 0 {
			return fmt.Sprintf("موصى به لأنه متوفر في التطبيق ويندرج تحت %s.", strings.Join(cats, " و"))
		}
		return "موصى به لأنه متوفر في تطبيق TrendFlix ويتناسب مع طلبك."
	}
	if len(cats) > 0 {
		return fmt.Sprintf("Recommended because it is available in the app and matches %s.", strings.Join(cats, ", "))
	}
	return "Recommended because it is available in TrendFlix and matches your request."
}

func defaultReply(recommendations []AIRecommendation, language string) string {
	dbCount := 0
	externalCount := 0
	for _, recommendation := range recommendations {
		if recommendation.Source == AIRecommendationSourceDatabase {
			dbCount++
		} else {
			externalCount++
		}
	}

	if isArabicLanguage(language) {
		return fmt.Sprintf("إليك %d توصية. %d متوفرة داخل التطبيق و%d اقتراحات خارجية.", len(recommendations), dbCount, externalCount)
	}
	return fmt.Sprintf("Here are %d recommendations. %d are available inside the app and %d are external suggestions.", len(recommendations), dbCount, externalCount)
}

func isArabicLanguage(language string) bool {
	return strings.EqualFold(language, "Arabic")
}

func truncate(value string, max int) string {
	value = strings.TrimSpace(value)
	if len([]rune(value)) <= max {
		return value
	}
	runes := []rune(value)
	return string(runes[:max]) + "…"
}

func escapeCSV(value string) string {
	value = strings.TrimSpace(value)
	if strings.ContainsAny(value, ",\n") {
		return "\"" + strings.ReplaceAll(value, "\"", "'") + "\""
	}
	return value
}

func uintPtr(value uint) *uint {
	return &value
}

func stringPtrOrNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

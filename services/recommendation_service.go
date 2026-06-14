package services

import (
	"fmt"
	"math"
	"sort"

	"trendflix/models"
)

const (
	FavoriteWeight   = 3
	WatchLaterWeight = 2
	ListWeight       = 1

	DefaultLimit = 10
	MaxLimit     = 50

	CategoryScoreWeight   = 5
	PopularityScoreWeight = 2
	RatingScoreWeight     = 2
)

const (
	sourceFavorite   = "favorite"
	sourceWatchLater = "watch_later"
	sourceList       = "list"
)

type Recommendation struct {
	ID         uint     `json:"id"`
	Title      string   `json:"title"`
	Type       string   `json:"type"`
	Categories []string `json:"categories"`
	Score      int      `json:"score"`
	Reason     string   `json:"reason"`
	CoverImage string   `json:"cover_image,omitempty"`
	Rating     float64  `json:"rating,omitempty"`
}

type Repository interface {
	FetchFavorites(userID uint) ([]models.Item, error)
	FetchWatchLater(userID uint) ([]models.Item, error)
	FetchListItems(userID uint) ([]models.Item, error)
	FetchReviewedItemIDs(userID uint) ([]uint, error)
	FetchCandidates(categoryIDs, excludedIDs []uint) ([]models.Item, error)
	FetchPopularity(itemIDs []uint) (map[uint]int, error)
	FetchTopRated(excludedIDs []uint, limit int) ([]models.Item, error)
}

type RecommendationService struct {
	repo Repository
}

func NewRecommendationService(repo Repository) *RecommendationService {
	return &RecommendationService{repo: repo}
}

func (s *RecommendationService) ForUser(userID uint, limit int) ([]Recommendation, error) {
	limit = normalizeLimit(limit)

	favorites, err := s.repo.FetchFavorites(userID)
	if err != nil {
		return nil, err
	}
	watchLater, err := s.repo.FetchWatchLater(userID)
	if err != nil {
		return nil, err
	}
	listItems, err := s.repo.FetchListItems(userID)
	if err != nil {
		return nil, err
	}
	reviewedIDs, err := s.repo.FetchReviewedItemIDs(userID)
	if err != nil {
		return nil, err
	}

	excluded := buildExclusionSet(favorites, watchLater, listItems, reviewedIDs)
	profile := buildInterestProfile(favorites, watchLater, listItems)

	if profile.empty() {
		return s.fallbackRecommendations(excluded, limit)
	}

	candidates, err := s.repo.FetchCandidates(profile.categoryIDs(), excluded.slice())
	if err != nil {
		return nil, err
	}

	popularity, err := s.repo.FetchPopularity(collectItemIDs(candidates))
	if err != nil {
		return nil, err
	}

	scored := scoreCandidates(candidates, profile, popularity)
	selected := selectDiverse(scored, limit)

	recommendations := make([]Recommendation, 0, len(selected))
	for _, item := range selected {
		recommendations = append(recommendations, item.toRecommendation())
	}
	return recommendations, nil
}

func (s *RecommendationService) fallbackRecommendations(excluded exclusionSet, limit int) ([]Recommendation, error) {
	pool, err := s.repo.FetchTopRated(excluded.slice(), limit*3)
	if err != nil {
		return nil, err
	}

	popularity, err := s.repo.FetchPopularity(collectItemIDs(pool))
	if err != nil {
		return nil, err
	}

	scored := make([]scoredItem, 0, len(pool))
	for _, item := range pool {
		pop := popularity[item.ID]
		score := item.Rating*RatingScoreWeight + float64(pop)*PopularityScoreWeight
		scored = append(scored, scoredItem{item: item, score: score, popularity: pop})
	}

	selected := selectDiverse(scored, limit)

	recommendations := make([]Recommendation, 0, len(selected))
	for _, sc := range selected {
		recommendations = append(recommendations, sc.toFallbackRecommendation())
	}
	return recommendations, nil
}

type categoryStat struct {
	id              uint
	name            string
	score           int
	favoritesCount  int
	watchLaterCount int
	listCount       int
}

type interestProfile struct {
	stats map[uint]*categoryStat
}

func buildInterestProfile(favorites, watchLater, listItems []models.Item) *interestProfile {
	profile := &interestProfile{stats: map[uint]*categoryStat{}}
	profile.add(favorites, FavoriteWeight, sourceFavorite)
	profile.add(watchLater, WatchLaterWeight, sourceWatchLater)
	profile.add(listItems, ListWeight, sourceList)
	return profile
}

func (p *interestProfile) add(items []models.Item, weight int, source string) {
	for _, item := range items {
		for _, category := range item.Categories {
			stat := p.stats[category.ID]
			if stat == nil {
				stat = &categoryStat{id: category.ID, name: category.Name}
				p.stats[category.ID] = stat
			}
			stat.score += weight
			switch source {
			case sourceFavorite:
				stat.favoritesCount++
			case sourceWatchLater:
				stat.watchLaterCount++
			case sourceList:
				stat.listCount++
			}
		}
	}
}

func (p *interestProfile) empty() bool {
	return len(p.stats) == 0
}

func (p *interestProfile) categoryIDs() []uint {
	ids := make([]uint, 0, len(p.stats))
	for id := range p.stats {
		ids = append(ids, id)
	}
	return ids
}

type exclusionSet struct {
	ids map[uint]struct{}
}

func buildExclusionSet(favorites, watchLater, listItems []models.Item, reviewedIDs []uint) exclusionSet {
	excluded := exclusionSet{ids: map[uint]struct{}{}}
	for _, item := range favorites {
		excluded.ids[item.ID] = struct{}{}
	}
	for _, item := range watchLater {
		excluded.ids[item.ID] = struct{}{}
	}
	for _, item := range listItems {
		excluded.ids[item.ID] = struct{}{}
	}
	for _, id := range reviewedIDs {
		excluded.ids[id] = struct{}{}
	}
	return excluded
}

func (e exclusionSet) slice() []uint {
	if len(e.ids) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(e.ids))
	for id := range e.ids {
		ids = append(ids, id)
	}
	return ids
}

type scoredItem struct {
	item       models.Item
	score      float64
	matched    []categoryStat
	popularity int
}

func scoreCandidates(candidates []models.Item, profile *interestProfile, popularity map[uint]int) []scoredItem {
	scored := make([]scoredItem, 0, len(candidates))
	for _, candidate := range candidates {
		matched := make([]categoryStat, 0, len(candidate.Categories))
		matchScore := 0
		for _, category := range candidate.Categories {
			stat, ok := profile.stats[category.ID]
			if !ok {
				continue
			}
			matched = append(matched, *stat)
			matchScore += stat.score
		}
		if len(matched) == 0 {
			continue
		}

		sort.SliceStable(matched, func(i, j int) bool {
			if matched[i].score != matched[j].score {
				return matched[i].score > matched[j].score
			}
			return matched[i].id < matched[j].id
		})

		pop := popularity[candidate.ID]
		total := float64(matchScore)*CategoryScoreWeight +
			float64(pop)*PopularityScoreWeight +
			candidate.Rating*RatingScoreWeight

		scored = append(scored, scoredItem{
			item:       candidate,
			score:      total,
			matched:    matched,
			popularity: pop,
		})
	}
	return scored
}

func selectDiverse(items []scoredItem, limit int) []scoredItem {
	if len(items) <= limit {
		return sortByScore(items)
	}

	groups := map[string][]scoredItem{}
	types := make([]string, 0)
	for _, item := range items {
		itemType := item.item.Type
		if _, ok := groups[itemType]; !ok {
			types = append(types, itemType)
		}
		groups[itemType] = append(groups[itemType], item)
	}

	for _, itemType := range types {
		groups[itemType] = sortByScore(groups[itemType])
	}
	sort.Strings(types)

	selected := make([]scoredItem, 0, limit)
	cursor := map[string]int{}
	for len(selected) < limit {
		progressed := false
		for _, itemType := range types {
			if cursor[itemType] >= len(groups[itemType]) {
				continue
			}
			selected = append(selected, groups[itemType][cursor[itemType]])
			cursor[itemType]++
			progressed = true
			if len(selected) == limit {
				break
			}
		}
		if !progressed {
			break
		}
	}

	return sortByScore(selected)
}

func sortByScore(items []scoredItem) []scoredItem {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].score != items[j].score {
			return items[i].score > items[j].score
		}
		return items[i].item.ID < items[j].item.ID
	})
	return items
}

func (si scoredItem) toRecommendation() Recommendation {
	return Recommendation{
		ID:         si.item.ID,
		Title:      si.item.Title,
		Type:       si.item.Type,
		Categories: categoryNames(si.item.Categories),
		Score:      int(math.Round(si.score)),
		Reason:     buildReason(si.matched),
		CoverImage: si.item.CoverImage,
		Rating:     si.item.Rating,
	}
}

func (si scoredItem) toFallbackRecommendation() Recommendation {
	return Recommendation{
		ID:         si.item.ID,
		Title:      si.item.Title,
		Type:       si.item.Type,
		Categories: categoryNames(si.item.Categories),
		Score:      int(math.Round(si.score)),
		Reason:     buildFallbackReason(si.item, si.popularity),
		CoverImage: si.item.CoverImage,
		Rating:     si.item.Rating,
	}
}

func buildReason(matched []categoryStat) string {
	if len(matched) == 0 {
		return "Recommended for you"
	}

	top := matched[0]
	switch {
	case top.favoritesCount >= 2:
		return fmt.Sprintf("Because you added several %s items to your favorites", top.name)
	case top.watchLaterCount >= 2:
		return fmt.Sprintf("Because you added several %s items to your Watch Later list", top.name)
	case top.watchLaterCount >= 1:
		return "Similar to items in your Watch Later list"
	case top.favoritesCount >= 1:
		return fmt.Sprintf("Because you favorited a %s item", top.name)
	default:
		names := statNames(matched, 2)
		if len(names) >= 2 {
			return fmt.Sprintf("Because you are interested in %s and %s", names[0], names[1])
		}
		return fmt.Sprintf("Because you are interested in %s", names[0])
	}
}

func buildFallbackReason(item models.Item, popularity int) string {
	switch {
	case popularity > 0 && item.Rating >= 8:
		return "Trending and top rated on TrendFlix"
	case popularity > 0:
		return "Trending on TrendFlix"
	case item.Rating >= 8:
		return "Top rated on TrendFlix"
	default:
		return "Popular on TrendFlix"
	}
}

func categoryNames(categories []models.Category) []string {
	names := make([]string, 0, len(categories))
	for _, category := range categories {
		names = append(names, category.Name)
	}
	return names
}

func statNames(matched []categoryStat, limit int) []string {
	if limit > len(matched) {
		limit = len(matched)
	}
	names := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		names = append(names, matched[i].name)
	}
	return names
}

func collectItemIDs(items []models.Item) []uint {
	ids := make([]uint, 0, len(items))
	seen := map[uint]struct{}{}
	for _, item := range items {
		if _, ok := seen[item.ID]; ok {
			continue
		}
		seen[item.ID] = struct{}{}
		ids = append(ids, item.ID)
	}
	return ids
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return DefaultLimit
	}
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}

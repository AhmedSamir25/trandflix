package admincontroller

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"trendflix/database"
	"trendflix/models"
)

type typeCount struct {
	Type  string `json:"type" gorm:"column:type"`
	Count int64  `json:"count" gorm:"column:count"`
}

type categoryCount struct {
	ID        uint   `json:"id" gorm:"column:id"`
	Name      string `json:"name" gorm:"column:name"`
	NameAr    string `json:"name_ar" gorm:"column:name_ar"`
	Slug      string `json:"slug" gorm:"column:slug"`
	ItemCount int64  `json:"item_count" gorm:"column:item_count"`
}

type roleCount struct {
	Role  string `json:"role" gorm:"column:role"`
	Count int64  `json:"count" gorm:"column:count"`
}

type periodStats struct {
	Today int64 `json:"today"`
	Month int64 `json:"month"`
	Year  int64 `json:"year"`
	Total int64 `json:"total"`
}

type amountStats struct {
	Today    float64 `json:"today"`
	Month    float64 `json:"month"`
	Year     float64 `json:"year"`
	Total    float64 `json:"total"`
	Currency string  `json:"currency"`
}

func ensureDatabase(context fiber.Map) error {
	if database.DbConn == nil {
		log.Println("database connection is not initialized")
		context["statusText"] = "bad"
		context["msg"] = "Database error"
		return errors.New("database is not initialized")
	}

	return nil
}

func failWithDatabaseError(c *fiber.Ctx, context fiber.Map, logMessage string, err error) error {
	log.Println(logMessage, err)
	context["statusText"] = "bad"
	context["msg"] = "Database error"
	return c.Status(fiber.StatusInternalServerError).JSON(context)
}

func fetchOverviewStats() (fiber.Map, error) {
	var totalItems int64
	var totalUsers int64
	var totalCommunities int64
	var totalCategories int64

	if err := database.DbConn.Model(&models.Item{}).Count(&totalItems).Error; err != nil {
		return nil, err
	}

	if err := database.DbConn.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		return nil, err
	}

	if err := database.DbConn.Model(&models.Community{}).Count(&totalCommunities).Error; err != nil {
		return nil, err
	}

	if err := database.DbConn.Model(&models.Category{}).Count(&totalCategories).Error; err != nil {
		return nil, err
	}

	var averageRating float64
	if err := database.DbConn.
		Model(&models.Item{}).
		Select("COALESCE(AVG(NULLIF(rating, 0)), 0)").
		Scan(&averageRating).Error; err != nil {
		return nil, err
	}

	var latestItem models.Item
	latestResult := database.DbConn.Order("id DESC").First(&latestItem)
	if latestResult.Error != nil && !errors.Is(latestResult.Error, gorm.ErrRecordNotFound) {
		return nil, latestResult.Error
	}

	overview := fiber.Map{
		"total_items":       totalItems,
		"total_users":       totalUsers,
		"total_communities": totalCommunities,
		"total_categories":  totalCategories,
		"average_rating":    averageRating,
		"latest_item":       latestItem,
	}

	return overview, nil
}

func fetchTypeCounts() ([]typeCount, error) {
	var typeCounts []typeCount
	if err := database.DbConn.
		Model(&models.Item{}).
		Select("type, COUNT(*) AS count").
		Group("type").
		Scan(&typeCounts).Error; err != nil {
		return nil, err
	}

	return typeCounts, nil
}

func fetchCategoryCounts() ([]categoryCount, error) {
	var categoryCounts []categoryCount
	if err := database.DbConn.
		Table("categories").
		Select("categories.id, categories.name, categories.name_ar, categories.slug, COUNT(category_item.item_id) AS item_count").
		Joins("LEFT JOIN category_item ON category_item.category_id = categories.id").
		Group("categories.id, categories.name, categories.name_ar, categories.slug").
		Order("categories.name ASC").
		Scan(&categoryCounts).Error; err != nil {
		return nil, err
	}

	return categoryCounts, nil
}

func countByCreatedAt(table string, since *time.Time) (int64, error) {
	query := database.DbConn.Table(table)
	if since != nil {
		query = query.Where("created_at >= ?", *since)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func countMetricByCreatedAt(spec metricSpec, since *time.Time) (int64, error) {
	query := database.DbConn.Table(spec.table)
	if since != nil {
		query = query.Where("created_at >= ?", *since)
	}
	if spec.typeFilter != "" {
		query = query.Where("type = ?", spec.typeFilter)
	}
	if spec.statusFilter != "" {
		query = query.Where("status = ?", spec.statusFilter)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func sumMetricByCreatedAt(spec metricSpec, since *time.Time) (float64, error) {
	query := database.DbConn.Table(spec.table)
	if since != nil {
		query = query.Where("created_at >= ?", *since)
	}
	if spec.typeFilter != "" {
		query = query.Where("type = ?", spec.typeFilter)
	}
	if spec.statusFilter != "" {
		query = query.Where("status = ?", spec.statusFilter)
	}

	var amount float64
	if err := query.Select("COALESCE(SUM(" + spec.amountColumn + "), 0)").Scan(&amount).Error; err != nil {
		return 0, err
	}

	return amount, nil
}

func fetchPeriodStats(table string, now time.Time) (periodStats, error) {
	todayStart := startOfDay(now)
	monthStart := firstDayOfMonth(now)
	yearStart := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())

	today, err := countByCreatedAt(table, &todayStart)
	if err != nil {
		return periodStats{}, err
	}

	month, err := countByCreatedAt(table, &monthStart)
	if err != nil {
		return periodStats{}, err
	}

	year, err := countByCreatedAt(table, &yearStart)
	if err != nil {
		return periodStats{}, err
	}

	total, err := countByCreatedAt(table, nil)
	if err != nil {
		return periodStats{}, err
	}

	return periodStats{Today: today, Month: month, Year: year, Total: total}, nil
}

func fetchMetricPeriodStats(spec metricSpec, now time.Time) (periodStats, error) {
	todayStart := startOfDay(now)
	monthStart := firstDayOfMonth(now)
	yearStart := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())

	today, err := countMetricByCreatedAt(spec, &todayStart)
	if err != nil {
		return periodStats{}, err
	}

	month, err := countMetricByCreatedAt(spec, &monthStart)
	if err != nil {
		return periodStats{}, err
	}

	year, err := countMetricByCreatedAt(spec, &yearStart)
	if err != nil {
		return periodStats{}, err
	}

	total, err := countMetricByCreatedAt(spec, nil)
	if err != nil {
		return periodStats{}, err
	}

	return periodStats{Today: today, Month: month, Year: year, Total: total}, nil
}

func fetchMetricAmountStats(spec metricSpec, now time.Time) (amountStats, error) {
	todayStart := startOfDay(now)
	monthStart := firstDayOfMonth(now)
	yearStart := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())

	today, err := sumMetricByCreatedAt(spec, &todayStart)
	if err != nil {
		return amountStats{}, err
	}

	month, err := sumMetricByCreatedAt(spec, &monthStart)
	if err != nil {
		return amountStats{}, err
	}

	year, err := sumMetricByCreatedAt(spec, &yearStart)
	if err != nil {
		return amountStats{}, err
	}

	total, err := sumMetricByCreatedAt(spec, nil)
	if err != nil {
		return amountStats{}, err
	}

	currency, err := activePlanCurrency()
	if err != nil {
		return amountStats{}, err
	}

	return amountStats{Today: today, Month: month, Year: year, Total: total, Currency: currency}, nil
}

func sumPaidPaymentsSince(since *time.Time) (float64, error) {
	query := database.DbConn.Model(&models.Payment{}).Where("status = ?", models.PaymentStatusPaid)
	if since != nil {
		query = query.Where("created_at >= ?", *since)
	}

	var amount float64
	if err := query.Select("COALESCE(SUM(amount), 0)").Scan(&amount).Error; err != nil {
		return 0, err
	}

	return amount, nil
}

func activePlanCurrency() (string, error) {
	var activePlan models.SubscriptionPlan
	result := database.DbConn.Where("is_active = ?", true).Order("id DESC").First(&activePlan)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "USD", nil
		}
		return "", result.Error
	}

	if strings.TrimSpace(activePlan.Currency) == "" {
		return "USD", nil
	}
	return activePlan.Currency, nil
}

func fetchPaymentAmountStats(now time.Time) (amountStats, error) {
	todayStart := startOfDay(now)
	monthStart := firstDayOfMonth(now)
	yearStart := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())

	today, err := sumPaidPaymentsSince(&todayStart)
	if err != nil {
		return amountStats{}, err
	}

	month, err := sumPaidPaymentsSince(&monthStart)
	if err != nil {
		return amountStats{}, err
	}

	year, err := sumPaidPaymentsSince(&yearStart)
	if err != nil {
		return amountStats{}, err
	}

	total, err := sumPaidPaymentsSince(nil)
	if err != nil {
		return amountStats{}, err
	}

	currency, err := activePlanCurrency()
	if err != nil {
		return amountStats{}, err
	}

	return amountStats{Today: today, Month: month, Year: year, Total: total, Currency: currency}, nil
}

func fetchUserStats() (fiber.Map, error) {
	var totalUsers int64
	if err := database.DbConn.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	registrationStats, err := fetchPeriodStats("users", now)
	if err != nil {
		return nil, err
	}

	subscriptionStats, err := fetchPeriodStats("subscriptions", now)
	if err != nil {
		return nil, err
	}

	subscriptionRevenue, err := fetchPaymentAmountStats(now)
	if err != nil {
		return nil, err
	}

	var activePlan models.SubscriptionPlan
	planPrice := 0.0
	planCurrency := "USD"
	planResult := database.DbConn.Where("is_active = ?", true).Order("id DESC").First(&activePlan)
	if planResult.Error != nil && !errors.Is(planResult.Error, gorm.ErrRecordNotFound) {
		return nil, planResult.Error
	}
	if planResult.Error == nil {
		planPrice = activePlan.Price
		if strings.TrimSpace(activePlan.Currency) != "" {
			planCurrency = activePlan.Currency
		}
	}

	var roleCounts []roleCount
	if err := database.DbConn.
		Model(&models.User{}).
		Select("role, COUNT(*) AS count").
		Group("role").
		Scan(&roleCounts).Error; err != nil {
		return nil, err
	}

	var recentUsers []models.User
	if err := database.DbConn.Order("id DESC").Limit(5).Find(&recentUsers).Error; err != nil {
		return nil, err
	}

	return fiber.Map{
		"total_users":        totalUsers,
		"registration_stats": registrationStats,
		"subscription_stats": fiber.Map{
			"today":         subscriptionStats.Today,
			"month":         subscriptionStats.Month,
			"year":          subscriptionStats.Year,
			"total":         subscriptionStats.Total,
			"revenue_today": subscriptionRevenue.Today,
			"revenue_month": subscriptionRevenue.Month,
			"revenue_year":  subscriptionRevenue.Year,
			"revenue_total": subscriptionRevenue.Total,
			"plan_price":    planPrice,
			"currency":      planCurrency,
		},
		"role_counts":  roleCounts,
		"recent_users": recentUsers,
	}, nil
}

func fetchRecentItems() ([]models.Item, error) {
	var recentItems []models.Item
	if err := database.DbConn.Preload("Categories").Order("id DESC").Limit(5).Find(&recentItems).Error; err != nil {
		return nil, err
	}

	return recentItems, nil
}

func GetOverviewStats(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Overview stats fetched successfully",
	}

	if err := ensureDatabase(context); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	overview, err := fetchOverviewStats()
	if err != nil {
		return failWithDatabaseError(c, context, "Error fetching overview stats:", err)
	}

	context["overview"] = overview
	return c.Status(fiber.StatusOK).JSON(context)
}

func GetTypeStats(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Type stats fetched successfully",
	}

	if err := ensureDatabase(context); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	typeCounts, err := fetchTypeCounts()
	if err != nil {
		return failWithDatabaseError(c, context, "Error counting items by type:", err)
	}

	context["type_counts"] = typeCounts
	return c.Status(fiber.StatusOK).JSON(context)
}

func GetCategoryStats(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Category stats fetched successfully",
	}

	if err := ensureDatabase(context); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	categoryCounts, err := fetchCategoryCounts()
	if err != nil {
		return failWithDatabaseError(c, context, "Error counting items by category:", err)
	}

	context["category_counts"] = categoryCounts
	return c.Status(fiber.StatusOK).JSON(context)
}

func GetUserStats(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "User stats fetched successfully",
	}

	if err := ensureDatabase(context); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	userStats, err := fetchUserStats()
	if err != nil {
		return failWithDatabaseError(c, context, "Error fetching user stats:", err)
	}

	context["user_stats"] = userStats
	return c.Status(fiber.StatusOK).JSON(context)
}

func GetStats(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Admin stats fetched successfully",
	}

	if err := ensureDatabase(context); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	overview, err := fetchOverviewStats()
	if err != nil {
		return failWithDatabaseError(c, context, "Error fetching overview stats:", err)
	}

	typeCounts, err := fetchTypeCounts()
	if err != nil {
		return failWithDatabaseError(c, context, "Error counting items by type:", err)
	}

	categoryCounts, err := fetchCategoryCounts()
	if err != nil {
		return failWithDatabaseError(c, context, "Error counting items by category:", err)
	}

	userStats, err := fetchUserStats()
	if err != nil {
		return failWithDatabaseError(c, context, "Error fetching user stats:", err)
	}

	recentItems, err := fetchRecentItems()
	if err != nil {
		return failWithDatabaseError(c, context, "Error fetching recent items:", err)
	}

	context["stats"] = fiber.Map{
		"total_items":       overview["total_items"],
		"total_users":       overview["total_users"],
		"total_communities": overview["total_communities"],
		"total_categories":  overview["total_categories"],
		"average_rating":    overview["average_rating"],
		"latest_item":       overview["latest_item"],
		"type_counts":       typeCounts,
		"category_counts":   categoryCounts,
		"user_stats":        userStats,
		"recent_items":      recentItems,
	}

	return c.Status(fiber.StatusOK).JSON(context)
}

// ─── Timeseries (line chart data) ────────────────────────────────────────

type timeseriesPoint struct {
	Period string  `json:"period"`
	Count  float64 `json:"count"`
	Total  float64 `json:"total"`
}

type metricSpec struct {
	table        string
	typeFilter   string
	amountColumn string
	statusFilter string
}

var metricSpecs = map[string]metricSpec{
	"items":                {table: "items"},
	"users":                {table: "users"},
	"communities":          {table: "communities"},
	"categories":           {table: "categories"},
	"movies":               {table: "items", typeFilter: "movie"},
	"tv_shows":             {table: "items", typeFilter: "tv_show"},
	"games":                {table: "items", typeFilter: "game"},
	"books":                {table: "items", typeFilter: "book"},
	"subscriptions":        {table: "subscriptions"},
	"subscription_revenue": {table: "payments", amountColumn: "amount", statusFilter: models.PaymentStatusPaid},
}

func firstDayOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// GetTimeseriesStats returns daily or monthly counts (and cumulative totals)
// for a given metric, used to render admin line charts.
//
// GET /admin/stats/timeseries?metric=items&days=30
func GetTimeseriesStats(c *fiber.Ctx) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        "Timeseries stats fetched successfully",
	}

	if err := ensureDatabase(context); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	metric := strings.TrimSpace(c.Query("metric", "items"))
	spec, ok := metricSpecs[metric]
	if !ok {
		context["statusText"] = "bad"
		context["msg"] = "Invalid metric. Allowed: items, users, communities, categories, movies, tv_shows, games, books, subscriptions, subscription_revenue"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	days, _ := strconv.Atoi(strings.TrimSpace(c.Query("days", "30")))
	if days <= 0 {
		days = 30
	}
	if days > 365 {
		days = 365
	}

	now := time.Now()
	monthly := days > 90

	periodExpr := "DATE_FORMAT(created_at, '%Y-%m-%d')"
	periodLayout := "2006-01-02"
	var start time.Time
	if monthly {
		months := days / 30
		if months < 1 {
			months = 1
		}
		start = firstDayOfMonth(now).AddDate(0, -months, 0)
		periodExpr = "DATE_FORMAT(created_at, '%Y-%m')"
		periodLayout = "2006-01"
	} else {
		start = startOfDay(now.AddDate(0, 0, -(days - 1)))
	}

	type tsRow struct {
		Period string
		Cnt    float64
	}

	valueExpr := "COUNT(*)"
	if spec.amountColumn != "" {
		valueExpr = "COALESCE(SUM(" + spec.amountColumn + "), 0)"
	}
	query := database.DbConn.Table(spec.table).
		Select(periodExpr+" AS period, "+valueExpr+" AS cnt").
		Where("created_at >= ?", start)
	if spec.typeFilter != "" {
		query = query.Where("type = ?", spec.typeFilter)
	}
	if spec.statusFilter != "" {
		query = query.Where("status = ?", spec.statusFilter)
	}

	var rows []tsRow
	if err := query.Group("period").Order("period ASC").Scan(&rows).Error; err != nil {
		return failWithDatabaseError(c, context, "Error fetching timeseries:", err)
	}

	countByPeriod := make(map[string]float64, len(rows))
	for _, row := range rows {
		countByPeriod[row.Period] = row.Cnt
	}

	// Cumulative base: everything created before the window start.
	baseQuery := database.DbConn.Table(spec.table).Where("created_at < ?", start)
	if spec.typeFilter != "" {
		baseQuery = baseQuery.Where("type = ?", spec.typeFilter)
	}
	if spec.statusFilter != "" {
		baseQuery = baseQuery.Where("status = ?", spec.statusFilter)
	}
	var cumulative float64
	if spec.amountColumn != "" {
		if err := baseQuery.Select("COALESCE(SUM(" + spec.amountColumn + "), 0)").Scan(&cumulative).Error; err != nil {
			return failWithDatabaseError(c, context, "Error counting timeseries base:", err)
		}
	} else {
		var cumulativeCount int64
		if err := baseQuery.Count(&cumulativeCount).Error; err != nil {
			return failWithDatabaseError(c, context, "Error counting timeseries base:", err)
		}
		cumulative = float64(cumulativeCount)
	}

	points := make([]timeseriesPoint, 0, days+2)
	if monthly {
		cursor := firstDayOfMonth(start)
		end := firstDayOfMonth(now)
		for !cursor.After(end) {
			key := cursor.Format(periodLayout)
			count := countByPeriod[key]
			cumulative += count
			points = append(points, timeseriesPoint{Period: key, Count: count, Total: cumulative})
			cursor = cursor.AddDate(0, 1, 0)
		}
	} else {
		cursor := start
		for !cursor.After(now) {
			key := cursor.Format(periodLayout)
			count := countByPeriod[key]
			cumulative += count
			points = append(points, timeseriesPoint{Period: key, Count: count, Total: cumulative})
			cursor = cursor.AddDate(0, 0, 1)
		}
	}

	granularity := "daily"
	if monthly {
		granularity = "monthly"
	}

	context["metric"] = metric
	context["granularity"] = granularity
	context["points"] = points
	if spec.amountColumn != "" {
		metricStats, err := fetchMetricAmountStats(spec, now)
		if err != nil {
			return failWithDatabaseError(c, context, "Error fetching metric amount stats:", err)
		}
		context["metric_stats"] = metricStats
	} else {
		metricStats, err := fetchMetricPeriodStats(spec, now)
		if err != nil {
			return failWithDatabaseError(c, context, "Error fetching metric period stats:", err)
		}
		context["metric_stats"] = metricStats
	}
	if metric == "users" {
		userStats, err := fetchUserStats()
		if err != nil {
			return failWithDatabaseError(c, context, "Error fetching user chart stats:", err)
		}
		context["user_stats"] = userStats
	}
	return c.Status(fiber.StatusOK).JSON(context)
}

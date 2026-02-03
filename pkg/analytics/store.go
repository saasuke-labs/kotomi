package analytics

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

// Store provides database operations for analytics
type Store struct {
	db *sql.DB
}

// NewStore creates a new analytics store
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// GetCommentMetrics retrieves comment statistics for a site
func (s *Store) GetCommentMetrics(siteID string, dateRange DateRange) (CommentMetrics, error) {
	var metrics CommentMetrics
	
	// Get total counts by status
	query := `
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
			SUM(CASE WHEN status = 'approved' THEN 1 ELSE 0 END) as approved,
			SUM(CASE WHEN status = 'rejected' THEN 1 ELSE 0 END) as rejected
		FROM comments
		WHERE site_id = ? AND created_at BETWEEN ? AND ?
	`
	
	err := s.db.QueryRow(query, siteID, dateRange.From, dateRange.To).Scan(
		&metrics.Total,
		&metrics.Pending,
		&metrics.Approved,
		&metrics.Rejected,
	)
	if err != nil {
		return metrics, fmt.Errorf("failed to get comment counts: %w", err)
	}
	
	// Calculate rates
	if metrics.Total > 0 {
		metrics.ApprovalRate = float64(metrics.Approved) / float64(metrics.Total) * 100
		metrics.RejectionRate = float64(metrics.Rejected) / float64(metrics.Total) * 100
	}
	
	// Get today's count
	today := time.Now().Truncate(24 * time.Hour)
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM comments 
		WHERE site_id = ? AND created_at >= ?
	`, siteID, today).Scan(&metrics.TotalToday)
	if err != nil {
		log.Printf("Failed to get today's comment count: %v", err)
	}
	
	// Get this week's count
	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
	weekStart = weekStart.Truncate(24 * time.Hour)
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM comments 
		WHERE site_id = ? AND created_at >= ?
	`, siteID, weekStart).Scan(&metrics.TotalThisWeek)
	if err != nil {
		log.Printf("Failed to get this week's comment count: %v", err)
	}
	
	// Get this month's count
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location())
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM comments 
		WHERE site_id = ? AND created_at >= ?
	`, siteID, monthStart).Scan(&metrics.TotalThisMonth)
	if err != nil {
		log.Printf("Failed to get this month's comment count: %v", err)
	}
	
	return metrics, nil
}

// GetUserMetrics retrieves user statistics for a site
func (s *Store) GetUserMetrics(siteID string, dateRange DateRange) (UserMetrics, error) {
	var metrics UserMetrics
	
	// Get total unique users
	err := s.db.QueryRow(`
		SELECT COUNT(DISTINCT id) FROM users WHERE site_id = ?
	`, siteID).Scan(&metrics.TotalUsers)
	if err != nil {
		return metrics, fmt.Errorf("failed to get total users: %w", err)
	}
	
	// Get active users today
	today := time.Now().Truncate(24 * time.Hour)
	err = s.db.QueryRow(`
		SELECT COUNT(DISTINCT author_id) FROM comments 
		WHERE site_id = ? AND created_at >= ?
	`, siteID, today).Scan(&metrics.ActiveUsersToday)
	if err != nil {
		log.Printf("Failed to get active users today: %v", err)
	}
	
	// Get active users this week
	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
	weekStart = weekStart.Truncate(24 * time.Hour)
	err = s.db.QueryRow(`
		SELECT COUNT(DISTINCT author_id) FROM comments 
		WHERE site_id = ? AND created_at >= ?
	`, siteID, weekStart).Scan(&metrics.ActiveUsersWeek)
	if err != nil {
		log.Printf("Failed to get active users this week: %v", err)
	}
	
	// Get active users this month
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location())
	err = s.db.QueryRow(`
		SELECT COUNT(DISTINCT author_id) FROM comments 
		WHERE site_id = ? AND created_at >= ?
	`, siteID, monthStart).Scan(&metrics.ActiveUsersMonth)
	if err != nil {
		log.Printf("Failed to get active users this month: %v", err)
	}
	
	// Get top contributors
	query := `
		SELECT c.author, c.author_email, COUNT(*) as comment_count
		FROM comments c
		WHERE c.site_id = ? AND c.created_at BETWEEN ? AND ?
		GROUP BY c.author_id, c.author, c.author_email
		ORDER BY comment_count DESC
		LIMIT 10
	`
	
	rows, err := s.db.Query(query, siteID, dateRange.From, dateRange.To)
	if err != nil {
		return metrics, fmt.Errorf("failed to get top contributors: %w", err)
	}
	defer rows.Close()
	
	metrics.TopContributors = []TopContributor{}
	for rows.Next() {
		var contributor TopContributor
		var email sql.NullString
		if err := rows.Scan(&contributor.Name, &email, &contributor.CommentCount); err != nil {
			log.Printf("Failed to scan top contributor: %v", err)
			continue
		}
		if email.Valid {
			contributor.Email = email.String
		}
		metrics.TopContributors = append(metrics.TopContributors, contributor)
	}
	
	return metrics, nil
}

// GetReactionMetrics retrieves reaction statistics for a site
func (s *Store) GetReactionMetrics(siteID string, dateRange DateRange) (ReactionMetrics, error) {
	var metrics ReactionMetrics
	
	// Get total reactions
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM reactions r
		INNER JOIN allowed_reactions ar ON r.allowed_reaction_id = ar.id
		WHERE ar.site_id = ? AND r.created_at BETWEEN ? AND ?
	`, siteID, dateRange.From, dateRange.To).Scan(&metrics.Total)
	if err != nil {
		return metrics, fmt.Errorf("failed to get total reactions: %w", err)
	}
	
	// Get today's count
	today := time.Now().Truncate(24 * time.Hour)
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM reactions r
		INNER JOIN allowed_reactions ar ON r.allowed_reaction_id = ar.id
		WHERE ar.site_id = ? AND r.created_at >= ?
	`, siteID, today).Scan(&metrics.TotalToday)
	if err != nil {
		log.Printf("Failed to get today's reaction count: %v", err)
	}
	
	// Get this week's count
	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
	weekStart = weekStart.Truncate(24 * time.Hour)
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM reactions r
		INNER JOIN allowed_reactions ar ON r.allowed_reaction_id = ar.id
		WHERE ar.site_id = ? AND r.created_at >= ?
	`, siteID, weekStart).Scan(&metrics.TotalThisWeek)
	if err != nil {
		log.Printf("Failed to get this week's reaction count: %v", err)
	}
	
	// Get this month's count
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location())
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM reactions r
		INNER JOIN allowed_reactions ar ON r.allowed_reaction_id = ar.id
		WHERE ar.site_id = ? AND r.created_at >= ?
	`, siteID, monthStart).Scan(&metrics.TotalThisMonth)
	if err != nil {
		log.Printf("Failed to get this month's reaction count: %v", err)
	}
	
	// Get reactions by type
	query := `
		SELECT ar.name, ar.emoji, COUNT(*) as count
		FROM reactions r
		INNER JOIN allowed_reactions ar ON r.allowed_reaction_id = ar.id
		WHERE ar.site_id = ? AND r.created_at BETWEEN ? AND ?
		GROUP BY ar.id, ar.name, ar.emoji
		ORDER BY count DESC
	`
	
	rows, err := s.db.Query(query, siteID, dateRange.From, dateRange.To)
	if err != nil {
		return metrics, fmt.Errorf("failed to get reaction breakdown: %w", err)
	}
	defer rows.Close()
	
	metrics.ByType = []ReactionBreakdown{}
	for rows.Next() {
		var breakdown ReactionBreakdown
		if err := rows.Scan(&breakdown.Name, &breakdown.Emoji, &breakdown.Count); err != nil {
			log.Printf("Failed to scan reaction breakdown: %v", err)
			continue
		}
		metrics.ByType = append(metrics.ByType, breakdown)
	}
	
	// Get most reacted items (pages and comments combined)
	metrics.MostReacted = []MostReactedItem{}
	
	// Most reacted pages
	pageQuery := `
		SELECT 'page' as type, p.path, COUNT(*) as reaction_count
		FROM reactions r
		INNER JOIN pages p ON r.page_id = p.id
		WHERE p.site_id = ? AND r.created_at BETWEEN ? AND ?
		GROUP BY r.page_id, p.path
		ORDER BY reaction_count DESC
		LIMIT 5
	`
	
	rows, err = s.db.Query(pageQuery, siteID, dateRange.From, dateRange.To)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item MostReactedItem
			if err := rows.Scan(&item.Type, &item.PagePath, &item.ReactionCount); err != nil {
				log.Printf("Failed to scan most reacted page: %v", err)
				continue
			}
			metrics.MostReacted = append(metrics.MostReacted, item)
		}
	}
	
	// Most reacted comments
	commentQuery := `
		SELECT 'comment' as type, SUBSTR(c.text, 1, 50) as comment_text, COUNT(*) as reaction_count
		FROM reactions r
		INNER JOIN comments c ON r.comment_id = c.id
		WHERE c.site_id = ? AND r.created_at BETWEEN ? AND ?
		GROUP BY r.comment_id, c.text
		ORDER BY reaction_count DESC
		LIMIT 5
	`
	
	rows, err = s.db.Query(commentQuery, siteID, dateRange.From, dateRange.To)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item MostReactedItem
			if err := rows.Scan(&item.Type, &item.CommentText, &item.ReactionCount); err != nil {
				log.Printf("Failed to scan most reacted comment: %v", err)
				continue
			}
			// Add ellipsis if text was truncated
			if len(item.CommentText) == 50 {
				item.CommentText += "..."
			}
			metrics.MostReacted = append(metrics.MostReacted, item)
		}
	}
	
	return metrics, nil
}

// GetModerationMetrics retrieves moderation statistics for a site
func (s *Store) GetModerationMetrics(siteID string, dateRange DateRange) (ModerationMetrics, error) {
	var metrics ModerationMetrics
	
	// Get total moderated comments
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM comments
		WHERE site_id = ? AND moderated_at IS NOT NULL AND created_at BETWEEN ? AND ?
	`, siteID, dateRange.From, dateRange.To).Scan(&metrics.TotalModerated)
	if err != nil {
		return metrics, fmt.Errorf("failed to get total moderated: %w", err)
	}
	
	// Get auto-rejected and auto-approved counts (moderated within 1 second indicates automated)
	query := `
		SELECT 
			SUM(CASE WHEN status = 'rejected' AND 
				(julianday(moderated_at) - julianday(created_at)) * 86400 < 1 THEN 1 ELSE 0 END) as auto_rejected,
			SUM(CASE WHEN status = 'approved' AND 
				(julianday(moderated_at) - julianday(created_at)) * 86400 < 1 THEN 1 ELSE 0 END) as auto_approved,
			SUM(CASE WHEN moderated_at IS NOT NULL AND 
				(julianday(moderated_at) - julianday(created_at)) * 86400 >= 1 THEN 1 ELSE 0 END) as manual_reviews
		FROM comments
		WHERE site_id = ? AND created_at BETWEEN ? AND ?
	`
	
	var autoRejected, autoApproved, manualReviews sql.NullInt64
	err = s.db.QueryRow(query, siteID, dateRange.From, dateRange.To).Scan(&autoRejected, &autoApproved, &manualReviews)
	if err != nil {
		return metrics, fmt.Errorf("failed to get moderation breakdown: %w", err)
	}
	
	if autoRejected.Valid {
		metrics.AutoRejected = int(autoRejected.Int64)
	}
	if autoApproved.Valid {
		metrics.AutoApproved = int(autoApproved.Int64)
	}
	if manualReviews.Valid {
		metrics.ManualReviews = int(manualReviews.Int64)
	}
	
	// Calculate average moderation time
	var avgSeconds sql.NullFloat64
	err = s.db.QueryRow(`
		SELECT AVG((julianday(moderated_at) - julianday(created_at)) * 86400)
		FROM comments
		WHERE site_id = ? AND moderated_at IS NOT NULL AND created_at BETWEEN ? AND ?
	`, siteID, dateRange.From, dateRange.To).Scan(&avgSeconds)
	if err == nil && avgSeconds.Valid {
		metrics.AverageModerationSec = avgSeconds.Float64
	}
	
	// Calculate spam detection rate (rejected / total moderated)
	if metrics.TotalModerated > 0 {
		totalRejected := 0
		s.db.QueryRow(`
			SELECT COUNT(*) FROM comments
			WHERE site_id = ? AND status = 'rejected' AND created_at BETWEEN ? AND ?
		`, siteID, dateRange.From, dateRange.To).Scan(&totalRejected)
		metrics.SpamDetectionRate = float64(totalRejected) / float64(metrics.TotalModerated) * 100
	}
	
	return metrics, nil
}

// GetCommentsTrend retrieves time series data for comments
func (s *Store) GetCommentsTrend(siteID string, dateRange DateRange) (TimeSeriesData, error) {
	var trend TimeSeriesData
	
	// Generate daily buckets
	daysDiff := int(dateRange.To.Sub(dateRange.From).Hours() / 24)
	if daysDiff > 90 {
		// For more than 90 days, group by week
		return s.getWeeklyTrend(siteID, dateRange, "comments")
	}
	
	// Daily trend
	query := `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM comments
		WHERE site_id = ? AND created_at BETWEEN ? AND ?
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`
	
	rows, err := s.db.Query(query, siteID, dateRange.From, dateRange.To)
	if err != nil {
		return trend, fmt.Errorf("failed to get comments trend: %w", err)
	}
	defer rows.Close()
	
	dateMap := make(map[string]int)
	for rows.Next() {
		var date string
		var count int
		if err := rows.Scan(&date, &count); err != nil {
			continue
		}
		dateMap[date] = count
	}
	
	// Fill in all dates in range
	trend.Labels = []string{}
	trend.Values = []int{}
	for d := dateRange.From; !d.After(dateRange.To); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		trend.Labels = append(trend.Labels, dateStr)
		trend.Values = append(trend.Values, dateMap[dateStr])
	}
	
	return trend, nil
}

// GetReactionsTrend retrieves time series data for reactions
func (s *Store) GetReactionsTrend(siteID string, dateRange DateRange) (TimeSeriesData, error) {
	var trend TimeSeriesData
	
	// Generate daily buckets
	daysDiff := int(dateRange.To.Sub(dateRange.From).Hours() / 24)
	if daysDiff > 90 {
		// For more than 90 days, group by week
		return s.getWeeklyTrend(siteID, dateRange, "reactions")
	}
	
	// Daily trend
	query := `
		SELECT DATE(r.created_at) as date, COUNT(*) as count
		FROM reactions r
		INNER JOIN allowed_reactions ar ON r.allowed_reaction_id = ar.id
		WHERE ar.site_id = ? AND r.created_at BETWEEN ? AND ?
		GROUP BY DATE(r.created_at)
		ORDER BY date ASC
	`
	
	rows, err := s.db.Query(query, siteID, dateRange.From, dateRange.To)
	if err != nil {
		return trend, fmt.Errorf("failed to get reactions trend: %w", err)
	}
	defer rows.Close()
	
	dateMap := make(map[string]int)
	for rows.Next() {
		var date string
		var count int
		if err := rows.Scan(&date, &count); err != nil {
			continue
		}
		dateMap[date] = count
	}
	
	// Fill in all dates in range
	trend.Labels = []string{}
	trend.Values = []int{}
	for d := dateRange.From; !d.After(dateRange.To); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		trend.Labels = append(trend.Labels, dateStr)
		trend.Values = append(trend.Values, dateMap[dateStr])
	}
	
	return trend, nil
}

// getWeeklyTrend is a helper to get weekly aggregated data
func (s *Store) getWeeklyTrend(siteID string, dateRange DateRange, dataType string) (TimeSeriesData, error) {
	var trend TimeSeriesData
	var query string
	
	if dataType == "comments" {
		query = `
			SELECT strftime('%Y-W%W', created_at) as week, COUNT(*) as count
			FROM comments
			WHERE site_id = ? AND created_at BETWEEN ? AND ?
			GROUP BY week
			ORDER BY week ASC
		`
	} else {
		query = `
			SELECT strftime('%Y-W%W', r.created_at) as week, COUNT(*) as count
			FROM reactions r
			INNER JOIN allowed_reactions ar ON r.allowed_reaction_id = ar.id
			WHERE ar.site_id = ? AND r.created_at BETWEEN ? AND ?
			GROUP BY week
			ORDER BY week ASC
		`
	}
	
	rows, err := s.db.Query(query, siteID, dateRange.From, dateRange.To)
	if err != nil {
		return trend, fmt.Errorf("failed to get weekly trend: %w", err)
	}
	defer rows.Close()
	
	trend.Labels = []string{}
	trend.Values = []int{}
	for rows.Next() {
		var week string
		var count int
		if err := rows.Scan(&week, &count); err != nil {
			continue
		}
		// Format week label nicely
		weekLabel := strings.Replace(week, "-W", " Week ", 1)
		trend.Labels = append(trend.Labels, weekLabel)
		trend.Values = append(trend.Values, count)
	}
	
	return trend, nil
}

// GetAnalyticsDashboard retrieves complete analytics data for a site
func (s *Store) GetAnalyticsDashboard(siteID string, dateRange DateRange) (*AnalyticsDashboard, error) {
	dashboard := &AnalyticsDashboard{
		SiteID:   siteID,
		DateFrom: dateRange.From,
		DateTo:   dateRange.To,
	}
	
	var err error
	
	// Get comment metrics
	dashboard.Comments, err = s.GetCommentMetrics(siteID, dateRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment metrics: %w", err)
	}
	
	// Get user metrics
	dashboard.Users, err = s.GetUserMetrics(siteID, dateRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get user metrics: %w", err)
	}
	
	// Get reaction metrics
	dashboard.Reactions, err = s.GetReactionMetrics(siteID, dateRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get reaction metrics: %w", err)
	}
	
	// Get moderation metrics
	dashboard.Moderation, err = s.GetModerationMetrics(siteID, dateRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get moderation metrics: %w", err)
	}
	
	// Get comments trend
	dashboard.CommentsTrend, err = s.GetCommentsTrend(siteID, dateRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments trend: %w", err)
	}
	
	// Get reactions trend
	dashboard.ReactionsTrend, err = s.GetReactionsTrend(siteID, dateRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get reactions trend: %w", err)
	}
	
	return dashboard, nil
}

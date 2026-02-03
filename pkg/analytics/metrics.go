package analytics

import (
	"time"
)

// CommentMetrics represents comment-related statistics
type CommentMetrics struct {
	Total          int     `json:"total"`
	Pending        int     `json:"pending"`
	Approved       int     `json:"approved"`
	Rejected       int     `json:"rejected"`
	ApprovalRate   float64 `json:"approval_rate"`
	RejectionRate  float64 `json:"rejection_rate"`
	TotalToday     int     `json:"total_today"`
	TotalThisWeek  int     `json:"total_this_week"`
	TotalThisMonth int     `json:"total_this_month"`
}

// UserMetrics represents user-related statistics
type UserMetrics struct {
	TotalUsers        int              `json:"total_users"`
	ActiveUsersToday  int              `json:"active_users_today"`
	ActiveUsersWeek   int              `json:"active_users_week"`
	ActiveUsersMonth  int              `json:"active_users_month"`
	TopContributors   []TopContributor `json:"top_contributors"`
}

// TopContributor represents a user with their contribution count
type TopContributor struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	CommentCount int    `json:"comment_count"`
}

// ReactionMetrics represents reaction-related statistics
type ReactionMetrics struct {
	Total          int                   `json:"total"`
	TotalToday     int                   `json:"total_today"`
	TotalThisWeek  int                   `json:"total_this_week"`
	TotalThisMonth int                   `json:"total_this_month"`
	ByType         []ReactionBreakdown   `json:"by_type"`
	MostReacted    []MostReactedItem     `json:"most_reacted"`
}

// ReactionBreakdown represents reactions by type
type ReactionBreakdown struct {
	Name  string `json:"name"`
	Emoji string `json:"emoji"`
	Count int    `json:"count"`
}

// MostReactedItem represents items with most reactions
type MostReactedItem struct {
	Type          string `json:"type"` // "page" or "comment"
	PagePath      string `json:"page_path,omitempty"`
	CommentText   string `json:"comment_text,omitempty"`
	ReactionCount int    `json:"reaction_count"`
}

// ModerationMetrics represents moderation-related statistics
type ModerationMetrics struct {
	TotalModerated       int     `json:"total_moderated"`
	AutoRejected         int     `json:"auto_rejected"`
	AutoApproved         int     `json:"auto_approved"`
	ManualReviews        int     `json:"manual_reviews"`
	AverageModerationSec float64 `json:"average_moderation_sec"`
	SpamDetectionRate    float64 `json:"spam_detection_rate"`
}

// TimeSeriesData represents time-series data for charts
type TimeSeriesData struct {
	Labels []string `json:"labels"`
	Values []int    `json:"values"`
}

// AnalyticsDashboard represents complete analytics data for a site
type AnalyticsDashboard struct {
	SiteID            string            `json:"site_id"`
	DateFrom          time.Time         `json:"date_from"`
	DateTo            time.Time         `json:"date_to"`
	Comments          CommentMetrics    `json:"comments"`
	Users             UserMetrics       `json:"users"`
	Reactions         ReactionMetrics   `json:"reactions"`
	Moderation        ModerationMetrics `json:"moderation"`
	CommentsTrend     TimeSeriesData    `json:"comments_trend"`
	ReactionsTrend    TimeSeriesData    `json:"reactions_trend"`
}

// DateRange represents a date range for filtering
type DateRange struct {
	From time.Time
	To   time.Time
}

// GetDefaultDateRange returns last 30 days as default date range
func GetDefaultDateRange() DateRange {
	now := time.Now()
	return DateRange{
		From: now.AddDate(0, 0, -30),
		To:   now,
	}
}

// ParseDateRange parses from and to strings into a DateRange
func ParseDateRange(from, to string) (DateRange, error) {
	defaultRange := GetDefaultDateRange()
	
	var dateFrom, dateTo time.Time
	var err error
	
	if from != "" {
		dateFrom, err = time.Parse("2006-01-02", from)
		if err != nil {
			return defaultRange, err
		}
	} else {
		dateFrom = defaultRange.From
	}
	
	if to != "" {
		dateTo, err = time.Parse("2006-01-02", to)
		if err != nil {
			return defaultRange, err
		}
		// Set to end of day
		dateTo = dateTo.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	} else {
		dateTo = defaultRange.To
	}
	
	return DateRange{From: dateFrom, To: dateTo}, nil
}

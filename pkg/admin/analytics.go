package admin

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/saasuke-labs/kotomi/pkg/analytics"
	"github.com/saasuke-labs/kotomi/pkg/auth"
	"github.com/saasuke-labs/kotomi/pkg/models"
)

// AnalyticsHandler handles analytics-related admin requests
type AnalyticsHandler struct {
	db        *sql.DB
	templates *template.Template
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(db *sql.DB, templates *template.Template) *AnalyticsHandler {
	return &AnalyticsHandler{
		db:        db,
		templates: templates,
	}
}

// ShowDashboard displays the analytics dashboard for a site
func (h *AnalyticsHandler) ShowDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Verify user owns the site
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(siteID)
	if err != nil || site.OwnerID != userID {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	// Parse date range from query parameters
	fromParam := r.URL.Query().Get("from")
	toParam := r.URL.Query().Get("to")
	
	dateRange, err := analytics.ParseDateRange(fromParam, toParam)
	if err != nil {
		log.Printf("Error parsing date range: %v", err)
		dateRange = analytics.GetDefaultDateRange()
	}

	// Get analytics data
	store := analytics.NewStore(h.db)
	dashboard, err := store.GetAnalyticsDashboard(siteID, dateRange)
	if err != nil {
		log.Printf("Error fetching analytics: %v", err)
		http.Error(w, "Failed to fetch analytics", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Site":      site,
		"Dashboard": dashboard,
		"DateFrom":  dateRange.From.Format("2006-01-02"),
		"DateTo":    dateRange.To.Format("2006-01-02"),
	}

	if err := h.templates.ExecuteTemplate(w, "admin/analytics/dashboard.html", data); err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// GetAnalyticsData returns analytics data as JSON (for AJAX requests)
func (h *AnalyticsHandler) GetAnalyticsData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Verify user owns the site
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(siteID)
	if err != nil || site.OwnerID != userID {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	// Parse date range from query parameters
	fromParam := r.URL.Query().Get("from")
	toParam := r.URL.Query().Get("to")
	
	dateRange, err := analytics.ParseDateRange(fromParam, toParam)
	if err != nil {
		log.Printf("Error parsing date range: %v", err)
		dateRange = analytics.GetDefaultDateRange()
	}

	// Get analytics data
	store := analytics.NewStore(h.db)
	dashboard, err := store.GetAnalyticsDashboard(siteID, dateRange)
	if err != nil {
		log.Printf("Error fetching analytics: %v", err)
		http.Error(w, "Failed to fetch analytics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

// ExportCSV exports analytics data to CSV format
func (h *AnalyticsHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Verify user owns the site
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(siteID)
	if err != nil || site.OwnerID != userID {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	// Parse date range from query parameters
	fromParam := r.URL.Query().Get("from")
	toParam := r.URL.Query().Get("to")
	
	dateRange, err := analytics.ParseDateRange(fromParam, toParam)
	if err != nil {
		log.Printf("Error parsing date range: %v", err)
		dateRange = analytics.GetDefaultDateRange()
	}

	// Get analytics data
	store := analytics.NewStore(h.db)
	dashboard, err := store.GetAnalyticsDashboard(siteID, dateRange)
	if err != nil {
		log.Printf("Error fetching analytics: %v", err)
		http.Error(w, "Failed to fetch analytics", http.StatusInternalServerError)
		return
	}

	// Set CSV headers
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=analytics-%s.csv", siteID))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write comment metrics
	writer.Write([]string{"Comment Metrics"})
	writer.Write([]string{"Metric", "Value"})
	writer.Write([]string{"Total Comments", strconv.Itoa(dashboard.Comments.Total)})
	writer.Write([]string{"Pending", strconv.Itoa(dashboard.Comments.Pending)})
	writer.Write([]string{"Approved", strconv.Itoa(dashboard.Comments.Approved)})
	writer.Write([]string{"Rejected", strconv.Itoa(dashboard.Comments.Rejected)})
	writer.Write([]string{"Approval Rate", fmt.Sprintf("%.2f%%", dashboard.Comments.ApprovalRate)})
	writer.Write([]string{"Today", strconv.Itoa(dashboard.Comments.TotalToday)})
	writer.Write([]string{"This Week", strconv.Itoa(dashboard.Comments.TotalThisWeek)})
	writer.Write([]string{"This Month", strconv.Itoa(dashboard.Comments.TotalThisMonth)})
	writer.Write([]string{})

	// Write user metrics
	writer.Write([]string{"User Metrics"})
	writer.Write([]string{"Metric", "Value"})
	writer.Write([]string{"Total Users", strconv.Itoa(dashboard.Users.TotalUsers)})
	writer.Write([]string{"Active Today", strconv.Itoa(dashboard.Users.ActiveUsersToday)})
	writer.Write([]string{"Active This Week", strconv.Itoa(dashboard.Users.ActiveUsersWeek)})
	writer.Write([]string{"Active This Month", strconv.Itoa(dashboard.Users.ActiveUsersMonth)})
	writer.Write([]string{})

	// Write top contributors
	writer.Write([]string{"Top Contributors"})
	writer.Write([]string{"Name", "Email", "Comments"})
	for _, contributor := range dashboard.Users.TopContributors {
		writer.Write([]string{contributor.Name, contributor.Email, strconv.Itoa(contributor.CommentCount)})
	}
	writer.Write([]string{})

	// Write reaction metrics
	writer.Write([]string{"Reaction Metrics"})
	writer.Write([]string{"Metric", "Value"})
	writer.Write([]string{"Total Reactions", strconv.Itoa(dashboard.Reactions.Total)})
	writer.Write([]string{"Today", strconv.Itoa(dashboard.Reactions.TotalToday)})
	writer.Write([]string{"This Week", strconv.Itoa(dashboard.Reactions.TotalThisWeek)})
	writer.Write([]string{"This Month", strconv.Itoa(dashboard.Reactions.TotalThisMonth)})
	writer.Write([]string{})

	// Write reaction breakdown
	writer.Write([]string{"Reactions by Type"})
	writer.Write([]string{"Name", "Emoji", "Count"})
	for _, reaction := range dashboard.Reactions.ByType {
		writer.Write([]string{reaction.Name, reaction.Emoji, strconv.Itoa(reaction.Count)})
	}
	writer.Write([]string{})

	// Write moderation metrics
	writer.Write([]string{"Moderation Metrics"})
	writer.Write([]string{"Metric", "Value"})
	writer.Write([]string{"Total Moderated", strconv.Itoa(dashboard.Moderation.TotalModerated)})
	writer.Write([]string{"Auto Rejected", strconv.Itoa(dashboard.Moderation.AutoRejected)})
	writer.Write([]string{"Auto Approved", strconv.Itoa(dashboard.Moderation.AutoApproved)})
	writer.Write([]string{"Manual Reviews", strconv.Itoa(dashboard.Moderation.ManualReviews)})
	writer.Write([]string{"Avg Moderation Time (sec)", fmt.Sprintf("%.2f", dashboard.Moderation.AverageModerationSec)})
	writer.Write([]string{"Spam Detection Rate", fmt.Sprintf("%.2f%%", dashboard.Moderation.SpamDetectionRate)})
	writer.Write([]string{})

	// Write time series data
	writer.Write([]string{"Comments Trend"})
	writer.Write([]string{"Date", "Count"})
	for i, label := range dashboard.CommentsTrend.Labels {
		if i < len(dashboard.CommentsTrend.Values) {
			writer.Write([]string{label, strconv.Itoa(dashboard.CommentsTrend.Values[i])})
		}
	}
	writer.Write([]string{})

	writer.Write([]string{"Reactions Trend"})
	writer.Write([]string{"Date", "Count"})
	for i, label := range dashboard.ReactionsTrend.Labels {
		if i < len(dashboard.ReactionsTrend.Values) {
			writer.Write([]string{label, strconv.Itoa(dashboard.ReactionsTrend.Values[i])})
		}
	}
}

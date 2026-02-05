package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/saasuke-labs/kotomi/pkg/auth"
	"github.com/saasuke-labs/kotomi/pkg/models"
)

// UserManagementHandler handles admin user management endpoints
type UserManagementHandler struct {
	db        *sql.DB
	templates *template.Template
}

// NewUserManagementHandler creates a new user management handler
func NewUserManagementHandler(db *sql.DB, templates *template.Template) *UserManagementHandler {
	return &UserManagementHandler{
		db:        db,
		templates: templates,
	}
}

// UserStats represents statistics about users
type UserStats struct {
	TotalUsers        int
	ActiveUsers       int
	VerifiedUsers     int
	AverageReputation int
}

// UserWithStats extends User with additional statistics
type UserWithStats struct {
	*models.User
	CommentCount int
}

// UserActivity represents user activity details
type UserActivity struct {
	TotalComments     int
	ApprovedComments  int
	PendingComments   int
	RejectedComments  int
	TotalReactions    int
	ApprovalRate      int
	AccountAgeDays    int
	ActivityLevel     string
}

// CommentInfo represents a comment with page information
type CommentInfo struct {
	ID        string
	Text      string
	PagePath  string
	Status    string
	CreatedAt time.Time
}

// ListUsersHandler handles GET /api/v1/admin/sites/{siteId}/users
func (h *UserManagementHandler) ListUsersHandler(w http.ResponseWriter, r *http.Request) {
	adminUserID := auth.GetUserIDFromContext(r.Context())
	if adminUserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Verify user owns the site
	if !h.verifySiteOwnership(r.Context(), siteID, adminUserID, w) {
		return
	}

	// Get all users for the site
	userStore := models.NewUserStore(h.db)
	users, err := userStore.ListBySite(r.Context(), siteID)
	if err != nil {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// ListUsersPage handles GET /admin/sites/{siteId}/users (HTML view)
func (h *UserManagementHandler) ListUsersPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	adminUserID := auth.GetUserIDFromContext(ctx)
	if adminUserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Verify user owns the site
	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(ctx, siteID)
	if err != nil || site == nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}
	if site.OwnerID != adminUserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get search and filter parameters
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	statusFilter := r.URL.Query().Get("status")

	// Get all users
	userStore := models.NewUserStore(h.db)
	users, err := userStore.ListBySite(ctx, siteID)
	if err != nil {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	// Enhance users with comment counts
	usersWithStats := make([]*UserWithStats, 0, len(users))
	for _, user := range users {
		commentCount := h.getCommentCount(ctx, siteID, user.ID)
		
		// Apply filters
		if statusFilter == "verified" && !user.IsVerified {
			continue
		}
		if statusFilter == "new" && time.Since(user.FirstSeen) > 7*24*time.Hour {
			continue
		}
		if statusFilter == "active" && time.Since(user.LastSeen) > 7*24*time.Hour {
			continue
		}
		
		// Apply search
		if search != "" {
			searchLower := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(user.Name), searchLower) &&
				!strings.Contains(strings.ToLower(user.Email), searchLower) {
				continue
			}
		}
		
		usersWithStats = append(usersWithStats, &UserWithStats{
			User:         user,
			CommentCount: commentCount,
		})
	}

	// Calculate stats
	stats := h.calculateUserStats(ctx, siteID, users)

	data := map[string]interface{}{
		"Site":         site,
		"Users":        usersWithStats,
		"Stats":        stats,
		"Search":       search,
		"StatusFilter": statusFilter,
	}

	if err := h.templates.ExecuteTemplate(w, "admin/users/list.html", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// GetUserHandler handles GET /api/v1/admin/sites/{siteId}/users/{userId}
func (h *UserManagementHandler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	adminUserID := auth.GetUserIDFromContext(r.Context())
	if adminUserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	siteID := vars["siteId"]
	userID := vars["userId"]

	// Verify user owns the site
	if !h.verifySiteOwnership(r.Context(), siteID, adminUserID, w) {
		return
	}

	// Get user
	userStore := models.NewUserStore(h.db)
	user, err := userStore.GetBySiteAndID(r.Context(), siteID, userID)
	if err != nil {
		http.Error(w, "Failed to retrieve user", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// GetUserDetailPage handles GET /admin/sites/{siteId}/users/{userId} (HTML view)
func (h *UserManagementHandler) GetUserDetailPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	adminUserID := auth.GetUserIDFromContext(ctx)
	if adminUserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	siteID := vars["siteId"]
	userID := vars["userId"]

	// Get site
	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(ctx, siteID)
	if err != nil || site == nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}
	if site.OwnerID != adminUserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get user
	userStore := models.NewUserStore(h.db)
	user, err := userStore.GetBySiteAndID(ctx, siteID, userID)
	if err != nil {
		http.Error(w, "Failed to retrieve user", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get user activity
	activity := h.getUserActivity(ctx, siteID, userID)
	
	// Get recent comments
	recentComments := h.getRecentComments(ctx, siteID, userID, 10)

	data := map[string]interface{}{
		"Site":           site,
		"User":           user,
		"Activity":       activity,
		"RecentComments": recentComments,
	}

	if err := h.templates.ExecuteTemplate(w, "admin/users/detail.html", data); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
	}
}

// DeleteUserHandler handles DELETE /api/v1/admin/sites/{siteId}/users/{userId}
func (h *UserManagementHandler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	adminUserID := auth.GetUserIDFromContext(r.Context())
	if adminUserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	siteID := vars["siteId"]
	userID := vars["userId"]

	// Verify user owns the site
	if !h.verifySiteOwnership(r.Context(), siteID, adminUserID, w) {
		return
	}

	// Delete user (cascade deletes comments and reactions)
	userStore := models.NewUserStore(h.db)
	if err := userStore.Delete(r.Context(), siteID, userID); err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// verifySiteOwnership checks if the authenticated admin user owns the specified site
func (h *UserManagementHandler) verifySiteOwnership(ctx context.Context, siteID, adminUserID string, w http.ResponseWriter) bool {
	// Check if site exists and belongs to admin user
	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(ctx, siteID)
	if err != nil || site == nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return false
	}

	if site.OwnerID != adminUserID {
		http.Error(w, "Forbidden: You do not own this site", http.StatusForbidden)
		return false
	}

	return true
}

// calculateUserStats calculates user statistics
func (h *UserManagementHandler) calculateUserStats(ctx context.Context, siteID string, users []*models.User) UserStats {
	stats := UserStats{}
	stats.TotalUsers = len(users)
	
	totalReputation := 0
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
	
	for _, user := range users {
		if user.IsVerified {
			stats.VerifiedUsers++
		}
		if user.LastSeen.After(sevenDaysAgo) {
			stats.ActiveUsers++
		}
		totalReputation += user.ReputationScore
	}
	
	if len(users) > 0 {
		stats.AverageReputation = totalReputation / len(users)
	}
	
	return stats
}

// getCommentCount gets the number of comments for a user
func (h *UserManagementHandler) getCommentCount(ctx context.Context, siteID, userID string) int {
	query := `SELECT COUNT(*) FROM comments WHERE site_id = ? AND author_id = ?`
	var count int
	err := h.db.QueryRowContext(ctx, query, siteID, userID).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

// getUserActivity gets detailed activity for a user
func (h *UserManagementHandler) getUserActivity(ctx context.Context, siteID, userID string) UserActivity {
	activity := UserActivity{}
	
	// Get comment stats
	query := `
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN status = 'approved' THEN 1 ELSE 0 END) as approved,
			SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
			SUM(CASE WHEN status = 'rejected' THEN 1 ELSE 0 END) as rejected
		FROM comments 
		WHERE site_id = ? AND author_id = ?
	`
	h.db.QueryRowContext(ctx, query, siteID, userID).Scan(
		&activity.TotalComments,
		&activity.ApprovedComments,
		&activity.PendingComments,
		&activity.RejectedComments,
	)
	
	// Get reaction count
	query = `SELECT COUNT(*) FROM reactions WHERE user_id = ?`
	h.db.QueryRowContext(ctx, query, userID).Scan(&activity.TotalReactions)
	
	// Calculate approval rate
	if activity.TotalComments > 0 {
		activity.ApprovalRate = (activity.ApprovedComments * 100) / activity.TotalComments
	}
	
	// Get account age
	userStore := models.NewUserStore(h.db)
	user, _ := userStore.GetBySiteAndID(ctx, siteID, userID)
	if user != nil {
		activity.AccountAgeDays = int(time.Since(user.FirstSeen).Hours() / 24)
		
		// Determine activity level
		if activity.TotalComments == 0 {
			activity.ActivityLevel = "Inactive"
		} else if activity.TotalComments < 5 {
			activity.ActivityLevel = "Low"
		} else if activity.TotalComments < 20 {
			activity.ActivityLevel = "Medium"
		} else {
			activity.ActivityLevel = "High"
		}
	}
	
	return activity
}

// getRecentComments gets recent comments for a user
func (h *UserManagementHandler) getRecentComments(ctx context.Context, siteID, userID string, limit int) []CommentInfo {
	query := `
		SELECT c.id, c.text, p.path, c.status, c.created_at
		FROM comments c
		JOIN pages p ON c.page_id = p.id
		WHERE c.site_id = ? AND c.author_id = ?
		ORDER BY c.created_at DESC
		LIMIT ?
	`
	
	rows, err := h.db.QueryContext(ctx, query, siteID, userID, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()
	
	var comments []CommentInfo
	for rows.Next() {
		var c CommentInfo
		if err := rows.Scan(&c.ID, &c.Text, &c.PagePath, &c.Status, &c.CreatedAt); err != nil {
			continue
		}
		comments = append(comments, c)
	}
	
	return comments
}

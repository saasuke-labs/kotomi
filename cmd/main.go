package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/saasuke-labs/kotomi/pkg/admin"
	"github.com/saasuke-labs/kotomi/pkg/auth"
	"github.com/saasuke-labs/kotomi/pkg/comments"
	"github.com/saasuke-labs/kotomi/pkg/middleware"
	"github.com/saasuke-labs/kotomi/pkg/models"
	"github.com/saasuke-labs/kotomi/pkg/moderation"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	_ "github.com/saasuke-labs/kotomi/docs" // Import generated docs
)

// @title Kotomi API
// @version 1.0
// @description A comment and reaction system API for static sites
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/saasuke-labs/kotomi
// @contact.email support@kotomi.dev

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"

// CommentStore interface for abstracting storage implementation
type CommentStore interface {
	AddPageComment(site, page string, comment comments.Comment) error
	GetPageComments(site, page string) ([]comments.Comment, error)
	Close() error
}

// Adapter for in-memory store to match CommentStore interface
type InMemoryStoreAdapter struct {
	*comments.SitePagesIndex
}

func (a *InMemoryStoreAdapter) AddPageComment(site, page string, comment comments.Comment) error {
	a.SitePagesIndex.AddPageComment(site, page, comment)
	return nil
}

func (a *InMemoryStoreAdapter) GetPageComments(site, page string) ([]comments.Comment, error) {
	return a.SitePagesIndex.GetPageComments(site, page), nil
}

func (a *InMemoryStoreAdapter) Close() error {
	return nil // No cleanup needed for in-memory
}

var commentStore CommentStore
var db *sql.DB
var templates *template.Template
var auth0Config *auth.Auth0Config
var moderator moderation.Moderator
var moderationConfigStore *moderation.ConfigStore

// postCommentsHandler creates a new comment for a page
// @Summary Create a comment
// @Description Add a new comment to a page (requires JWT authentication)
// @Tags comments
// @Accept json
// @Produce json
// @Param siteId path string true "Site ID"
// @Param pageId path string true "Page ID"
// @Param comment body comments.Comment true "Comment to create"
// @Success 200 {object} comments.Comment
// @Failure 400 {string} string "Invalid JSON or missing required fields"
// @Failure 401 {string} string "Authentication required"
// @Failure 500 {string} string "Failed to add comment"
// @Security BearerAuth
// @Router /site/{siteId}/page/{pageId}/comments [post]
func postCommentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	
	// Fallback to manual parsing if vars is empty (e.g., in unit tests)
	if len(vars) == 0 {
		parsedVars, err := getUrlParams(r)
		if err != nil {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}
		vars = parsedVars
	}
	
	siteId := vars["siteId"]
	pageId := vars["pageId"]

	// Get authenticated user from context (set by JWT middleware)
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Decode body as a Comment
	var comment comments.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if comment.Text == "" {
		http.Error(w, "Text is required", http.StatusBadRequest)
		return
	}
	
	// Set user information from authenticated user
	comment.ID = uuid.NewString()
	comment.AuthorID = user.ID
	comment.Author = user.Name
	comment.AuthorEmail = user.Email
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	// Apply AI moderation if enabled
	if moderator != nil && moderationConfigStore != nil {
		config, err := moderationConfigStore.GetBySiteID(siteId)
		if err == nil && config != nil && config.Enabled {
			// Analyze comment with AI moderation
			result, err := moderator.AnalyzeComment(comment.Text, *config)
			if err != nil {
				log.Printf("AI moderation failed: %v", err)
				// Continue with default status on error
			} else {
				// Determine status based on moderation result
				comment.Status = moderation.DetermineStatus(result, *config)
				log.Printf("AI moderation result for comment %s: decision=%s, confidence=%.2f, reason=%s",
					comment.ID, result.Decision, result.Confidence, result.Reason)
			}
		}
	}

	// Set default status if not set by moderation
	if comment.Status == "" {
		comment.Status = "pending"
	}

	if err := commentStore.AddPageComment(siteId, pageId, comment); err != nil {
		log.Printf("Error adding comment: %v", err)
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	writeJsonResponse(w, comment)
}

// getUrlParams extracts site and page IDs from the request URL
// This function provides a wrapper around mux.Vars() with fallback to manual parsing
// for unit tests that call handlers directly without using the router.
func getUrlParams(r *http.Request) (map[string]string, error) {
	// Use mux.Vars if available (when using gorilla mux router)
	vars := mux.Vars(r)
	if len(vars) > 0 {
		return vars, nil
	}
	
	// Fallback to manual parsing for legacy/test code
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	// Expected: ["api", "site", "{siteId}", "page", "{pageId}", "comments"]
	// Or: ["api", "v1", "site", "{siteId}", "page", "{pageId}", "comments"]
	if len(parts) == 7 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "site" && parts[4] == "page" && parts[6] == "comments" {
		// Versioned path
		return map[string]string{
			"siteId": parts[3],
			"pageId": parts[5],
		}, nil
	} else if len(parts) == 6 && parts[0] == "api" && parts[1] == "site" && parts[3] == "page" && parts[5] == "comments" {
		// Legacy path
		return map[string]string{
			"siteId": parts[2],
			"pageId": parts[4],
		}, nil
	}

	return nil, fmt.Errorf("invalid path")

}

// getCommentsHandler retrieves all comments for a page
// @Summary Get comments for a page
// @Description Retrieve all comments for a specific page
// @Tags comments
// @Produce json
// @Param siteId path string true "Site ID"
// @Param pageId path string true "Page ID"
// @Success 200 {array} comments.Comment
// @Failure 400 {string} string "Invalid URL"
// @Failure 500 {string} string "Failed to retrieve comments"
// @Router /site/{siteId}/page/{pageId}/comments [get]
func getCommentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	
	// Fallback to manual parsing if vars is empty (e.g., in unit tests)
	if len(vars) == 0 {
		parsedVars, err := getUrlParams(r)
		if err != nil {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}
		vars = parsedVars
	}
	
	siteId := vars["siteId"]
	pageId := vars["pageId"]
	
	comments, err := commentStore.GetPageComments(siteId, pageId)

	if err != nil {
		log.Printf("Error retrieving comments: %v", err)
		http.Error(w, "Failed to retrieve comments", http.StatusInternalServerError)
		return
	}

	writeJsonResponse(w, comments)
}

// updateCommentHandler updates a comment's text (owner only)
// @Summary Update a comment
// @Description Update the text of an existing comment (requires JWT authentication and ownership)
// @Tags comments
// @Accept json
// @Produce json
// @Param siteId path string true "Site ID"
// @Param commentId path string true "Comment ID"
// @Param update body object{text=string} true "Updated comment text"
// @Success 200 {object} comments.Comment
// @Failure 400 {string} string "Invalid JSON or missing required fields"
// @Failure 401 {string} string "Authentication required"
// @Failure 403 {string} string "Forbidden - not the comment owner"
// @Failure 404 {string} string "Comment not found"
// @Failure 500 {string} string "Failed to update comment"
// @Security BearerAuth
// @Router /site/{siteId}/comments/{commentId} [put]
func updateCommentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["commentId"]
	siteID := vars["siteId"]

	// Get authenticated user from context
	user, ok := r.Context().Value(middleware.ContextKeyUser).(*models.KotomiUser)
	if !ok || user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var updateReq struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if updateReq.Text == "" {
		http.Error(w, "Text is required", http.StatusBadRequest)
		return
	}

	// Get the comment to verify ownership
	sqliteStore, ok := commentStore.(*comments.SQLiteStore)
	if !ok {
		http.Error(w, "Storage backend not supported", http.StatusInternalServerError)
		return
	}

	comment, err := sqliteStore.GetCommentByID(commentID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Verify the comment belongs to this site
	if comment.SiteID != siteID {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Verify ownership - user can only edit their own comments
	if comment.AuthorID != user.ID {
		http.Error(w, "Forbidden - you can only edit your own comments", http.StatusForbidden)
		return
	}

	// Update the comment text
	if err := sqliteStore.UpdateCommentText(commentID, updateReq.Text); err != nil {
		log.Printf("Error updating comment: %v", err)
		http.Error(w, "Failed to update comment", http.StatusInternalServerError)
		return
	}

	// Retrieve and return the updated comment
	updatedComment, err := sqliteStore.GetCommentByID(commentID)
	if err != nil {
		http.Error(w, "Failed to retrieve updated comment", http.StatusInternalServerError)
		return
	}

	writeJsonResponse(w, updatedComment)
}

// deleteCommentHandler deletes a comment (owner only)
// @Summary Delete a comment
// @Description Delete an existing comment (requires JWT authentication and ownership)
// @Tags comments
// @Param siteId path string true "Site ID"
// @Param commentId path string true "Comment ID"
// @Success 204 "Comment deleted successfully"
// @Failure 401 {string} string "Authentication required"
// @Failure 403 {string} string "Forbidden - not the comment owner"
// @Failure 404 {string} string "Comment not found"
// @Failure 500 {string} string "Failed to delete comment"
// @Security BearerAuth
// @Router /site/{siteId}/comments/{commentId} [delete]
func deleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["commentId"]
	siteID := vars["siteId"]

	// Get authenticated user from context
	user, ok := r.Context().Value(middleware.ContextKeyUser).(*models.KotomiUser)
	if !ok || user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Get the comment to verify ownership
	sqliteStore, ok := commentStore.(*comments.SQLiteStore)
	if !ok {
		http.Error(w, "Storage backend not supported", http.StatusInternalServerError)
		return
	}

	comment, err := sqliteStore.GetCommentByID(commentID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Verify the comment belongs to this site
	if comment.SiteID != siteID {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	// Verify ownership - user can only delete their own comments
	if comment.AuthorID != user.ID {
		http.Error(w, "Forbidden - you can only delete your own comments", http.StatusForbidden)
		return
	}

	// Delete the comment
	if err := sqliteStore.DeleteComment(commentID); err != nil {
		log.Printf("Error deleting comment: %v", err)
		http.Error(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// getAllowedReactionsHandler retrieves allowed reactions for a site
// @Summary Get allowed reactions
// @Description Retrieve all allowed reactions for a site, optionally filtered by type
// @Tags reactions
// @Produce json
// @Param siteId path string true "Site ID"
// @Param type query string false "Reaction type filter (page or comment)"
// @Success 200 {array} models.AllowedReaction
// @Failure 500 {string} string "Failed to retrieve allowed reactions"
// @Router /site/{siteId}/allowed-reactions [get]
func getAllowedReactionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Check if type filter is provided
	reactionType := r.URL.Query().Get("type")

	allowedReactionStore := models.NewAllowedReactionStore(db)
	var reactions []models.AllowedReaction
	var err error

	if reactionType != "" && (reactionType == "page" || reactionType == "comment") {
		reactions, err = allowedReactionStore.GetBySiteAndType(siteID, reactionType)
	} else {
		reactions, err = allowedReactionStore.GetBySite(siteID)
	}

	if err != nil {
		log.Printf("Error retrieving allowed reactions: %v", err)
		http.Error(w, "Failed to retrieve allowed reactions", http.StatusInternalServerError)
		return
	}

	writeJsonResponse(w, reactions)
}

func addReactionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["commentId"]

	// Get authenticated user from context (set by JWT middleware)
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var req struct {
		AllowedReactionID string `json:"allowed_reaction_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.AllowedReactionID == "" {
		http.Error(w, "allowed_reaction_id is required", http.StatusBadRequest)
		return
	}

	reactionStore := models.NewReactionStore(db)
	reaction, err := reactionStore.AddReaction(commentID, req.AllowedReactionID, user.ID)
	if err != nil {
		log.Printf("Error adding reaction: %v", err)
		http.Error(w, "Failed to add reaction", http.StatusInternalServerError)
		return
	}

	// If reaction is nil, it means the user toggled off their reaction
	if reaction == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	writeJsonResponse(w, reaction)
}

func getReactionsByCommentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["commentId"]

	reactionStore := models.NewReactionStore(db)
	reactions, err := reactionStore.GetReactionsByComment(commentID)
	if err != nil {
		log.Printf("Error retrieving reactions: %v", err)
		http.Error(w, "Failed to retrieve reactions", http.StatusInternalServerError)
		return
	}

	writeJsonResponse(w, reactions)
}

func getReactionCountsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID := vars["commentId"]

	reactionStore := models.NewReactionStore(db)
	counts, err := reactionStore.GetReactionCounts(commentID)
	if err != nil {
		log.Printf("Error retrieving reaction counts: %v", err)
		http.Error(w, "Failed to retrieve reaction counts", http.StatusInternalServerError)
		return
	}

	writeJsonResponse(w, counts)
}

// Page reaction handlers
func addPageReactionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageID := vars["pageId"]

	// Get authenticated user from context (set by JWT middleware)
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var req struct {
		AllowedReactionID string `json:"allowed_reaction_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.AllowedReactionID == "" {
		http.Error(w, "allowed_reaction_id is required", http.StatusBadRequest)
		return
	}

	reactionStore := models.NewReactionStore(db)
	reaction, err := reactionStore.AddPageReaction(pageID, req.AllowedReactionID, user.ID)
	if err != nil {
		log.Printf("Error adding page reaction: %v", err)
		http.Error(w, "Failed to add reaction", http.StatusInternalServerError)
		return
	}

	// If reaction is nil, it means the user toggled off their reaction
	if reaction == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	writeJsonResponse(w, reaction)
}

func getReactionsByPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageID := vars["pageId"]

	reactionStore := models.NewReactionStore(db)
	reactions, err := reactionStore.GetReactionsByPage(pageID)
	if err != nil {
		log.Printf("Error retrieving page reactions: %v", err)
		http.Error(w, "Failed to retrieve reactions", http.StatusInternalServerError)
		return
	}

	writeJsonResponse(w, reactions)
}

func getPageReactionCountsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageID := vars["pageId"]

	reactionStore := models.NewReactionStore(db)
	counts, err := reactionStore.GetPageReactionCounts(pageID)
	if err != nil {
		log.Printf("Error retrieving page reaction counts: %v", err)
		http.Error(w, "Failed to retrieve reaction counts", http.StatusInternalServerError)
		return
	}

	writeJsonResponse(w, counts)
}

func removeReactionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reactionID := vars["reactionId"]

	reactionStore := models.NewReactionStore(db)
	if err := reactionStore.RemoveReaction(reactionID); err != nil {
		log.Printf("Error removing reaction: %v", err)
		http.Error(w, "Failed to remove reaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func getUserIdentifier(r *http.Request) string {
	// Extract IP address without port
	userIdentifier := r.RemoteAddr
	if idx := strings.LastIndex(userIdentifier, ":"); idx != -1 {
		userIdentifier = userIdentifier[:idx]
	}
	
	// Prefer X-Real-IP or X-Forwarded-For if behind a reverse proxy
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		userIdentifier = realIP
	} else if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(forwarded, ","); idx != -1 {
			userIdentifier = strings.TrimSpace(forwarded[:idx])
		} else {
			userIdentifier = strings.TrimSpace(forwarded)
		}
	}
	
	return userIdentifier
}

// deprecationMiddleware adds deprecation headers to legacy API routes
func deprecationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-API-Warn", "Deprecated API endpoint. Please use /api/v1/ prefix instead.")
		w.Header().Set("Deprecation", "true")
		w.Header().Set("Sunset", "Sun, 01 Jun 2026 00:00:00 GMT") // 5 months deprecation period
		next.ServeHTTP(w, r)
	})
}

func writeJsonResponse(w http.ResponseWriter, data interface{}) {

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(data)

}

// getHealthz health check endpoint
// @Summary Health check
// @Description Check if the API is running
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /healthz [get]
func getHealthz(w http.ResponseWriter, r *http.Request) {

	jsonResponse := struct {
		Message string `json:"message,omitempty"`
	}{
		Message: "OK",
	}

	writeJsonResponse(w, jsonResponse)
}

// Auth handlers
func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Generate random state
	state, err := auth.GenerateRandomState()
	if err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}

	// Store state in session
	session, err := auth.GetSession(r)
	if err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	session.Values[auth.SessionKeyState] = state
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Redirect to Auth0
	loginURL := auth0Config.GetLoginURL(state)
	http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	// Verify state
	session, err := auth.GetSession(r)
	if err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	savedState, ok := session.Values[auth.SessionKeyState].(string)
	if !ok || savedState == "" {
		http.Error(w, "Invalid session state", http.StatusBadRequest)
		return
	}

	if r.URL.Query().Get("state") != savedState {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code in request", http.StatusBadRequest)
		return
	}

	token, err := auth0Config.ExchangeCode(r.Context(), code)
	if err != nil {
		log.Printf("Failed to exchange code: %v", err)
		http.Error(w, "Failed to exchange code", http.StatusInternalServerError)
		return
	}

	// Get user info
	userInfo, err := auth0Config.GetUserInfo(r.Context(), token)
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	// Get or create admin user
	adminUserStore := models.NewAdminUserStore(db)
	user, err := adminUserStore.GetByAuth0Sub(userInfo.Sub)
	if err != nil {
		log.Printf("Error checking user: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if user == nil {
		// Create new user
		user, err = adminUserStore.Create(userInfo.Email, userInfo.Name, userInfo.Sub)
		if err != nil {
			log.Printf("Failed to create user: %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
		log.Printf("Created new user: %s", user.Email)
	}

	// Store user info in session
	session.Values[auth.SessionKeyUserID] = user.ID
	session.Values[auth.SessionKeyAuth0Sub] = user.Auth0Sub
	session.Values[auth.SessionKeyEmail] = user.Email
	session.Values[auth.SessionKeyName] = user.Name
	delete(session.Values, auth.SessionKeyState) // Clear the state

	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Redirect to dashboard
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear session
	if err := auth.ClearSession(w, r); err != nil {
		log.Printf("Error clearing session: %v", err)
	}

	// Redirect to Auth0 logout
	returnTo := fmt.Sprintf("http://localhost:%s/", os.Getenv("PORT"))
	if returnTo == "http://localhost:/" {
		returnTo = "http://localhost:8080/"
	}
	logoutURL := auth0Config.GetLogoutURL(returnTo)
	http.Redirect(w, r, logoutURL, http.StatusTemporaryRedirect)
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get admin user
	adminUserStore := models.NewAdminUserStore(db)
	user, err := adminUserStore.GetByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get sites count
	siteStore := models.NewSiteStore(db)
	sites, err := siteStore.GetByOwner(userID)
	if err != nil {
		http.Error(w, "Failed to fetch sites", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"User":         user,
		"SitesCount":   len(sites),
		"PendingCount": 0, // TODO: implement pending comments count
	}

	if err := templates.ExecuteTemplate(w, "admin/dashboard.html", data); err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func showLoginPage(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "login.html", nil); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize the comment store
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./kotomi.db"
	}

	var err error
	sqliteStore, err := comments.NewSQLiteStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize SQLite store: %v", err)
	}
	commentStore = sqliteStore
	db = sqliteStore.GetDB()
	log.Printf("Using SQLite database at: %s", dbPath)

	// Initialize Auth0 config (optional, won't fail if not configured)
	auth0Config, err = auth.NewAuth0Config()
	if err != nil {
		log.Printf("Auth0 not configured: %v", err)
		log.Println("Admin panel will not be available. Set AUTH0_* environment variables to enable.")
	}

	// Initialize session store
	if err := auth.InitSessionStore(); err != nil {
		log.Printf("Session store initialization warning: %v", err)
	}

	// Load templates
	templates, err = template.ParseGlob("templates/**/*.html")
	if err != nil {
		log.Printf("Warning: Failed to load templates: %v", err)
		// Try alternative pattern
		templates, err = template.ParseGlob("templates/*.html")
		if err != nil {
			log.Printf("Warning: Failed to load templates with alternative pattern: %v", err)
		}
	}

	// If templates still not loaded, try loading individually
	if templates == nil {
		templates = template.New("main")
		templateFiles := []string{
			"templates/base.html",
			"templates/login.html",
			"templates/admin/dashboard.html",
			"templates/admin/sites/list.html",
			"templates/admin/sites/detail.html",
			"templates/admin/sites/form.html",
			"templates/admin/pages/list.html",
			"templates/admin/pages/form.html",
			"templates/admin/comments/list.html",
			"templates/admin/comments/row.html",
			"templates/admin/reactions/list.html",
			"templates/admin/reactions/form.html",
			"templates/admin/moderation/form.html",
		}
		for _, file := range templateFiles {
			_, err := templates.ParseFiles(file)
			if err != nil {
				log.Printf("Warning: Could not load template %s: %v", file, err)
			}
		}
	}

	// Initialize AI moderation
	moderationConfigStore = moderation.NewConfigStore(db)
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey != "" {
		moderator = moderation.NewOpenAIModerator(openaiAPIKey)
		log.Println("AI moderation enabled with OpenAI")
	} else {
		moderator = moderation.NewMockModerator()
		log.Println("Using mock moderation (set OPENAI_API_KEY for AI moderation)")
	}

	// Set up graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Create router
	router := mux.NewRouter()

	// Create CORS middleware
	corsMiddleware := middleware.NewCORSMiddleware()

	// Create rate limiter middleware
	rateLimiter := middleware.NewRateLimiter()

	// API v1 routes (with CORS and rate limiting enabled)
	apiV1Router := router.PathPrefix("/api/v1").Subrouter()
	apiV1Router.Use(corsMiddleware.Handler)
	apiV1Router.Use(rateLimiter.Handler)
	
	// Read-only routes (no auth required for phase 1)
	apiV1Router.HandleFunc("/site/{siteId}/page/{pageId}/comments", getCommentsHandler).Methods("GET")
	apiV1Router.HandleFunc("/site/{siteId}/allowed-reactions", getAllowedReactionsHandler).Methods("GET")
	apiV1Router.HandleFunc("/site/{siteId}/comments/{commentId}/reactions", getReactionsByCommentHandler).Methods("GET")
	apiV1Router.HandleFunc("/site/{siteId}/comments/{commentId}/reactions/counts", getReactionCountsHandler).Methods("GET")
	apiV1Router.HandleFunc("/site/{siteId}/pages/{pageId}/reactions", getReactionsByPageHandler).Methods("GET")
	apiV1Router.HandleFunc("/site/{siteId}/pages/{pageId}/reactions/counts", getPageReactionCountsHandler).Methods("GET")
	
	// Protected routes requiring JWT authentication
	apiV1AuthRouter := apiV1Router.PathPrefix("").Subrouter()
	apiV1AuthRouter.Use(middleware.JWTAuthMiddleware(db))
	apiV1AuthRouter.HandleFunc("/site/{siteId}/page/{pageId}/comments", postCommentsHandler).Methods("POST")
	apiV1AuthRouter.HandleFunc("/site/{siteId}/comments/{commentId}", updateCommentHandler).Methods("PUT")
	apiV1AuthRouter.HandleFunc("/site/{siteId}/comments/{commentId}", deleteCommentHandler).Methods("DELETE")
	apiV1AuthRouter.HandleFunc("/site/{siteId}/comments/{commentId}/reactions", addReactionHandler).Methods("POST")
	apiV1AuthRouter.HandleFunc("/site/{siteId}/pages/{pageId}/reactions", addPageReactionHandler).Methods("POST")
	apiV1AuthRouter.HandleFunc("/site/{siteId}/reactions/{reactionId}", removeReactionHandler).Methods("DELETE")

	// Legacy API routes (backward compatibility with deprecation warning)
	legacyAPIRouter := router.PathPrefix("/api").Subrouter()
	legacyAPIRouter.Use(corsMiddleware.Handler)
	legacyAPIRouter.Use(rateLimiter.Handler)
	legacyAPIRouter.Use(deprecationMiddleware)
	
	// Read-only routes
	legacyAPIRouter.HandleFunc("/site/{siteId}/page/{pageId}/comments", getCommentsHandler).Methods("GET")
	legacyAPIRouter.HandleFunc("/site/{siteId}/allowed-reactions", getAllowedReactionsHandler).Methods("GET")
	legacyAPIRouter.HandleFunc("/site/{siteId}/comments/{commentId}/reactions", getReactionsByCommentHandler).Methods("GET")
	legacyAPIRouter.HandleFunc("/site/{siteId}/comments/{commentId}/reactions/counts", getReactionCountsHandler).Methods("GET")
	legacyAPIRouter.HandleFunc("/site/{siteId}/pages/{pageId}/reactions", getReactionsByPageHandler).Methods("GET")
	legacyAPIRouter.HandleFunc("/site/{siteId}/pages/{pageId}/reactions/counts", getPageReactionCountsHandler).Methods("GET")
	
	// Protected write routes
	legacyAuthRouter := legacyAPIRouter.PathPrefix("").Subrouter()
	legacyAuthRouter.Use(middleware.JWTAuthMiddleware(db))
	legacyAuthRouter.HandleFunc("/site/{siteId}/page/{pageId}/comments", postCommentsHandler).Methods("POST")
	legacyAuthRouter.HandleFunc("/site/{siteId}/comments/{commentId}", updateCommentHandler).Methods("PUT")
	legacyAuthRouter.HandleFunc("/site/{siteId}/comments/{commentId}", deleteCommentHandler).Methods("DELETE")
	legacyAuthRouter.HandleFunc("/site/{siteId}/comments/{commentId}/reactions", addReactionHandler).Methods("POST")
	legacyAuthRouter.HandleFunc("/site/{siteId}/pages/{pageId}/reactions", addPageReactionHandler).Methods("POST")
	legacyAuthRouter.HandleFunc("/site/{siteId}/reactions/{reactionId}", removeReactionHandler).Methods("DELETE")

	// Health check endpoint (no CORS needed, but harmless if included)
	router.HandleFunc("/healthz", getHealthz).Methods("GET")

	// Static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Auth routes (if Auth0 is configured)
	if auth0Config != nil {
		router.HandleFunc("/login", loginHandler).Methods("GET")
		router.HandleFunc("/callback", callbackHandler).Methods("GET")
		router.HandleFunc("/logout", logoutHandler).Methods("GET")

		// Admin routes (protected)
		adminRouter := router.PathPrefix("/admin").Subrouter()
		adminRouter.Use(auth.RequireAuth)

		// Dashboard
		adminRouter.HandleFunc("/dashboard", dashboardHandler).Methods("GET")

		// Sites handlers
		sitesHandler := admin.NewSitesHandler(db, templates)
		adminRouter.HandleFunc("/sites", sitesHandler.ListSites).Methods("GET")
		adminRouter.HandleFunc("/sites/new", sitesHandler.ShowSiteForm).Methods("GET")
		adminRouter.HandleFunc("/sites", sitesHandler.CreateSite).Methods("POST")
		adminRouter.HandleFunc("/sites/{siteId}", sitesHandler.GetSite).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/edit", sitesHandler.ShowSiteForm).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}", sitesHandler.UpdateSite).Methods("PUT")
		adminRouter.HandleFunc("/sites/{siteId}", sitesHandler.DeleteSite).Methods("DELETE")

		// Pages handlers
		pagesHandler := admin.NewPagesHandler(db, templates)
		adminRouter.HandleFunc("/sites/{siteId}/pages", pagesHandler.ListPages).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/pages/new", pagesHandler.ShowPageForm).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/pages", pagesHandler.CreatePage).Methods("POST")
		adminRouter.HandleFunc("/sites/{siteId}/pages/{pageId}", pagesHandler.GetPage).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/pages/{pageId}/edit", pagesHandler.ShowPageForm).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/pages/{pageId}", pagesHandler.UpdatePage).Methods("PUT")
		adminRouter.HandleFunc("/sites/{siteId}/pages/{pageId}", pagesHandler.DeletePage).Methods("DELETE")

		// Comments handlers
		commentsHandler := admin.NewCommentsHandler(db, sqliteStore, templates)
		adminRouter.HandleFunc("/sites/{siteId}/comments", commentsHandler.ListComments).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/pages/{pageId}/comments", commentsHandler.ListPageComments).Methods("GET")
		adminRouter.HandleFunc("/comments/{commentId}/approve", commentsHandler.ApproveComment).Methods("POST")
		adminRouter.HandleFunc("/comments/{commentId}/reject", commentsHandler.RejectComment).Methods("POST")
		adminRouter.HandleFunc("/comments/{commentId}", commentsHandler.DeleteComment).Methods("DELETE")

		// Reactions handlers
		reactionsHandler := admin.NewReactionsHandler(db, templates)
		adminRouter.HandleFunc("/sites/{siteId}/reactions", reactionsHandler.ListAllowedReactions).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/reactions/new", reactionsHandler.ShowReactionForm).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/reactions", reactionsHandler.CreateAllowedReaction).Methods("POST")
		adminRouter.HandleFunc("/sites/{siteId}/reactions/{reactionId}/edit", reactionsHandler.ShowReactionForm).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/reactions/{reactionId}", reactionsHandler.UpdateAllowedReaction).Methods("PUT")
		adminRouter.HandleFunc("/sites/{siteId}/reactions/{reactionId}", reactionsHandler.DeleteAllowedReaction).Methods("DELETE")
		adminRouter.HandleFunc("/sites/{siteId}/reactions/stats", reactionsHandler.GetReactionStats).Methods("GET")

		// Moderation handlers
		moderationHandler := admin.NewModerationHandler(db, templates)
		adminRouter.HandleFunc("/sites/{siteId}/moderation", moderationHandler.HandleModerationForm).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/moderation", moderationHandler.HandleModerationUpdate).Methods("POST")

		// Auth configuration handlers
		authConfigHandler := admin.NewAuthConfigHandler(db)
		adminRouter.HandleFunc("/sites/{siteId}/auth/config", authConfigHandler.GetAuthConfig).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/auth/config", authConfigHandler.CreateAuthConfig).Methods("POST")
		adminRouter.HandleFunc("/sites/{siteId}/auth/config", authConfigHandler.UpdateAuthConfig).Methods("PUT")
		adminRouter.HandleFunc("/sites/{siteId}/auth/config", authConfigHandler.DeleteAuthConfig).Methods("DELETE")

		// User management handlers (Phase 2)
		userMgmtHandler := admin.NewUserManagementHandler(db)
		adminRouter.HandleFunc("/sites/{siteId}/users", userMgmtHandler.ListUsersHandler).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/users/{userId}", userMgmtHandler.GetUserHandler).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/users/{userId}", userMgmtHandler.DeleteUserHandler).Methods("DELETE")

		// Redirect /admin to dashboard
		router.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		}).Methods("GET")
	} else {
		// Show login page that explains Auth0 is not configured
		router.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Admin panel is not configured. Please set AUTH0_* environment variables."))
		}).Methods("GET")
	}

	// Swagger UI - Only available in development (when ENV != "production")
	env := os.Getenv("ENV")
	if env != "production" {
		router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
		log.Println("Swagger UI enabled at /swagger/index.html")
	} else {
		log.Println("Swagger UI disabled in production mode")
	}

	// Root handler
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if auth0Config != nil {
			if err := templates.ExecuteTemplate(w, "login.html", nil); err != nil {
				http.Error(w, "Template error", http.StatusInternalServerError)
			}
		} else {
			w.Write([]byte("Kotomi API Server - Visit /healthz to check status"))
		}
	}).Methods("GET")

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second, // Protection against Slowloris attacks
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server running at http://localhost:%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	stop()

	log.Println("Shutting down gracefully...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	// Close database connection
	if err := commentStore.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	log.Println("Server stopped")
}

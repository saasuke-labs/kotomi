package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/saasuke-labs/kotomi/pkg/auth"
	"github.com/saasuke-labs/kotomi/pkg/comments"
	"github.com/saasuke-labs/kotomi/pkg/moderation"
	"github.com/saasuke-labs/kotomi/pkg/notifications"
)

// ServerHandlers wraps the server dependencies for handler methods
type ServerHandlers struct {
	CommentStore          *comments.SQLiteStore
	DB                    *sql.DB
	Templates             *template.Template
	Auth0Config           *auth.Auth0Config
	Moderator             moderation.Moderator
	ModerationConfigStore *moderation.ConfigStore
	NotificationQueue     *notifications.Queue
}

// NewHandlers creates a new ServerHandlers instance
func NewHandlers(
	commentStore *comments.SQLiteStore,
	db *sql.DB,
	templates *template.Template,
	auth0Config *auth.Auth0Config,
	moderator moderation.Moderator,
	moderationConfigStore *moderation.ConfigStore,
	notificationQueue *notifications.Queue,
) *ServerHandlers {
	return &ServerHandlers{
		CommentStore:          commentStore,
		DB:                    db,
		Templates:             templates,
		Auth0Config:           auth0Config,
		Moderator:             moderator,
		ModerationConfigStore: moderationConfigStore,
		NotificationQueue:     notificationQueue,
	}
}

// WriteJsonResponse writes a JSON response to the http.ResponseWriter
func WriteJsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if data == nil {
		data = map[string]interface{}{}
	}
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Note: WriteHeader was already called, so we can't change status code
		log.Printf("Failed to encode response: %v", err)
	}
}

// GetUrlParams extracts site and page IDs from the request URL
// This function provides a wrapper around mux.Vars() with fallback to manual parsing
// for unit tests that call handlers directly without using the router.
func GetUrlParams(r *http.Request) (map[string]string, error) {
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

// GetUserIdentifier extracts a user identifier from the request
// WARNING: This function reads client-provided headers which can be spoofed.
// Only use when behind a trusted reverse proxy that sanitizes these headers.
// - X-User-ID: Should only be set by trusted middleware, not from client requests
// - X-Forwarded-For/X-Real-IP: Only reliable when behind properly configured reverse proxy
func GetUserIdentifier(r *http.Request) string {
	// Try to get user from Auth (preferred) - NOTE: This header can be spoofed if not validated
	// TODO: This should only be read if set by internal middleware, not from client
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return userID
	}
	
	// Fall back to IP address - only reliable behind trusted proxy
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	
	return ip
}

package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/saasuke-labs/kotomi/pkg/auth"
	"github.com/saasuke-labs/kotomi/pkg/models"
)

// UserManagementHandler handles admin user management endpoints
type UserManagementHandler struct {
	db *sql.DB
}

// NewUserManagementHandler creates a new user management handler
func NewUserManagementHandler(db *sql.DB) *UserManagementHandler {
	return &UserManagementHandler{db: db}
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

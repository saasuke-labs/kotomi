package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/saasuke-labs/kotomi/pkg/auth"
	"github.com/saasuke-labs/kotomi/pkg/models"
)

// AuthConfigHandler handles site authentication configuration requests
type AuthConfigHandler struct {
	db        *sql.DB
	templates *template.Template
}

// NewAuthConfigHandler creates a new auth config handler
func NewAuthConfigHandler(db *sql.DB, templates *template.Template) *AuthConfigHandler {
	return &AuthConfigHandler{
		db:        db,
		templates: templates,
	}
}

// HandleAuthConfigForm displays the authentication configuration form
func (h *AuthConfigHandler) HandleAuthConfigForm(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Verify site ownership
	if !h.verifySiteOwnership(r.Context(), siteID, userID, w) {
		return
	}

	// Get auth configuration or use defaults
	authConfigStore := models.NewSiteAuthConfigStore(h.db)
	config, err := authConfigStore.GetBySiteID(r.Context(), siteID)
	if err != nil {
		// Config doesn't exist yet, use defaults
		config = &models.SiteAuthConfig{
			SiteID:                siteID,
			AuthMode:              "external",
			JWTValidationType:     "hmac",
			TokenExpirationBuffer: 30,
		}
	}

	data := map[string]interface{}{
		"SiteID": siteID,
		"Config": config,
	}

	if err := h.templates.ExecuteTemplate(w, "auth/form.html", data); err != nil {
		log.Printf("Error rendering auth config form: %v", err)
		http.Error(w, "Failed to render form", http.StatusInternalServerError)
	}
}

// GetAuthConfig handles GET /admin/sites/{siteId}/auth/config (JSON API)
func (h *AuthConfigHandler) GetAuthConfig(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Verify site ownership
	if !h.verifySiteOwnership(r.Context(), siteID, userID, w) {
		return
	}

	// Get auth configuration
	authConfigStore := models.NewSiteAuthConfigStore(h.db)
	config, err := authConfigStore.GetBySiteID(r.Context(), siteID)
	if err != nil {
		// If not found, return a default configuration
		if err.Error() == "site auth config not found" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "Authentication not configured",
				"site_id": siteID,
			})
			return
		}
		http.Error(w, "Failed to fetch auth configuration", http.StatusInternalServerError)
		return
	}

	// Don't expose secret in response
	config.JWTSecret = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// CreateAuthConfig handles POST /admin/sites/{siteId}/auth/config
func (h *AuthConfigHandler) CreateAuthConfig(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Verify site ownership
	if !h.verifySiteOwnership(r.Context(), siteID, userID, w) {
		return
	}

	// Parse request body
	var config models.SiteAuthConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set site ID from path parameter
	config.SiteID = siteID

	// Validate configuration
	if err := h.validateAuthConfig(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create configuration
	authConfigStore := models.NewSiteAuthConfigStore(h.db)
	if err := authConfigStore.Create(r.Context(), &config); err != nil {
		http.Error(w, "Failed to create auth configuration", http.StatusInternalServerError)
		return
	}

	// Don't expose secret in response
	config.JWTSecret = ""

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(config)
}

// UpdateAuthConfig handles PUT /admin/sites/{siteId}/auth/config
func (h *AuthConfigHandler) UpdateAuthConfig(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Verify site ownership
	if !h.verifySiteOwnership(r.Context(), siteID, userID, w) {
		return
	}

	// Get existing config
	authConfigStore := models.NewSiteAuthConfigStore(h.db)
	existingConfig, err := authConfigStore.GetBySiteID(r.Context(), siteID)
	if err != nil {
		http.Error(w, "Auth configuration not found", http.StatusNotFound)
		return
	}

	// Parse request body
	var updates models.SiteAuthConfig
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update fields (keep ID and timestamps from existing config)
	existingConfig.AuthMode = updates.AuthMode
	existingConfig.JWTValidationType = updates.JWTValidationType
	
	// Only update secret if provided (allow empty string to keep existing)
	if updates.JWTSecret != "" {
		existingConfig.JWTSecret = updates.JWTSecret
	}
	
	existingConfig.JWTPublicKey = updates.JWTPublicKey
	existingConfig.JWKSEndpoint = updates.JWKSEndpoint
	existingConfig.JWTIssuer = updates.JWTIssuer
	existingConfig.JWTAudience = updates.JWTAudience
	
	if updates.TokenExpirationBuffer > 0 {
		existingConfig.TokenExpirationBuffer = updates.TokenExpirationBuffer
	}

	// Validate configuration
	if err := h.validateAuthConfig(existingConfig); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update configuration
	if err := authConfigStore.Update(r.Context(), existingConfig); err != nil {
		http.Error(w, "Failed to update auth configuration", http.StatusInternalServerError)
		return
	}

	// Don't expose secret in response
	existingConfig.JWTSecret = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingConfig)
}

// DeleteAuthConfig handles DELETE /admin/sites/{siteId}/auth/config
func (h *AuthConfigHandler) DeleteAuthConfig(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Verify site ownership
	if !h.verifySiteOwnership(r.Context(), siteID, userID, w) {
		return
	}

	// Get existing config to get its ID
	authConfigStore := models.NewSiteAuthConfigStore(h.db)
	config, err := authConfigStore.GetBySiteID(r.Context(), siteID)
	if err != nil {
		http.Error(w, "Auth configuration not found", http.StatusNotFound)
		return
	}

	// Delete configuration
	if err := authConfigStore.Delete(r.Context(), config.ID); err != nil {
		http.Error(w, "Failed to delete auth configuration", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// verifySiteOwnership verifies that the user owns the specified site
func (h *AuthConfigHandler) verifySiteOwnership(ctx context.Context, siteID, userID string, w http.ResponseWriter) bool {
	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(ctx, siteID)
	if err != nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return false
	}

	if site.OwnerID != userID {
		http.Error(w, "Forbidden: You do not own this site", http.StatusForbidden)
		return false
	}

	return true
}

// validateAuthConfig validates the auth configuration
func (h *AuthConfigHandler) validateAuthConfig(config *models.SiteAuthConfig) error {
	// Validate auth mode
	validAuthModes := map[string]bool{
		"external": true,
		"kotomi":   true,
	}
	if !validAuthModes[config.AuthMode] {
		return fmt.Errorf("invalid auth_mode: must be either 'external' or 'kotomi'")
	}

	// For kotomi auth mode, no JWT validation settings are required (uses internal auth)
	if config.AuthMode == "kotomi" {
		return nil
	}

	// For external auth mode, validate JWT validation type
	validTypes := map[string]bool{
		"hmac":  true,
		"rsa":   true,
		"ecdsa": true,
		"jwks":  true,
	}
	if !validTypes[config.JWTValidationType] {
		return fmt.Errorf("invalid jwt_validation_type: must be one of hmac, rsa, ecdsa, or jwks")
	}

	// Validate required fields based on validation type
	switch config.JWTValidationType {
	case "hmac":
		if config.JWTSecret == "" {
			return http.ErrBodyNotAllowed
		}
	case "rsa", "ecdsa":
		if config.JWTPublicKey == "" {
			return http.ErrBodyNotAllowed
		}
	case "jwks":
		if config.JWKSEndpoint == "" {
			return http.ErrBodyNotAllowed
		}
	}

	return nil
}

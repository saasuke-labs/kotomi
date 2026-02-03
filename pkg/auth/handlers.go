package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/saasuke-labs/kotomi/pkg/models"
)

// AuthHandler handles kotomi authentication API endpoints using Auth0
type AuthHandler struct {
	authStore   *KotomiAuthStore
	db          *sql.DB
	auth0Config *Auth0Config  // Shared Auth0 config for kotomi auth
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *sql.DB, auth0Config *Auth0Config) *AuthHandler {
	return &AuthHandler{
		authStore:   NewKotomiAuthStore(db),
		db:          db,
		auth0Config: auth0Config,
	}
}

// AuthResponse represents an auth response with tokens
type AuthResponse struct {
	User         *KotomiAuthUser `json:"user"`
	Token        string          `json:"token"`
	RefreshToken string          `json:"refresh_token"`
	ExpiresAt    time.Time       `json:"expires_at"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// GetJWTSecret retrieves the JWT secret for a site (internal key for kotomi mode)
func (h *AuthHandler) GetJWTSecret(siteID string) (string, error) {
	// For kotomi auth mode, we use a site-specific internal secret
	// In production, this should be stored in environment variable or secure key storage
	// TODO: Move to environment variable or secure key management
	return "kotomi-internal-jwt-secret-change-in-production-" + siteID, nil
}

// Login redirects to Auth0 login
// @Summary Login with Auth0
// @Description Redirects to Auth0 Universal Login for authentication
// @Tags auth
// @Param siteId query string true "Site ID"
// @Param redirect_uri query string false "Redirect URI after login"
// @Success 302 {string} string "Redirect to Auth0"
// @Failure 400 {object} ErrorResponse
// @Router /auth/login [get]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Get site ID from query parameter
	siteID := r.URL.Query().Get("siteId")
	if siteID == "" {
		http.Error(w, `{"error": "siteId is required"}`, http.StatusBadRequest)
		return
	}

	// Store site ID in session state for callback
	state, err := GenerateRandomState()
	if err != nil {
		http.Error(w, `{"error": "Failed to generate state"}`, http.StatusInternalServerError)
		return
	}
	state = fmt.Sprintf("%s:%s", siteID, state)
	
	// Get Auth0 login URL
	loginURL := h.auth0Config.GetLoginURL(state)
	
	// Redirect to Auth0
	http.Redirect(w, r, loginURL, http.StatusFound)
}

// Callback handles Auth0 callback
// @Summary Auth0 callback
// @Description Handles OAuth callback from Auth0
// @Tags auth
// @Param code query string true "Authorization code"
// @Param state query string true "State parameter"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/callback [get]
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	// Get authorization code and state
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	
	if code == "" || state == "" {
		http.Error(w, `{"error": "Missing code or state"}`, http.StatusBadRequest)
		return
	}
	
	// Extract site ID from state
	parts := strings.SplitN(state, ":", 2)
	if len(parts) != 2 {
		http.Error(w, `{"error": "Invalid state parameter"}`, http.StatusBadRequest)
		return
	}
	siteID := parts[0]
	
	// Exchange code for token
	token, err := h.auth0Config.ExchangeCode(r.Context(), code)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to exchange code: %v"}`, err), http.StatusInternalServerError)
		return
	}
	
	// Get user info from Auth0
	userInfo, err := h.auth0Config.GetUserInfo(r.Context(), token)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to get user info: %v"}`, err), http.StatusInternalServerError)
		return
	}
	
	// Create or update user in database
	user, err := h.authStore.CreateOrUpdateUserFromAuth0(siteID, userInfo)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to create user: %v"}`, err), http.StatusInternalServerError)
		return
	}
	
	// Get JWT secret for this site
	jwtSecret, err := h.GetJWTSecret(siteID)
	if err != nil {
		http.Error(w, `{"error": "Failed to generate token"}`, http.StatusInternalServerError)
		return
	}
	
	// Create session with our own JWT token
	session, err := h.authStore.CreateSession(user, jwtSecret)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to create session: %v"}`, err), http.StatusInternalServerError)
		return
	}
	
	// Set HTTP-only cookie for web clients
	http.SetCookie(w, &http.Cookie{
		Name:     "kotomi_auth_token",
		Value:    session.Token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   r.TLS != nil, // Only secure in HTTPS
		SameSite: http.SameSiteLaxMode,
	})
	
	// Return response
	response := AuthResponse{
		User:         user,
		Token:        session.Token,
		RefreshToken: session.RefreshToken,
		ExpiresAt:    session.ExpiresAt,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Logout handles user logout
// @Summary Logout
// @Description Logout and invalidate session
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get token from Authorization header or cookie
	token := h.extractToken(r)
	if token == "" {
		http.Error(w, `{"error": "No token provided"}`, http.StatusUnauthorized)
		return
	}

	// Delete session
	if err := h.authStore.DeleteSessionByToken(token); err != nil {
		// Even if delete fails, clear the cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "kotomi_auth_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		})
		http.Error(w, fmt.Sprintf(`{"error": "Failed to logout: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "kotomi_auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

// GetCurrentUser returns the current authenticated user
// @Summary Get current user
// @Description Get current user profile
// @Tags auth
// @Produce json
// @Success 200 {object} KotomiAuthUser
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /auth/user [get]
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get token from Authorization header or cookie
	token := h.extractToken(r)
	if token == "" {
		http.Error(w, `{"error": "No token provided"}`, http.StatusUnauthorized)
		return
	}

	// Get session
	session, err := h.authStore.GetSessionByToken(token)
	if err != nil {
		http.Error(w, `{"error": "Invalid or expired token"}`, http.StatusUnauthorized)
		return
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		http.Error(w, `{"error": "Token expired"}`, http.StatusUnauthorized)
		return
	}

	// Get user
	user, err := h.authStore.GetUserByID(session.UserID)
	if err != nil {
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// GetAuthConfig returns the auth configuration for a site
// @Summary Get auth config
// @Description Get authentication configuration for a site (helps clients know which auth flow to use)
// @Tags auth
// @Produce json
// @Param siteId query string true "Site ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Router /auth/config [get]
func (h *AuthHandler) GetAuthConfig(w http.ResponseWriter, r *http.Request) {
	// Get site ID from query parameter
	siteID := r.URL.Query().Get("siteId")
	if siteID == "" {
		http.Error(w, `{"error": "siteId is required"}`, http.StatusBadRequest)
		return
	}
	
	// Get site auth config
	authConfigStore := models.NewSiteAuthConfigStore(h.db)
	authConfig, err := authConfigStore.GetBySiteID(r.Context(), siteID)
	if err != nil {
		http.Error(w, `{"error": "Site not found or auth not configured"}`, http.StatusNotFound)
		return
	}
	
	// Return public auth config info
	response := map[string]interface{}{
		"site_id":   siteID,
		"auth_mode": authConfig.AuthMode,
	}
	
	// Add Auth0 domain if kotomi mode
	if authConfig.AuthMode == "kotomi" && h.auth0Config != nil {
		response["auth0_domain"] = h.auth0Config.Domain
		response["auth0_client_id"] = h.auth0Config.ClientID
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// extractToken extracts JWT token from Authorization header or cookie
func (h *AuthHandler) extractToken(r *http.Request) string {
	// Try Authorization header first
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// Try cookie
	cookie, err := r.Cookie("kotomi_auth_token")
	if err == nil {
		return cookie.Value
	}

	return ""
}

// RegisterRoutes registers auth routes with the router
func (h *AuthHandler) RegisterRoutes(router *mux.Router) {
	// Create subrouter for auth endpoints
	authRouter := router.PathPrefix("/api/v1/auth").Subrouter()

	authRouter.HandleFunc("/login", h.Login).Methods("GET")
	authRouter.HandleFunc("/callback", h.Callback).Methods("GET")
	authRouter.HandleFunc("/logout", h.Logout).Methods("POST")
	authRouter.HandleFunc("/user", h.GetCurrentUser).Methods("GET")
	authRouter.HandleFunc("/config", h.GetAuthConfig).Methods("GET")
}

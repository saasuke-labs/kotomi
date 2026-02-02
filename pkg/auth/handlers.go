package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// AuthHandler handles kotomi authentication API endpoints
type AuthHandler struct {
	authStore *KotomiAuthStore
	db        *sql.DB
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{
		authStore: NewKotomiAuthStore(db),
		db:        db,
	}
}

// SignupRequest represents a signup request
type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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
	// This is stored in an environment variable or generated per site
	// For now, we'll use a simple approach: generate a consistent secret per site
	
	// In production, this should be:
	// 1. Stored in environment variable (KOTOMI_JWT_SECRET)
	// 2. Or stored in a secure key management system
	// 3. Or generated and stored per-site in database
	
	// For MVP, we'll use a hardcoded secret (should be env variable in production)
	// TODO: Move to environment variable or secure key storage
	return "kotomi-internal-jwt-secret-change-in-production-" + siteID, nil
}

// Signup handles user registration
// @Summary Sign up a new user
// @Description Create a new user account for kotomi authentication
// @Tags auth
// @Accept json
// @Produce json
// @Param siteId query string true "Site ID"
// @Param request body SignupRequest true "Signup request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/signup [post]
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	// Get site ID from query parameter
	siteID := r.URL.Query().Get("siteId")
	if siteID == "" {
		http.Error(w, `{"error": "siteId is required"}`, http.StatusBadRequest)
		return
	}

	// Parse request body
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" || req.Name == "" {
		http.Error(w, `{"error": "Email, password, and name are required"}`, http.StatusBadRequest)
		return
	}

	// Basic email validation
	if !strings.Contains(req.Email, "@") {
		http.Error(w, `{"error": "Invalid email format"}`, http.StatusBadRequest)
		return
	}

	// Password strength check (minimum 8 characters)
	if len(req.Password) < 8 {
		http.Error(w, `{"error": "Password must be at least 8 characters"}`, http.StatusBadRequest)
		return
	}

	// Create user
	user, err := h.authStore.CreateUser(siteID, req.Email, req.Password, req.Name)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, `{"error": "Email already exists"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf(`{"error": "Failed to create user: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Get JWT secret for this site
	jwtSecret, err := h.GetJWTSecret(siteID)
	if err != nil {
		http.Error(w, `{"error": "Failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	// Create session
	session, err := h.authStore.CreateSession(user, jwtSecret)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to create session: %v"}`, err), http.StatusInternalServerError)
		return
	}

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

// Login handles user login
// @Summary Login
// @Description Login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param siteId query string true "Site ID"
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Get site ID from query parameter
	siteID := r.URL.Query().Get("siteId")
	if siteID == "" {
		http.Error(w, `{"error": "siteId is required"}`, http.StatusBadRequest)
		return
	}

	// Parse request body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error": "Email and password are required"}`, http.StatusBadRequest)
		return
	}

	// Authenticate user
	user, err := h.authStore.AuthenticateUser(siteID, req.Email, req.Password)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "invalid password") {
			http.Error(w, `{"error": "Invalid email or password"}`, http.StatusUnauthorized)
			return
		}
		http.Error(w, fmt.Sprintf(`{"error": "Authentication failed: %v"}`, err), http.StatusInternalServerError)
		return
	}

	// Get JWT secret for this site
	jwtSecret, err := h.GetJWTSecret(siteID)
	if err != nil {
		http.Error(w, `{"error": "Failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	// Create session
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

	authRouter.HandleFunc("/signup", h.Signup).Methods("POST")
	authRouter.HandleFunc("/login", h.Login).Methods("POST")
	authRouter.HandleFunc("/logout", h.Logout).Methods("POST")
	authRouter.HandleFunc("/user", h.GetCurrentUser).Methods("GET")
}

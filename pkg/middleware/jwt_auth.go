package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/saasuke-labs/kotomi/pkg/auth"
	"github.com/saasuke-labs/kotomi/pkg/models"
)

// ContextKey is the type for context keys
type ContextKey string

const (
	// ContextKeyUser is the context key for authenticated user
	ContextKeyUser ContextKey = "authenticated_user"
)

// JWTAuthMiddleware creates a middleware that validates JWT tokens for a specific site
func JWTAuthMiddleware(db *sql.DB) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract site ID from request path
			vars := mux.Vars(r)
			siteID := vars["siteId"]
			if siteID == "" {
				// Try alternative variable names
				siteID = vars["site_id"]
			}

			if siteID == "" {
				writeJSONError(w, "Site ID not found in request", http.StatusBadRequest)
				return
			}

			// Get site auth configuration
			authConfigStore := models.NewSiteAuthConfigStore(db)
			authConfig, err := authConfigStore.GetBySiteID(siteID)
			if err != nil {
				writeJSONError(w, "Authentication not configured for this site", http.StatusUnauthorized)
				return
			}

			// Extract JWT token from Authorization header
			authHeader := r.Header.Get("Authorization")
			token := auth.ExtractTokenFromHeader(authHeader)
			if token == "" {
				writeJSONError(w, "Authorization token required", http.StatusUnauthorized)
				return
			}

			// Validate JWT token
			validator := auth.NewJWTValidator(authConfig)
			user, err := validator.ValidateToken(token)
			if err != nil {
				writeJSONError(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
				return
			}

			// Add user to request context
			ctx := context.WithValue(r.Context(), ContextKeyUser, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext retrieves the authenticated user from the request context
func GetUserFromContext(ctx context.Context) *models.KotomiUser {
	user, ok := ctx.Value(ContextKeyUser).(*models.KotomiUser)
	if !ok {
		return nil
	}
	return user
}

// RequireAuth is a simple middleware that just checks if user is in context
// Use this after JWTAuthMiddleware to ensure user is authenticated
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			writeJSONError(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// OptionalAuth is a middleware that extracts user if JWT token is present, but doesn't require it
func OptionalAuth(db *sql.DB) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract site ID from request path
			vars := mux.Vars(r)
			siteID := vars["siteId"]
			if siteID == "" {
				siteID = vars["site_id"]
			}

			// If no site ID, or no auth header, just continue without user
			authHeader := r.Header.Get("Authorization")
			if siteID == "" || authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Try to authenticate, but don't fail if it doesn't work
			authConfigStore := models.NewSiteAuthConfigStore(db)
			authConfig, err := authConfigStore.GetBySiteID(siteID)
			if err != nil {
				// No auth config, continue without user
				next.ServeHTTP(w, r)
				return
			}

			token := auth.ExtractTokenFromHeader(authHeader)
			if token != "" {
				validator := auth.NewJWTValidator(authConfig)
				user, err := validator.ValidateToken(token)
				if err == nil && user != nil {
					// Add user to context if validation succeeded
					ctx := context.WithValue(r.Context(), ContextKeyUser, user)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

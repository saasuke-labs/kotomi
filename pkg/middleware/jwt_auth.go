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
			authConfig, err := authConfigStore.GetBySiteID(r.Context(), siteID)
			if err != nil {
				writeJSONError(w, "Authentication not configured for this site", http.StatusUnauthorized)
				return
			}

			// Extract JWT token from Authorization header or cookie (for kotomi mode)
			authHeader := r.Header.Get("Authorization")
			token := auth.ExtractTokenFromHeader(authHeader)
			
			// If no token in header and auth mode is kotomi, try cookie
			if token == "" && authConfig.AuthMode == "kotomi" {
				cookie, err := r.Cookie("kotomi_auth_token")
				if err == nil {
					token = cookie.Value
				}
			}
			
			if token == "" {
				writeJSONError(w, "Authorization token required", http.StatusUnauthorized)
				return
			}

			// Validate JWT token
			validator := auth.NewJWTValidator(authConfig)
			kotomiUser, err := validator.ValidateToken(token)
			if err != nil {
				writeJSONError(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
				return
			}

			// Persist/update user in database (Phase 2)
			userStore := models.NewUserStore(db)
			user := &models.User{
				ID:         kotomiUser.ID,
				SiteID:     siteID,
				Name:       kotomiUser.Name,
				Email:      kotomiUser.Email,
				AvatarURL:  kotomiUser.AvatarURL,
				ProfileURL: kotomiUser.ProfileURL,
				IsVerified: kotomiUser.Verified,
				Roles:      kotomiUser.Roles,
			}
			
			if err := userStore.CreateOrUpdate(r.Context(), user); err != nil {
				// Log error but don't fail the request
				// User data will still be available from JWT
				fmt.Printf("Warning: failed to persist user: %v\n", err)
			}

			// Add kotomi user to request context
			ctx := context.WithValue(r.Context(), ContextKeyUser, kotomiUser)
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
			authConfig, err := authConfigStore.GetBySiteID(r.Context(), siteID)
			if err != nil {
				// No auth config, continue without user
				next.ServeHTTP(w, r)
				return
			}

			token := auth.ExtractTokenFromHeader(authHeader)
			if token != "" {
				validator := auth.NewJWTValidator(authConfig)
				kotomiUser, err := validator.ValidateToken(token)
				if err == nil && kotomiUser != nil {
					// Persist/update user in database (Phase 2)
					userStore := models.NewUserStore(db)
					user := &models.User{
						ID:         kotomiUser.ID,
						SiteID:     siteID,
						Name:       kotomiUser.Name,
						Email:      kotomiUser.Email,
						AvatarURL:  kotomiUser.AvatarURL,
						ProfileURL: kotomiUser.ProfileURL,
						IsVerified: kotomiUser.Verified,
						Roles:      kotomiUser.Roles,
					}
					
					if err := userStore.CreateOrUpdate(r.Context(), user); err != nil {
						// Log error but don't fail the request
						fmt.Printf("Warning: failed to persist user: %v\n", err)
					}

					// Add user to context if validation succeeded
					ctx := context.WithValue(r.Context(), ContextKeyUser, kotomiUser)
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

package auth

import (
	"context"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

const (
	// SessionName is the name of the session cookie
	SessionName = "kotomi-session"
	// SessionKeyUserID is the session key for user ID
	SessionKeyUserID = "user_id"
	// SessionKeyAuth0Sub is the session key for Auth0 subject
	SessionKeyAuth0Sub = "auth0_sub"
	// SessionKeyEmail is the session key for user email
	SessionKeyEmail = "email"
	// SessionKeyName is the session key for user name
	SessionKeyName = "name"
	// SessionKeyState is the session key for OAuth state
	SessionKeyState = "oauth_state"
)

var (
	// Store is the session store
	Store *sessions.CookieStore
)

// InitSessionStore initializes the session store
func InitSessionStore() error {
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		return nil // Will be initialized with a default for development
	}

	Store = sessions.NewCookieStore([]byte(secret))
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		// Set Secure to true in production
		Secure: os.Getenv("ENV") == "production",
	}

	return nil
}

// GetSession returns the session for a request
func GetSession(r *http.Request) (*sessions.Session, error) {
	if Store == nil {
		// Initialize with a default secret for development
		Store = sessions.NewCookieStore([]byte("dev-secret-change-in-production"))
		Store.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 7,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		}
	}
	return Store.Get(r, SessionName)
}

// RequireAuth is a middleware that requires authentication
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := GetSession(r)
		if err != nil {
			http.Error(w, "Session error", http.StatusInternalServerError)
			return
		}

		userID, ok := session.Values[SessionKeyUserID].(string)
		if !ok || userID == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Add user ID to request context
		ctx := context.WithValue(r.Context(), SessionKeyUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext retrieves the user ID from the request context
func GetUserIDFromContext(ctx context.Context) string {
	userID, ok := ctx.Value(SessionKeyUserID).(string)
	if !ok {
		return ""
	}
	return userID
}

// ClearSession clears all session data
func ClearSession(w http.ResponseWriter, r *http.Request) error {
	session, err := GetSession(r)
	if err != nil {
		return err
	}

	// Clear all values
	for key := range session.Values {
		delete(session.Values, key)
	}

	// Set MaxAge to -1 to delete the cookie
	session.Options.MaxAge = -1

	return session.Save(r, w)
}

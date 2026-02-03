package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/saasuke-labs/kotomi/pkg/auth"
	"github.com/saasuke-labs/kotomi/pkg/models"
)

// GetHealthz is a health check handler
// @Summary Health check
// @Description Check if the service is healthy
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /healthz [get]
func (s *ServerHandlers) GetHealthz(w http.ResponseWriter, r *http.Request) {
	jsonResponse := struct {
		Message string `json:"message,omitempty"`
	}{
		Message: "OK",
	}

	WriteJsonResponse(w, jsonResponse)
}

// Login initiates the Auth0 login flow
func (s *ServerHandlers) Login(w http.ResponseWriter, r *http.Request) {
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
	loginURL := s.Auth0Config.GetLoginURL(state)
	http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
}

// Callback handles the Auth0 callback
func (s *ServerHandlers) Callback(w http.ResponseWriter, r *http.Request) {
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

	token, err := s.Auth0Config.ExchangeCode(r.Context(), code)
	if err != nil {
		log.Printf("Failed to exchange code: %v", err)
		http.Error(w, "Failed to exchange code", http.StatusInternalServerError)
		return
	}

	// Get user info
	userInfo, err := s.Auth0Config.GetUserInfo(r.Context(), token)
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	// Get or create admin user
	adminUserStore := models.NewAdminUserStore(s.DB)
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

// Logout handles user logout
func (s *ServerHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear session
	if err := auth.ClearSession(w, r); err != nil {
		log.Printf("Error clearing session: %v", err)
	}

	// Redirect to Auth0 logout
	returnTo := fmt.Sprintf("http://localhost:%s/", os.Getenv("PORT"))
	if returnTo == "http://localhost:/" {
		returnTo = "http://localhost:8080/"
	}
	logoutURL := s.Auth0Config.GetLogoutURL(returnTo)
	http.Redirect(w, r, logoutURL, http.StatusTemporaryRedirect)
}

// Dashboard renders the admin dashboard
func (s *ServerHandlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get admin user
	adminUserStore := models.NewAdminUserStore(s.DB)
	user, err := adminUserStore.GetByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get sites count
	siteStore := models.NewSiteStore(s.DB)
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

	if err := s.Templates.ExecuteTemplate(w, "admin/dashboard.html", data); err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// ShowLoginPage renders the login page
func (s *ServerHandlers) ShowLoginPage(w http.ResponseWriter, r *http.Request) {
	if err := s.Templates.ExecuteTemplate(w, "login.html", nil); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// DeprecationMiddleware adds deprecation headers to legacy API routes
func DeprecationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Deprecation", "true")
		w.Header().Set("Sunset", "2027-12-31")
		w.Header().Set("Link", `</api/v1>; rel="alternate"`)
		next.ServeHTTP(w, r)
	})
}

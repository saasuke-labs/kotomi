package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/saasuke-labs/kotomi/pkg/auth"
	apierrors "github.com/saasuke-labs/kotomi/pkg/errors"
	"github.com/saasuke-labs/kotomi/pkg/middleware"
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

	s.WriteJsonResponse(w, jsonResponse)
}

// Login initiates the Auth0 login flow
func (s *ServerHandlers) Login(w http.ResponseWriter, r *http.Request) {
	// Generate random state
	state, err := auth.GenerateRandomState()
	if err != nil {
		apierrors.WriteError(w, apierrors.InternalServerError("Failed to generate state").WithRequestID(middleware.GetRequestID(r)))
		return
	}

	// Store state in session
	session, err := auth.GetSession(r)
	if err != nil {
		apierrors.WriteError(w, apierrors.InternalServerError("Session error").WithRequestID(middleware.GetRequestID(r)))
		return
	}

	session.Values[auth.SessionKeyState] = state
	if err := session.Save(r, w); err != nil {
		apierrors.WriteError(w, apierrors.InternalServerError("Failed to save session").WithRequestID(middleware.GetRequestID(r)))
		return
	}

	// Redirect to Auth0
	loginURL := s.Auth0Config.GetLoginURL(state)
	http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
}

// Callback handles the Auth0 callback
func (s *ServerHandlers) Callback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Verify state
	session, err := auth.GetSession(r)
	if err != nil {
		apierrors.WriteError(w, apierrors.InternalServerError("Session error").WithRequestID(middleware.GetRequestID(r)))
		return
	}

	savedState, ok := session.Values[auth.SessionKeyState].(string)
	if !ok || savedState == "" {
		apierrors.WriteError(w, apierrors.BadRequest("Invalid session state").WithRequestID(middleware.GetRequestID(r)))
		return
	}

	if r.URL.Query().Get("state") != savedState {
		apierrors.WriteError(w, apierrors.BadRequest("Invalid state parameter").WithRequestID(middleware.GetRequestID(r)))
		return
	}

	// Exchange code for token
	code := r.URL.Query().Get("code")
	if code == "" {
		apierrors.WriteError(w, apierrors.BadRequest("No code in request").WithRequestID(middleware.GetRequestID(r)))
		return
	}

	token, err := s.Auth0Config.ExchangeCode(ctx, code)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to exchange code", "error", err)
		apierrors.WriteError(w, apierrors.InternalServerError("Failed to exchange code").WithRequestID(middleware.GetRequestID(r)))
		return
	}

	// Get user info
	userInfo, err := s.Auth0Config.GetUserInfo(ctx, token)
	if err != nil {
		s.Logger.ErrorContext(ctx, "failed to get user info", "error", err)
		apierrors.WriteError(w, apierrors.InternalServerError("Failed to get user info").WithRequestID(middleware.GetRequestID(r)))
		return
	}

	// Get or create admin user
	adminUserStore := models.NewAdminUserStore(s.DB)
	user, err := adminUserStore.GetByAuth0Sub(ctx, userInfo.Sub)
	if err != nil {
		s.Logger.ErrorContext(ctx, "error checking user", "error", err, "auth0_sub", userInfo.Sub)
		apierrors.WriteError(w, apierrors.DatabaseError("Database error").WithRequestID(middleware.GetRequestID(r)))
		return
	}

	if user == nil {
		// Create new user
		user, err = adminUserStore.Create(ctx, userInfo.Email, userInfo.Name, userInfo.Sub)
		if err != nil {
			s.Logger.ErrorContext(ctx, "failed to create user", "error", err, "email", userInfo.Email)
			apierrors.WriteError(w, apierrors.DatabaseError("Failed to create user").WithRequestID(middleware.GetRequestID(r)))
			return
		}
		s.Logger.InfoContext(ctx, "created new user", "email", user.Email, "user_id", user.ID)
	}

	// Store user info in session
	session.Values[auth.SessionKeyUserID] = user.ID
	session.Values[auth.SessionKeyAuth0Sub] = user.Auth0Sub
	session.Values[auth.SessionKeyEmail] = user.Email
	session.Values[auth.SessionKeyName] = user.Name
	delete(session.Values, auth.SessionKeyState) // Clear the state

	if err := session.Save(r, w); err != nil {
		apierrors.WriteError(w, apierrors.InternalServerError("Failed to save session").WithRequestID(middleware.GetRequestID(r)))
		return
	}

	// Redirect to dashboard
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

// Logout handles user logout
func (s *ServerHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Clear session
	if err := auth.ClearSession(w, r); err != nil {
		s.Logger.ErrorContext(ctx, "error clearing session", "error", err)
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
	ctx := r.Context()
	userID := auth.GetUserIDFromContext(ctx)
	if userID == "" {
		apierrors.WriteError(w, apierrors.Unauthorized("Unauthorized").WithRequestID(middleware.GetRequestID(r)))
		return
	}

	// Get admin user
	adminUserStore := models.NewAdminUserStore(s.DB)
	user, err := adminUserStore.GetByID(ctx, userID)
	if err != nil {
		apierrors.WriteError(w, apierrors.NotFound("User not found").WithRequestID(middleware.GetRequestID(r)))
		return
	}

	// Get sites
	siteStore := models.NewSiteStore(s.DB)
	sites, err := siteStore.GetByOwner(ctx, userID)
	if err != nil {
		apierrors.WriteError(w, apierrors.DatabaseError("Failed to fetch sites").WithRequestID(middleware.GetRequestID(r)))
		return
	}

	// Calculate aggregate statistics across all sites
	var pendingCount, totalComments, totalUsers, totalReactions int
	for _, site := range sites {
		// Pending comments
		var pending int
		s.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM comments WHERE site_id = ? AND status = 'pending'", site.ID).Scan(&pending)
		pendingCount += pending
		
		// Total comments
		var total int
		s.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM comments WHERE site_id = ?", site.ID).Scan(&total)
		totalComments += total
		
		// Total users
		var users int
		s.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE site_id = ?", site.ID).Scan(&users)
		totalUsers += users
		
		// Total reactions
		var reactions int
		s.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM reactions r JOIN pages p ON r.page_id = p.id WHERE p.site_id = ?", site.ID).Scan(&reactions)
		var commentReactions int
		s.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM reactions r JOIN comments c ON r.comment_id = c.id WHERE c.site_id = ?", site.ID).Scan(&commentReactions)
		totalReactions += reactions + commentReactions
	}

	data := map[string]interface{}{
		"User":           user,
		"Sites":          sites,
		"SitesCount":     len(sites),
		"PendingCount":   pendingCount,
		"TotalComments":  totalComments,
		"TotalUsers":     totalUsers,
		"TotalReactions": totalReactions,
	}

	if err := s.Templates.ExecuteTemplate(w, "admin/dashboard.html", data); err != nil {
		s.Logger.ErrorContext(ctx, "template error", "error", err, "template", "admin/dashboard.html")
		apierrors.WriteError(w, apierrors.InternalServerError("Template error").WithRequestID(middleware.GetRequestID(r)))
	}
}

// ShowLoginPage renders the login page
func (s *ServerHandlers) ShowLoginPage(w http.ResponseWriter, r *http.Request) {
	if err := s.Templates.ExecuteTemplate(w, "login.html", nil); err != nil {
		apierrors.WriteError(w, apierrors.InternalServerError("Template error").WithRequestID(middleware.GetRequestID(r)))
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

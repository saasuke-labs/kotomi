package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/saasuke-labs/kotomi/cmd/server/handlers"
	"github.com/saasuke-labs/kotomi/pkg/admin"
	"github.com/saasuke-labs/kotomi/pkg/auth"
	"github.com/saasuke-labs/kotomi/pkg/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	_ "github.com/saasuke-labs/kotomi/docs" // Import generated docs
)

// RegisterRoutes registers all HTTP routes for the server
func (s *Server) RegisterRoutes(router *mux.Router) {
	h := handlers.NewHandlers(
		s.CommentStore,
		s.DB,
		s.Templates,
		s.Auth0Config,
		s.Moderator,
		s.ModerationConfigStore,
		s.NotificationQueue,
		s.Logger,
	)
	
	logger := middleware.NewLogger()

	// Apply global middleware (request ID and logging)
	router.Use(middleware.RequestIDMiddleware)
	router.Use(middleware.LoggingMiddleware(logger))

	// Create CORS middleware
	corsMiddleware := middleware.NewCORSMiddleware()

	// Create rate limiter middleware
	rateLimiter := middleware.NewRateLimiter()

	// API v1 routes (with CORS and rate limiting enabled)
	apiV1Router := router.PathPrefix("/api/v1").Subrouter()
	apiV1Router.Use(corsMiddleware.Handler)
	apiV1Router.Use(rateLimiter.Handler)
	
	// Kotomi authentication routes (no JWT auth required for these endpoints)
	// Use the same Auth0 config as admin panel for kotomi auth mode
	authHandler := auth.NewAuthHandler(s.DB, s.Auth0Config)
	authHandler.RegisterRoutes(router)
	
	// Read-only routes (no auth required for phase 1)
	apiV1Router.HandleFunc("/site/{siteId}/page/{pageId}/comments", h.GetComments).Methods("GET")
	apiV1Router.HandleFunc("/site/{siteId}/allowed-reactions", h.GetAllowedReactions).Methods("GET")
	apiV1Router.HandleFunc("/site/{siteId}/comments/{commentId}/reactions", h.GetReactionsByComment).Methods("GET")
	apiV1Router.HandleFunc("/site/{siteId}/comments/{commentId}/reactions/counts", h.GetReactionCounts).Methods("GET")
	apiV1Router.HandleFunc("/site/{siteId}/pages/{pageId}/reactions", h.GetReactionsByPage).Methods("GET")
	apiV1Router.HandleFunc("/site/{siteId}/pages/{pageId}/reactions/counts", h.GetPageReactionCounts).Methods("GET")
	
	// Protected routes requiring JWT authentication
	apiV1AuthRouter := apiV1Router.PathPrefix("").Subrouter()
	apiV1AuthRouter.Use(middleware.JWTAuthMiddleware(s.DB))
	apiV1AuthRouter.HandleFunc("/site/{siteId}/page/{pageId}/comments", h.PostComments).Methods("POST")
	apiV1AuthRouter.HandleFunc("/site/{siteId}/comments/{commentId}", h.UpdateComment).Methods("PUT")
	apiV1AuthRouter.HandleFunc("/site/{siteId}/comments/{commentId}", h.DeleteComment).Methods("DELETE")
	apiV1AuthRouter.HandleFunc("/site/{siteId}/comments/{commentId}/reactions", h.AddReaction).Methods("POST")
	apiV1AuthRouter.HandleFunc("/site/{siteId}/pages/{pageId}/reactions", h.AddPageReaction).Methods("POST")
	apiV1AuthRouter.HandleFunc("/site/{siteId}/reactions/{reactionId}", h.RemoveReaction).Methods("DELETE")

	// Legacy API routes (backward compatibility with deprecation warning)
	legacyAPIRouter := router.PathPrefix("/api").Subrouter()
	legacyAPIRouter.Use(corsMiddleware.Handler)
	legacyAPIRouter.Use(rateLimiter.Handler)
	legacyAPIRouter.Use(handlers.DeprecationMiddleware)
	
	// Read-only routes
	legacyAPIRouter.HandleFunc("/site/{siteId}/page/{pageId}/comments", h.GetComments).Methods("GET")
	legacyAPIRouter.HandleFunc("/site/{siteId}/allowed-reactions", h.GetAllowedReactions).Methods("GET")
	legacyAPIRouter.HandleFunc("/site/{siteId}/comments/{commentId}/reactions", h.GetReactionsByComment).Methods("GET")
	legacyAPIRouter.HandleFunc("/site/{siteId}/comments/{commentId}/reactions/counts", h.GetReactionCounts).Methods("GET")
	legacyAPIRouter.HandleFunc("/site/{siteId}/pages/{pageId}/reactions", h.GetReactionsByPage).Methods("GET")
	legacyAPIRouter.HandleFunc("/site/{siteId}/pages/{pageId}/reactions/counts", h.GetPageReactionCounts).Methods("GET")
	
	// Protected write routes
	legacyAuthRouter := legacyAPIRouter.PathPrefix("").Subrouter()
	legacyAuthRouter.Use(middleware.JWTAuthMiddleware(s.DB))
	legacyAuthRouter.HandleFunc("/site/{siteId}/page/{pageId}/comments", h.PostComments).Methods("POST")
	legacyAuthRouter.HandleFunc("/site/{siteId}/comments/{commentId}", h.UpdateComment).Methods("PUT")
	legacyAuthRouter.HandleFunc("/site/{siteId}/comments/{commentId}", h.DeleteComment).Methods("DELETE")
	legacyAuthRouter.HandleFunc("/site/{siteId}/comments/{commentId}/reactions", h.AddReaction).Methods("POST")
	legacyAuthRouter.HandleFunc("/site/{siteId}/pages/{pageId}/reactions", h.AddPageReaction).Methods("POST")
	legacyAuthRouter.HandleFunc("/site/{siteId}/reactions/{reactionId}", h.RemoveReaction).Methods("DELETE")

	// Health check endpoint (no CORS needed, but harmless if included)
	router.HandleFunc("/healthz", h.GetHealthz).Methods("GET")

	// Static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Auth routes (if Auth0 is configured)
	if s.Auth0Config != nil {
		router.HandleFunc("/login", h.Login).Methods("GET")
		router.HandleFunc("/callback", h.Callback).Methods("GET")
		router.HandleFunc("/logout", h.Logout).Methods("GET")

		// Admin routes (protected)
		adminRouter := router.PathPrefix("/admin").Subrouter()
		adminRouter.Use(auth.RequireAuth)

		// Dashboard
		adminRouter.HandleFunc("/dashboard", h.Dashboard).Methods("GET")

		// Comments handlers
		commentsHandler := admin.NewCommentsHandler(s.DB, s.CommentStore, s.Templates)
		commentsHandler.SetNotificationQueue(s.NotificationQueue)
		// Sites handlers
		sitesHandler := admin.NewSitesHandler(s.DB, s.Templates)
		adminRouter.HandleFunc("/sites", sitesHandler.ListSites).Methods("GET")
		adminRouter.HandleFunc("/sites/new", sitesHandler.ShowSiteForm).Methods("GET")
		adminRouter.HandleFunc("/sites", sitesHandler.CreateSite).Methods("POST")
		adminRouter.HandleFunc("/sites/{siteId}", sitesHandler.GetSite).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/edit", sitesHandler.ShowSiteForm).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}", sitesHandler.UpdateSite).Methods("PUT")
		adminRouter.HandleFunc("/sites/{siteId}", sitesHandler.DeleteSite).Methods("DELETE")

		// Pages handlers
		pagesHandler := admin.NewPagesHandler(s.DB, s.Templates)
		adminRouter.HandleFunc("/sites/{siteId}/pages", pagesHandler.ListPages).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/pages/new", pagesHandler.ShowPageForm).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/pages", pagesHandler.CreatePage).Methods("POST")
		adminRouter.HandleFunc("/sites/{siteId}/pages/{pageId}", pagesHandler.GetPage).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/pages/{pageId}/edit", pagesHandler.ShowPageForm).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/pages/{pageId}", pagesHandler.UpdatePage).Methods("PUT")
		adminRouter.HandleFunc("/sites/{siteId}/pages/{pageId}", pagesHandler.DeletePage).Methods("DELETE")

		// Comments handlers already added earlier
		adminRouter.HandleFunc("/sites/{siteId}/comments", commentsHandler.ListComments).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/pages/{pageId}/comments", commentsHandler.ListPageComments).Methods("GET")
		adminRouter.HandleFunc("/comments/{commentId}/approve", commentsHandler.ApproveComment).Methods("POST")
		adminRouter.HandleFunc("/comments/{commentId}/reject", commentsHandler.RejectComment).Methods("POST")
		adminRouter.HandleFunc("/comments/{commentId}", commentsHandler.DeleteComment).Methods("DELETE")
		
		// Bulk comment actions
		adminRouter.HandleFunc("/comments/bulk/approve", commentsHandler.BulkApprove).Methods("POST")
		adminRouter.HandleFunc("/comments/bulk/reject", commentsHandler.BulkReject).Methods("POST")
		adminRouter.HandleFunc("/comments/bulk/delete", commentsHandler.BulkDelete).Methods("POST")

		// Reactions handlers
		reactionsHandler := admin.NewReactionsHandler(s.DB, s.Templates)
		adminRouter.HandleFunc("/sites/{siteId}/reactions", reactionsHandler.ListAllowedReactions).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/reactions/new", reactionsHandler.ShowReactionForm).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/reactions", reactionsHandler.CreateAllowedReaction).Methods("POST")
		adminRouter.HandleFunc("/sites/{siteId}/reactions/{reactionId}/edit", reactionsHandler.ShowReactionForm).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/reactions/{reactionId}", reactionsHandler.UpdateAllowedReaction).Methods("PUT")
		adminRouter.HandleFunc("/sites/{siteId}/reactions/{reactionId}", reactionsHandler.DeleteAllowedReaction).Methods("DELETE")
		adminRouter.HandleFunc("/sites/{siteId}/reactions/stats", reactionsHandler.GetReactionStats).Methods("GET")

		// Moderation handlers
		moderationHandler := admin.NewModerationHandler(s.DB, s.Templates)
		adminRouter.HandleFunc("/sites/{siteId}/moderation", moderationHandler.HandleModerationForm).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/moderation", moderationHandler.HandleModerationUpdate).Methods("POST")

		// Notifications handlers
		notificationsHandler := admin.NewNotificationsHandler(s.DB, s.Templates)
		adminRouter.HandleFunc("/sites/{siteId}/notifications", notificationsHandler.HandleNotificationsForm).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/notifications", notificationsHandler.HandleNotificationsUpdate).Methods("POST")
		adminRouter.HandleFunc("/sites/{siteId}/notifications/test", notificationsHandler.HandleTestEmail).Methods("POST")

		// Auth configuration handlers
		authConfigHandler := admin.NewAuthConfigHandler(s.DB, s.Templates)
		adminRouter.HandleFunc("/sites/{siteId}/auth", authConfigHandler.HandleAuthConfigForm).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/auth/config", authConfigHandler.GetAuthConfig).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/auth/config", authConfigHandler.CreateAuthConfig).Methods("POST")
		adminRouter.HandleFunc("/sites/{siteId}/auth/config", authConfigHandler.UpdateAuthConfig).Methods("PUT")
		adminRouter.HandleFunc("/sites/{siteId}/auth/config", authConfigHandler.DeleteAuthConfig).Methods("DELETE")

		// User management handlers (Phase 2)
		userMgmtHandler := admin.NewUserManagementHandler(s.DB, s.Templates)
		adminRouter.HandleFunc("/sites/{siteId}/users", userMgmtHandler.ListUsersPage).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/users/{userId}", userMgmtHandler.GetUserDetailPage).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/users/{userId}", userMgmtHandler.DeleteUserHandler).Methods("DELETE")

		// Export/Import handlers
		exportImportHandler := admin.NewExportImportHandler(s.DB, s.Templates)
		adminRouter.HandleFunc("/sites/{siteId}/export", exportImportHandler.ShowExportForm).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/export", exportImportHandler.ExportData).Methods("POST")
		adminRouter.HandleFunc("/sites/{siteId}/import", exportImportHandler.ShowImportForm).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/import", exportImportHandler.ImportData).Methods("POST")

		// Analytics handlers
		analyticsHandler := admin.NewAnalyticsHandler(s.DB, s.Templates)
		adminRouter.HandleFunc("/sites/{siteId}/analytics", analyticsHandler.ShowDashboard).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/analytics/data", analyticsHandler.GetAnalyticsData).Methods("GET")
		adminRouter.HandleFunc("/sites/{siteId}/analytics/export", analyticsHandler.ExportCSV).Methods("GET")

		// Redirect /admin to dashboard
		router.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		}).Methods("GET")
	} else {
		// Show login page when Auth0 is not configured
		router.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
			h.ShowLoginPage(w, r)
		}).Methods("GET")
	}

	// Swagger documentation (only in development)
	if s.Auth0Config != nil {
		router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	}

	// Root handler - show login or info page
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	}).Methods("GET")
}

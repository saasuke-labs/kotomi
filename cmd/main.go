package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/saasuke-labs/kotomi/cmd/server"
	"github.com/saasuke-labs/kotomi/pkg/auth"
	"github.com/saasuke-labs/kotomi/pkg/comments"
	"github.com/saasuke-labs/kotomi/pkg/db"
	"github.com/saasuke-labs/kotomi/pkg/logging"
	"github.com/saasuke-labs/kotomi/pkg/moderation"
	"github.com/saasuke-labs/kotomi/pkg/notifications"
)

// @title Kotomi API
// @version 1.0
// @description A comment and reaction system API for static sites
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/saasuke-labs/kotomi
// @contact.email support@kotomi.dev

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"

// InMemoryStoreAdapter adapts the in-memory store to match CommentStore interface
type InMemoryStoreAdapter struct {
	*comments.SitePagesIndex
}

func (a *InMemoryStoreAdapter) AddPageComment(site, page string, comment interface{}) error {
	c, ok := comment.(comments.Comment)
	if !ok {
		return fmt.Errorf("expected comments.Comment, got %T", comment)
	}
	a.SitePagesIndex.AddPageComment(site, page, c)
	return nil
}

func (a *InMemoryStoreAdapter) GetPageComments(site, page string) ([]interface{}, error) {
	commentsList := a.SitePagesIndex.GetPageComments(site, page)
	result := make([]interface{}, len(commentsList))
	for i, c := range commentsList {
		result[i] = c
	}
	return result, nil
}

func (a *InMemoryStoreAdapter) Close() error {
	return nil // No cleanup needed for in-memory
}

func main() {
	// Initialize structured logger with context-aware handler
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	contextHandler := logging.NewContextHandler(jsonHandler)
	logger := slog.New(contextHandler)
	slog.SetDefault(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize the database store based on configuration
	dbConfig := db.ConfigFromEnv()
	logger.Info("initializing database", "provider", dbConfig.Provider)
	
	store, err := db.NewStore(context.Background(), dbConfig)
	if err != nil {
		logger.Error("failed to initialize database store", "error", err, "provider", dbConfig.Provider)
		log.Fatalf("Failed to initialize database store: %v", err)
	}
	logger.Info("database initialized", "provider", dbConfig.Provider)

	// Get database connection (will be nil for non-SQL databases like Firestore)
	sqlDB := store.GetDB()

	// Initialize Auth0 config (optional, won't fail if not configured)
	auth0Config, err := auth.NewAuth0Config()
	if err != nil {
		logger.Warn("Auth0 not configured", "error", err)
		logger.Info("admin panel will not be available - set AUTH0_* environment variables to enable")
	}

	// Initialize session store
	if err := auth.InitSessionStore(); err != nil {
		logger.Warn("session store initialization warning", "error", err)
	}

	// Load templates
	templates, err := template.ParseGlob("templates/**/*.html")
	if err != nil {
		logger.Warn("failed to load templates", "error", err)
		// Try alternative pattern
		templates, err = template.ParseGlob("templates/*.html")
		if err != nil {
			logger.Warn("failed to load templates with alternative pattern", "error", err)
		}
	}

	// If templates still not loaded, try loading individually
	if templates == nil {
		templates = template.New("main")
		templateFiles := []string{
			"templates/base.html",
			"templates/login.html",
			"templates/admin/dashboard.html",
			"templates/admin/sites/list.html",
			"templates/admin/sites/detail.html",
			"templates/admin/sites/form.html",
			"templates/admin/pages/list.html",
			"templates/admin/pages/form.html",
			"templates/admin/comments/list.html",
			"templates/admin/comments/row.html",
			"templates/admin/reactions/list.html",
			"templates/admin/reactions/form.html",
			"templates/admin/moderation/form.html",
			"templates/admin/notifications/form.html",
			"templates/admin/users/list.html",
			"templates/admin/users/detail.html",
		}
		for _, file := range templateFiles {
			_, err := templates.ParseFiles(file)
			if err != nil {
				logger.Warn("could not load template", "file", file, "error", err)
			}
		}
	}

	// Initialize AI moderation
	// Note: Moderation requires SQL database (not available with Firestore)
	var moderationConfigStore *moderation.ConfigStore
	if sqlDB != nil {
		moderationConfigStore = moderation.NewConfigStore(sqlDB)
	}
	var moderator moderation.Moderator
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey != "" {
		moderator = moderation.NewOpenAIModerator(openaiAPIKey)
		logger.Info("AI moderation enabled", "provider", "openai")
	} else {
		moderator = moderation.NewMockModerator()
		logger.Info("using mock moderation - set OPENAI_API_KEY for AI moderation")
	}

	// Set up graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize notification queue
	// Note: Notifications require SQL database (not available with Firestore)
	var notificationQueue *notifications.Queue
	if sqlDB != nil {
		notificationQueue = notifications.NewQueue(sqlDB, 30*time.Second, 10)
		go notificationQueue.Start(ctx)
		logger.Info("notification queue processor started")
	} else {
		logger.Warn("notification queue disabled - requires SQL database")
	}

	// Create server configuration
	cfg := server.Config{
		CommentStore:          store,
		DB:                    sqlDB,
		Templates:             templates,
		Auth0Config:           auth0Config,
		Moderator:             moderator,
		ModerationConfigStore: moderationConfigStore,
		NotificationQueue:     notificationQueue,
		Logger:                logger,
	}

	// Create server
	srv, err := server.New(cfg)
	if err != nil {
		logger.Error("failed to create server", "error", err)
		log.Fatalf("Failed to create server: %v", err)
	}

	// Create HTTP server
	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second, // Protection against Slowloris attacks
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("server starting", "port", port, "address", "http://localhost:"+port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	stop()

	logger.Info("shutting down gracefully...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown server
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", "error", err)
	}

	// Close database connection
	if err := store.Close(); err != nil {
		logger.Error("error closing database", "error", err)
	}

	logger.Info("server stopped")
}

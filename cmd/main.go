package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/saasuke-labs/kotomi/cmd/server"
	"github.com/saasuke-labs/kotomi/pkg/auth"
	"github.com/saasuke-labs/kotomi/pkg/comments"
	"github.com/saasuke-labs/kotomi/pkg/logging"
	"github.com/saasuke-labs/kotomi/pkg/moderation"
	"github.com/saasuke-labs/kotomi/pkg/notifications"
)

// maskConnectionString masks sensitive information in connection strings for logging
func maskConnectionString(connStr string) string {
	// For postgres connection strings, mask password
	if strings.Contains(connStr, "postgres://") {
		parts := strings.Split(connStr, "@")
		if len(parts) == 2 {
			return "postgres://***:***@" + parts[1]
		}
	}
	return "***masked***"
}

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

	// Initialize the comment store based on DB_TYPE environment variable
	// DB_TYPE can be "sqlite" (default) or "postgres"
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "sqlite"
	}

	var commentStore comments.Store
	var db *sql.DB
	var err error

	switch dbType {
	case "postgres":
		// PostgreSQL setup using DATABASE_URL
		databaseURL := os.Getenv("DATABASE_URL")
		if databaseURL == "" {
			logger.Error("DATABASE_URL is required when DB_TYPE=postgres")
			log.Fatalf("DATABASE_URL is required when DB_TYPE=postgres")
		}
		
		postgresStore, err := comments.NewPostgresStore(databaseURL)
		if err != nil {
			logger.Error("failed to initialize PostgreSQL store", "error", err)
			log.Fatalf("Failed to initialize PostgreSQL store: %v", err)
		}
		commentStore = postgresStore
		db = postgresStore.GetDB()
		logger.Info("database initialized", "type", "postgres", "url", maskConnectionString(databaseURL))

	case "sqlite":
		// SQLite setup using DB_PATH
		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = "./kotomi.db"
		}
		
		sqliteStore, err := comments.NewSQLiteStore(dbPath)
		if err != nil {
			logger.Error("failed to initialize SQLite store", "error", err)
			log.Fatalf("Failed to initialize SQLite store: %v", err)
		}
		commentStore = sqliteStore
		db = sqliteStore.GetDB()
		logger.Info("database initialized", "type", "sqlite", "path", dbPath)

	default:
		logger.Error("unsupported DB_TYPE", "type", dbType)
		log.Fatalf("Unsupported DB_TYPE: %s (supported: sqlite, postgres)", dbType)
	}

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
	moderationConfigStore := moderation.NewConfigStore(db)
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
	notificationQueue := notifications.NewQueue(db, 30*time.Second, 10)
	go notificationQueue.Start(ctx)
	logger.Info("notification queue processor started")

	// Create server configuration
	cfg := server.Config{
		CommentStore:          commentStore,
		DB:                    db,
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
	if err := commentStore.Close(); err != nil {
		logger.Error("error closing database", "error", err)
	}

	logger.Info("server stopped")
}

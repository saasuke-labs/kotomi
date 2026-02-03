package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/saasuke-labs/kotomi/cmd/server"
	"github.com/saasuke-labs/kotomi/pkg/auth"
	"github.com/saasuke-labs/kotomi/pkg/comments"
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
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize the comment store
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./kotomi.db"
	}

	sqliteStore, err := comments.NewSQLiteStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize SQLite store: %v", err)
	}
	log.Printf("Using SQLite database at: %s", dbPath)

	// Get database connection
	db := sqliteStore.GetDB()

	// Initialize Auth0 config (optional, won't fail if not configured)
	auth0Config, err := auth.NewAuth0Config()
	if err != nil {
		log.Printf("Auth0 not configured: %v", err)
		log.Println("Admin panel will not be available. Set AUTH0_* environment variables to enable.")
	}

	// Initialize session store
	if err := auth.InitSessionStore(); err != nil {
		log.Printf("Session store initialization warning: %v", err)
	}

	// Load templates
	templates, err := template.ParseGlob("templates/**/*.html")
	if err != nil {
		log.Printf("Warning: Failed to load templates: %v", err)
		// Try alternative pattern
		templates, err = template.ParseGlob("templates/*.html")
		if err != nil {
			log.Printf("Warning: Failed to load templates with alternative pattern: %v", err)
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
		}
		for _, file := range templateFiles {
			_, err := templates.ParseFiles(file)
			if err != nil {
				log.Printf("Warning: Could not load template %s: %v", file, err)
			}
		}
	}

	// Initialize AI moderation
	moderationConfigStore := moderation.NewConfigStore(db)
	var moderator moderation.Moderator
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey != "" {
		moderator = moderation.NewOpenAIModerator(openaiAPIKey)
		log.Println("AI moderation enabled with OpenAI")
	} else {
		moderator = moderation.NewMockModerator()
		log.Println("Using mock moderation (set OPENAI_API_KEY for AI moderation)")
	}

	// Set up graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize notification queue
	notificationQueue := notifications.NewQueue(db, 30*time.Second, 10)
	go notificationQueue.Start(ctx)
	log.Println("Notification queue processor started")

	// Create server configuration
	cfg := server.Config{
		CommentStore:          sqliteStore,
		DB:                    db,
		Templates:             templates,
		Auth0Config:           auth0Config,
		Moderator:             moderator,
		ModerationConfigStore: moderationConfigStore,
		NotificationQueue:     notificationQueue,
	}

	// Create server
	srv, err := server.New(cfg)
	if err != nil {
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
		log.Printf("Server running at http://localhost:%s", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	stop()

	log.Println("Shutting down gracefully...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown server
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	// Close database connection
	if err := sqliteStore.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	log.Println("Server stopped")
}

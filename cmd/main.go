package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/saasuke-labs/kotomi/pkg/comments"
)

// CommentStore interface for abstracting storage implementation
type CommentStore interface {
	AddPageComment(site, page string, comment comments.Comment) error
	GetPageComments(site, page string) ([]comments.Comment, error)
	Close() error
}

// Adapter for in-memory store to match CommentStore interface
type InMemoryStoreAdapter struct {
	*comments.SitePagesIndex
}

func (a *InMemoryStoreAdapter) AddPageComment(site, page string, comment comments.Comment) error {
	a.SitePagesIndex.AddPageComment(site, page, comment)
	return nil
}

func (a *InMemoryStoreAdapter) GetPageComments(site, page string) ([]comments.Comment, error) {
	return a.SitePagesIndex.GetPageComments(site, page), nil
}

func (a *InMemoryStoreAdapter) Close() error {
	return nil // No cleanup needed for in-memory
}

var commentStore CommentStore

// /api/site/:site-id/page/:page-id/comments
func postCommentsHandler(w http.ResponseWriter, r *http.Request) {
	vars, err := getUrlParams(r)

	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	siteId := vars["siteId"]
	pageId := vars["pageId"]

	// Decode body as a Comment
	var comment comments.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	comment.ID = uuid.NewString()
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	if err := commentStore.AddPageComment(siteId, pageId, comment); err != nil {
		log.Printf("Error adding comment: %v", err)
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(comment)
}

// Expecting  /api/site/:site-id/page/:page-id/comments
func getUrlParams(r *http.Request) (map[string]string, error) {
	// Parse the path manually
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	// Expected: ["api", "site", "{siteId}", "page", "{pageId}", "comments"]
	if len(parts) != 6 || parts[0] != "api" || parts[1] != "site" || parts[3] != "page" || parts[5] != "comments" {
		return nil, fmt.Errorf("invalid path")
	}

	siteId := parts[2]
	pageId := parts[4]

	return map[string]string{
		"siteId": siteId,
		"pageId": pageId,
	}, nil

}

func getCommentsHandler(w http.ResponseWriter, r *http.Request) {

	vars, err := getUrlParams(r)

	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	siteId := vars["siteId"]
	pageId := vars["pageId"]
	comments, err := commentStore.GetPageComments(siteId, pageId)

	if err != nil {
		log.Printf("Error retrieving comments: %v", err)
		http.Error(w, "Failed to retrieve comments", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(comments)
}

func writeJsonResponse(w http.ResponseWriter, data interface{}) {

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(data)

}

func getHealthz(w http.ResponseWriter, r *http.Request) {

	jsonResponse := struct {
		Message string `json:"message,omitempty"`
	}{
		Message: "OK",
	}

	writeJsonResponse(w, jsonResponse)
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

	var err error
	sqliteStore, err := comments.NewSQLiteStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize SQLite store: %v", err)
	}
	commentStore = sqliteStore
	log.Printf("Using SQLite database at: %s", dbPath)

	// Set up graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", getHealthz)
	mux.HandleFunc("GET /api/site/{siteId}/page/{pageId}/comments", getCommentsHandler)
	mux.HandleFunc("POST /api/site/{siteId}/page/{pageId}/comments", postCommentsHandler)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server running at http://localhost:%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	// Close database connection
	if err := commentStore.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	log.Println("Server stopped")
}

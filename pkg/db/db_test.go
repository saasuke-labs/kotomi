package db

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/saasuke-labs/kotomi/pkg/comments"
)

func TestSQLiteAdapter(t *testing.T) {
	// Create a temporary database file
	dbPath := "/tmp/test_adapter_" + time.Now().Format("20060102150405") + ".db"
	defer os.Remove(dbPath)

	// Create adapter
	adapter, err := NewSQLiteAdapter(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite adapter: %v", err)
	}
	defer adapter.Close()

	ctx := context.Background()

	// Test AddPageComment
	comment := comments.Comment{
		ID:       "test-1",
		Author:   "Test User",
		AuthorID: "user-1",
		Text:     "Test comment",
		Status:   "pending",
	}

	err = adapter.AddPageComment(ctx, "test-site", "test-page", comment)
	if err != nil {
		t.Fatalf("Failed to add comment: %v", err)
	}

	// Test GetPageComments
	results, err := adapter.GetPageComments(ctx, "test-site", "test-page")
	if err != nil {
		t.Fatalf("Failed to get comments: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 comment, got %d", len(results))
	}

	// Test GetCommentByID
	result, err := adapter.GetCommentByID(ctx, "test-1")
	if err != nil {
		t.Fatalf("Failed to get comment by ID: %v", err)
	}

	c, ok := result.(*comments.Comment)
	if !ok {
		t.Fatalf("Expected *comments.Comment, got %T", result)
	}

	if c.ID != "test-1" {
		t.Errorf("Expected ID 'test-1', got '%s'", c.ID)
	}

	// Test UpdateCommentStatus
	err = adapter.UpdateCommentStatus(ctx, "test-1", "approved", "moderator-1")
	if err != nil {
		t.Fatalf("Failed to update comment status: %v", err)
	}

	// Verify status update
	result, err = adapter.GetCommentByID(ctx, "test-1")
	if err != nil {
		t.Fatalf("Failed to get comment after status update: %v", err)
	}

	c, _ = result.(*comments.Comment)
	if c.Status != "approved" {
		t.Errorf("Expected status 'approved', got '%s'", c.Status)
	}

	// Test GetDB
	db := adapter.GetDB()
	if db == nil {
		t.Error("Expected non-nil database connection")
	}
}

func TestConfigFromEnv(t *testing.T) {
	// Test default (SQLite)
	os.Unsetenv("DB_PROVIDER")
	os.Unsetenv("DB_PATH")

	cfg := ConfigFromEnv()
	if cfg.Provider != ProviderSQLite {
		t.Errorf("Expected provider SQLite, got %s", cfg.Provider)
	}
	if cfg.SQLitePath != "./kotomi.db" {
		t.Errorf("Expected default path './kotomi.db', got '%s'", cfg.SQLitePath)
	}

	// Test SQLite with custom path
	os.Setenv("DB_PROVIDER", "sqlite")
	os.Setenv("DB_PATH", "/custom/path.db")
	defer os.Unsetenv("DB_PROVIDER")
	defer os.Unsetenv("DB_PATH")

	cfg = ConfigFromEnv()
	if cfg.Provider != ProviderSQLite {
		t.Errorf("Expected provider SQLite, got %s", cfg.Provider)
	}
	if cfg.SQLitePath != "/custom/path.db" {
		t.Errorf("Expected path '/custom/path.db', got '%s'", cfg.SQLitePath)
	}

	// Test Firestore
	os.Setenv("DB_PROVIDER", "firestore")
	os.Setenv("FIRESTORE_PROJECT_ID", "test-project")
	defer os.Unsetenv("FIRESTORE_PROJECT_ID")

	cfg = ConfigFromEnv()
	if cfg.Provider != ProviderFirestore {
		t.Errorf("Expected provider Firestore, got %s", cfg.Provider)
	}
	if cfg.FirestoreProjectID != "test-project" {
		t.Errorf("Expected project ID 'test-project', got '%s'", cfg.FirestoreProjectID)
	}

	// Test Firestore with GCP_PROJECT
	os.Unsetenv("FIRESTORE_PROJECT_ID")
	os.Setenv("GCP_PROJECT", "gcp-project")
	defer os.Unsetenv("GCP_PROJECT")

	cfg = ConfigFromEnv()
	if cfg.FirestoreProjectID != "gcp-project" {
		t.Errorf("Expected project ID 'gcp-project', got '%s'", cfg.FirestoreProjectID)
	}
}

func TestNewStore(t *testing.T) {
	ctx := context.Background()

	// Test SQLite store creation
	dbPath := "/tmp/test_factory_" + time.Now().Format("20060102150405") + ".db"
	defer os.Remove(dbPath)

	cfg := Config{
		Provider:   ProviderSQLite,
		SQLitePath: dbPath,
	}

	store, err := NewStore(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to create SQLite store: %v", err)
	}
	defer store.Close()

	if store.GetDB() == nil {
		t.Error("Expected non-nil database connection for SQLite")
	}

	// Test invalid provider
	cfg = Config{
		Provider: Provider("invalid"),
	}

	_, err = NewStore(ctx, cfg)
	if err == nil {
		t.Error("Expected error for invalid provider")
	}

	// Test missing SQLite path
	cfg = Config{
		Provider: ProviderSQLite,
	}

	_, err = NewStore(ctx, cfg)
	if err == nil {
		t.Error("Expected error for missing SQLite path")
	}

	// Test missing Firestore project ID
	cfg = Config{
		Provider: ProviderFirestore,
	}

	_, err = NewStore(ctx, cfg)
	if err == nil {
		t.Error("Expected error for missing Firestore project ID")
	}
}

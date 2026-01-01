package models

import (
	"path/filepath"
	"testing"

	"github.com/saasuke-labs/kotomi/pkg/comments"
)

// Helper function to create a test database
func createTestDB(t *testing.T) *comments.SQLiteStore {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := comments.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	return store
}

func TestUserStore_CreateAndGet(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	userStore := NewUserStore(db)

	// Create user
	user, err := userStore.Create("test@example.com", "Test User", "auth0|12345")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if user.ID == "" {
		t.Error("Expected non-empty ID")
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}

	// Get by Auth0 sub
	retrieved, err := userStore.GetByAuth0Sub("auth0|12345")
	if err != nil {
		t.Fatalf("GetByAuth0Sub failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Expected user, got nil")
	}
	if retrieved.Email != user.Email {
		t.Errorf("Expected email '%s', got '%s'", user.Email, retrieved.Email)
	}

	// Get by ID
	retrieved2, err := userStore.GetByID(user.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved2.Email != user.Email {
		t.Errorf("Expected email '%s', got '%s'", user.Email, retrieved2.Email)
	}
}

func TestUserStore_Update(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	userStore := NewUserStore(db)

	// Create user
	user, err := userStore.Create("test@example.com", "Test User", "auth0|12345")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update user
	err = userStore.Update(user.ID, "updated@example.com", "Updated Name")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	retrieved, err := userStore.GetByID(user.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Email != "updated@example.com" {
		t.Errorf("Expected email 'updated@example.com', got '%s'", retrieved.Email)
	}
	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", retrieved.Name)
	}
}

func TestSiteStore_CreateAndGet(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	userStore := NewUserStore(db)
	siteStore := NewSiteStore(db)

	// Create user first
	user, err := userStore.Create("test@example.com", "Test User", "auth0|12345")
	if err != nil {
		t.Fatalf("Create user failed: %v", err)
	}

	// Create site
	site, err := siteStore.Create(user.ID, "Test Site", "example.com", "A test site")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if site.ID == "" {
		t.Error("Expected non-empty ID")
	}
	if site.Name != "Test Site" {
		t.Errorf("Expected name 'Test Site', got '%s'", site.Name)
	}

	// Get by ID
	retrieved, err := siteStore.GetByID(site.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Name != site.Name {
		t.Errorf("Expected name '%s', got '%s'", site.Name, retrieved.Name)
	}

	// Get by owner
	sites, err := siteStore.GetByOwner(user.ID)
	if err != nil {
		t.Fatalf("GetByOwner failed: %v", err)
	}
	if len(sites) != 1 {
		t.Errorf("Expected 1 site, got %d", len(sites))
	}
}

func TestSiteStore_Update(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	userStore := NewUserStore(db)
	siteStore := NewSiteStore(db)

	user, _ := userStore.Create("test@example.com", "Test User", "auth0|12345")
	site, _ := siteStore.Create(user.ID, "Test Site", "example.com", "A test site")

	// Update site
	err := siteStore.Update(site.ID, "Updated Site", "newdomain.com", "Updated description")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	retrieved, err := siteStore.GetByID(site.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Name != "Updated Site" {
		t.Errorf("Expected name 'Updated Site', got '%s'", retrieved.Name)
	}
}

func TestSiteStore_Delete(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	userStore := NewUserStore(db)
	siteStore := NewSiteStore(db)

	user, _ := userStore.Create("test@example.com", "Test User", "auth0|12345")
	site, _ := siteStore.Create(user.ID, "Test Site", "example.com", "A test site")

	// Delete site
	err := siteStore.Delete(site.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = siteStore.GetByID(site.ID)
	if err == nil {
		t.Error("Expected error for deleted site, got nil")
	}
}

func TestPageStore_CreateAndGet(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	userStore := NewUserStore(db)
	siteStore := NewSiteStore(db)
	pageStore := NewPageStore(db)

	user, _ := userStore.Create("test@example.com", "Test User", "auth0|12345")
	site, _ := siteStore.Create(user.ID, "Test Site", "example.com", "A test site")

	// Create page
	page, err := pageStore.Create(site.ID, "/blog/post-1", "Post 1")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if page.ID == "" {
		t.Error("Expected non-empty ID")
	}
	if page.Path != "/blog/post-1" {
		t.Errorf("Expected path '/blog/post-1', got '%s'", page.Path)
	}

	// Get by ID
	retrieved, err := pageStore.GetByID(page.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Path != page.Path {
		t.Errorf("Expected path '%s', got '%s'", page.Path, retrieved.Path)
	}

	// Get by site
	pages, err := pageStore.GetBySite(site.ID)
	if err != nil {
		t.Fatalf("GetBySite failed: %v", err)
	}
	if len(pages) != 1 {
		t.Errorf("Expected 1 page, got %d", len(pages))
	}
}

func TestPageStore_GetBySitePath(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	userStore := NewUserStore(db)
	siteStore := NewSiteStore(db)
	pageStore := NewPageStore(db)

	user, _ := userStore.Create("test@example.com", "Test User", "auth0|12345")
	site, _ := siteStore.Create(user.ID, "Test Site", "example.com", "A test site")
	page, _ := pageStore.Create(site.ID, "/blog/post-1", "Post 1")

	// Get by site and path
	retrieved, err := pageStore.GetBySitePath(site.ID, "/blog/post-1")
	if err != nil {
		t.Fatalf("GetBySitePath failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Expected page, got nil")
	}
	if retrieved.ID != page.ID {
		t.Errorf("Expected ID '%s', got '%s'", page.ID, retrieved.ID)
	}

	// Try non-existent path
	retrieved2, err := pageStore.GetBySitePath(site.ID, "/nonexistent")
	if err != nil {
		t.Fatalf("GetBySitePath failed: %v", err)
	}
	if retrieved2 != nil {
		t.Error("Expected nil for non-existent path")
	}
}

func TestPageStore_Update(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	userStore := NewUserStore(db)
	siteStore := NewSiteStore(db)
	pageStore := NewPageStore(db)

	user, _ := userStore.Create("test@example.com", "Test User", "auth0|12345")
	site, _ := siteStore.Create(user.ID, "Test Site", "example.com", "A test site")
	page, _ := pageStore.Create(site.ID, "/blog/post-1", "Post 1")

	// Update page
	err := pageStore.Update(page.ID, "/blog/updated-post", "Updated Post")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	retrieved, err := pageStore.GetByID(page.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Path != "/blog/updated-post" {
		t.Errorf("Expected path '/blog/updated-post', got '%s'", retrieved.Path)
	}
	if retrieved.Title != "Updated Post" {
		t.Errorf("Expected title 'Updated Post', got '%s'", retrieved.Title)
	}
}

func TestPageStore_Delete(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	userStore := NewUserStore(db)
	siteStore := NewSiteStore(db)
	pageStore := NewPageStore(db)

	user, _ := userStore.Create("test@example.com", "Test User", "auth0|12345")
	site, _ := siteStore.Create(user.ID, "Test Site", "example.com", "A test site")
	page, _ := pageStore.Create(site.ID, "/blog/post-1", "Post 1")

	// Delete page
	err := pageStore.Delete(page.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = pageStore.GetByID(page.ID)
	if err == nil {
		t.Error("Expected error for deleted page, got nil")
	}
}

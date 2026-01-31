package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/saasuke-labs/kotomi/pkg/auth"
	"github.com/saasuke-labs/kotomi/pkg/comments"
	"github.com/saasuke-labs/kotomi/pkg/models"
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

// Helper to create a context with user ID
func contextWithUser(userID string) context.Context {
	return context.WithValue(context.Background(), auth.SessionKeyUserID, userID)
}

func TestSitesHandler_NewSitesHandler(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	handler := NewSitesHandler(db, nil)

	if handler == nil {
		t.Fatal("NewSitesHandler returned nil")
	}
	if handler.db != db {
		t.Error("Handler db not set correctly")
	}
}

func TestSitesHandler_ListSites_Unauthorized(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	handler := NewSitesHandler(db, nil)

	req := httptest.NewRequest("GET", "/admin/sites", nil)
	w := httptest.NewRecorder()

	handler.ListSites(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestSitesHandler_ListSites_Success(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	handler := NewSitesHandler(db, nil)

	// Create a test user and site
	userStore := models.NewUserStore(db)
	user, err := userStore.Create("test@example.com", "Test User", "auth0|123")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	siteStore := models.NewSiteStore(db)
	_, err = siteStore.Create(user.ID, "Test Site", "example.com", "Test description")
	if err != nil {
		t.Fatalf("Failed to create site: %v", err)
	}

	req := httptest.NewRequest("GET", "/admin/sites", nil)
	req = req.WithContext(contextWithUser(user.ID))
	w := httptest.NewRecorder()

	handler.ListSites(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var sites []models.Site
	err = json.NewDecoder(w.Body).Decode(&sites)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(sites) != 1 {
		t.Errorf("Expected 1 site, got %d", len(sites))
	}
	if sites[0].Name != "Test Site" {
		t.Errorf("Expected site name 'Test Site', got '%s'", sites[0].Name)
	}
}

func TestSitesHandler_GetSite_Unauthorized(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	handler := NewSitesHandler(db, nil)

	req := httptest.NewRequest("GET", "/admin/sites/site-123", nil)
	w := httptest.NewRecorder()

	handler.GetSite(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestSitesHandler_GetSite_NotFound(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	handler := NewSitesHandler(db, nil)

	req := httptest.NewRequest("GET", "/admin/sites/nonexistent", nil)
	req = mux.SetURLVars(req, map[string]string{"siteId": "nonexistent"})
	req = req.WithContext(contextWithUser("user-123"))
	w := httptest.NewRecorder()

	handler.GetSite(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestSitesHandler_GetSite_Forbidden(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	handler := NewSitesHandler(db, nil)

	// Create two users
	userStore := models.NewUserStore(db)
	user1, _ := userStore.Create("user1@example.com", "User 1", "auth0|1")
	user2, _ := userStore.Create("user2@example.com", "User 2", "auth0|2")

	// User1 creates a site
	siteStore := models.NewSiteStore(db)
	site, _ := siteStore.Create(user1.ID, "User1 Site", "example.com", "User1 description")

	// Create a router with the proper route
	router := mux.NewRouter()
	router.HandleFunc("/admin/sites/{siteId}", handler.GetSite).Methods("GET")

	// User2 tries to access User1's site
	req := httptest.NewRequest("GET", "/admin/sites/"+site.ID, nil)
	req = req.WithContext(contextWithUser(user2.ID))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestSitesHandler_GetSite_Success(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	handler := NewSitesHandler(db, nil)

	// Create a user and site
	userStore := models.NewUserStore(db)
	user, _ := userStore.Create("test@example.com", "Test User", "auth0|123")

	siteStore := models.NewSiteStore(db)
	site, _ := siteStore.Create(user.ID, "Test Site", "example.com", "Test description")

	// Create a router with the proper route
	router := mux.NewRouter()
	router.HandleFunc("/admin/sites/{siteId}", handler.GetSite).Methods("GET")

	req := httptest.NewRequest("GET", "/admin/sites/"+site.ID, nil)
	req = req.WithContext(contextWithUser(user.ID))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	siteData := response["site"].(map[string]interface{})
	if siteData["name"] != "Test Site" {
		t.Errorf("Expected site name 'Test Site', got '%s'", siteData["name"])
	}
}

func TestSitesHandler_CreateSite_Unauthorized(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	handler := NewSitesHandler(db, nil)

	body := `{"name":"Test Site","url":"https://example.com"}`
	req := httptest.NewRequest("POST", "/admin/sites", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateSite(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestSitesHandler_CreateSite_Success(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	handler := NewSitesHandler(db, nil)

	// Create a user
	userStore := models.NewUserStore(db)
	user, _ := userStore.Create("test@example.com", "Test User", "auth0|123")

	body := bytes.NewBufferString("name=New Site&domain=newsite.com&description=Test description")
	req := httptest.NewRequest("POST", "/admin/sites", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(contextWithUser(user.ID))
	w := httptest.NewRecorder()

	handler.CreateSite(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var site models.Site
	err := json.NewDecoder(w.Body).Decode(&site)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if site.Name != "New Site" {
		t.Errorf("Expected site name 'New Site', got '%s'", site.Name)
	}
	if site.Domain != "newsite.com" {
		t.Errorf("Expected site domain 'newsite.com', got '%s'", site.Domain)
	}
}

func TestSitesHandler_CreateSite_InvalidJSON(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	handler := NewSitesHandler(db, nil)

	// Create a user
	userStore := models.NewUserStore(db)
	user, _ := userStore.Create("test@example.com", "Test User", "auth0|123")

	body := bytes.NewBufferString("")  // Empty name
	req := httptest.NewRequest("POST", "/admin/sites", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(contextWithUser(user.ID))
	w := httptest.NewRecorder()

	handler.CreateSite(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSitesHandler_UpdateSite_Unauthorized(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	handler := NewSitesHandler(db, nil)

	body := `{"name":"Updated Site","url":"https://updated.com"}`
	req := httptest.NewRequest("PUT", "/admin/sites/site-123", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateSite(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestSitesHandler_UpdateSite_Success(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	handler := NewSitesHandler(db, nil)

	// Create a user and site
	userStore := models.NewUserStore(db)
	user, _ := userStore.Create("test@example.com", "Test User", "auth0|123")

	siteStore := models.NewSiteStore(db)
	site, _ := siteStore.Create(user.ID, "Original Site", "original.com", "Original description")

	// Create a router with the proper route
	router := mux.NewRouter()
	router.HandleFunc("/admin/sites/{siteId}", handler.UpdateSite).Methods("PUT")

	body := bytes.NewBufferString("name=Updated Site&domain=updated.com&description=Updated description")
	req := httptest.NewRequest("PUT", "/admin/sites/"+site.ID, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(contextWithUser(user.ID))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Verify update in database
	updated, _ := siteStore.GetByID(site.ID)
	if updated.Name != "Updated Site" {
		t.Errorf("Expected site name 'Updated Site', got '%s'", updated.Name)
	}
}

func TestSitesHandler_DeleteSite_Unauthorized(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	handler := NewSitesHandler(db, nil)

	req := httptest.NewRequest("DELETE", "/admin/sites/site-123", nil)
	w := httptest.NewRecorder()

	handler.DeleteSite(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestSitesHandler_DeleteSite_Success(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	handler := NewSitesHandler(db, nil)

	// Create a user and site
	userStore := models.NewUserStore(db)
	user, _ := userStore.Create("test@example.com", "Test User", "auth0|123")

	siteStore := models.NewSiteStore(db)
	site, _ := siteStore.Create(user.ID, "Site to Delete", "delete.com", "Delete description")

	// Create a router with the proper route
	router := mux.NewRouter()
	router.HandleFunc("/admin/sites/{siteId}", handler.DeleteSite).Methods("DELETE")

	req := httptest.NewRequest("DELETE", "/admin/sites/"+site.ID, nil)
	req = req.WithContext(contextWithUser(user.ID))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	// Verify site is deleted
	_, err := siteStore.GetByID(site.ID)
	if err == nil {
		t.Error("Expected error when getting deleted site, got nil")
	}
}

func TestSitesHandler_ShowSiteForm_NoTemplates(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	handler := NewSitesHandler(db, nil) // No templates

	req := httptest.NewRequest("GET", "/admin/sites/form", nil)
	w := httptest.NewRecorder()

	handler.ShowSiteForm(w, req)

	// Should return 500 when templates are not available
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestSitesHandler_HTMX_Requests(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	handler := NewSitesHandler(db, nil)

	// Create a user and site
	userStore := models.NewUserStore(db)
	user, _ := userStore.Create("test@example.com", "Test User", "auth0|123")

	siteStore := models.NewSiteStore(db)
	_, _ = siteStore.Create(user.ID, "Test Site", "example.com", "Test description")

	// Test HTMX request without templates (should return without error)
	req := httptest.NewRequest("GET", "/admin/sites", nil)
	req.Header.Set("HX-Request", "true")
	req = req.WithContext(contextWithUser(user.ID))
	w := httptest.NewRecorder()

	handler.ListSites(w, req)

	// Should return without template since templates is nil
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestSitesHandler_CreateSite_FormEncoded(t *testing.T) {
	sqliteStore := createTestDB(t)
	defer sqliteStore.Close()

	db := sqliteStore.GetDB()
	handler := NewSitesHandler(db, nil)

	// Create a user
	userStore := models.NewUserStore(db)
	user, _ := userStore.Create("test@example.com", "Test User", "auth0|123")

	body := bytes.NewBufferString("name=Form Site&domain=form.com")
	req := httptest.NewRequest("POST", "/admin/sites", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(contextWithUser(user.ID))
	w := httptest.NewRecorder()

	handler.CreateSite(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Verify site was created
	siteStore := models.NewSiteStore(db)
	sites, _ := siteStore.GetByOwner(user.ID)
	found := false
	for _, s := range sites {
		if s.Name == "Form Site" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected site 'Form Site' to be created")
	}
}

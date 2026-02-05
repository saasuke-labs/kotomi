package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/saasuke-labs/kotomi/pkg/auth"
	"github.com/saasuke-labs/kotomi/pkg/comments"
	"github.com/saasuke-labs/kotomi/pkg/models"
)

// setupUserManagementTest creates a test environment for user management handler tests
// Returns: handler, store, siteID, adminUserID
func setupUserManagementTest(t *testing.T) (*UserManagementHandler, *comments.SQLiteStore, string, string) {
	t.Helper()

	// Create test database
	store, err := comments.NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	db := store.GetDB()

	// Create test admin user
	adminUserStore := models.NewAdminUserStore(db)
	adminUser, err := adminUserStore.Create(context.Background(), "admin@test.com", "Admin User", "auth0|test")
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	// Create test site owned by admin user
	siteStore := models.NewSiteStore(db)
	site, err := siteStore.Create(context.Background(), adminUser.ID, "Test Site", "", "Test site for user management")
	if err != nil {
		t.Fatalf("Failed to create site: %v", err)
	}
	testSiteID := site.ID

	// Create auth config for the site
	authConfigStore := models.NewSiteAuthConfigStore(db)
	authConfig := &models.SiteAuthConfig{
		ID:                "auth-config-1",
		SiteID:            testSiteID,
		AuthMode:          "external",
		JWTValidationType: "hmac",
		JWTSecret:         "test-secret",
		JWTIssuer:         "test-issuer",
		JWTAudience:       "kotomi",
	}
	if err := authConfigStore.Create(context.Background(), authConfig); err != nil {
		t.Fatalf("Failed to create auth config: %v", err)
	}

	// Create some test JWT users
	userStore := models.NewUserStore(db)
	user1 := &models.User{
		ID:         "user-1",
		SiteID:     testSiteID,
		Name:       "Test User 1",
		Email:      "user1@test.com",
		IsVerified: true,
	}
	if err := userStore.CreateOrUpdate(context.Background(), user1); err != nil {
		t.Fatalf("Failed to create test user 1: %v", err)
	}

	user2 := &models.User{
		ID:     "user-2",
		SiteID: testSiteID,
		Name:   "Test User 2",
		Email:  "user2@test.com",
		Roles:  []string{"premium"},
	}
	if err := userStore.CreateOrUpdate(context.Background(), user2); err != nil {
		t.Fatalf("Failed to create test user 2: %v", err)
	}

	handler := NewUserManagementHandler(db, nil) // Pass nil for templates in tests

	// Return handler, store, siteID, and adminUserID for tests to use
	return handler, store, testSiteID, adminUser.ID
}

func TestUserManagementHandler_ListUsersHandler(t *testing.T) {
	handler, store, siteID, adminUserID := setupUserManagementTest(t)
	defer store.Close()

	// Create request
	req := httptest.NewRequest("GET", "/admin/sites/"+siteID+"/users", nil)
	req = mux.SetURLVars(req, map[string]string{"siteId": siteID})

	// Add admin user ID to context
	ctx := auth.SetUserIDInContext(req.Context(), adminUserID)
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.ListUsersHandler(rr, req)

	// Check response
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	// Parse response
	var users []*models.User
	if err := json.NewDecoder(rr.Body).Decode(&users); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify we got both users
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

func TestUserManagementHandler_GetUserHandler(t *testing.T) {
	handler, store, siteID, adminUserID := setupUserManagementTest(t)
	defer store.Close()

	// Create request
	req := httptest.NewRequest("GET", "/admin/sites/"+siteID+"/users/user-1", nil)
	req = mux.SetURLVars(req, map[string]string{
		"siteId": siteID,
		"userId": "user-1",
	})

	// Add admin user ID to context
	ctx := auth.SetUserIDInContext(req.Context(), adminUserID)
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.GetUserHandler(rr, req)

	// Check response
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	// Parse response
	var user models.User
	if err := json.NewDecoder(rr.Body).Decode(&user); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify user details
	if user.ID != "user-1" {
		t.Errorf("Expected user ID 'user-1', got '%s'", user.ID)
	}
	if user.Name != "Test User 1" {
		t.Errorf("Expected name 'Test User 1', got '%s'", user.Name)
	}
	if user.Email != "user1@test.com" {
		t.Errorf("Expected email 'user1@test.com', got '%s'", user.Email)
	}
}

func TestUserManagementHandler_DeleteUserHandler(t *testing.T) {
	handler, store, siteID, adminUserID := setupUserManagementTest(t)
	defer store.Close()

	// Create request
	req := httptest.NewRequest("DELETE", "/admin/sites/"+siteID+"/users/user-1", nil)
	req = mux.SetURLVars(req, map[string]string{
		"siteId": siteID,
		"userId": "user-1",
	})

	// Add admin user ID to context
	ctx := auth.SetUserIDInContext(req.Context(), adminUserID)
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.DeleteUserHandler(rr, req)

	// Check response
	if rr.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", rr.Code)
	}

	// Verify user was deleted
	userStore := models.NewUserStore(store.GetDB())
	user, err := userStore.GetBySiteAndID(context.Background(), siteID, "user-1")
	if err != nil {
		t.Fatalf("Failed to check if user was deleted: %v", err)
	}
	if user != nil {
		t.Error("Expected user to be deleted, but it still exists")
	}
}

func TestUserManagementHandler_UnauthorizedAccess(t *testing.T) {
	handler, store, siteID, _ := setupUserManagementTest(t)
	defer store.Close()

	// Create request without auth context
	req := httptest.NewRequest("GET", "/admin/sites/"+siteID+"/users", nil)
	req = mux.SetURLVars(req, map[string]string{"siteId": siteID})

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.ListUsersHandler(rr, req)

	// Check response
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rr.Code)
	}
}

func TestUserManagementHandler_ForbiddenAccess(t *testing.T) {
	handler, store, siteID, _ := setupUserManagementTest(t)
	defer store.Close()

	// Create another admin user who doesn't own the site
	adminUserStore := models.NewAdminUserStore(store.GetDB())
	otherAdmin, err := adminUserStore.Create(context.Background(), "other@test.com", "Other Admin", "auth0|other")
	if err != nil {
		t.Fatalf("Failed to create other admin: %v", err)
	}

	// Create request with other admin's context
	req := httptest.NewRequest("GET", "/admin/sites/"+siteID+"/users", nil)
	req = mux.SetURLVars(req, map[string]string{"siteId": siteID})
	ctx := auth.SetUserIDInContext(req.Context(), otherAdmin.ID)
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.ListUsersHandler(rr, req)

	// Check response
	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", rr.Code)
	}
}

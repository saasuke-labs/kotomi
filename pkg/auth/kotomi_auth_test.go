package auth

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	// Create temporary database
	dbPath := "/tmp/test_kotomi_auth_" + time.Now().Format("20060102150405") + ".db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		db.Close()
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Create schema - updated for Auth0
	schema := `
	CREATE TABLE IF NOT EXISTS sites (
		id TEXT PRIMARY KEY,
		owner_id TEXT NOT NULL,
		name TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS kotomi_auth_users (
		id TEXT PRIMARY KEY,
		site_id TEXT NOT NULL,
		email TEXT NOT NULL,
		auth0_sub TEXT NOT NULL,
		name TEXT,
		avatar_url TEXT,
		is_verified INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE,
		UNIQUE(site_id, auth0_sub)
	);

	CREATE TABLE IF NOT EXISTS kotomi_auth_sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		site_id TEXT NOT NULL,
		token TEXT NOT NULL UNIQUE,
		refresh_token TEXT NOT NULL UNIQUE,
		expires_at TIMESTAMP NOT NULL,
		refresh_expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES kotomi_auth_users(id) ON DELETE CASCADE,
		FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE
	);
	`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		t.Fatalf("Failed to create schema: %v", err)
	}

	// Insert test site
	_, err = db.Exec(`INSERT INTO sites (id, owner_id, name) VALUES (?, ?, ?)`, "test-site", "test-owner", "Test Site")
	if err != nil {
		db.Close()
		t.Fatalf("Failed to insert test site: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return db, cleanup
}

func TestGenerateRandomToken(t *testing.T) {
	token1, err := GenerateRandomToken()
	if err != nil {
		t.Fatalf("GenerateRandomToken failed: %v", err)
	}

	token2, err := GenerateRandomToken()
	if err != nil {
		t.Fatalf("GenerateRandomToken failed: %v", err)
	}

	if token1 == "" || token2 == "" {
		t.Error("Expected non-empty tokens")
	}

	if token1 == token2 {
		t.Error("Generated tokens should be unique")
	}
}

func TestCreateOrUpdateUserFromAuth0(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewKotomiAuthStore(db)

	// Create mock Auth0 user info
	userInfo := &UserInfo{
		Sub:           "auth0|12345",
		Email:         "test@example.com",
		Name:          "Test User",
		Picture:       "https://example.com/avatar.jpg",
		EmailVerified: true,
	}

	// Create user from Auth0
	user, err := store.CreateOrUpdateUserFromAuth0("test-site", userInfo)
	if err != nil {
		t.Fatalf("CreateOrUpdateUserFromAuth0 failed: %v", err)
	}

	// Verify user properties
	if user.ID == "" {
		t.Error("Expected user to have an ID")
	}
	if user.SiteID != "test-site" {
		t.Errorf("Expected SiteID to be 'test-site', got %s", user.SiteID)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email to be 'test@example.com', got %s", user.Email)
	}
	if user.Auth0Sub != "auth0|12345" {
		t.Errorf("Expected Auth0Sub to be 'auth0|12345', got %s", user.Auth0Sub)
	}
	if user.Name != "Test User" {
		t.Errorf("Expected name to be 'Test User', got %s", user.Name)
	}
	if !user.IsVerified {
		t.Error("Expected IsVerified to be true")
	}

	// Update the same user
	userInfo.Name = "Updated Name"
	updatedUser, err := store.CreateOrUpdateUserFromAuth0("test-site", userInfo)
	if err != nil {
		t.Fatalf("CreateOrUpdateUserFromAuth0 (update) failed: %v", err)
	}

	// Should be same user ID
	if updatedUser.ID != user.ID {
		t.Errorf("Expected same user ID, got %s vs %s", updatedUser.ID, user.ID)
	}
	if updatedUser.Name != "Updated Name" {
		t.Errorf("Expected name to be 'Updated Name', got %s", updatedUser.Name)
	}
}

func TestGetUserByAuth0Sub(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewKotomiAuthStore(db)

	// Create user
	userInfo := &UserInfo{
		Sub:           "auth0|12345",
		Email:         "test@example.com",
		Name:          "Test User",
		EmailVerified: true,
	}
	created, err := store.CreateOrUpdateUserFromAuth0("test-site", userInfo)
	if err != nil {
		t.Fatalf("CreateOrUpdateUserFromAuth0 failed: %v", err)
	}

	// Get user by Auth0 sub
	user, err := store.GetUserByAuth0Sub("test-site", "auth0|12345")
	if err != nil {
		t.Fatalf("GetUserByAuth0Sub failed: %v", err)
	}

	if user.ID != created.ID {
		t.Errorf("Expected user ID %s, got %s", created.ID, user.ID)
	}
	if user.Auth0Sub != "auth0|12345" {
		t.Errorf("Expected Auth0Sub 'auth0|12345', got %s", user.Auth0Sub)
	}

	// Test non-existent user
	_, err = store.GetUserByAuth0Sub("test-site", "nonexistent-sub")
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

func TestGetUserByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewKotomiAuthStore(db)

	// Create user
	userInfo := &UserInfo{
		Sub:           "auth0|12345",
		Email:         "test@example.com",
		Name:          "Test User",
		EmailVerified: true,
	}
	created, err := store.CreateOrUpdateUserFromAuth0("test-site", userInfo)
	if err != nil {
		t.Fatalf("CreateOrUpdateUserFromAuth0 failed: %v", err)
	}

	// Get user by ID
	user, err := store.GetUserByID(created.ID)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}

	if user.ID != created.ID {
		t.Errorf("Expected user ID %s, got %s", created.ID, user.ID)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", user.Email)
	}

	// Test non-existent user
	_, err = store.GetUserByID("nonexistent-id")
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

func TestGenerateJWTToken(t *testing.T) {
	user := &KotomiAuthUser{
		ID:         "user-123",
		SiteID:     "test-site",
		Email:      "test@example.com",
		Auth0Sub:   "auth0|12345",
		Name:       "Test User",
		IsVerified: true,
	}

	secret := "test-secret"
	token, err := GenerateJWTToken(user, "test-site", secret, 60)
	if err != nil {
		t.Fatalf("GenerateJWTToken failed: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	// Verify token has reasonable length for a JWT
	if len(token) < 100 {
		t.Error("Token seems too short to be valid JWT")
	}
}

func TestCreateSession(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewKotomiAuthStore(db)

	// Create user
	userInfo := &UserInfo{
		Sub:           "auth0|12345",
		Email:         "test@example.com",
		Name:          "Test User",
		EmailVerified: true,
	}
	user, err := store.CreateOrUpdateUserFromAuth0("test-site", userInfo)
	if err != nil {
		t.Fatalf("CreateOrUpdateUserFromAuth0 failed: %v", err)
	}

	// Create session
	secret := "test-secret"
	session, err := store.CreateSession(user, secret)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Verify session properties
	if session.ID == "" {
		t.Error("Expected session to have an ID")
	}
	if session.UserID != user.ID {
		t.Errorf("Expected UserID %s, got %s", user.ID, session.UserID)
	}
	if session.SiteID != "test-site" {
		t.Errorf("Expected SiteID 'test-site', got %s", session.SiteID)
	}
	if session.Token == "" {
		t.Error("Expected non-empty token")
	}
	if session.RefreshToken == "" {
		t.Error("Expected non-empty refresh token")
	}
	if session.ExpiresAt.Before(time.Now()) {
		t.Error("Expected ExpiresAt to be in the future")
	}
	if session.RefreshExpiresAt.Before(time.Now()) {
		t.Error("Expected RefreshExpiresAt to be in the future")
	}
}

func TestGetSessionByToken(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewKotomiAuthStore(db)

	// Create user and session
	userInfo := &UserInfo{
		Sub:           "auth0|12345",
		Email:         "test@example.com",
		Name:          "Test User",
		EmailVerified: true,
	}
	user, err := store.CreateOrUpdateUserFromAuth0("test-site", userInfo)
	if err != nil {
		t.Fatalf("CreateOrUpdateUserFromAuth0 failed: %v", err)
	}

	created, err := store.CreateSession(user, "test-secret")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Get session by token
	session, err := store.GetSessionByToken(created.Token)
	if err != nil {
		t.Fatalf("GetSessionByToken failed: %v", err)
	}

	if session.ID != created.ID {
		t.Errorf("Expected session ID %s, got %s", created.ID, session.ID)
	}
	if session.Token != created.Token {
		t.Errorf("Expected token %s, got %s", created.Token, session.Token)
	}

	// Test non-existent token
	_, err = store.GetSessionByToken("nonexistent-token")
	if err == nil {
		t.Error("Expected error for non-existent token")
	}
}

func TestGetSessionByRefreshToken(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewKotomiAuthStore(db)

	// Create user and session
	userInfo := &UserInfo{
		Sub:           "auth0|12345",
		Email:         "test@example.com",
		Name:          "Test User",
		EmailVerified: true,
	}
	user, err := store.CreateOrUpdateUserFromAuth0("test-site", userInfo)
	if err != nil {
		t.Fatalf("CreateOrUpdateUserFromAuth0 failed: %v", err)
	}

	created, err := store.CreateSession(user, "test-secret")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Get session by refresh token
	session, err := store.GetSessionByRefreshToken(created.RefreshToken)
	if err != nil {
		t.Fatalf("GetSessionByRefreshToken failed: %v", err)
	}

	if session.ID != created.ID {
		t.Errorf("Expected session ID %s, got %s", created.ID, session.ID)
	}
	if session.RefreshToken != created.RefreshToken {
		t.Errorf("Expected refresh token %s, got %s", created.RefreshToken, session.RefreshToken)
	}

	// Test non-existent refresh token
	_, err = store.GetSessionByRefreshToken("nonexistent-refresh-token")
	if err == nil {
		t.Error("Expected error for non-existent refresh token")
	}
}

func TestDeleteSession(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewKotomiAuthStore(db)

	// Create user and session
	userInfo := &UserInfo{
		Sub:           "auth0|12345",
		Email:         "test@example.com",
		Name:          "Test User",
		EmailVerified: true,
	}
	user, err := store.CreateOrUpdateUserFromAuth0("test-site", userInfo)
	if err != nil {
		t.Fatalf("CreateOrUpdateUserFromAuth0 failed: %v", err)
	}

	session, err := store.CreateSession(user, "test-secret")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Delete session
	err = store.DeleteSession(session.ID)
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	// Verify session is deleted
	_, err = store.GetSessionByToken(session.Token)
	if err == nil {
		t.Error("Expected error when getting deleted session")
	}
}

func TestDeleteSessionByToken(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewKotomiAuthStore(db)

	// Create user and session
	userInfo := &UserInfo{
		Sub:           "auth0|12345",
		Email:         "test@example.com",
		Name:          "Test User",
		EmailVerified: true,
	}
	user, err := store.CreateOrUpdateUserFromAuth0("test-site", userInfo)
	if err != nil {
		t.Fatalf("CreateOrUpdateUserFromAuth0 failed: %v", err)
	}

	session, err := store.CreateSession(user, "test-secret")
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Delete session by token
	err = store.DeleteSessionByToken(session.Token)
	if err != nil {
		t.Fatalf("DeleteSessionByToken failed: %v", err)
	}

	// Verify session is deleted
	_, err = store.GetSessionByToken(session.Token)
	if err == nil {
		t.Error("Expected error when getting deleted session")
	}
}

func TestUpdateUser(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewKotomiAuthStore(db)

	// Create user
	userInfo := &UserInfo{
		Sub:           "auth0|12345",
		Email:         "test@example.com",
		Name:          "Test User",
		EmailVerified: false,
	}
	user, err := store.CreateOrUpdateUserFromAuth0("test-site", userInfo)
	if err != nil {
		t.Fatalf("CreateOrUpdateUserFromAuth0 failed: %v", err)
	}

	// Update user
	user.Name = "Updated Name"
	user.AvatarURL = "https://example.com/avatar.jpg"
	user.IsVerified = true

	err = store.UpdateUser(user)
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	// Verify updates
	updated, err := store.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got %s", updated.Name)
	}
	if updated.AvatarURL != "https://example.com/avatar.jpg" {
		t.Errorf("Expected avatar URL 'https://example.com/avatar.jpg', got %s", updated.AvatarURL)
	}
	if !updated.IsVerified {
		t.Error("Expected IsVerified to be true")
	}
}

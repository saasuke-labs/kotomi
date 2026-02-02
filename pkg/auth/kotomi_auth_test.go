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

	// Create schema
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
		password_hash TEXT NOT NULL,
		name TEXT,
		avatar_url TEXT,
		is_verified INTEGER DEFAULT 0,
		verification_token TEXT,
		reset_token TEXT,
		reset_token_expires TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE,
		UNIQUE(site_id, email)
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

func TestHashPassword(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("Expected non-empty hash")
	}

	if hash == password {
		t.Error("Hash should not equal plain password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Test correct password
	err = CheckPassword(password, hash)
	if err != nil {
		t.Errorf("CheckPassword failed for correct password: %v", err)
	}

	// Test incorrect password
	err = CheckPassword("wrongpassword", hash)
	if err == nil {
		t.Error("CheckPassword should fail for incorrect password")
	}
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

func TestCreateUser(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewKotomiAuthStore(db)

	// Create user
	user, err := store.CreateUser("test-site", "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
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
	if user.Name != "Test User" {
		t.Errorf("Expected name to be 'Test User', got %s", user.Name)
	}
	if user.IsVerified {
		t.Error("Expected IsVerified to be false by default")
	}
	if user.VerificationToken == "" {
		t.Error("Expected verification token to be generated")
	}

	// Test duplicate email
	_, err = store.CreateUser("test-site", "test@example.com", "password456", "Another User")
	if err == nil {
		t.Error("Expected error when creating user with duplicate email")
	}
}

func TestGetUserByEmail(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewKotomiAuthStore(db)

	// Create user
	created, err := store.CreateUser("test-site", "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Get user by email
	user, err := store.GetUserByEmail("test-site", "test@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}

	if user.ID != created.ID {
		t.Errorf("Expected user ID %s, got %s", created.ID, user.ID)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", user.Email)
	}

	// Test non-existent user
	_, err = store.GetUserByEmail("test-site", "nonexistent@example.com")
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

func TestGetUserByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewKotomiAuthStore(db)

	// Create user
	created, err := store.CreateUser("test-site", "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
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

func TestAuthenticateUser(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewKotomiAuthStore(db)

	// Create user
	_, err := store.CreateUser("test-site", "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Test successful authentication
	user, err := store.AuthenticateUser("test-site", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("AuthenticateUser failed: %v", err)
	}
	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", user.Email)
	}

	// Test wrong password
	_, err = store.AuthenticateUser("test-site", "test@example.com", "wrongpassword")
	if err == nil {
		t.Error("Expected error for wrong password")
	}

	// Test non-existent user
	_, err = store.AuthenticateUser("test-site", "nonexistent@example.com", "password123")
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

func TestGenerateJWTToken(t *testing.T) {
	user := &KotomiAuthUser{
		ID:         "user-123",
		SiteID:     "test-site",
		Email:      "test@example.com",
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

	// Verify token can be parsed
	// (This is a basic check, full validation is done in jwt_validator_test.go)
	if len(token) < 100 {
		t.Error("Token seems too short to be valid JWT")
	}
}

func TestCreateSession(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewKotomiAuthStore(db)

	// Create user
	user, err := store.CreateUser("test-site", "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
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
	user, err := store.CreateUser("test-site", "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
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
	user, err := store.CreateUser("test-site", "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
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
	user, err := store.CreateUser("test-site", "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
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
	user, err := store.CreateUser("test-site", "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
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
	user, err := store.CreateUser("test-site", "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
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

func TestSetPasswordResetToken(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewKotomiAuthStore(db)

	// Create user
	user, err := store.CreateUser("test-site", "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Set reset token
	token := "reset-token-123"
	expiresAt := time.Now().Add(1 * time.Hour)
	err = store.SetPasswordResetToken(user.ID, token, expiresAt)
	if err != nil {
		t.Fatalf("SetPasswordResetToken failed: %v", err)
	}

	// Verify reset token is set
	updated, err := store.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}

	if updated.ResetToken != token {
		t.Errorf("Expected reset token '%s', got '%s'", token, updated.ResetToken)
	}
}

func TestResetPassword(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewKotomiAuthStore(db)

	// Create user
	user, err := store.CreateUser("test-site", "test@example.com", "oldpassword", "Test User")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Set reset token
	token := "reset-token-123"
	expiresAt := time.Now().Add(1 * time.Hour)
	err = store.SetPasswordResetToken(user.ID, token, expiresAt)
	if err != nil {
		t.Fatalf("SetPasswordResetToken failed: %v", err)
	}

	// Reset password
	newPassword := "newpassword123"
	err = store.ResetPassword(token, newPassword)
	if err != nil {
		t.Fatalf("ResetPassword failed: %v", err)
	}

	// Verify new password works
	_, err = store.AuthenticateUser("test-site", "test@example.com", newPassword)
	if err != nil {
		t.Errorf("Authentication with new password failed: %v", err)
	}

	// Verify old password doesn't work
	_, err = store.AuthenticateUser("test-site", "test@example.com", "oldpassword")
	if err == nil {
		t.Error("Expected authentication with old password to fail")
	}

	// Verify reset token is cleared
	updated, err := store.GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}
	if updated.ResetToken != "" {
		t.Error("Expected reset token to be cleared after password reset")
	}
}

func TestResetPassword_ExpiredToken(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewKotomiAuthStore(db)

	// Create user
	user, err := store.CreateUser("test-site", "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Set expired reset token
	token := "reset-token-123"
	expiresAt := time.Now().Add(-1 * time.Hour) // Expired
	err = store.SetPasswordResetToken(user.ID, token, expiresAt)
	if err != nil {
		t.Fatalf("SetPasswordResetToken failed: %v", err)
	}

	// Try to reset password with expired token
	err = store.ResetPassword(token, "newpassword")
	if err == nil {
		t.Error("Expected error when using expired reset token")
	}
}

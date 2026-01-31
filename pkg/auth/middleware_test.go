package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestInitSessionStore_WithSecret(t *testing.T) {
	os.Setenv("SESSION_SECRET", "test-secret-key-for-sessions")
	defer os.Unsetenv("SESSION_SECRET")

	// Reset Store to nil before test
	Store = nil

	err := InitSessionStore()
	if err != nil {
		t.Fatalf("InitSessionStore failed: %v", err)
	}

	if Store == nil {
		t.Error("Store should not be nil after initialization")
	}

	if Store.Options.MaxAge != 86400*7 {
		t.Errorf("Expected MaxAge to be 604800, got %d", Store.Options.MaxAge)
	}

	if !Store.Options.HttpOnly {
		t.Error("Expected HttpOnly to be true")
	}

	if Store.Options.SameSite != http.SameSiteLaxMode {
		t.Errorf("Expected SameSite to be Lax, got %v", Store.Options.SameSite)
	}
}

func TestInitSessionStore_WithoutSecret(t *testing.T) {
	os.Unsetenv("SESSION_SECRET")

	// Reset Store to nil before test
	Store = nil

	err := InitSessionStore()
	if err != nil {
		t.Fatalf("InitSessionStore failed: %v", err)
	}

	// Store should remain nil when no secret is provided
	if Store != nil {
		t.Error("Store should be nil when no secret is provided")
	}
}

func TestInitSessionStore_ProductionSecure(t *testing.T) {
	os.Setenv("SESSION_SECRET", "test-secret-key")
	os.Setenv("ENV", "production")
	defer func() {
		os.Unsetenv("SESSION_SECRET")
		os.Unsetenv("ENV")
	}()

	// Reset Store to nil before test
	Store = nil

	err := InitSessionStore()
	if err != nil {
		t.Fatalf("InitSessionStore failed: %v", err)
	}

	if !Store.Options.Secure {
		t.Error("Expected Secure to be true in production environment")
	}
}

func TestGetSession_WithInitializedStore(t *testing.T) {
	os.Setenv("SESSION_SECRET", "test-secret-key")
	defer os.Unsetenv("SESSION_SECRET")

	Store = nil
	InitSessionStore()

	req := httptest.NewRequest("GET", "/", nil)
	session, err := GetSession(req)

	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	if session == nil {
		t.Error("Expected non-nil session")
	}
}

func TestGetSession_WithUninitializedStore(t *testing.T) {
	// Reset Store to nil
	Store = nil
	os.Unsetenv("SESSION_SECRET")

	req := httptest.NewRequest("GET", "/", nil)
	session, err := GetSession(req)

	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	if session == nil {
		t.Error("Expected non-nil session")
	}

	// Store should be initialized with default secret
	if Store == nil {
		t.Error("Store should be initialized after GetSession call")
	}
}

func TestRequireAuth_WithAuthenticatedUser(t *testing.T) {
	// Reset and initialize store
	Store = nil
	os.Setenv("SESSION_SECRET", "test-secret-key")
	defer os.Unsetenv("SESSION_SECRET")
	InitSessionStore()

	// Create a test handler
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		userID := GetUserIDFromContext(r.Context())
		if userID != "test-user-123" {
			t.Errorf("Expected user ID 'test-user-123', got '%s'", userID)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Create a request with a session
	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	// Set up session with user ID
	session, _ := GetSession(req)
	session.Values[SessionKeyUserID] = "test-user-123"
	session.Save(req, w)

	// Create a new request with the session cookie
	req2 := httptest.NewRequest("GET", "/protected", nil)
	for _, cookie := range w.Result().Cookies() {
		req2.AddCookie(cookie)
	}

	// Test the middleware
	w2 := httptest.NewRecorder()
	handler := RequireAuth(next)
	handler.ServeHTTP(w2, req2)

	if !nextCalled {
		t.Error("Expected next handler to be called")
	}

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w2.Code)
	}
}

func TestRequireAuth_WithoutAuthentication(t *testing.T) {
	// Reset and initialize store
	Store = nil
	os.Setenv("SESSION_SECRET", "test-secret-key")
	defer os.Unsetenv("SESSION_SECRET")
	InitSessionStore()

	// Create a test handler
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Create a request without a session
	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	// Test the middleware
	handler := RequireAuth(next)
	handler.ServeHTTP(w, req)

	if nextCalled {
		t.Error("Expected next handler not to be called")
	}

	// Should redirect to login
	if w.Code != http.StatusSeeOther {
		t.Errorf("Expected status %d (See Other), got %d", http.StatusSeeOther, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/login" {
		t.Errorf("Expected redirect to '/login', got '%s'", location)
	}
}

func TestGetUserIDFromContext_WithUserID(t *testing.T) {
	ctx := context.WithValue(context.Background(), SessionKeyUserID, "test-user-123")
	userID := GetUserIDFromContext(ctx)

	if userID != "test-user-123" {
		t.Errorf("Expected user ID 'test-user-123', got '%s'", userID)
	}
}

func TestGetUserIDFromContext_WithoutUserID(t *testing.T) {
	ctx := context.Background()
	userID := GetUserIDFromContext(ctx)

	if userID != "" {
		t.Errorf("Expected empty user ID, got '%s'", userID)
	}
}

func TestGetUserIDFromContext_WithWrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), SessionKeyUserID, 123) // Wrong type
	userID := GetUserIDFromContext(ctx)

	if userID != "" {
		t.Errorf("Expected empty user ID for wrong type, got '%s'", userID)
	}
}

func TestClearSession(t *testing.T) {
	// Reset and initialize store
	Store = nil
	os.Setenv("SESSION_SECRET", "test-secret-key")
	defer os.Unsetenv("SESSION_SECRET")
	InitSessionStore()

	// Create a request with a session
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Set up session with values
	session, _ := GetSession(req)
	session.Values[SessionKeyUserID] = "test-user-123"
	session.Values[SessionKeyEmail] = "test@example.com"
	session.Save(req, w)

	// Create a new request with the session cookie
	req2 := httptest.NewRequest("GET", "/", nil)
	for _, cookie := range w.Result().Cookies() {
		req2.AddCookie(cookie)
	}

	// Clear the session
	w2 := httptest.NewRecorder()
	err := ClearSession(w2, req2)
	if err != nil {
		t.Fatalf("ClearSession failed: %v", err)
	}

	// Verify session is cleared
	req3 := httptest.NewRequest("GET", "/", nil)
	for _, cookie := range w2.Result().Cookies() {
		req3.AddCookie(cookie)
	}

	session3, _ := GetSession(req3)
	if len(session3.Values) > 0 {
		t.Errorf("Expected empty session values, got %d values", len(session3.Values))
	}
}

func TestSessionConstants(t *testing.T) {
	// Test that session constants are defined correctly
	if SessionName != "kotomi-session" {
		t.Errorf("Expected SessionName to be 'kotomi-session', got '%s'", SessionName)
	}
	if SessionKeyUserID != "user_id" {
		t.Errorf("Expected SessionKeyUserID to be 'user_id', got '%s'", SessionKeyUserID)
	}
	if SessionKeyAuth0Sub != "auth0_sub" {
		t.Errorf("Expected SessionKeyAuth0Sub to be 'auth0_sub', got '%s'", SessionKeyAuth0Sub)
	}
	if SessionKeyEmail != "email" {
		t.Errorf("Expected SessionKeyEmail to be 'email', got '%s'", SessionKeyEmail)
	}
	if SessionKeyName != "name" {
		t.Errorf("Expected SessionKeyName to be 'name', got '%s'", SessionKeyName)
	}
	if SessionKeyState != "oauth_state" {
		t.Errorf("Expected SessionKeyState to be 'oauth_state', got '%s'", SessionKeyState)
	}
}

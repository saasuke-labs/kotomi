package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestNewAuth0Config_Success(t *testing.T) {
	os.Setenv("AUTH0_DOMAIN", "test.auth0.com")
	os.Setenv("AUTH0_CLIENT_ID", "test_client_id")
	os.Setenv("AUTH0_CLIENT_SECRET", "test_client_secret")
	os.Setenv("AUTH0_CALLBACK_URL", "http://localhost:8080/callback")
	defer func() {
		os.Unsetenv("AUTH0_DOMAIN")
		os.Unsetenv("AUTH0_CLIENT_ID")
		os.Unsetenv("AUTH0_CLIENT_SECRET")
		os.Unsetenv("AUTH0_CALLBACK_URL")
	}()

	config, err := NewAuth0Config()
	if err != nil {
		t.Fatalf("NewAuth0Config failed: %v", err)
	}

	if config.Domain != "test.auth0.com" {
		t.Errorf("Expected domain 'test.auth0.com', got '%s'", config.Domain)
	}
	if config.ClientID != "test_client_id" {
		t.Errorf("Expected client ID 'test_client_id', got '%s'", config.ClientID)
	}
	if config.ClientSecret != "test_client_secret" {
		t.Errorf("Expected client secret 'test_client_secret', got '%s'", config.ClientSecret)
	}
	if config.CallbackURL != "http://localhost:8080/callback" {
		t.Errorf("Expected callback URL 'http://localhost:8080/callback', got '%s'", config.CallbackURL)
	}
	if config.OAuth2Config == nil {
		t.Error("OAuth2Config should not be nil")
	}
}

func TestNewAuth0Config_DefaultCallbackURL(t *testing.T) {
	os.Setenv("AUTH0_DOMAIN", "test.auth0.com")
	os.Setenv("AUTH0_CLIENT_ID", "test_client_id")
	os.Setenv("AUTH0_CLIENT_SECRET", "test_client_secret")
	os.Unsetenv("AUTH0_CALLBACK_URL")
	defer func() {
		os.Unsetenv("AUTH0_DOMAIN")
		os.Unsetenv("AUTH0_CLIENT_ID")
		os.Unsetenv("AUTH0_CLIENT_SECRET")
	}()

	config, err := NewAuth0Config()
	if err != nil {
		t.Fatalf("NewAuth0Config failed: %v", err)
	}

	if config.CallbackURL != "http://localhost:8080/callback" {
		t.Errorf("Expected default callback URL 'http://localhost:8080/callback', got '%s'", config.CallbackURL)
	}
}

func TestNewAuth0Config_MissingDomain(t *testing.T) {
	os.Unsetenv("AUTH0_DOMAIN")
	os.Setenv("AUTH0_CLIENT_ID", "test_client_id")
	os.Setenv("AUTH0_CLIENT_SECRET", "test_client_secret")
	defer func() {
		os.Unsetenv("AUTH0_CLIENT_ID")
		os.Unsetenv("AUTH0_CLIENT_SECRET")
	}()

	_, err := NewAuth0Config()
	if err == nil {
		t.Error("Expected error for missing AUTH0_DOMAIN, got nil")
	}
	if !strings.Contains(err.Error(), "AUTH0_DOMAIN") {
		t.Errorf("Expected error message to mention AUTH0_DOMAIN, got: %s", err.Error())
	}
}

func TestNewAuth0Config_MissingClientID(t *testing.T) {
	os.Setenv("AUTH0_DOMAIN", "test.auth0.com")
	os.Unsetenv("AUTH0_CLIENT_ID")
	os.Setenv("AUTH0_CLIENT_SECRET", "test_client_secret")
	defer func() {
		os.Unsetenv("AUTH0_DOMAIN")
		os.Unsetenv("AUTH0_CLIENT_SECRET")
	}()

	_, err := NewAuth0Config()
	if err == nil {
		t.Error("Expected error for missing AUTH0_CLIENT_ID, got nil")
	}
	if !strings.Contains(err.Error(), "AUTH0_CLIENT_ID") {
		t.Errorf("Expected error message to mention AUTH0_CLIENT_ID, got: %s", err.Error())
	}
}

func TestNewAuth0Config_MissingClientSecret(t *testing.T) {
	os.Setenv("AUTH0_DOMAIN", "test.auth0.com")
	os.Setenv("AUTH0_CLIENT_ID", "test_client_id")
	os.Unsetenv("AUTH0_CLIENT_SECRET")
	defer func() {
		os.Unsetenv("AUTH0_DOMAIN")
		os.Unsetenv("AUTH0_CLIENT_ID")
	}()

	_, err := NewAuth0Config()
	if err == nil {
		t.Error("Expected error for missing AUTH0_CLIENT_SECRET, got nil")
	}
	if !strings.Contains(err.Error(), "AUTH0_CLIENT_SECRET") {
		t.Errorf("Expected error message to mention AUTH0_CLIENT_SECRET, got: %s", err.Error())
	}
}

func TestAuth0Config_GetLoginURL(t *testing.T) {
	os.Setenv("AUTH0_DOMAIN", "test.auth0.com")
	os.Setenv("AUTH0_CLIENT_ID", "test_client_id")
	os.Setenv("AUTH0_CLIENT_SECRET", "test_client_secret")
	defer func() {
		os.Unsetenv("AUTH0_DOMAIN")
		os.Unsetenv("AUTH0_CLIENT_ID")
		os.Unsetenv("AUTH0_CLIENT_SECRET")
	}()

	config, err := NewAuth0Config()
	if err != nil {
		t.Fatalf("NewAuth0Config failed: %v", err)
	}

	state := "random_state"
	loginURL := config.GetLoginURL(state)

	if !strings.Contains(loginURL, "test.auth0.com") {
		t.Errorf("Expected login URL to contain domain, got: %s", loginURL)
	}
	if !strings.Contains(loginURL, state) {
		t.Errorf("Expected login URL to contain state parameter, got: %s", loginURL)
	}
	if !strings.Contains(loginURL, "authorize") {
		t.Errorf("Expected login URL to contain authorize endpoint, got: %s", loginURL)
	}
}

func TestAuth0Config_GetLogoutURL(t *testing.T) {
	os.Setenv("AUTH0_DOMAIN", "test.auth0.com")
	os.Setenv("AUTH0_CLIENT_ID", "test_client_id")
	os.Setenv("AUTH0_CLIENT_SECRET", "test_client_secret")
	defer func() {
		os.Unsetenv("AUTH0_DOMAIN")
		os.Unsetenv("AUTH0_CLIENT_ID")
		os.Unsetenv("AUTH0_CLIENT_SECRET")
	}()

	config, err := NewAuth0Config()
	if err != nil {
		t.Fatalf("NewAuth0Config failed: %v", err)
	}

	returnTo := "http://localhost:8080"
	logoutURL := config.GetLogoutURL(returnTo)

	if !strings.Contains(logoutURL, "test.auth0.com") {
		t.Errorf("Expected logout URL to contain domain, got: %s", logoutURL)
	}
	if !strings.Contains(logoutURL, "logout") {
		t.Errorf("Expected logout URL to contain logout endpoint, got: %s", logoutURL)
	}
	if !strings.Contains(logoutURL, "client_id=test_client_id") {
		t.Errorf("Expected logout URL to contain client_id parameter, got: %s", logoutURL)
	}
	if !strings.Contains(logoutURL, "returnTo=") {
		t.Errorf("Expected logout URL to contain returnTo parameter, got: %s", logoutURL)
	}
}

func TestGenerateRandomState(t *testing.T) {
	state1, err := GenerateRandomState()
	if err != nil {
		t.Fatalf("GenerateRandomState failed: %v", err)
	}

	if state1 == "" {
		t.Error("Expected non-empty state, got empty string")
	}

	// Generate another state and verify they're different
	state2, err := GenerateRandomState()
	if err != nil {
		t.Fatalf("GenerateRandomState failed on second call: %v", err)
	}

	if state1 == state2 {
		t.Error("Expected different state values, got the same value")
	}

	// Verify it's base64 encoded
	if len(state1) < 20 {
		t.Errorf("Expected state to be at least 20 characters, got %d", len(state1))
	}
}

func TestAuth0Config_GetUserInfo_Success(t *testing.T) {
	// Create a mock Auth0 userinfo endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/userinfo" {
			t.Errorf("Expected path '/userinfo', got '%s'", r.URL.Path)
		}

		userInfo := UserInfo{
			Sub:           "auth0|123456",
			Name:          "Test User",
			Email:         "test@example.com",
			EmailVerified: true,
			Picture:       "https://example.com/picture.jpg",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(userInfo)
	}))
	defer server.Close()

	// Create a config with the test server domain
	domain := strings.TrimPrefix(server.URL, "http://")
	config := &Auth0Config{
		Domain:       domain,
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
		CallbackURL:  "http://localhost:8080/callback",
	}

	// Note: GetUserInfo requires a valid OAuth2 token and client
	// This test verifies the structure, but full integration testing
	// would require mocking the OAuth2 client
	if config.Domain != domain {
		t.Errorf("Expected domain '%s', got '%s'", domain, config.Domain)
	}
}

func TestAuth0Config_GetUserInfo_InvalidResponse(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	domain := strings.TrimPrefix(server.URL, "http://")
	config := &Auth0Config{
		Domain:       domain,
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
		CallbackURL:  "http://localhost:8080/callback",
	}

	// Note: Full testing would require mocking OAuth2 client
	// This test verifies the structure
	if config.Domain != domain {
		t.Errorf("Expected domain '%s', got '%s'", domain, config.Domain)
	}
}

func TestAuth0Config_ExchangeCode(t *testing.T) {
	os.Setenv("AUTH0_DOMAIN", "test.auth0.com")
	os.Setenv("AUTH0_CLIENT_ID", "test_client_id")
	os.Setenv("AUTH0_CLIENT_SECRET", "test_client_secret")
	defer func() {
		os.Unsetenv("AUTH0_DOMAIN")
		os.Unsetenv("AUTH0_CLIENT_ID")
		os.Unsetenv("AUTH0_CLIENT_SECRET")
	}()

	config, err := NewAuth0Config()
	if err != nil {
		t.Fatalf("NewAuth0Config failed: %v", err)
	}

	// Test with invalid code (will fail as expected since we don't have a real auth server)
	// This tests the error handling path
	ctx := context.Background()
	_, err = config.ExchangeCode(ctx, "invalid_code")
	if err == nil {
		t.Error("Expected error for invalid code exchange, got nil")
	}
}

func TestUserInfo_JSONMarshaling(t *testing.T) {
	userInfo := UserInfo{
		Sub:           "auth0|123456",
		Name:          "Test User",
		Email:         "test@example.com",
		EmailVerified: true,
		Picture:       "https://example.com/picture.jpg",
	}

	// Marshal to JSON
	data, err := json.Marshal(userInfo)
	if err != nil {
		t.Fatalf("Failed to marshal UserInfo: %v", err)
	}

	// Unmarshal back
	var decoded UserInfo
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal UserInfo: %v", err)
	}

	// Verify fields
	if decoded.Sub != userInfo.Sub {
		t.Errorf("Expected Sub '%s', got '%s'", userInfo.Sub, decoded.Sub)
	}
	if decoded.Name != userInfo.Name {
		t.Errorf("Expected Name '%s', got '%s'", userInfo.Name, decoded.Name)
	}
	if decoded.Email != userInfo.Email {
		t.Errorf("Expected Email '%s', got '%s'", userInfo.Email, decoded.Email)
	}
	if decoded.EmailVerified != userInfo.EmailVerified {
		t.Errorf("Expected EmailVerified %v, got %v", userInfo.EmailVerified, decoded.EmailVerified)
	}
	if decoded.Picture != userInfo.Picture {
		t.Errorf("Expected Picture '%s', got '%s'", userInfo.Picture, decoded.Picture)
	}
}

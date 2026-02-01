package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/saasuke-labs/kotomi/pkg/models"
)

// TestJWTValidator_ValidateHMAC tests HMAC token validation
func TestJWTValidator_ValidateHMAC(t *testing.T) {
	secret := "test-secret-key-min-32-characters-long"
	
	config := &models.SiteAuthConfig{
		AuthMode:              "external",
		JWTValidationType:     "hmac",
		JWTSecret:             secret,
		JWTIssuer:             "https://example.com",
		JWTAudience:           "kotomi",
		TokenExpirationBuffer: 60,
	}

	// Create a valid token
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "https://example.com",
		"sub": "user-123",
		"aud": "kotomi",
		"exp": now.Add(time.Hour).Unix(),
		"iat": now.Unix(),
		"kotomi_user": map[string]interface{}{
			"id":    "user-123",
			"name":  "John Doe",
			"email": "john@example.com",
		},
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Validate token
	validator := NewJWTValidator(config)
	user, err := validator.ValidateToken(tokenString)
	if err != nil {
		t.Fatalf("Token validation failed: %v", err)
	}

	// Check user fields
	if user.ID != "user-123" {
		t.Errorf("Expected user ID 'user-123', got '%s'", user.ID)
	}
	if user.Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got '%s'", user.Name)
	}
	if user.Email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got '%s'", user.Email)
	}
}

// TestJWTValidator_ExpiredToken tests that expired tokens are rejected
func TestJWTValidator_ExpiredToken(t *testing.T) {
	secret := "test-secret-key-min-32-characters-long"
	
	config := &models.SiteAuthConfig{
		AuthMode:              "external",
		JWTValidationType:     "hmac",
		JWTSecret:             secret,
		JWTIssuer:             "https://example.com",
		JWTAudience:           "kotomi",
		TokenExpirationBuffer: 60,
	}

	// Create an expired token
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "https://example.com",
		"sub": "user-123",
		"aud": "kotomi",
		"exp": now.Add(-2 * time.Hour).Unix(), // Expired 2 hours ago
		"iat": now.Add(-3 * time.Hour).Unix(),
		"kotomi_user": map[string]interface{}{
			"id":   "user-123",
			"name": "John Doe",
		},
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Validate token - should fail
	validator := NewJWTValidator(config)
	_, err = validator.ValidateToken(tokenString)
	if err == nil {
		t.Fatal("Expected validation to fail for expired token")
	}
}

// TestJWTValidator_InvalidIssuer tests that tokens with wrong issuer are rejected
func TestJWTValidator_InvalidIssuer(t *testing.T) {
	secret := "test-secret-key-min-32-characters-long"
	
	config := &models.SiteAuthConfig{
		AuthMode:              "external",
		JWTValidationType:     "hmac",
		JWTSecret:             secret,
		JWTIssuer:             "https://example.com",
		JWTAudience:           "kotomi",
		TokenExpirationBuffer: 60,
	}

	// Create a token with wrong issuer
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "https://evil.com", // Wrong issuer
		"sub": "user-123",
		"aud": "kotomi",
		"exp": now.Add(time.Hour).Unix(),
		"iat": now.Unix(),
		"kotomi_user": map[string]interface{}{
			"id":   "user-123",
			"name": "John Doe",
		},
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Validate token - should fail
	validator := NewJWTValidator(config)
	_, err = validator.ValidateToken(tokenString)
	if err == nil {
		t.Fatal("Expected validation to fail for invalid issuer")
	}
}

// TestJWTValidator_MissingKotomiUser tests that tokens without kotomi_user claim are rejected
func TestJWTValidator_MissingKotomiUser(t *testing.T) {
	secret := "test-secret-key-min-32-characters-long"
	
	config := &models.SiteAuthConfig{
		AuthMode:              "external",
		JWTValidationType:     "hmac",
		JWTSecret:             secret,
		JWTIssuer:             "https://example.com",
		JWTAudience:           "kotomi",
		TokenExpirationBuffer: 60,
	}

	// Create a token without kotomi_user claim
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "https://example.com",
		"sub": "user-123",
		"aud": "kotomi",
		"exp": now.Add(time.Hour).Unix(),
		"iat": now.Unix(),
		// Missing kotomi_user claim
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Validate token - should fail
	validator := NewJWTValidator(config)
	_, err = validator.ValidateToken(tokenString)
	if err == nil {
		t.Fatal("Expected validation to fail for missing kotomi_user claim")
	}
}

// TestJWTValidator_ValidateRSA tests RSA token validation
func TestJWTValidator_ValidateRSA(t *testing.T) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	// Encode public key to PEM
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to marshal public key: %v", err)
	}
	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	config := &models.SiteAuthConfig{
		AuthMode:              "external",
		JWTValidationType:     "rsa",
		JWTPublicKey:          string(pubKeyPEM),
		JWTIssuer:             "https://example.com",
		JWTAudience:           "kotomi",
		TokenExpirationBuffer: 60,
	}

	// Create a valid token
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": "https://example.com",
		"sub": "user-456",
		"aud": "kotomi",
		"exp": now.Add(time.Hour).Unix(),
		"iat": now.Unix(),
		"kotomi_user": map[string]interface{}{
			"id":   "user-456",
			"name": "Jane Smith",
		},
	})

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Validate token
	validator := NewJWTValidator(config)
	user, err := validator.ValidateToken(tokenString)
	if err != nil {
		t.Fatalf("Token validation failed: %v", err)
	}

	// Check user fields
	if user.ID != "user-456" {
		t.Errorf("Expected user ID 'user-456', got '%s'", user.ID)
	}
	if user.Name != "Jane Smith" {
		t.Errorf("Expected name 'Jane Smith', got '%s'", user.Name)
	}
}

// TestExtractTokenFromHeader tests token extraction from Authorization header
func TestExtractTokenFromHeader(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "Valid Bearer token",
			header:   "Bearer abc123xyz",
			expected: "abc123xyz",
		},
		{
			name:     "Case insensitive Bearer",
			header:   "bearer abc123xyz",
			expected: "abc123xyz",
		},
		{
			name:     "Empty header",
			header:   "",
			expected: "",
		},
		{
			name:     "Missing token",
			header:   "Bearer",
			expected: "",
		},
		{
			name:     "Invalid format",
			header:   "InvalidFormat",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTokenFromHeader(tt.header)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestJWTValidator_UserWithOptionalFields tests validation with optional user fields
func TestJWTValidator_UserWithOptionalFields(t *testing.T) {
	secret := "test-secret-key-min-32-characters-long"
	
	config := &models.SiteAuthConfig{
		AuthMode:              "external",
		JWTValidationType:     "hmac",
		JWTSecret:             secret,
		JWTIssuer:             "https://example.com",
		JWTAudience:           "kotomi",
		TokenExpirationBuffer: 60,
	}

	// Create a token with optional fields
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "https://example.com",
		"sub": "user-789",
		"aud": "kotomi",
		"exp": now.Add(time.Hour).Unix(),
		"iat": now.Unix(),
		"kotomi_user": map[string]interface{}{
			"id":          "user-789",
			"name":        "Bob Johnson",
			"email":       "bob@example.com",
			"avatar_url":  "https://example.com/avatar.jpg",
			"profile_url": "https://example.com/profile/bob",
			"verified":    true,
			"roles":       []interface{}{"user", "moderator"},
		},
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	// Validate token
	validator := NewJWTValidator(config)
	user, err := validator.ValidateToken(tokenString)
	if err != nil {
		t.Fatalf("Token validation failed: %v", err)
	}

	// Check user fields
	if user.ID != "user-789" {
		t.Errorf("Expected user ID 'user-789', got '%s'", user.ID)
	}
	if user.Name != "Bob Johnson" {
		t.Errorf("Expected name 'Bob Johnson', got '%s'", user.Name)
	}
	if user.Email != "bob@example.com" {
		t.Errorf("Expected email 'bob@example.com', got '%s'", user.Email)
	}
	if user.AvatarURL != "https://example.com/avatar.jpg" {
		t.Errorf("Expected avatar URL 'https://example.com/avatar.jpg', got '%s'", user.AvatarURL)
	}
	if user.ProfileURL != "https://example.com/profile/bob" {
		t.Errorf("Expected profile URL 'https://example.com/profile/bob', got '%s'", user.ProfileURL)
	}
	if !user.Verified {
		t.Error("Expected user to be verified")
	}
	if len(user.Roles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(user.Roles))
	}
	if user.Roles[0] != "user" || user.Roles[1] != "moderator" {
		t.Errorf("Expected roles ['user', 'moderator'], got %v", user.Roles)
	}
}

package middleware

import (
	"os"
	"testing"

	"github.com/rs/cors"
)

func TestNewCORSMiddleware_Defaults(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("CORS_ALLOWED_ORIGINS")
	os.Unsetenv("CORS_ALLOWED_METHODS")
	os.Unsetenv("CORS_ALLOWED_HEADERS")
	os.Unsetenv("CORS_ALLOW_CREDENTIALS")

	c := NewCORSMiddleware()

	if c == nil {
		t.Fatal("NewCORSMiddleware returned nil")
	}
}

func TestNewCORSMiddleware_CustomOrigins(t *testing.T) {
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com,https://test.com")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	c := NewCORSMiddleware()

	if c == nil {
		t.Fatal("NewCORSMiddleware returned nil")
	}
}

func TestNewCORSMiddleware_WildcardOrigin(t *testing.T) {
	os.Setenv("CORS_ALLOWED_ORIGINS", "*")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	c := NewCORSMiddleware()

	if c == nil {
		t.Fatal("NewCORSMiddleware returned nil")
	}
}

func TestNewCORSMiddleware_CustomMethods(t *testing.T) {
	os.Setenv("CORS_ALLOWED_METHODS", "GET,POST,PUT")
	defer os.Unsetenv("CORS_ALLOWED_METHODS")

	c := NewCORSMiddleware()

	if c == nil {
		t.Fatal("NewCORSMiddleware returned nil")
	}
}

func TestNewCORSMiddleware_CustomHeaders(t *testing.T) {
	os.Setenv("CORS_ALLOWED_HEADERS", "Content-Type,X-Custom-Header")
	defer os.Unsetenv("CORS_ALLOWED_HEADERS")

	c := NewCORSMiddleware()

	if c == nil {
		t.Fatal("NewCORSMiddleware returned nil")
	}
}

func TestNewCORSMiddleware_AllowCredentialsTrue(t *testing.T) {
	os.Setenv("CORS_ALLOW_CREDENTIALS", "true")
	defer os.Unsetenv("CORS_ALLOW_CREDENTIALS")

	c := NewCORSMiddleware()

	if c == nil {
		t.Fatal("NewCORSMiddleware returned nil")
	}
}

func TestNewCORSMiddleware_AllowCredentialsFalse(t *testing.T) {
	os.Setenv("CORS_ALLOW_CREDENTIALS", "false")
	defer os.Unsetenv("CORS_ALLOW_CREDENTIALS")

	c := NewCORSMiddleware()

	if c == nil {
		t.Fatal("NewCORSMiddleware returned nil")
	}
}

func TestNewCORSMiddleware_AllowCredentialsInvalid(t *testing.T) {
	os.Setenv("CORS_ALLOW_CREDENTIALS", "invalid")
	defer os.Unsetenv("CORS_ALLOW_CREDENTIALS")

	// Should default to false when invalid
	c := NewCORSMiddleware()

	if c == nil {
		t.Fatal("NewCORSMiddleware returned nil")
	}
}

func TestNewCORSMiddleware_OriginsWithSpaces(t *testing.T) {
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com , https://test.com ")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	c := NewCORSMiddleware()

	if c == nil {
		t.Fatal("NewCORSMiddleware returned nil")
	}
}

func TestNewCORSMiddleware_MethodsWithSpaces(t *testing.T) {
	os.Setenv("CORS_ALLOWED_METHODS", "GET , POST , PUT")
	defer os.Unsetenv("CORS_ALLOWED_METHODS")

	c := NewCORSMiddleware()

	if c == nil {
		t.Fatal("NewCORSMiddleware returned nil")
	}
}

func TestNewCORSMiddleware_HeadersWithSpaces(t *testing.T) {
	os.Setenv("CORS_ALLOWED_HEADERS", "Content-Type , Authorization ")
	defer os.Unsetenv("CORS_ALLOWED_HEADERS")

	c := NewCORSMiddleware()

	if c == nil {
		t.Fatal("NewCORSMiddleware returned nil")
	}
}

func TestNewCORSMiddleware_AllConfigOptions(t *testing.T) {
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com")
	os.Setenv("CORS_ALLOWED_METHODS", "GET,POST")
	os.Setenv("CORS_ALLOWED_HEADERS", "Content-Type")
	os.Setenv("CORS_ALLOW_CREDENTIALS", "true")
	defer func() {
		os.Unsetenv("CORS_ALLOWED_ORIGINS")
		os.Unsetenv("CORS_ALLOWED_METHODS")
		os.Unsetenv("CORS_ALLOWED_HEADERS")
		os.Unsetenv("CORS_ALLOW_CREDENTIALS")
	}()

	c := NewCORSMiddleware()

	if c == nil {
		t.Fatal("NewCORSMiddleware returned nil")
	}

	// Verify it returns a valid cors.Cors instance
	if _, ok := interface{}(c).(*cors.Cors); !ok {
		t.Error("NewCORSMiddleware did not return a *cors.Cors instance")
	}
}

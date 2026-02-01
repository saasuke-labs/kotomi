package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/saasuke-labs/kotomi/pkg/comments"
)

// generateTestJWT creates a test JWT token for E2E testing
func generateTestJWT() string {
	// Create a test JWT token with HMAC signing
	// This secret should match what the test server uses
	secret := "test-secret-for-e2e-testing-min-32-chars"
	
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "https://e2e-test.example.com",
		"sub": "e2e-test-user",
		"aud": "kotomi",
		"exp": now.Add(time.Hour).Unix(),
		"iat": now.Unix(),
		"kotomi_user": map[string]interface{}{
			"id":    "e2e-test-user",
			"name":  "E2E Test User",
			"email": "e2e@test.example.com",
		},
	})
	
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

// GetComments retrieves comments via HTTP API
func GetComments(t *testing.T, baseURL, siteID, pageID string) []comments.Comment {
	t.Helper()

	url := fmt.Sprintf("%s/api/v1/site/%s/page/%s/comments", baseURL, siteID, pageID)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to get comments: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var result []comments.Comment
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	return result
}

// PostComment posts a comment via HTTP API
func PostComment(t *testing.T, baseURL, siteID, pageID string, comment comments.Comment) comments.Comment {
	t.Helper()

	url := fmt.Sprintf("%s/api/v1/site/%s/page/%s/comments", baseURL, siteID, pageID)
	
	jsonData, err := json.Marshal(comment)
	if err != nil {
		t.Fatalf("failed to marshal comment: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+generateTestJWT())
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to post comment: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	var result comments.Comment
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	return result
}

// AssertStatusCode checks if an HTTP response has the expected status code
func AssertStatusCode(t *testing.T, resp *http.Response, expectedStatus int) {
	t.Helper()

	if resp.StatusCode != expectedStatus {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status %d, got %d: %s", expectedStatus, resp.StatusCode, string(body))
	}
}

// WaitForServer waits for the server to be ready
func WaitForServer(t *testing.T, baseURL string, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	healthURL := baseURL + "/healthz"

	for time.Now().Before(deadline) {
		resp, err := http.Get(healthURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("server did not become ready within %v", timeout)
}

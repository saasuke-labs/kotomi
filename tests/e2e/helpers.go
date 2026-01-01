package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/saasuke-labs/kotomi/pkg/comments"
)

// TestServer holds the server configuration for E2E tests
type TestServer struct {
	BaseURL string
	DBPath  string
}

// SetupTestServer initializes a test server configuration
func SetupTestServer(t *testing.T) *TestServer {
	t.Helper()

	// Create temporary database for testing
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "kotomi_test.db")

	// Get test server URL from environment or use default
	baseURL := os.Getenv("TEST_SERVER_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8888"
	}

	return &TestServer{
		BaseURL: baseURL,
		DBPath:  dbPath,
	}
}

// SetupBrowser initializes a headless browser for testing
func SetupBrowser(t *testing.T) (*rod.Browser, context.CancelFunc) {
	t.Helper()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	// Launch browser in headless mode
	l := launcher.New().
		Headless(true).
		NoSandbox(true)

	url, err := l.Launch()
	if err != nil {
		cancel()
		t.Fatalf("failed to launch browser: %v", err)
	}

	browser := rod.New().
		Context(ctx).
		ControlURL(url).
		MustConnect()

	// Return both browser and cancel function
	return browser, cancel
}

// NavigateTo navigates the page to a specific path
func NavigateTo(t *testing.T, page *rod.Page, baseURL, path string) {
	t.Helper()

	fullURL := baseURL + path
	err := page.Navigate(fullURL)
	if err != nil {
		t.Fatalf("failed to navigate to %s: %v", fullURL, err)
	}
	err = page.WaitLoad()
	if err != nil {
		t.Fatalf("failed to wait for page load at %s: %v", fullURL, err)
	}
}

// CreateTestComment creates a test comment in the database
func CreateTestComment(siteID, pageID string, comment comments.Comment, dbPath string) error {
	store, err := comments.NewSQLiteStore(dbPath)
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}
	defer store.Close()

	if err := store.AddPageComment(siteID, pageID, comment); err != nil {
		return fmt.Errorf("failed to add comment: %w", err)
	}

	return nil
}

// GetComments retrieves comments via HTTP API
func GetComments(t *testing.T, baseURL, siteID, pageID string) []comments.Comment {
	t.Helper()

	url := fmt.Sprintf("%s/api/site/%s/page/%s/comments", baseURL, siteID, pageID)
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

	url := fmt.Sprintf("%s/api/site/%s/page/%s/comments", baseURL, siteID, pageID)
	
	jsonData, err := json.Marshal(comment)
	if err != nil {
		t.Fatalf("failed to marshal comment: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
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

// AssertElementExists checks if an element exists on the page
func AssertElementExists(t *testing.T, page *rod.Page, selector string) {
	t.Helper()

	has, _, err := page.Has(selector)
	if err != nil {
		t.Fatalf("error checking for element %s: %v", selector, err)
	}
	if !has {
		t.Fatalf("expected element %s to exist, but it was not found", selector)
	}
}

// AssertElementText checks if an element contains expected text
func AssertElementText(t *testing.T, page *rod.Page, selector, expectedText string) {
	t.Helper()

	element, err := page.Element(selector)
	if err != nil {
		t.Fatalf("failed to find element %s: %v", selector, err)
	}

	text, err := element.Text()
	if err != nil {
		t.Fatalf("failed to get text from element %s: %v", selector, err)
	}

	if text != expectedText {
		t.Fatalf("expected element %s to contain '%s', but got '%s'", selector, expectedText, text)
	}
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

package e2e

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/saasuke-labs/kotomi/pkg/comments"
	"github.com/saasuke-labs/kotomi/pkg/models"
)

var (
	testServerCmd *exec.Cmd
	testDBPath    string
	testLogFile   *os.File
	testPort      = "8888"
	testBaseURL   = "http://localhost:8888"
)

// TestMain sets up and tears down the test environment
func TestMain(m *testing.M) {
	// Skip E2E tests if not in test mode
	if os.Getenv("RUN_E2E_TESTS") == "" {
		log.Println("Skipping E2E tests. Set RUN_E2E_TESTS=true to run them.")
		os.Exit(0)
	}

	// Setup
	if err := setupTestEnvironment(); err != nil {
		log.Fatalf("Failed to setup test environment: %v", err)
	}

	// Run tests
	code := m.Run()

	// Teardown
	teardownTestEnvironment()

	os.Exit(code)
}

func setupTestEnvironment() error {
	// Create temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "kotomi-e2e-*")
	if err != nil {
		return err
	}
	testDBPath = filepath.Join(tmpDir, "kotomi_test.db")

	// Get the project root directory (go up from tests/e2e/ to module root)
	// This works because tests are always run from within the module
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	projectRoot := filepath.Join(wd, "../..")

	// Create log file for server output
	logFile, err := os.Create(filepath.Join(tmpDir, "server.log"))
	if err != nil {
		return err
	}
	testLogFile = logFile

	// Start the server using go run
	// We use 'go run' instead of pre-building to avoid managing build artifacts
	// and to ensure we're always testing the current source code
	testServerCmd = exec.Command("go", "run", "cmd/main.go")
	testServerCmd.Dir = projectRoot
	testServerCmd.Env = append(os.Environ(),
		"PORT="+testPort,
		"DB_PATH="+testDBPath,
		"TEST_MODE=true",
		"RATE_LIMIT_GET=1000",  // High limit for E2E testing
		"RATE_LIMIT_POST=1000", // High limit for E2E testing
	)
	
	// Always log to file for debugging
	testServerCmd.Stdout = logFile
	testServerCmd.Stderr = logFile

	if err = testServerCmd.Start(); err != nil {
		logFile.Close()
		return err
	}

	log.Printf("Started test server with PID %d, DB at %s, logs at %s", 
		testServerCmd.Process.Pid, testDBPath, logFile.Name())

	// Wait for server to be ready
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	healthURL := testBaseURL + "/healthz"
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			resp, httpErr := http.Get(healthURL)
			if httpErr == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				log.Println("Test server is ready")
				
				// Seed the database with auth configurations
				seedAuthConfigurations(testDBPath)
				
				return nil
			}
			if resp != nil {
				resp.Body.Close()
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func teardownTestEnvironment() {
	if testServerCmd != nil && testServerCmd.Process != nil {
		log.Println("Stopping test server...")
		
		// Try graceful shutdown first with SIGINT (os.Interrupt)
		// Note: On Unix, this sends SIGINT; on Windows, it generates a Ctrl+C event
		if err := testServerCmd.Process.Signal(os.Interrupt); err != nil {
			// If signal fails, fall back to Kill (SIGKILL)
			testServerCmd.Process.Kill()
		}
		
		// Wait for process to exit with timeout
		done := make(chan error, 1)
		go func() {
			done <- testServerCmd.Wait()
		}()
		
		select {
		case <-time.After(5 * time.Second):
			// Timeout - force kill
			log.Println("Server did not stop gracefully, forcing kill")
			testServerCmd.Process.Kill()
			testServerCmd.Wait()
		case <-done:
			// Process exited successfully
		}
	}

	// Close log file
	if testLogFile != nil {
		testLogFile.Close()
	}

	// Clean up test database
	if testDBPath != "" {
		os.Remove(testDBPath)
		os.RemoveAll(filepath.Dir(testDBPath))
	}
}

// seedAuthConfigurations seeds auth configurations for E2E test sites
// This is called during setup and should not fail the setup if errors occur
func seedAuthConfigurations(dbPath string) {
	store, err := comments.NewSQLiteStore(dbPath)
	if err != nil {
		log.Printf("Warning: failed to open store for seeding auth configs: %v", err)
		return
	}
	defer store.Close()
	
	db := store.GetDB()
	
	// Create auth configurations for all potential test sites
	// These match the JWT secret used in helpers.go
	authConfigStore := models.NewSiteAuthConfigStore(db)
	
	// List of all site IDs that might be used in E2E tests
	testSites := []string{
		"e2e-site-1", "e2e-site-2", "test-site-1",
		"reactions-site-1", "reaction-isolation-1", "reaction-isolation-2",
		"multiple-reactions-site", "remove-reaction-site",
		// Add more generic patterns for dynamically created sites
		"site-1", "site-2", "isolation-site-1", "isolation-site-2",
	}
	
	for _, siteID := range testSites {
		authConfig := &models.SiteAuthConfig{
			SiteID:                siteID,
			AuthMode:              "external",
			JWTValidationType:     "hmac",
			JWTSecret:             "test-secret-for-e2e-testing-min-32-chars",
			JWTIssuer:             "https://e2e-test.example.com",
			JWTAudience:           "kotomi",
			TokenExpirationBuffer: 60,
		}
		
		// Try to create, ignore errors if already exists
		if err := authConfigStore.Create(authConfig); err != nil {
			// Silently ignore - config might already exist or site might not exist yet
			log.Printf("Debug: Could not create auth config for %s (this is normal): %v", siteID, err)
		}
	}
	
	log.Println("Auth configurations seeded for E2E tests")
}

// SeedTestData seeds the database with test data
func SeedTestData(t *testing.T, dbPath string) {
	t.Helper()

	store, err := comments.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()
	
	db := store.GetDB()
	
	// Create auth configurations for test sites
	authConfigStore := models.NewSiteAuthConfigStore(db)
	testSites := []string{"e2e-site-1", "e2e-site-2", "test-site-1"}
	
	for _, siteID := range testSites {
		authConfig := &models.SiteAuthConfig{
			SiteID:                siteID,
			AuthMode:              "external",
			JWTValidationType:     "hmac",
			JWTSecret:             "test-secret-for-e2e-testing-min-32-chars",
			JWTIssuer:             "https://e2e-test.example.com",
			JWTAudience:           "kotomi",
			TokenExpirationBuffer: 60,
		}
		
		// Check if auth config already exists before creating
		existing, _ := authConfigStore.GetBySiteID(siteID)
		if existing == nil {
			if err := authConfigStore.Create(authConfig); err != nil {
				// Ignore errors if auth config already exists
				log.Printf("Warning: failed to create auth config for %s: %v", siteID, err)
			}
		}
	}

	// Add some test comments
	testComments := []struct {
		siteID  string
		pageID  string
		comment comments.Comment
	}{
		{
			siteID: "test-site-1",
			pageID: "test-page-1",
			comment: comments.Comment{
				ID:        "comment-1",
				Author:    "Alice",
				AuthorID:  "alice-123",
				Text:      "This is a test comment",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		{
			siteID: "test-site-1",
			pageID: "test-page-1",
			comment: comments.Comment{
				ID:        "comment-2",
				Author:    "Bob",
				AuthorID:  "bob-456",
				Text:      "This is another test comment",
				ParentID:  "comment-1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
	}

	for _, tc := range testComments {
		if err := store.AddPageComment(tc.siteID, tc.pageID, tc.comment); err != nil {
			t.Fatalf("failed to seed test data: %v", err)
		}
	}
}

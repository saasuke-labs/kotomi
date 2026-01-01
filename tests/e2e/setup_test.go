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

// SeedTestData seeds the database with test data
func SeedTestData(t *testing.T, dbPath string) {
	t.Helper()

	store, err := comments.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

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

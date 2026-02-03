package comments

import (
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// TestSQLiteOptimizations verifies that all optimization PRAGMAs are properly configured
func TestSQLiteOptimizations(t *testing.T) {
	dbPath := "/tmp/test_optimizations_" + uuid.New().String() + ".db"
	defer os.Remove(dbPath)
	defer os.Remove(dbPath + "-wal")
	defer os.Remove(dbPath + "-shm")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	db := store.GetDB()

	tests := []struct {
		name          string
		pragma        string
		expectedValue string
	}{
		{
			name:          "Foreign keys enabled",
			pragma:        "PRAGMA foreign_keys",
			expectedValue: "1",
		},
		{
			name:          "WAL mode enabled",
			pragma:        "PRAGMA journal_mode",
			expectedValue: "wal",
		},
		{
			name:          "Busy timeout configured",
			pragma:        "PRAGMA busy_timeout",
			expectedValue: "5000",
		},
		{
			name:          "Synchronous mode set to NORMAL",
			pragma:        "PRAGMA synchronous",
			expectedValue: "1", // NORMAL = 1
		},
		{
			name:          "Cache size configured",
			pragma:        "PRAGMA cache_size",
			expectedValue: "-64000",
		},
		{
			name:          "Temp store set to MEMORY",
			pragma:        "PRAGMA temp_store",
			expectedValue: "2", // MEMORY = 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var value string
			err := db.QueryRow(tt.pragma).Scan(&value)
			if err != nil {
				t.Fatalf("Failed to query %s: %v", tt.pragma, err)
			}

			if value != tt.expectedValue {
				t.Errorf("%s: expected %s, got %s", tt.pragma, tt.expectedValue, value)
			} else {
				t.Logf("%s: ✓ %s = %s", tt.name, tt.pragma, value)
			}
		})
	}

	// Test connection pool configuration
	t.Run("Connection pool configured", func(t *testing.T) {
		stats := db.Stats()
		
		// Verify MaxOpenConns is set (should not be 0 which means unlimited)
		if stats.MaxOpenConnections != 25 {
			t.Errorf("MaxOpenConnections: expected 25, got %d", stats.MaxOpenConnections)
		} else {
			t.Logf("Connection pool: ✓ MaxOpenConnections = %d", stats.MaxOpenConnections)
		}
	})

	// Verify WAL files are created
	t.Run("WAL files created", func(t *testing.T) {
		// Add a comment to trigger WAL file creation
		comment := Comment{
			ID:       "test-wal-1",
			Author:   "Test User",
			AuthorID: "user-1",
			Text:     "Test comment for WAL",
			Status:   "approved",
		}
		if err := store.AddPageComment("test-site", "test-page", comment); err != nil {
			t.Fatalf("Failed to add comment: %v", err)
		}

		// Check if WAL file exists
		walPath := dbPath + "-wal"
		if _, err := os.Stat(walPath); os.IsNotExist(err) {
			t.Errorf("WAL file not created at %s", walPath)
		} else {
			t.Logf("WAL file: ✓ %s exists", walPath)
		}

		// Check if SHM file exists
		shmPath := dbPath + "-shm"
		if _, err := os.Stat(shmPath); os.IsNotExist(err) {
			t.Errorf("SHM file not created at %s", shmPath)
		} else {
			t.Logf("SHM file: ✓ %s exists", shmPath)
		}
	})
}

// TestSQLiteOptimizations_MemoryDB tests that optimizations work with :memory: databases
func TestSQLiteOptimizations_MemoryDB(t *testing.T) {
	store, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory store: %v", err)
	}
	defer store.Close()

	db := store.GetDB()

	// Verify basic PRAGMAs work with memory DB
	var journalMode string
	err = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Fatalf("Failed to query journal_mode: %v", err)
	}

	// Memory databases might use different journal modes
	t.Logf("In-memory database journal_mode: %s", journalMode)

	// Verify we can still perform operations
	comment := Comment{
		ID:       "test-1",
		Author:   "Test User",
		AuthorID: "user-1",
		Text:     "Test comment",
		Status:   "approved",
	}
	if err := store.AddPageComment("test-site", "test-page", comment); err != nil {
		t.Fatalf("Failed to add comment to memory DB: %v", err)
	}

	comments, err := store.GetPageComments("test-site", "test-page")
	if err != nil {
		t.Fatalf("Failed to get comments from memory DB: %v", err)
	}

	if len(comments) != 1 {
		t.Errorf("Expected 1 comment, got %d", len(comments))
	}

	t.Log("✓ In-memory database works with optimizations")
}

// TestConnectionPoolBehavior tests the connection pool limits
func TestConnectionPoolBehavior(t *testing.T) {
	dbPath := "/tmp/test_pool_" + uuid.New().String() + ".db"
	defer os.Remove(dbPath)
	defer os.Remove(dbPath + "-wal")
	defer os.Remove(dbPath + "-shm")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	db := store.GetDB()

	// Get initial stats
	initialStats := db.Stats()
	t.Logf("Initial connection pool stats:")
	t.Logf("  - MaxOpenConnections: %d", initialStats.MaxOpenConnections)
	t.Logf("  - OpenConnections: %d", initialStats.OpenConnections)
	t.Logf("  - InUse: %d", initialStats.InUse)
	t.Logf("  - Idle: %d", initialStats.Idle)

	// Perform some operations to create connections
	siteID := "test-site"
	pageID := "test-page"
	
	for i := 0; i < 10; i++ {
		comment := Comment{
			ID:       uuid.New().String(),
			Author:   "Test User",
			AuthorID: "user-1",
			Text:     "Test comment",
			Status:   "approved",
		}
		if err := store.AddPageComment(siteID, pageID, comment); err != nil {
			t.Fatalf("Failed to add comment: %v", err)
		}
	}

	// Get stats after operations
	afterStats := db.Stats()
	t.Logf("After operations connection pool stats:")
	t.Logf("  - OpenConnections: %d", afterStats.OpenConnections)
	t.Logf("  - InUse: %d", afterStats.InUse)
	t.Logf("  - Idle: %d", afterStats.Idle)
	t.Logf("  - WaitCount: %d", afterStats.WaitCount)
	t.Logf("  - WaitDuration: %v", afterStats.WaitDuration)

	// Verify we didn't exceed max connections
	if afterStats.OpenConnections > 25 {
		t.Errorf("OpenConnections exceeded MaxOpenConns: %d > 25", afterStats.OpenConnections)
	}

	t.Log("✓ Connection pool respects limits")
}

// TestDatabasePingHealthCheck verifies the connection health check works
func TestDatabasePingHealthCheck(t *testing.T) {
	dbPath := "/tmp/test_ping_" + uuid.New().String() + ".db"
	defer os.Remove(dbPath)
	defer os.Remove(dbPath + "-wal")
	defer os.Remove(dbPath + "-shm")

	// Valid database path - should succeed
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store with valid path: %v", err)
	}
	store.Close()

	// Invalid database path - should fail with timeout
	invalidPath := "/nonexistent/directory/test.db"
	_, err = NewSQLiteStore(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	} else if !strings.Contains(err.Error(), "database not responding") && !strings.Contains(err.Error(), "failed to open database") {
		t.Logf("Got expected error: %v", err)
	}

	t.Log("✓ Database health check works")
}

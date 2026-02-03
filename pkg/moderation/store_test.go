package moderation

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestModerationConfigStore(t *testing.T) {
	// Create a temporary test database
	dbPath := "/tmp/test_moderation_" + time.Now().Format("20060102150405") + ".db"
	defer os.Remove(dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create moderation_config table
	schema := `
	CREATE TABLE IF NOT EXISTS sites (
		id TEXT PRIMARY KEY,
		owner_id TEXT NOT NULL,
		name TEXT NOT NULL,
		domain TEXT,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS moderation_config (
		id TEXT PRIMARY KEY,
		site_id TEXT NOT NULL UNIQUE,
		enabled INTEGER DEFAULT 0,
		auto_reject_threshold REAL DEFAULT 0.85,
		auto_approve_threshold REAL DEFAULT 0.30,
		check_spam INTEGER DEFAULT 1,
		check_offensive INTEGER DEFAULT 1,
		check_aggressive INTEGER DEFAULT 1,
		check_off_topic INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE
	);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	// Create a test site
	_, err = db.Exec("INSERT INTO sites (id, owner_id, name) VALUES (?, ?, ?)", "site1", "owner1", "Test Site")
	if err != nil {
		t.Fatalf("Failed to create test site: %v", err)
	}

	store := NewConfigStore(db)

	t.Run("CreateAndGetConfig", func(t *testing.T) {
		config := ModerationConfig{
			Enabled:              true,
			AutoRejectThreshold:  0.9,
			AutoApproveThreshold: 0.2,
			CheckSpam:            true,
			CheckOffensive:       false,
			CheckAggressive:      true,
			CheckOffTopic:        true,
		}

		// Create config
		err := store.Create(context.Background(), "site1", config)
		if err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}

		// Retrieve config
		retrieved, err := store.GetBySiteID(context.Background(), "site1")
		if err != nil {
			t.Fatalf("Failed to get config: %v", err)
		}

		if !retrieved.Enabled {
			t.Error("Expected Enabled to be true")
		}
		if retrieved.AutoRejectThreshold != 0.9 {
			t.Errorf("Expected AutoRejectThreshold 0.9, got %f", retrieved.AutoRejectThreshold)
		}
		if retrieved.AutoApproveThreshold != 0.2 {
			t.Errorf("Expected AutoApproveThreshold 0.2, got %f", retrieved.AutoApproveThreshold)
		}
		if !retrieved.CheckSpam {
			t.Error("Expected CheckSpam to be true")
		}
		if retrieved.CheckOffensive {
			t.Error("Expected CheckOffensive to be false")
		}
		if !retrieved.CheckAggressive {
			t.Error("Expected CheckAggressive to be true")
		}
		if !retrieved.CheckOffTopic {
			t.Error("Expected CheckOffTopic to be true")
		}
	})

	t.Run("UpdateConfig", func(t *testing.T) {
		config := ModerationConfig{
			Enabled:              false,
			AutoRejectThreshold:  0.8,
			AutoApproveThreshold: 0.3,
			CheckSpam:            false,
			CheckOffensive:       true,
			CheckAggressive:      false,
			CheckOffTopic:        false,
		}

		// Update config
		err := store.Update(context.Background(), "site1", config)
		if err != nil {
			t.Fatalf("Failed to update config: %v", err)
		}

		// Retrieve updated config
		retrieved, err := store.GetBySiteID(context.Background(), "site1")
		if err != nil {
			t.Fatalf("Failed to get config: %v", err)
		}

		if retrieved.Enabled {
			t.Error("Expected Enabled to be false")
		}
		if retrieved.AutoRejectThreshold != 0.8 {
			t.Errorf("Expected AutoRejectThreshold 0.8, got %f", retrieved.AutoRejectThreshold)
		}
		if retrieved.CheckSpam {
			t.Error("Expected CheckSpam to be false")
		}
		if !retrieved.CheckOffensive {
			t.Error("Expected CheckOffensive to be true")
		}
	})

	t.Run("GetNonExistentConfig", func(t *testing.T) {
		_, err := store.GetBySiteID(context.Background(), "nonexistent")
		if err == nil {
			t.Error("Expected error for non-existent site")
		}
	})

	t.Run("DeleteConfig", func(t *testing.T) {
		err := store.Delete(context.Background(), "site1")
		if err != nil {
			t.Fatalf("Failed to delete config: %v", err)
		}

		// Verify deletion
		_, err = store.GetBySiteID(context.Background(), "site1")
		if err == nil {
			t.Error("Expected error after deletion")
		}
	})
}

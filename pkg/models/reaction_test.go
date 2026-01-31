package models

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Enable foreign key constraints
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Create schema
	schema := `
	CREATE TABLE sites (
		id TEXT PRIMARY KEY,
		owner_id TEXT NOT NULL,
		name TEXT NOT NULL,
		domain TEXT,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE comments (
		id TEXT PRIMARY KEY,
		site_id TEXT NOT NULL,
		page_id TEXT NOT NULL,
		author TEXT NOT NULL,
		text TEXT NOT NULL,
		parent_id TEXT,
		status TEXT DEFAULT 'pending',
		moderated_by TEXT,
		moderated_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE allowed_reactions (
		id TEXT PRIMARY KEY,
		site_id TEXT NOT NULL,
		name TEXT NOT NULL,
		emoji TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE,
		UNIQUE(site_id, name)
	);

	CREATE TABLE reactions (
		id TEXT PRIMARY KEY,
		comment_id TEXT NOT NULL,
		allowed_reaction_id TEXT NOT NULL,
		user_identifier TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE,
		FOREIGN KEY (allowed_reaction_id) REFERENCES allowed_reactions(id) ON DELETE CASCADE,
		UNIQUE(comment_id, allowed_reaction_id, user_identifier)
	);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

func TestAllowedReactionStore_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a test site first
	_, err := db.Exec("INSERT INTO sites (id, owner_id, name) VALUES (?, ?, ?)",
		"site-1", "user-1", "Test Site")
	if err != nil {
		t.Fatalf("Failed to create test site: %v", err)
	}

	store := NewAllowedReactionStore(db)

	// Test creating an allowed reaction
	reaction, err := store.Create("site-1", "thumbs_up", "👍")
	if err != nil {
		t.Fatalf("Failed to create allowed reaction: %v", err)
	}

	if reaction.ID == "" {
		t.Error("Expected reaction ID to be set")
	}
	if reaction.SiteID != "site-1" {
		t.Errorf("Expected site_id to be 'site-1', got '%s'", reaction.SiteID)
	}
	if reaction.Name != "thumbs_up" {
		t.Errorf("Expected name to be 'thumbs_up', got '%s'", reaction.Name)
	}
	if reaction.Emoji != "👍" {
		t.Errorf("Expected emoji to be '👍', got '%s'", reaction.Emoji)
	}
}

func TestAllowedReactionStore_GetBySite(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test sites
	_, err := db.Exec("INSERT INTO sites (id, owner_id, name) VALUES (?, ?, ?)",
		"site-1", "user-1", "Test Site")
	if err != nil {
		t.Fatalf("Failed to create test site: %v", err)
	}

	store := NewAllowedReactionStore(db)

	// Create some reactions
	store.Create("site-1", "thumbs_up", "👍")
	store.Create("site-1", "heart", "❤️")

	// Get reactions for site
	reactions, err := store.GetBySite("site-1")
	if err != nil {
		t.Fatalf("Failed to get reactions: %v", err)
	}

	if len(reactions) != 2 {
		t.Errorf("Expected 2 reactions, got %d", len(reactions))
	}
}

func TestAllowedReactionStore_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test site
	_, err := db.Exec("INSERT INTO sites (id, owner_id, name) VALUES (?, ?, ?)",
		"site-1", "user-1", "Test Site")
	if err != nil {
		t.Fatalf("Failed to create test site: %v", err)
	}

	store := NewAllowedReactionStore(db)

	// Create a reaction
	reaction, _ := store.Create("site-1", "thumbs_up", "👍")

	// Update it
	err = store.Update(reaction.ID, "like", "👍")
	if err != nil {
		t.Fatalf("Failed to update reaction: %v", err)
	}

	// Verify update
	updated, err := store.GetByID(reaction.ID)
	if err != nil {
		t.Fatalf("Failed to get updated reaction: %v", err)
	}
	if updated.Name != "like" {
		t.Errorf("Expected name to be 'like', got '%s'", updated.Name)
	}
}

func TestAllowedReactionStore_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test site
	_, err := db.Exec("INSERT INTO sites (id, owner_id, name) VALUES (?, ?, ?)",
		"site-1", "user-1", "Test Site")
	if err != nil {
		t.Fatalf("Failed to create test site: %v", err)
	}

	store := NewAllowedReactionStore(db)

	// Create a reaction
	reaction, _ := store.Create("site-1", "thumbs_up", "👍")

	// Delete it
	err = store.Delete(reaction.ID)
	if err != nil {
		t.Fatalf("Failed to delete reaction: %v", err)
	}

	// Verify deletion
	_, err = store.GetByID(reaction.ID)
	if err == nil {
		t.Error("Expected error when getting deleted reaction")
	}
}

func TestReactionStore_AddReaction(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test data
	_, err := db.Exec("INSERT INTO sites (id, owner_id, name) VALUES (?, ?, ?)",
		"site-1", "user-1", "Test Site")
	if err != nil {
		t.Fatalf("Failed to create test site: %v", err)
	}

	_, err = db.Exec("INSERT INTO comments (id, site_id, page_id, author, text) VALUES (?, ?, ?, ?, ?)",
		"comment-1", "site-1", "page-1", "John", "Test comment")
	if err != nil {
		t.Fatalf("Failed to create test comment: %v", err)
	}

	allowedStore := NewAllowedReactionStore(db)
	allowed, _ := allowedStore.Create("site-1", "thumbs_up", "👍")

	reactionStore := NewReactionStore(db)

	// Add a reaction
	reaction, err := reactionStore.AddReaction("comment-1", allowed.ID, "user-123")
	if err != nil {
		t.Fatalf("Failed to add reaction: %v", err)
	}

	if reaction == nil {
		t.Fatal("Expected reaction to be created")
	}
	if reaction.CommentID != "comment-1" {
		t.Errorf("Expected comment_id to be 'comment-1', got '%s'", reaction.CommentID)
	}
	if reaction.UserIdentifier != "user-123" {
		t.Errorf("Expected user_identifier to be 'user-123', got '%s'", reaction.UserIdentifier)
	}
}

func TestReactionStore_AddReaction_Toggle(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test data
	_, err := db.Exec("INSERT INTO sites (id, owner_id, name) VALUES (?, ?, ?)",
		"site-1", "user-1", "Test Site")
	if err != nil {
		t.Fatalf("Failed to create test site: %v", err)
	}

	_, err = db.Exec("INSERT INTO comments (id, site_id, page_id, author, text) VALUES (?, ?, ?, ?, ?)",
		"comment-1", "site-1", "page-1", "John", "Test comment")
	if err != nil {
		t.Fatalf("Failed to create test comment: %v", err)
	}

	allowedStore := NewAllowedReactionStore(db)
	allowed, _ := allowedStore.Create("site-1", "thumbs_up", "👍")

	reactionStore := NewReactionStore(db)

	// Add a reaction
	reaction1, err := reactionStore.AddReaction("comment-1", allowed.ID, "user-123")
	if err != nil {
		t.Fatalf("Failed to add reaction: %v", err)
	}
	if reaction1 == nil {
		t.Fatal("Expected reaction to be created")
	}

	// Add same reaction again (should toggle it off)
	reaction2, err := reactionStore.AddReaction("comment-1", allowed.ID, "user-123")
	if err != nil {
		t.Fatalf("Failed to toggle reaction: %v", err)
	}
	if reaction2 != nil {
		t.Error("Expected reaction to be nil (toggled off)")
	}

	// Verify reaction was removed
	reactions, _ := reactionStore.GetReactionsByComment("comment-1")
	if len(reactions) != 0 {
		t.Errorf("Expected 0 reactions after toggle, got %d", len(reactions))
	}
}

func TestReactionStore_GetReactionCounts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test data
	_, err := db.Exec("INSERT INTO sites (id, owner_id, name) VALUES (?, ?, ?)",
		"site-1", "user-1", "Test Site")
	if err != nil {
		t.Fatalf("Failed to create test site: %v", err)
	}

	_, err = db.Exec("INSERT INTO comments (id, site_id, page_id, author, text) VALUES (?, ?, ?, ?, ?)",
		"comment-1", "site-1", "page-1", "John", "Test comment")
	if err != nil {
		t.Fatalf("Failed to create test comment: %v", err)
	}

	allowedStore := NewAllowedReactionStore(db)
	thumbsUp, _ := allowedStore.Create("site-1", "thumbs_up", "👍")
	heart, _ := allowedStore.Create("site-1", "heart", "❤️")

	reactionStore := NewReactionStore(db)

	// Add multiple reactions
	reactionStore.AddReaction("comment-1", thumbsUp.ID, "user-1")
	reactionStore.AddReaction("comment-1", thumbsUp.ID, "user-2")
	reactionStore.AddReaction("comment-1", thumbsUp.ID, "user-3")
	reactionStore.AddReaction("comment-1", heart.ID, "user-1")

	// Get counts
	counts, err := reactionStore.GetReactionCounts("comment-1")
	if err != nil {
		t.Fatalf("Failed to get reaction counts: %v", err)
	}

	if len(counts) != 2 {
		t.Fatalf("Expected 2 reaction types, got %d", len(counts))
	}

	// Check thumbs_up count
	if counts[0].Name != "thumbs_up" {
		t.Errorf("Expected first reaction to be 'thumbs_up', got '%s'", counts[0].Name)
	}
	if counts[0].Count != 3 {
		t.Errorf("Expected thumbs_up count to be 3, got %d", counts[0].Count)
	}

	// Check heart count
	if counts[1].Name != "heart" {
		t.Errorf("Expected second reaction to be 'heart', got '%s'", counts[1].Name)
	}
	if counts[1].Count != 1 {
		t.Errorf("Expected heart count to be 1, got %d", counts[1].Count)
	}
}

func TestReactionStore_CascadeDelete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test data
	_, err := db.Exec("INSERT INTO sites (id, owner_id, name) VALUES (?, ?, ?)",
		"site-1", "user-1", "Test Site")
	if err != nil {
		t.Fatalf("Failed to create test site: %v", err)
	}

	_, err = db.Exec("INSERT INTO comments (id, site_id, page_id, author, text) VALUES (?, ?, ?, ?, ?)",
		"comment-1", "site-1", "page-1", "John", "Test comment")
	if err != nil {
		t.Fatalf("Failed to create test comment: %v", err)
	}

	allowedStore := NewAllowedReactionStore(db)
	allowed, _ := allowedStore.Create("site-1", "thumbs_up", "👍")

	reactionStore := NewReactionStore(db)
	reactionStore.AddReaction("comment-1", allowed.ID, "user-1")

	// Delete the comment - reactions should cascade delete
	_, err = db.Exec("DELETE FROM comments WHERE id = ?", "comment-1")
	if err != nil {
		t.Fatalf("Failed to delete comment: %v", err)
	}

	// Verify reactions were deleted
	reactions, _ := reactionStore.GetReactionsByComment("comment-1")
	if len(reactions) != 0 {
		t.Errorf("Expected reactions to be cascade deleted, found %d", len(reactions))
	}
}

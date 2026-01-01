package comments

import (
	"path/filepath"
	"testing"
	"time"
)

// Integration test: Complete flow from in-memory to SQLite
func TestIntegration_AddRetrieveComments(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "integration.db")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Add multiple comments
	comments := []Comment{
		{ID: "1", Author: "Alice", Text: "First comment", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "2", Author: "Bob", Text: "Second comment", ParentID: "1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "3", Author: "Charlie", Text: "Third comment", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, c := range comments {
		if err := store.AddPageComment("mysite", "mypage", c); err != nil {
			t.Fatalf("failed to add comment: %v", err)
		}
	}

	// Retrieve and verify
	retrieved, err := store.GetPageComments("mysite", "mypage")
	if err != nil {
		t.Fatalf("failed to retrieve comments: %v", err)
	}

	if len(retrieved) != len(comments) {
		t.Fatalf("expected %d comments, got %d", len(comments), len(retrieved))
	}

	// Verify each comment
	for i, expected := range comments {
		found := false
		for _, actual := range retrieved {
			if actual.ID == expected.ID {
				found = true
				if actual.Author != expected.Author {
					t.Errorf("comment %s: expected author '%s', got '%s'", expected.ID, expected.Author, actual.Author)
				}
				if actual.Text != expected.Text {
					t.Errorf("comment %s: expected text '%s', got '%s'", expected.ID, expected.Text, actual.Text)
				}
				if actual.ParentID != expected.ParentID {
					t.Errorf("comment %s: expected ParentID '%s', got '%s'", expected.ID, expected.ParentID, actual.ParentID)
				}
				break
			}
		}
		if !found {
			t.Errorf("comment %d with ID '%s' not found in retrieved comments", i, expected.ID)
		}
	}
}

// Integration test: Verify persistence across store instances
func TestIntegration_PersistenceAcrossRestarts(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "persistence.db")

	// First session: add comments
	store1, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create first store: %v", err)
	}

	for i := 0; i < 10; i++ {
		comment := Comment{
			ID:        string(rune('A' + i)),
			Author:    "User",
			Text:      "Comment " + string(rune('A'+i)),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := store1.AddPageComment("site1", "page1", comment); err != nil {
			t.Fatalf("failed to add comment in first session: %v", err)
		}
	}

	if err := store1.Close(); err != nil {
		t.Fatalf("failed to close first store: %v", err)
	}

	// Second session: verify and add more
	store2, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create second store: %v", err)
	}

	comments, err := store2.GetPageComments("site1", "page1")
	if err != nil {
		t.Fatalf("failed to retrieve comments in second session: %v", err)
	}

	if len(comments) != 10 {
		t.Fatalf("expected 10 comments in second session, got %d", len(comments))
	}

	// Add more comments
	for i := 10; i < 15; i++ {
		comment := Comment{
			ID:        string(rune('A' + i)),
			Author:    "User2",
			Text:      "Comment " + string(rune('A'+i)),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := store2.AddPageComment("site1", "page1", comment); err != nil {
			t.Fatalf("failed to add comment in second session: %v", err)
		}
	}

	if err := store2.Close(); err != nil {
		t.Fatalf("failed to close second store: %v", err)
	}

	// Third session: verify all
	store3, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create third store: %v", err)
	}
	defer store3.Close()

	comments, err = store3.GetPageComments("site1", "page1")
	if err != nil {
		t.Fatalf("failed to retrieve comments in third session: %v", err)
	}

	if len(comments) != 15 {
		t.Fatalf("expected 15 comments in third session, got %d", len(comments))
	}
}

// Integration test: Error recovery
func TestIntegration_ErrorRecovery(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "recovery.db")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Add a comment successfully
	comment1 := Comment{
		ID:        "1",
		Author:    "Alice",
		Text:      "First comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := store.AddPageComment("site1", "page1", comment1); err != nil {
		t.Fatalf("failed to add first comment: %v", err)
	}

	// Try to add duplicate (should fail)
	duplicate := Comment{
		ID:        "1",
		Author:    "Bob",
		Text:      "Duplicate",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = store.AddPageComment("site1", "page1", duplicate)
	if err == nil {
		t.Error("expected error for duplicate ID, got nil")
	}

	// Verify original comment is still there and unchanged
	comments, err := store.GetPageComments("site1", "page1")
	if err != nil {
		t.Fatalf("failed to retrieve comments: %v", err)
	}

	if len(comments) != 1 {
		t.Fatalf("expected 1 comment after failed duplicate, got %d", len(comments))
	}

	if comments[0].Author != "Alice" {
		t.Errorf("expected author 'Alice', got '%s'", comments[0].Author)
	}

	// Add another comment successfully after error
	comment2 := Comment{
		ID:        "2",
		Author:    "Charlie",
		Text:      "Second comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := store.AddPageComment("site1", "page1", comment2); err != nil {
		t.Fatalf("failed to add comment after error: %v", err)
	}

	// Verify both comments exist
	comments, err = store.GetPageComments("site1", "page1")
	if err != nil {
		t.Fatalf("failed to retrieve comments after recovery: %v", err)
	}

	if len(comments) != 2 {
		t.Fatalf("expected 2 comments after recovery, got %d", len(comments))
	}
}

// Integration test: Multiple sites and pages
func TestIntegration_MultipleSitesAndPages(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "multisites.db")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Create a matrix of sites, pages, and comments
	sites := []string{"blog", "docs", "forum"}
	pages := []string{"home", "about", "contact"}

	expectedCounts := make(map[string]map[string]int)

	for _, site := range sites {
		expectedCounts[site] = make(map[string]int)
		for _, page := range pages {
			// Add 1-3 comments per page
			numComments := (len(site) + len(page)) % 3 + 1
			expectedCounts[site][page] = numComments

			for i := 0; i < numComments; i++ {
				comment := Comment{
					ID:        site + "-" + page + "-" + string(rune('A'+i)),
					Author:    "User",
					Text:      "Comment on " + site + "/" + page,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				if err := store.AddPageComment(site, page, comment); err != nil {
					t.Fatalf("failed to add comment to %s/%s: %v", site, page, err)
				}
			}
		}
	}

	// Verify each site/page has correct number of comments
	for _, site := range sites {
		for _, page := range pages {
			comments, err := store.GetPageComments(site, page)
			if err != nil {
				t.Fatalf("failed to get comments for %s/%s: %v", site, page, err)
			}

			expected := expectedCounts[site][page]
			if len(comments) != expected {
				t.Errorf("%s/%s: expected %d comments, got %d", site, page, expected, len(comments))
			}
		}
	}
}

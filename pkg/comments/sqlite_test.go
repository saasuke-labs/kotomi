package comments

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// Helper function to create a temporary test database
func createTestDB(t *testing.T) (*SQLiteStore, string) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	return store, dbPath
}

func TestNewSQLiteStore_Success(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()

	if store == nil {
		t.Error("expected non-nil store")
	}
	if store.db == nil {
		t.Error("expected non-nil database connection")
	}
}

func TestNewSQLiteStore_InvalidPath(t *testing.T) {
	// Try to create database in non-existent directory without proper permissions
	dbPath := "/nonexistent/impossible/path/test.db"

	_, err := NewSQLiteStore(dbPath)
	if err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}

func TestSQLiteStore_AddPageComment_Success(t *testing.T) {
	store, _ := createTestDB(t)
	defer store.Close()

	comment := Comment{
		ID:        "1",
		Author:    "John",
		Text:      "Test comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := store.AddPageComment(context.Background(), "site1", "page1", comment)
	if err != nil {
		t.Fatalf("AddPageComment failed: %v", err)
	}

	// Verify comment was added
	comments, err := store.GetPageComments(context.Background(), "site1", "page1")
	if err != nil {
		t.Fatalf("GetPageComments failed: %v", err)
	}

	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}

	if comments[0].ID != "1" {
		t.Errorf("expected ID '1', got '%s'", comments[0].ID)
	}
	if comments[0].Author != "John" {
		t.Errorf("expected author 'John', got '%s'", comments[0].Author)
	}
}

func TestSQLiteStore_AddPageComment_DuplicateID(t *testing.T) {
	store, _ := createTestDB(t)
	defer store.Close()

	comment := Comment{
		ID:        "1",
		Author:    "John",
		Text:      "First comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := store.AddPageComment(context.Background(), "site1", "page1", comment)
	if err != nil {
		t.Fatalf("first AddPageComment failed: %v", err)
	}

	// Try to add duplicate
	duplicate := Comment{
		ID:        "1",
		Author:    "Jane",
		Text:      "Duplicate comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = store.AddPageComment(context.Background(), "site1", "page1", duplicate)
	if err == nil {
		t.Error("expected error for duplicate ID, got nil")
	}
}

func TestSQLiteStore_AddPageComment_AutoTimestamps(t *testing.T) {
	store, _ := createTestDB(t)
	defer store.Close()

	comment := Comment{
		ID:     "1",
		Author: "John",
		Text:   "Test comment",
		// No timestamps set
	}

	err := store.AddPageComment(context.Background(), "site1", "page1", comment)
	if err != nil {
		t.Fatalf("AddPageComment failed: %v", err)
	}

	comments, err := store.GetPageComments(context.Background(), "site1", "page1")
	if err != nil {
		t.Fatalf("GetPageComments failed: %v", err)
	}

	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}

	if comments[0].CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set automatically")
	}
	if comments[0].UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set automatically")
	}
}

func TestSQLiteStore_GetPageComments_Empty(t *testing.T) {
	store, _ := createTestDB(t)
	defer store.Close()

	comments, err := store.GetPageComments(context.Background(), "nonexistent", "page")
	if err != nil {
		t.Fatalf("GetPageComments failed: %v", err)
	}

	if comments == nil {
		t.Error("expected non-nil slice")
	}
	if len(comments) != 0 {
		t.Errorf("expected empty slice, got %d comments", len(comments))
	}
}

func TestSQLiteStore_GetPageComments_MultipleComments(t *testing.T) {
	store, _ := createTestDB(t)
	defer store.Close()

	baseTime := time.Now()
	comments := []Comment{
		{ID: "1", Author: "John", Text: "First", CreatedAt: baseTime, UpdatedAt: baseTime},
		{ID: "2", Author: "Jane", Text: "Second", CreatedAt: baseTime.Add(1 * time.Second), UpdatedAt: baseTime.Add(1 * time.Second)},
		{ID: "3", Author: "Bob", Text: "Third", CreatedAt: baseTime.Add(2 * time.Second), UpdatedAt: baseTime.Add(2 * time.Second)},
	}

	for _, c := range comments {
		if err := store.AddPageComment(context.Background(), "site1", "page1", c); err != nil {
			t.Fatalf("failed to add comment: %v", err)
		}
	}

	retrieved, err := store.GetPageComments(context.Background(), "site1", "page1")
	if err != nil {
		t.Fatalf("GetPageComments failed: %v", err)
	}

	if len(retrieved) != 3 {
		t.Fatalf("expected 3 comments, got %d", len(retrieved))
	}

	// Verify ordering by CreatedAt
	for i, c := range retrieved {
		if c.ID != comments[i].ID {
			t.Errorf("comment %d: expected ID '%s', got '%s'", i, comments[i].ID, c.ID)
		}
	}
}

func TestSQLiteStore_GetPageComments_WithParentID(t *testing.T) {
	store, _ := createTestDB(t)
	defer store.Close()

	parent := Comment{
		ID:        "1",
		Author:    "John",
		Text:      "Parent",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	reply := Comment{
		ID:        "2",
		Author:    "Jane",
		Text:      "Reply",
		ParentID:  "1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := store.AddPageComment(context.Background(), "site1", "page1", parent); err != nil {
		t.Fatalf("failed to add parent: %v", err)
	}
	if err := store.AddPageComment(context.Background(), "site1", "page1", reply); err != nil {
		t.Fatalf("failed to add reply: %v", err)
	}

	comments, err := store.GetPageComments(context.Background(), "site1", "page1")
	if err != nil {
		t.Fatalf("GetPageComments failed: %v", err)
	}

	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}

	// Find reply
	var foundReply *Comment
	for _, c := range comments {
		if c.ID == "2" {
			foundReply = &c
			break
		}
	}

	if foundReply == nil {
		t.Fatal("reply not found")
	}
	if foundReply.ParentID != "1" {
		t.Errorf("expected ParentID '1', got '%s'", foundReply.ParentID)
	}
}

func TestSQLiteStore_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create store and add comment
	store1, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create first store: %v", err)
	}

	comment := Comment{
		ID:        "1",
		Author:    "John",
		Text:      "Persistent comment",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := store1.AddPageComment(context.Background(), "site1", "page1", comment); err != nil {
		t.Fatalf("failed to add comment: %v", err)
	}

	// Close first store
	if err := store1.Close(); err != nil {
		t.Fatalf("failed to close first store: %v", err)
	}

	// Reopen database
	store2, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to reopen store: %v", err)
	}
	defer store2.Close()

	// Verify comment persisted
	comments, err := store2.GetPageComments(context.Background(), "site1", "page1")
	if err != nil {
		t.Fatalf("GetPageComments failed: %v", err)
	}

	if len(comments) != 1 {
		t.Fatalf("expected 1 comment after reopen, got %d", len(comments))
	}

	if comments[0].ID != "1" {
		t.Errorf("expected ID '1', got '%s'", comments[0].ID)
	}
	if comments[0].Text != "Persistent comment" {
		t.Errorf("expected text 'Persistent comment', got '%s'", comments[0].Text)
	}
}

func TestSQLiteStore_ConcurrentWrites(t *testing.T) {
	store, _ := createTestDB(t)
	defer store.Close()

	numGoroutines := 10
	commentsPerGoroutine := 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < commentsPerGoroutine; j++ {
				comment := Comment{
					ID:        fmt.Sprintf("g%d-c%d", goroutineID, j),
					Author:    fmt.Sprintf("Author%d", goroutineID),
					Text:      fmt.Sprintf("Text from goroutine %d, comment %d", goroutineID, j),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				if err := store.AddPageComment(context.Background(), "site1", "page1", comment); err != nil {
					t.Errorf("failed to add comment: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify all comments were added
	comments, err := store.GetPageComments(context.Background(), "site1", "page1")
	if err != nil {
		t.Fatalf("GetPageComments failed: %v", err)
	}

	expectedCount := numGoroutines * commentsPerGoroutine
	if len(comments) != expectedCount {
		t.Errorf("expected %d comments, got %d", expectedCount, len(comments))
	}
}

func TestSQLiteStore_ConcurrentReadWrite(t *testing.T) {
	store, _ := createTestDB(t)
	defer store.Close()

	// Add initial comments
	for i := 0; i < 5; i++ {
		comment := Comment{
			ID:        fmt.Sprintf("%d", i),
			Author:    "Initial",
			Text:      "Initial comment",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := store.AddPageComment(context.Background(), "site1", "page1", comment); err != nil {
			t.Fatalf("failed to add initial comment: %v", err)
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Writer goroutine
	go func() {
		defer wg.Done()
		for i := 5; i < 10; i++ {
			comment := Comment{
				ID:        fmt.Sprintf("%d", i),
				Author:    "Writer",
				Text:      "New comment",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := store.AddPageComment(context.Background(), "site1", "page1", comment); err != nil {
				t.Errorf("failed to add comment: %v", err)
			}
			time.Sleep(1 * time.Millisecond)
		}
	}()

	// Reader goroutine
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			comments, err := store.GetPageComments(context.Background(), "site1", "page1")
			if err != nil {
				t.Errorf("GetPageComments failed: %v", err)
			}
			if comments == nil {
				t.Error("GetPageComments returned nil")
			}
			time.Sleep(1 * time.Millisecond)
		}
	}()

	wg.Wait()
}

func TestSQLiteStore_Close(t *testing.T) {
	store, _ := createTestDB(t)

	err := store.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Attempting operations after close should fail
	err = store.AddPageComment(context.Background(), "site1", "page1", Comment{
		ID:     "1",
		Author: "John",
		Text:   "Test",
	})
	if err == nil {
		t.Error("expected error when adding comment after close, got nil")
	}
}

func TestSQLiteStore_CloseNil(t *testing.T) {
	store := &SQLiteStore{db: nil}
	err := store.Close()
	if err != nil {
		t.Errorf("Close with nil db should not error, got: %v", err)
	}
}

func TestSQLiteStore_MultipleSitesAndPages(t *testing.T) {
	store, _ := createTestDB(t)
	defer store.Close()

	testCases := []struct {
		site    string
		page    string
		comment Comment
	}{
		{"site1", "page1", Comment{ID: "1", Author: "John", Text: "S1P1", CreatedAt: time.Now(), UpdatedAt: time.Now()}},
		{"site1", "page2", Comment{ID: "2", Author: "Jane", Text: "S1P2", CreatedAt: time.Now(), UpdatedAt: time.Now()}},
		{"site2", "page1", Comment{ID: "3", Author: "Bob", Text: "S2P1", CreatedAt: time.Now(), UpdatedAt: time.Now()}},
		{"site2", "page2", Comment{ID: "4", Author: "Alice", Text: "S2P2", CreatedAt: time.Now(), UpdatedAt: time.Now()}},
	}

	// Add all comments
	for _, tc := range testCases {
		if err := store.AddPageComment(context.Background(), tc.site, tc.page, tc.comment); err != nil {
			t.Fatalf("failed to add comment for %s/%s: %v", tc.site, tc.page, err)
		}
	}

	// Verify each site/page has exactly one comment
	for _, tc := range testCases {
		comments, err := store.GetPageComments(context.Background(), tc.site, tc.page)
		if err != nil {
			t.Fatalf("GetPageComments failed for %s/%s: %v", tc.site, tc.page, err)
		}
		if len(comments) != 1 {
			t.Errorf("%s/%s: expected 1 comment, got %d", tc.site, tc.page, len(comments))
			continue
		}
		if comments[0].Text != tc.comment.Text {
			t.Errorf("%s/%s: expected text '%s', got '%s'", tc.site, tc.page, tc.comment.Text, comments[0].Text)
		}
	}
}

// TestSQLiteStore_UpdateCommentText tests updating comment text
func TestSQLiteStore_UpdateCommentText(t *testing.T) {
	store, _ := createTestDB(t)
	defer store.Close()

	// Create a comment
	comment := Comment{
		ID:        "test-comment-1",
		Author:    "John Doe",
		AuthorID:  "user123",
		Text:      "Original text",
		Status:    "approved",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := store.AddPageComment(context.Background(), "site1", "page1", comment)
	if err != nil {
		t.Fatalf("failed to add comment: %v", err)
	}

	// Update the comment text
	newText := "Updated text content"
	err = store.UpdateCommentText(context.Background(), comment.ID, newText)
	if err != nil {
		t.Fatalf("UpdateCommentText failed: %v", err)
	}

	// Retrieve and verify the comment was updated
	updated, err := store.GetCommentByID(context.Background(), comment.ID)
	if err != nil {
		t.Fatalf("failed to get updated comment: %v", err)
	}

	if updated.Text != newText {
		t.Errorf("expected text '%s', got '%s'", newText, updated.Text)
	}

	// Verify other fields remain unchanged
	if updated.ID != comment.ID {
		t.Errorf("expected ID '%s', got '%s'", comment.ID, updated.ID)
	}
	if updated.Author != comment.Author {
		t.Errorf("expected author '%s', got '%s'", comment.Author, updated.Author)
	}
	if updated.Status != comment.Status {
		t.Errorf("expected status '%s', got '%s'", comment.Status, updated.Status)
	}
}

// TestSQLiteStore_UpdateCommentText_NotFound tests updating non-existent comment
func TestSQLiteStore_UpdateCommentText_NotFound(t *testing.T) {
	store, _ := createTestDB(t)
	defer store.Close()

	err := store.UpdateCommentText(context.Background(), "nonexistent-id", "Some text")
	if err == nil {
		t.Error("expected error for non-existent comment, got nil")
	}
	if err.Error() != "comment not found" {
		t.Errorf("expected 'comment not found' error, got: %v", err)
	}
}


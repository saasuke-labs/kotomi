package comments

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestPostgresStore_IntegrationTest tests basic PostgreSQL operations
// This test requires a PostgreSQL database to be available
// Set DATABASE_URL environment variable to run this test
func TestPostgresStore_IntegrationTest(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("Skipping PostgreSQL integration test: TEST_DATABASE_URL not set")
	}

	store, err := NewPostgresStore(databaseURL)
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Test 1: Add a comment
	siteID := "test-site"
	pageID := "test-page"
	commentID := uuid.New().String()
	
	comment := Comment{
		ID:          commentID,
		Author:      "Test User",
		AuthorID:    "test-user-1",
		AuthorEmail: "test@example.com",
		Text:        "This is a test comment",
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = store.AddPageComment(ctx, siteID, pageID, comment)
	if err != nil {
		t.Fatalf("Failed to add comment: %v", err)
	}

	// Test 2: Retrieve comments
	comments, err := store.GetPageComments(ctx, siteID, pageID)
	if err != nil {
		t.Fatalf("Failed to get comments: %v", err)
	}

	if len(comments) != 1 {
		t.Fatalf("Expected 1 comment, got %d", len(comments))
	}

	if comments[0].ID != commentID {
		t.Errorf("Expected comment ID %s, got %s", commentID, comments[0].ID)
	}

	// Test 3: Update comment status
	err = store.UpdateCommentStatus(ctx, commentID, "approved", "moderator-1")
	if err != nil {
		t.Fatalf("Failed to update comment status: %v", err)
	}

	// Test 4: Get comment by ID
	retrievedComment, err := store.GetCommentByID(ctx, commentID)
	if err != nil {
		t.Fatalf("Failed to get comment by ID: %v", err)
	}

	if retrievedComment.Status != "approved" {
		t.Errorf("Expected status 'approved', got '%s'", retrievedComment.Status)
	}

	if retrievedComment.ModeratedBy != "moderator-1" {
		t.Errorf("Expected moderated_by 'moderator-1', got '%s'", retrievedComment.ModeratedBy)
	}

	// Test 5: Delete comment
	err = store.DeleteComment(ctx, commentID)
	if err != nil {
		t.Fatalf("Failed to delete comment: %v", err)
	}

	// Verify deletion
	comments, err = store.GetPageComments(ctx, siteID, pageID)
	if err != nil {
		t.Fatalf("Failed to get comments after deletion: %v", err)
	}

	if len(comments) != 0 {
		t.Errorf("Expected 0 comments after deletion, got %d", len(comments))
	}

	// Cleanup: delete test site and page
	_, err = store.GetDB().ExecContext(ctx, "DELETE FROM pages WHERE id = $1", pageID)
	if err != nil {
		t.Logf("Warning: Failed to cleanup page: %v", err)
	}
	_, err = store.GetDB().ExecContext(ctx, "DELETE FROM sites WHERE id = $1", siteID)
	if err != nil {
		t.Logf("Warning: Failed to cleanup site: %v", err)
	}
	_, err = store.GetDB().ExecContext(ctx, "DELETE FROM admin_users WHERE id = 'system'")
	if err != nil {
		t.Logf("Warning: Failed to cleanup admin user: %v", err)
	}
}

// TestStoreInterface ensures both SQLite and PostgreSQL implement the Store interface
func TestStoreInterface(t *testing.T) {
	var _ Store = (*SQLiteStore)(nil)
	var _ Store = (*PostgresStore)(nil)
}

// TestPostgresStore_ParallelOperations tests concurrent access to PostgreSQL
func TestPostgresStore_ParallelOperations(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("Skipping PostgreSQL parallel test: TEST_DATABASE_URL not set")
	}

	store, err := NewPostgresStore(databaseURL)
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	siteID := "parallel-test-site"
	pageID := "parallel-test-page"

	// Add multiple comments in parallel
	numComments := 10
	errChan := make(chan error, numComments)

	for i := 0; i < numComments; i++ {
		go func(idx int) {
			comment := Comment{
				ID:        uuid.New().String(),
				Author:    "Test User",
				AuthorID:  "test-user-1",
				Text:      "Parallel test comment",
				Status:    "pending",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			errChan <- store.AddPageComment(ctx, siteID, pageID, comment)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numComments; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("Failed to add comment in parallel: %v", err)
		}
	}

	// Verify all comments were added
	comments, err := store.GetPageComments(ctx, siteID, pageID)
	if err != nil {
		t.Fatalf("Failed to get comments: %v", err)
	}

	if len(comments) != numComments {
		t.Errorf("Expected %d comments, got %d", numComments, len(comments))
	}

	// Cleanup
	for _, comment := range comments {
		_ = store.DeleteComment(ctx, comment.ID)
	}
	_, _ = store.GetDB().ExecContext(ctx, "DELETE FROM pages WHERE id = $1", pageID)
	_, _ = store.GetDB().ExecContext(ctx, "DELETE FROM sites WHERE id = $1", siteID)
	_, _ = store.GetDB().ExecContext(ctx, "DELETE FROM admin_users WHERE id = 'system'")
}

// TestPostgresStore_NullHandling tests proper handling of NULL values
func TestPostgresStore_NullHandling(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("Skipping PostgreSQL NULL handling test: TEST_DATABASE_URL not set")
	}

	store, err := NewPostgresStore(databaseURL)
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	siteID := "null-test-site"
	pageID := "null-test-page"
	commentID := uuid.New().String()

	// Add a comment with minimal fields (no parent, no email, not moderated)
	comment := Comment{
		ID:        commentID,
		Author:    "Test User",
		AuthorID:  "test-user-1",
		Text:      "Test comment with nulls",
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = store.AddPageComment(ctx, siteID, pageID, comment)
	if err != nil {
		t.Fatalf("Failed to add comment: %v", err)
	}

	// Retrieve and verify NULL fields are handled correctly
	retrievedComment, err := store.GetCommentByID(ctx, commentID)
	if err != nil {
		t.Fatalf("Failed to get comment: %v", err)
	}

	if retrievedComment.ParentID != "" {
		t.Errorf("Expected empty ParentID, got '%s'", retrievedComment.ParentID)
	}

	if retrievedComment.AuthorEmail != "" {
		t.Errorf("Expected empty AuthorEmail, got '%s'", retrievedComment.AuthorEmail)
	}

	if retrievedComment.ModeratedBy != "" {
		t.Errorf("Expected empty ModeratedBy, got '%s'", retrievedComment.ModeratedBy)
	}

	if !retrievedComment.ModeratedAt.IsZero() {
		t.Errorf("Expected zero ModeratedAt, got %v", retrievedComment.ModeratedAt)
	}

	// Cleanup
	_ = store.DeleteComment(ctx, commentID)
	_, _ = store.GetDB().ExecContext(ctx, "DELETE FROM pages WHERE id = $1", pageID)
	_, _ = store.GetDB().ExecContext(ctx, "DELETE FROM sites WHERE id = $1", siteID)
	_, _ = store.GetDB().ExecContext(ctx, "DELETE FROM admin_users WHERE id = 'system'")
}

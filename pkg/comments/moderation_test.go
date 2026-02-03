package comments

import (
	"context"
	"testing"
	"time"
)

func TestSQLiteStore_UpdateCommentStatus(t *testing.T) {
	store, _ := createTestDB(t)
	defer store.Close()

	// Create a comment
	comment := Comment{
		ID:        "1",
		Author:    "John",
		Text:      "Test comment",
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := store.AddPageComment(context.Background(), "site1", "page1", comment)
	if err != nil {
		t.Fatalf("AddPageComment failed: %v", err)
	}

	// Update status to approved
	err = store.UpdateCommentStatus(context.Background(), "1", "approved", "moderator123")
	if err != nil {
		t.Fatalf("UpdateCommentStatus failed: %v", err)
	}

	// Verify status update
	retrieved, err := store.GetCommentByID(context.Background(), "1")
	if err != nil {
		t.Fatalf("GetCommentByID failed: %v", err)
	}

	if retrieved.Status != "approved" {
		t.Errorf("Expected status 'approved', got '%s'", retrieved.Status)
	}
	if retrieved.ModeratedBy != "moderator123" {
		t.Errorf("Expected moderatedBy 'moderator123', got '%s'", retrieved.ModeratedBy)
	}
	if retrieved.ModeratedAt.IsZero() {
		t.Error("Expected non-zero ModeratedAt")
	}
}

func TestSQLiteStore_GetCommentsBySite(t *testing.T) {
	store, _ := createTestDB(t)
	defer store.Close()

	// Create comments with different statuses
	comments := []Comment{
		{ID: "1", Author: "John", Text: "Comment 1", Status: "pending", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "2", Author: "Jane", Text: "Comment 2", Status: "approved", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "3", Author: "Bob", Text: "Comment 3", Status: "rejected", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, c := range comments {
		if err := store.AddPageComment(context.Background(), "site1", "page1", c); err != nil {
			t.Fatalf("AddPageComment failed: %v", err)
		}
	}

	// Get all comments for site
	allComments, err := store.GetCommentsBySite(context.Background(), "site1", "")
	if err != nil {
		t.Fatalf("GetCommentsBySite failed: %v", err)
	}
	if len(allComments) != 3 {
		t.Errorf("Expected 3 comments, got %d", len(allComments))
	}

	// Get only pending comments
	pendingComments, err := store.GetCommentsBySite(context.Background(), "site1", "pending")
	if err != nil {
		t.Fatalf("GetCommentsBySite failed: %v", err)
	}
	if len(pendingComments) != 1 {
		t.Errorf("Expected 1 pending comment, got %d", len(pendingComments))
	}
	if pendingComments[0].Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", pendingComments[0].Status)
	}

	// Get only approved comments
	approvedComments, err := store.GetCommentsBySite(context.Background(), "site1", "approved")
	if err != nil {
		t.Fatalf("GetCommentsBySite failed: %v", err)
	}
	if len(approvedComments) != 1 {
		t.Errorf("Expected 1 approved comment, got %d", len(approvedComments))
	}
}

func TestSQLiteStore_DeleteComment(t *testing.T) {
	store, _ := createTestDB(t)
	defer store.Close()

	// Create a comment
	comment := Comment{
		ID:        "1",
		Author:    "John",
		Text:      "Test comment",
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := store.AddPageComment(context.Background(), "site1", "page1", comment)
	if err != nil {
		t.Fatalf("AddPageComment failed: %v", err)
	}

	// Delete the comment
	err = store.DeleteComment(context.Background(), "1")
	if err != nil {
		t.Fatalf("DeleteComment failed: %v", err)
	}

	// Verify deletion
	_, err = store.GetCommentByID(context.Background(), "1")
	if err == nil {
		t.Error("Expected error when getting deleted comment, got nil")
	}

	// Verify it's not in the page comments either
	comments, err := store.GetPageComments(context.Background(), "site1", "page1")
	if err != nil {
		t.Fatalf("GetPageComments failed: %v", err)
	}
	if len(comments) != 0 {
		t.Errorf("Expected 0 comments after deletion, got %d", len(comments))
	}
}

func TestSQLiteStore_GetCommentByID(t *testing.T) {
	store, _ := createTestDB(t)
	defer store.Close()

	// Create a comment
	comment := Comment{
		ID:        "1",
		Author:    "John",
		Text:      "Test comment",
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := store.AddPageComment(context.Background(), "site1", "page1", comment)
	if err != nil {
		t.Fatalf("AddPageComment failed: %v", err)
	}

	// Get by ID
	retrieved, err := store.GetCommentByID(context.Background(), "1")
	if err != nil {
		t.Fatalf("GetCommentByID failed: %v", err)
	}

	if retrieved.ID != "1" {
		t.Errorf("Expected ID '1', got '%s'", retrieved.ID)
	}
	if retrieved.Author != "John" {
		t.Errorf("Expected author 'John', got '%s'", retrieved.Author)
	}

	// Try non-existent ID
	_, err = store.GetCommentByID(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent ID, got nil")
	}
}

func TestSQLiteStore_CommentDefaultStatus(t *testing.T) {
	store, _ := createTestDB(t)
	defer store.Close()

	// Create comment without status
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

	// Verify default status is "pending"
	retrieved, err := store.GetCommentByID(context.Background(), "1")
	if err != nil {
		t.Fatalf("GetCommentByID failed: %v", err)
	}

	if retrieved.Status != "pending" {
		t.Errorf("Expected default status 'pending', got '%s'", retrieved.Status)
	}
}

func TestSQLiteStore_ModerationWorkflow(t *testing.T) {
	store, _ := createTestDB(t)
	defer store.Close()

	// Create a pending comment
	comment := Comment{
		ID:        "1",
		Author:    "John",
		Text:      "Test comment",
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := store.AddPageComment(context.Background(), "site1", "page1", comment)
	if err != nil {
		t.Fatalf("AddPageComment failed: %v", err)
	}

	// Approve it
	err = store.UpdateCommentStatus(context.Background(), "1", "approved", "moderator1")
	if err != nil {
		t.Fatalf("UpdateCommentStatus (approve) failed: %v", err)
	}

	// Verify it's approved
	retrieved, err := store.GetCommentByID(context.Background(), "1")
	if err != nil {
		t.Fatalf("GetCommentByID failed: %v", err)
	}
	if retrieved.Status != "approved" {
		t.Errorf("Expected status 'approved', got '%s'", retrieved.Status)
	}

	// Reject it (change from approved to rejected)
	err = store.UpdateCommentStatus(context.Background(), "1", "rejected", "moderator2")
	if err != nil {
		t.Fatalf("UpdateCommentStatus (reject) failed: %v", err)
	}

	// Verify it's rejected
	retrieved, err = store.GetCommentByID(context.Background(), "1")
	if err != nil {
		t.Fatalf("GetCommentByID failed: %v", err)
	}
	if retrieved.Status != "rejected" {
		t.Errorf("Expected status 'rejected', got '%s'", retrieved.Status)
	}
	if retrieved.ModeratedBy != "moderator2" {
		t.Errorf("Expected moderatedBy 'moderator2', got '%s'", retrieved.ModeratedBy)
	}
}

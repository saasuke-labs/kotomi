package e2e

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/saasuke-labs/kotomi/pkg/comments"
)

func TestE2E_HealthCheck(t *testing.T) {
	resp, err := http.Get(testBaseURL + "/healthz")
	if err != nil {
		t.Fatalf("failed to get health endpoint: %v", err)
	}
	defer resp.Body.Close()

	AssertStatusCode(t, resp, http.StatusOK)

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result["message"] != "OK" {
		t.Errorf("expected message 'OK', got '%s'", result["message"])
	}
}

func TestE2E_PostComment(t *testing.T) {
	siteID := "e2e-site-1"
	pageID := "e2e-page-1"

	comment := comments.Comment{
		Author: "Test User",
		Text:   "This is an E2E test comment",
	}

	result := PostComment(t, testBaseURL, siteID, pageID, comment)

	// Verify the returned comment has an ID and timestamps
	if result.ID == "" {
		t.Error("expected comment ID to be set, got empty string")
	}
	if result.Author != comment.Author {
		t.Errorf("expected author '%s', got '%s'", comment.Author, result.Author)
	}
	if result.Text != comment.Text {
		t.Errorf("expected text '%s', got '%s'", comment.Text, result.Text)
	}
	if result.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if result.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestE2E_GetComments(t *testing.T) {
	siteID := "e2e-site-2"
	pageID := "e2e-page-2"

	// Post a comment first
	comment1 := comments.Comment{
		Author: "User 1",
		Text:   "First comment",
	}
	posted1 := PostComment(t, testBaseURL, siteID, pageID, comment1)

	comment2 := comments.Comment{
		Author: "User 2",
		Text:   "Second comment",
	}
	posted2 := PostComment(t, testBaseURL, siteID, pageID, comment2)

	// Get comments
	result := GetComments(t, testBaseURL, siteID, pageID)

	// Verify we got both comments
	if len(result) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(result))
	}

	// Verify the comments match
	foundFirst := false
	foundSecond := false
	for _, c := range result {
		if c.ID == posted1.ID {
			foundFirst = true
			if c.Text != comment1.Text {
				t.Errorf("expected first comment text '%s', got '%s'", comment1.Text, c.Text)
			}
		}
		if c.ID == posted2.ID {
			foundSecond = true
			if c.Text != comment2.Text {
				t.Errorf("expected second comment text '%s', got '%s'", comment2.Text, c.Text)
			}
		}
	}

	if !foundFirst {
		t.Error("first comment not found in results")
	}
	if !foundSecond {
		t.Error("second comment not found in results")
	}
}

func TestE2E_GetCommentsForEmptyPage(t *testing.T) {
	siteID := "e2e-site-3"
	pageID := "e2e-page-3"

	result := GetComments(t, testBaseURL, siteID, pageID)

	// Should return an empty array, not null
	if result == nil {
		t.Error("expected empty array, got nil")
	}
	if len(result) != 0 {
		t.Errorf("expected 0 comments, got %d", len(result))
	}
}

func TestE2E_PostCommentWithParent(t *testing.T) {
	siteID := "e2e-site-4"
	pageID := "e2e-page-4"

	// Post parent comment
	parent := comments.Comment{
		Author: "Parent User",
		Text:   "Parent comment",
	}
	parentResult := PostComment(t, testBaseURL, siteID, pageID, parent)

	// Post reply
	reply := comments.Comment{
		Author:   "Reply User",
		Text:     "Reply to parent",
		ParentID: parentResult.ID,
	}
	replyResult := PostComment(t, testBaseURL, siteID, pageID, reply)

	// Verify parent ID is preserved
	if replyResult.ParentID != parentResult.ID {
		t.Errorf("expected ParentID '%s', got '%s'", parentResult.ID, replyResult.ParentID)
	}

	// Get all comments
	result := GetComments(t, testBaseURL, siteID, pageID)

	if len(result) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(result))
	}

	// Find the reply and verify its parent ID
	for _, c := range result {
		if c.ID == replyResult.ID {
			if c.ParentID != parentResult.ID {
				t.Errorf("expected reply ParentID '%s', got '%s'", parentResult.ID, c.ParentID)
			}
			break
		}
	}
}

func TestE2E_CommentsAreSiteIsolated(t *testing.T) {
	// Post comment to site 1
	comment1 := comments.Comment{
		Author: "User 1",
		Text:   "Comment on site 1",
	}
	PostComment(t, testBaseURL, "site-isolation-1", "page-1", comment1)

	// Post comment to site 2
	comment2 := comments.Comment{
		Author: "User 2",
		Text:   "Comment on site 2",
	}
	PostComment(t, testBaseURL, "site-isolation-2", "page-1", comment2)

	// Verify site 1 only has its comment
	site1Comments := GetComments(t, testBaseURL, "site-isolation-1", "page-1")
	if len(site1Comments) != 1 {
		t.Errorf("expected 1 comment for site 1, got %d", len(site1Comments))
	}
	if len(site1Comments) > 0 && site1Comments[0].Text != comment1.Text {
		t.Errorf("expected text '%s', got '%s'", comment1.Text, site1Comments[0].Text)
	}

	// Verify site 2 only has its comment
	site2Comments := GetComments(t, testBaseURL, "site-isolation-2", "page-1")
	if len(site2Comments) != 1 {
		t.Errorf("expected 1 comment for site 2, got %d", len(site2Comments))
	}
	if len(site2Comments) > 0 && site2Comments[0].Text != comment2.Text {
		t.Errorf("expected text '%s', got '%s'", comment2.Text, site2Comments[0].Text)
	}
}

func TestE2E_CommentsArePageIsolated(t *testing.T) {
	siteID := "page-isolation-site"

	// Post comment to page 1
	comment1 := comments.Comment{
		Author: "User 1",
		Text:   "Comment on page 1",
	}
	PostComment(t, testBaseURL, siteID, "page-1", comment1)

	// Post comment to page 2
	comment2 := comments.Comment{
		Author: "User 2",
		Text:   "Comment on page 2",
	}
	PostComment(t, testBaseURL, siteID, "page-2", comment2)

	// Verify page 1 only has its comment
	page1Comments := GetComments(t, testBaseURL, siteID, "page-1")
	if len(page1Comments) != 1 {
		t.Errorf("expected 1 comment for page 1, got %d", len(page1Comments))
	}
	if len(page1Comments) > 0 && page1Comments[0].Text != comment1.Text {
		t.Errorf("expected text '%s', got '%s'", comment1.Text, page1Comments[0].Text)
	}

	// Verify page 2 only has its comment
	page2Comments := GetComments(t, testBaseURL, siteID, "page-2")
	if len(page2Comments) != 1 {
		t.Errorf("expected 1 comment for page 2, got %d", len(page2Comments))
	}
	if len(page2Comments) > 0 && page2Comments[0].Text != comment2.Text {
		t.Errorf("expected text '%s', got '%s'", comment2.Text, page2Comments[0].Text)
	}
}

func TestE2E_CommentsPreserveTimestamps(t *testing.T) {
	siteID := "timestamp-site"
	pageID := "timestamp-page"

	comment := comments.Comment{
		Author: "Test User",
		Text:   "Testing timestamps",
	}

	posted := PostComment(t, testBaseURL, siteID, pageID, comment)

	// Wait a moment
	time.Sleep(100 * time.Millisecond)

	// Get the comment back
	result := GetComments(t, testBaseURL, siteID, pageID)

	if len(result) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(result))
	}

	retrieved := result[0]

	// Verify timestamps match
	if !retrieved.CreatedAt.Equal(posted.CreatedAt) {
		t.Errorf("CreatedAt changed: posted=%v, retrieved=%v", posted.CreatedAt, retrieved.CreatedAt)
	}
	if !retrieved.UpdatedAt.Equal(posted.UpdatedAt) {
		t.Errorf("UpdatedAt changed: posted=%v, retrieved=%v", posted.UpdatedAt, retrieved.UpdatedAt)
	}
}

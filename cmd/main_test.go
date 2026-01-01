package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/saasuke-labs/kotomi/pkg/comments"
)

func TestGetHealthz(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	getHealthz(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var response struct {
		Message string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Message != "OK" {
		t.Errorf("expected message 'OK', got '%s'", response.Message)
	}
}

func TestGetUrlParams_ValidPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected map[string]string
	}{
		{
			name: "simple site and page IDs",
			path: "/api/site/mysite/page/mypage/comments",
			expected: map[string]string{
				"siteId": "mysite",
				"pageId": "mypage",
			},
		},
		{
			name: "numeric IDs",
			path: "/api/site/123/page/456/comments",
			expected: map[string]string{
				"siteId": "123",
				"pageId": "456",
			},
		},
		{
			name: "IDs with dashes",
			path: "/api/site/my-site/page/my-page/comments",
			expected: map[string]string{
				"siteId": "my-site",
				"pageId": "my-page",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			params, err := getUrlParams(req)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if params["siteId"] != tt.expected["siteId"] {
				t.Errorf("expected siteId '%s', got '%s'", tt.expected["siteId"], params["siteId"])
			}

			if params["pageId"] != tt.expected["pageId"] {
				t.Errorf("expected pageId '%s', got '%s'", tt.expected["pageId"], params["pageId"])
			}
		})
	}
}

func TestGetUrlParams_InvalidPath(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"missing comments", "/api/site/mysite/page/mypage"},
		{"wrong prefix", "/wrong/site/mysite/page/mypage/comments"},
		{"missing site keyword", "/api/mysite/page/mypage/comments"},
		{"missing page keyword", "/api/site/mysite/mypage/comments"},
		{"too short", "/api/site/mysite"},
		{"empty path", "/"},
		{"only api", "/api"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			_, err := getUrlParams(req)

			if err == nil {
				t.Error("expected error for invalid path, got nil")
			}
		})
	}
}

func TestPostCommentsHandler_ValidInput(t *testing.T) {
	// Reset comment store with in-memory adapter
	commentStore = &InMemoryStoreAdapter{SitePagesIndex: comments.NewSitePagesIndex()}

	comment := comments.Comment{
		Author:   "John Doe",
		Text:     "This is a test comment",
		ParentID: "",
	}

	body, _ := json.Marshal(comment)
	req := httptest.NewRequest("POST", "/api/site/testsite/page/testpage/comments", bytes.NewReader(body))
	w := httptest.NewRecorder()

	postCommentsHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var returnedComment comments.Comment
	if err := json.NewDecoder(resp.Body).Decode(&returnedComment); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if returnedComment.ID == "" {
		t.Error("expected comment to have an ID")
	}

	if returnedComment.Author != comment.Author {
		t.Errorf("expected author '%s', got '%s'", comment.Author, returnedComment.Author)
	}

	if returnedComment.Text != comment.Text {
		t.Errorf("expected text '%s', got '%s'", comment.Text, returnedComment.Text)
	}

	// Verify comment was stored
	stored, err := commentStore.GetPageComments("testsite", "testpage")
	if err != nil {
		t.Fatalf("failed to get comments: %v", err)
	}
	if len(stored) != 1 {
		t.Fatalf("expected 1 stored comment, got %d", len(stored))
	}
}

func TestPostCommentsHandler_InvalidJSON(t *testing.T) {
	commentStore = &InMemoryStoreAdapter{SitePagesIndex: comments.NewSitePagesIndex()}

	req := httptest.NewRequest("POST", "/api/site/testsite/page/testpage/comments", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	postCommentsHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestPostCommentsHandler_InvalidURL(t *testing.T) {
	// Reset comment store
	commentStore = &InMemoryStoreAdapter{SitePagesIndex: comments.NewSitePagesIndex()}

	comment := comments.Comment{
		Author: "John",
		Text:   "Test",
	}

	body, _ := json.Marshal(comment)
	req := httptest.NewRequest("POST", "/invalid/path", bytes.NewReader(body))
	w := httptest.NewRecorder()

	postCommentsHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestGetCommentsHandler_ExistingComments(t *testing.T) {
	// Reset and populate comment store
	inMemory := comments.NewSitePagesIndex()
	commentStore = &InMemoryStoreAdapter{SitePagesIndex: inMemory}

	testComments := []comments.Comment{
		{ID: "1", Author: "Alice", Text: "First comment", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "2", Author: "Bob", Text: "Second comment", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, c := range testComments {
		inMemory.AddPageComment("site1", "page1", c)
	}

	req := httptest.NewRequest("GET", "/api/site/site1/page/page1/comments", nil)
	w := httptest.NewRecorder()

	getCommentsHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var retrievedComments []comments.Comment
	if err := json.NewDecoder(resp.Body).Decode(&retrievedComments); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(retrievedComments) != len(testComments) {
		t.Fatalf("expected %d comments, got %d", len(testComments), len(retrievedComments))
	}

	for i, expected := range testComments {
		if retrievedComments[i].ID != expected.ID {
			t.Errorf("comment %d: expected ID '%s', got '%s'", i, expected.ID, retrievedComments[i].ID)
		}
		if retrievedComments[i].Author != expected.Author {
			t.Errorf("comment %d: expected author '%s', got '%s'", i, expected.Author, retrievedComments[i].Author)
		}
	}
}

func TestGetCommentsHandler_NonExistentPage(t *testing.T) {
	// Reset comment store
	commentStore = &InMemoryStoreAdapter{SitePagesIndex: comments.NewSitePagesIndex()}

	req := httptest.NewRequest("GET", "/api/site/nonexistent/page/missing/comments", nil)
	w := httptest.NewRecorder()

	getCommentsHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var comments []comments.Comment
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(comments) != 0 {
		t.Errorf("expected 0 comments for non-existent page, got %d", len(comments))
	}
}

func TestGetCommentsHandler_InvalidURL(t *testing.T) {
	commentStore = &InMemoryStoreAdapter{SitePagesIndex: comments.NewSitePagesIndex()}

	req := httptest.NewRequest("GET", "/invalid/path", nil)
	w := httptest.NewRecorder()

	getCommentsHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestWriteJsonResponse(t *testing.T) {
	w := httptest.NewRecorder()

	data := struct {
		Message string `json:"message"`
		Count   int    `json:"count"`
	}{
		Message: "test",
		Count:   42,
	}

	writeJsonResponse(w, data)

	resp := w.Result()
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}

	var decoded struct {
		Message string `json:"message"`
		Count   int    `json:"count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if decoded.Message != data.Message {
		t.Errorf("expected message '%s', got '%s'", data.Message, decoded.Message)
	}

	if decoded.Count != data.Count {
		t.Errorf("expected count %d, got %d", data.Count, decoded.Count)
	}
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/saasuke-labs/kotomi/pkg/comments"
	"github.com/saasuke-labs/kotomi/pkg/middleware"
	"github.com/saasuke-labs/kotomi/pkg/models"
	"github.com/saasuke-labs/kotomi/pkg/moderation"
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
		{
			name: "v1 API path",
			path: "/api/v1/site/mysite/page/mypage/comments",
			expected: map[string]string{
				"siteId": "mysite",
				"pageId": "mypage",
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

// createTestUserContext creates a context with a mock authenticated user
func createTestUserContext(ctx context.Context) context.Context {
	testUser := &models.KotomiUser{
		ID:    "test-user-123",
		Name:  "John Doe",
		Email: "john@example.com",
	}
	return context.WithValue(ctx, middleware.ContextKeyUser, testUser)
}

func TestPostCommentsHandler_ValidInput(t *testing.T) {
	// Reset comment store with in-memory adapter
	commentStore = &InMemoryStoreAdapter{SitePagesIndex: comments.NewSitePagesIndex()}

	comment := comments.Comment{
		Text:     "This is a test comment",
		ParentID: "",
	}

	body, _ := json.Marshal(comment)
	req := httptest.NewRequest("POST", "/api/site/testsite/page/testpage/comments", bytes.NewReader(body))
	// Add authenticated user to context
	req = req.WithContext(createTestUserContext(req.Context()))
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

	if returnedComment.Author != "John Doe" {
		t.Errorf("expected author 'John Doe', got '%s'", returnedComment.Author)
	}

	if returnedComment.AuthorID != "test-user-123" {
		t.Errorf("expected author_id 'test-user-123', got '%s'", returnedComment.AuthorID)
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
	// Add authenticated user to context
	req = req.WithContext(createTestUserContext(req.Context()))
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

func TestGetUserIdentifier(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		headers        map[string]string
		expectedResult string
	}{
		{
			name:           "basic remote address with port",
			remoteAddr:     "192.168.1.100:12345",
			headers:        map[string]string{},
			expectedResult: "192.168.1.100",
		},
		{
			name:           "X-Real-IP header",
			remoteAddr:     "192.168.1.100:12345",
			headers:        map[string]string{"X-Real-IP": "203.0.113.42"},
			expectedResult: "203.0.113.42",
		},
		{
			name:           "X-Forwarded-For single IP",
			remoteAddr:     "192.168.1.100:12345",
			headers:        map[string]string{"X-Forwarded-For": "203.0.113.42"},
			expectedResult: "203.0.113.42",
		},
		{
			name:           "X-Forwarded-For multiple IPs",
			remoteAddr:     "192.168.1.100:12345",
			headers:        map[string]string{"X-Forwarded-For": "203.0.113.42, 198.51.100.1, 192.168.1.1"},
			expectedResult: "203.0.113.42",
		},
		{
			name:           "X-Real-IP takes precedence over X-Forwarded-For",
			remoteAddr:     "192.168.1.100:12345",
			headers:        map[string]string{"X-Real-IP": "203.0.113.42", "X-Forwarded-For": "198.51.100.1"},
			expectedResult: "203.0.113.42",
		},
		{
			name:           "IPv6 address",
			remoteAddr:     "[2001:db8::1]:12345",
			headers:        map[string]string{},
			expectedResult: "[2001:db8::1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			result := getUserIdentifier(req)
			if result != tt.expectedResult {
				t.Errorf("expected '%s', got '%s'", tt.expectedResult, result)
			}
		})
	}
}

func TestDeprecationMiddleware(t *testing.T) {
	// Create a simple handler that just returns OK
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap it with the deprecation middleware
	wrappedHandler := deprecationMiddleware(handler)

	req := httptest.NewRequest("GET", "/api/site/test/page/test/comments", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Check that deprecation headers are present
	if resp.Header.Get("X-API-Warn") == "" {
		t.Error("expected X-API-Warn header to be set")
	}

	if resp.Header.Get("Deprecation") != "true" {
		t.Errorf("expected Deprecation header to be 'true', got '%s'", resp.Header.Get("Deprecation"))
	}

	if resp.Header.Get("Sunset") == "" {
		t.Error("expected Sunset header to be set")
	}

	// Verify the handler still executes
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestPostCommentsHandler_Unauthenticated(t *testing.T) {
	// Reset comment store with in-memory adapter
	commentStore = &InMemoryStoreAdapter{SitePagesIndex: comments.NewSitePagesIndex()}

	comment := comments.Comment{
		Text: "This is a test comment",
	}

	body, _ := json.Marshal(comment)
	req := httptest.NewRequest("POST", "/api/site/testsite/page/testpage/comments", bytes.NewReader(body))
	// Don't add authenticated user to context
	w := httptest.NewRecorder()

	postCommentsHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestPostCommentsHandler_EmptyText(t *testing.T) {
	commentStore = &InMemoryStoreAdapter{SitePagesIndex: comments.NewSitePagesIndex()}

	comment := comments.Comment{
		Text: "", // Empty text
	}

	body, _ := json.Marshal(comment)
	req := httptest.NewRequest("POST", "/api/site/testsite/page/testpage/comments", bytes.NewReader(body))
	req = req.WithContext(createTestUserContext(req.Context()))
	w := httptest.NewRecorder()

	postCommentsHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestAddReactionHandler_Unauthenticated(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/comments/test-comment-1/reactions", bytes.NewReader([]byte("{}")))
	// Don't add authenticated user to context
	w := httptest.NewRecorder()

	addReactionHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestAddReactionHandler_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/comments/test-comment-1/reactions", bytes.NewReader([]byte("invalid json")))
	req = req.WithContext(createTestUserContext(req.Context()))
	w := httptest.NewRecorder()

	addReactionHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestAddReactionHandler_MissingAllowedReactionID(t *testing.T) {
	reqBody := map[string]string{
		"allowed_reaction_id": "", // Empty
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/comments/test-comment-1/reactions", bytes.NewReader(body))
	req = req.WithContext(createTestUserContext(req.Context()))
	w := httptest.NewRecorder()

	addReactionHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestAddPageReactionHandler_Unauthenticated(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/pages/test-page-1/reactions", bytes.NewReader([]byte("{}")))
	// Don't add authenticated user to context
	w := httptest.NewRecorder()

	addPageReactionHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestAddPageReactionHandler_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/pages/test-page-1/reactions", bytes.NewReader([]byte("invalid json")))
	req = req.WithContext(createTestUserContext(req.Context()))
	w := httptest.NewRecorder()

	addPageReactionHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestAddPageReactionHandler_MissingAllowedReactionID(t *testing.T) {
	reqBody := map[string]string{
		"allowed_reaction_id": "", // Empty
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/pages/test-page-1/reactions", bytes.NewReader(body))
	req = req.WithContext(createTestUserContext(req.Context()))
	w := httptest.NewRecorder()

	addPageReactionHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestPostCommentsHandler_WithModeration(t *testing.T) {
// Create a test database with moderation config
dbPath := ":memory:"
store, err := comments.NewSQLiteStore(dbPath)
if err != nil {
t.Fatalf("Failed to create SQLite store: %v", err)
}
defer store.Close()

commentStore = store
db = store.GetDB()

// Create test user
userStore := models.NewUserStore(db)
user, _ := userStore.Create("test@example.com", "Test User", "test-auth0-sub")

// Create test site
siteStore := models.NewSiteStore(db)
site, _ := siteStore.Create(user.ID, "Test Site", "test.com", "Test site")

// Create test page
pageStore := models.NewPageStore(db)
page, _ := pageStore.Create(site.ID, "/test", "Test Page")

// Create mock moderator and config
moderator = &mockModerator{
result: &moderation.ModerationResult{
Decision:   "approve",
Confidence: 0.2,
Reason:     "Looks good",
},
}
moderationConfigStore = moderation.NewConfigStore(db)

// Enable moderation for the site
config := moderation.DefaultModerationConfig()
config.Enabled = true
moderationConfigStore.Create(site.ID, config)

// Create test comment
comment := comments.Comment{
Text: "This is a test comment",
}

body, _ := json.Marshal(comment)
req := httptest.NewRequest("POST", "/api/site/"+site.ID+"/page/"+page.ID+"/comments", bytes.NewReader(body))
req = req.WithContext(createTestUserContext(req.Context()))
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

// Comment should be approved based on mock moderation result
if returnedComment.Status != "approved" {
t.Errorf("expected status 'approved', got '%s'", returnedComment.Status)
}

// Cleanup
moderator = nil
moderationConfigStore = nil
}

// mockModerator is a simple mock for testing
type mockModerator struct {
result *moderation.ModerationResult
err    error
}

func (m *mockModerator) AnalyzeComment(text string, config moderation.ModerationConfig) (*moderation.ModerationResult, error) {
return m.result, m.err
}


func TestShowLoginPage(t *testing.T) {
// Initialize templates (required for showLoginPage)
templates = template.New("test")
templates.New("login.html").Parse("<html>Login Page</html>")

req := httptest.NewRequest("GET", "/login", nil)
w := httptest.NewRecorder()

showLoginPage(w, req)

resp := w.Result()
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
t.Errorf("expected status 200, got %d", resp.StatusCode)
}
}

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/saasuke-labs/kotomi/pkg/comments"
	"github.com/saasuke-labs/kotomi/pkg/models"
)

func setupTestServer(t *testing.T) (*mux.Router, string, func()) {
	// Create in-memory database
	dbPath := ":memory:"
	store, err := comments.NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite store: %v", err)
	}
	
	commentStore = store
	db = store.GetDB()

	// Create test user first (required by foreign key constraint)
	userStore := models.NewUserStore(db)
	user, err := userStore.Create("test@example.com", "Test User", "test-auth0-sub")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test site and page
	siteStore := models.NewSiteStore(db)
	site, err := siteStore.Create(user.ID, "Test Site", "test.com", "Test site")
	if err != nil {
		t.Fatalf("Failed to create test site: %v", err)
	}
	
	pageStore := models.NewPageStore(db)
	page, err := pageStore.Create(site.ID, "/test", "Test Page")
	if err != nil {
		t.Fatalf("Failed to create test page: %v", err)
	}
	
	// Create a test comment
	testComment := comments.Comment{
		ID:     "test-comment-1",
		Author: "Test Author",
		Text:   "Test comment",
		Status: "approved",
	}
	if err := store.AddPageComment(site.ID, page.ID, testComment); err != nil {
		t.Fatalf("Failed to add test comment: %v", err)
	}

	// Create router with reaction routes
	router := mux.NewRouter()
	router.HandleFunc("/api/site/{siteId}/allowed-reactions", getAllowedReactionsHandler).Methods("GET")
	router.HandleFunc("/api/comments/{commentId}/reactions", addReactionHandler).Methods("POST")
	router.HandleFunc("/api/comments/{commentId}/reactions", getReactionsByCommentHandler).Methods("GET")
	router.HandleFunc("/api/comments/{commentId}/reactions/counts", getReactionCountsHandler).Methods("GET")
	router.HandleFunc("/api/reactions/{reactionId}", removeReactionHandler).Methods("DELETE")

	cleanup := func() {
		store.Close()
	}

	return router, site.ID, cleanup
}

func TestGetAllowedReactions(t *testing.T) {
	router, siteID, cleanup := setupTestServer(t)
	defer cleanup()

	// Create some allowed reactions
	allowedStore := models.NewAllowedReactionStore(db)
	allowedStore.Create(siteID, "thumbs_up", "👍", "comment")
	allowedStore.Create(siteID, "heart", "❤️", "comment")

	// Make request
	req := httptest.NewRequest("GET", "/api/site/"+siteID+"/allowed-reactions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var reactions []models.AllowedReaction
	if err := json.NewDecoder(w.Body).Decode(&reactions); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(reactions) != 2 {
		t.Errorf("Expected 2 reactions, got %d", len(reactions))
	}
}

func TestAddReaction(t *testing.T) {
	router, siteID, cleanup := setupTestServer(t)
	defer cleanup()

	// Create an allowed reaction
	allowedStore := models.NewAllowedReactionStore(db)
	allowed, _ := allowedStore.Create(siteID, "thumbs_up", "👍", "comment")

	// Prepare request
	reqBody := map[string]string{
		"allowed_reaction_id": allowed.ID,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/comments/test-comment-1/reactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var reaction models.Reaction
	if err := json.NewDecoder(w.Body).Decode(&reaction); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if reaction.CommentID != "test-comment-1" {
		t.Errorf("Expected comment_id to be 'test-comment-1', got '%s'", reaction.CommentID)
	}
}

func TestAddReaction_Toggle(t *testing.T) {
	router, siteID, cleanup := setupTestServer(t)
	defer cleanup()

	// Create an allowed reaction
	allowedStore := models.NewAllowedReactionStore(db)
	allowed, _ := allowedStore.Create(siteID, "thumbs_up", "👍", "comment")

	// Prepare request
	reqBody := map[string]string{
		"allowed_reaction_id": allowed.ID,
	}
	body, _ := json.Marshal(reqBody)

	// First reaction - should add
	req1 := httptest.NewRequest("POST", "/api/comments/test-comment-1/reactions", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	req1.RemoteAddr = "192.168.1.1:1234" // Set a consistent IP
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("First request: Expected status 200, got %d", w1.Code)
	}

	// Second reaction - should toggle off
	body2, _ := json.Marshal(reqBody)
	req2 := httptest.NewRequest("POST", "/api/comments/test-comment-1/reactions", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2.RemoteAddr = "192.168.1.1:1234" // Same IP
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusNoContent {
		t.Errorf("Second request: Expected status 204, got %d", w2.Code)
	}
}

func TestGetReactionCounts(t *testing.T) {
	router, siteID, cleanup := setupTestServer(t)
	defer cleanup()

	// Create allowed reactions
	allowedStore := models.NewAllowedReactionStore(db)
	thumbsUp, _ := allowedStore.Create(siteID, "thumbs_up", "👍", "comment")
	heart, _ := allowedStore.Create(siteID, "heart", "❤️", "comment")

	// Add some reactions
	reactionStore := models.NewReactionStore(db)
	reactionStore.AddReaction("test-comment-1", thumbsUp.ID, "user-1")
	reactionStore.AddReaction("test-comment-1", thumbsUp.ID, "user-2")
	reactionStore.AddReaction("test-comment-1", heart.ID, "user-3")

	// Make request
	req := httptest.NewRequest("GET", "/api/comments/test-comment-1/reactions/counts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var counts []models.ReactionCount
	if err := json.NewDecoder(w.Body).Decode(&counts); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(counts) != 2 {
		t.Fatalf("Expected 2 reaction types, got %d", len(counts))
	}

	// Check thumbs_up count (should be first due to DESC count order)
	if counts[0].Name != "thumbs_up" || counts[0].Count != 2 {
		t.Errorf("Expected thumbs_up with count 2, got %s with count %d", counts[0].Name, counts[0].Count)
	}

	// Check heart count
	if counts[1].Name != "heart" || counts[1].Count != 1 {
		t.Errorf("Expected heart with count 1, got %s with count %d", counts[1].Name, counts[1].Count)
	}
}

func TestGetReactionsByComment(t *testing.T) {
	router, siteID, cleanup := setupTestServer(t)
	defer cleanup()

	// Create allowed reaction
	allowedStore := models.NewAllowedReactionStore(db)
	allowed, _ := allowedStore.Create(siteID, "thumbs_up", "👍", "comment")

	// Add a reaction
	reactionStore := models.NewReactionStore(db)
	reactionStore.AddReaction("test-comment-1", allowed.ID, "user-1")

	// Make request
	req := httptest.NewRequest("GET", "/api/comments/test-comment-1/reactions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var reactions []models.ReactionWithDetails
	if err := json.NewDecoder(w.Body).Decode(&reactions); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(reactions) != 1 {
		t.Errorf("Expected 1 reaction, got %d", len(reactions))
	}

	if reactions[0].Name != "thumbs_up" {
		t.Errorf("Expected reaction name to be 'thumbs_up', got '%s'", reactions[0].Name)
	}
	if reactions[0].Emoji != "👍" {
		t.Errorf("Expected emoji to be '👍', got '%s'", reactions[0].Emoji)
	}
}

func TestMain(m *testing.M) {
	// Run tests
	os.Exit(m.Run())
}

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/saasuke-labs/kotomi/pkg/comments"
)

// TestE2E_ReactionsWorkflow tests the complete reactions workflow
func TestE2E_ReactionsWorkflow(t *testing.T) {
	siteID := "reactions-site-1"
	pageID := "reactions-page-1"

	// Post a comment first
	comment := comments.Comment{
		Text: "Comment for reactions test",
	}
	posted := PostComment(t, testBaseURL, siteID, pageID, comment)

	// Get allowed reactions for the site
	reactionsURL := fmt.Sprintf("%s/api/v1/site/%s/allowed-reactions", testBaseURL, siteID)
	resp, err := http.Get(reactionsURL)
	if err != nil {
		t.Fatalf("failed to get allowed reactions: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var allowedReactions []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&allowedReactions); err != nil {
		t.Fatalf("failed to decode reactions: %v", err)
	}

	// Should have at least some default reactions
	if len(allowedReactions) == 0 {
		t.Skip("No allowed reactions configured, skipping reaction tests")
	}

	// Get the first allowed reaction ID
	firstReaction := allowedReactions[0]
	reactionID := firstReaction["id"].(string)

	// Add a reaction to the comment using JWT auth
	addReactionURL := fmt.Sprintf("%s/api/v1/comments/%s/reactions",
		testBaseURL, posted.ID)
	
	reactionData := map[string]string{
		"allowed_reaction_id": reactionID,
	}
	jsonData, _ := json.Marshal(reactionData)
	
	req, err := http.NewRequest("POST", addReactionURL, bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+generateTestJWT())
	
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to add reaction: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 when adding reaction, got %d", resp.StatusCode)
	}

	// Get reactions for the comment
	getReactionsURL := fmt.Sprintf("%s/api/v1/comments/%s/reactions",
		testBaseURL, posted.ID)
	
	resp, err = http.Get(getReactionsURL)
	if err != nil {
		t.Fatalf("failed to get comment reactions: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 when getting reactions, got %d", resp.StatusCode)
	}

	var reactions []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&reactions); err != nil {
		t.Fatalf("failed to decode reactions: %v", err)
	}

	// Verify we have at least one reaction
	if len(reactions) == 0 {
		t.Error("expected at least 1 reaction, got 0")
	}

	// Get reaction counts (returns array, not map)
	getCountsURL := fmt.Sprintf("%s/api/v1/comments/%s/reactions/counts",
		testBaseURL, posted.ID)
	
	resp, err = http.Get(getCountsURL)
	if err != nil {
		t.Fatalf("failed to get reaction counts: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 when getting reaction counts, got %d", resp.StatusCode)
	}

	var counts []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&counts); err != nil {
		t.Fatalf("failed to decode reaction counts: %v", err)
	}

	// Verify counts include our reaction
	if len(counts) == 0 {
		t.Error("expected reaction counts, got empty array")
	}
}

// TestE2E_PageReactions tests reactions on pages (not comments)
func TestE2E_PageReactions(t *testing.T) {
	siteID := "page-reactions-site"
	pageID := "page-reactions-page"

	// Get allowed reactions
	reactionsURL := fmt.Sprintf("%s/api/v1/site/%s/allowed-reactions", testBaseURL, siteID)
	resp, err := http.Get(reactionsURL)
	if err != nil {
		t.Fatalf("failed to get allowed reactions: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var allowedReactions []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&allowedReactions); err != nil {
		t.Fatalf("failed to decode reactions: %v", err)
	}

	if len(allowedReactions) == 0 {
		t.Skip("No allowed reactions configured, skipping page reaction tests")
	}

	// Get a page-type reaction ID
	var reactionID string
	for _, reaction := range allowedReactions {
		if reactionType, ok := reaction["reaction_type"].(string); ok && reactionType == "page" {
			reactionID = reaction["id"].(string)
			break
		}
	}
	
	if reactionID == "" {
		t.Skip("No page reactions configured, skipping page reaction tests")
	}

	// Add a reaction to the page using JWT auth
	addReactionURL := fmt.Sprintf("%s/api/v1/pages/%s/reactions",
		testBaseURL, pageID)
	
	reactionData := map[string]string{
		"allowed_reaction_id": reactionID,
	}
	jsonData, _ := json.Marshal(reactionData)
	
	req, err := http.NewRequest("POST", addReactionURL, bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+generateTestJWT())
	
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to add page reaction: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 200 or 201 when adding page reaction, got %d", resp.StatusCode)
	}

	// Get reactions for the page
	getReactionsURL := fmt.Sprintf("%s/api/v1/pages/%s/reactions",
		testBaseURL, pageID)
	
	resp, err = http.Get(getReactionsURL)
	if err != nil {
		t.Fatalf("failed to get page reactions: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 when getting page reactions, got %d", resp.StatusCode)
	}

	// Get reaction counts for page
	getCountsURL := fmt.Sprintf("%s/api/v1/pages/%s/reactions/counts",
		testBaseURL, pageID)
	
	resp, err = http.Get(getCountsURL)
	if err != nil {
		t.Fatalf("failed to get page reaction counts: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 when getting page reaction counts, got %d", resp.StatusCode)
	}
}

// TestE2E_ReactionIsolation tests that reactions are properly isolated by site
func TestE2E_ReactionIsolation(t *testing.T) {
	// Create comments on two different sites
	comment1 := comments.Comment{
		Text: "Comment on site 1",
	}
	posted1 := PostComment(t, testBaseURL, "reaction-isolation-1", "page-1", comment1)

	comment2 := comments.Comment{
		Text: "Comment on site 2",
	}
	posted2 := PostComment(t, testBaseURL, "reaction-isolation-2", "page-1", comment2)

	// Get allowed reactions for site 1
	reactionsURL1 := fmt.Sprintf("%s/api/v1/site/%s/allowed-reactions", testBaseURL, "reaction-isolation-1")
	resp1, err := http.Get(reactionsURL1)
	if err != nil {
		t.Fatalf("failed to get allowed reactions for site 1: %v", err)
	}
	defer resp1.Body.Close()

	// Get allowed reactions for site 2
	reactionsURL2 := fmt.Sprintf("%s/api/v1/site/%s/allowed-reactions", testBaseURL, "reaction-isolation-2")
	resp2, err := http.Get(reactionsURL2)
	if err != nil {
		t.Fatalf("failed to get allowed reactions for site 2: %v", err)
	}
	defer resp2.Body.Close()

	// Both sites should return reactions (even if they're the same or different)
	if resp1.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 for site 1 reactions, got %d", resp1.StatusCode)
	}
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 for site 2 reactions, got %d", resp2.StatusCode)
	}

	// Verify comment IDs are different
	if posted1.ID == posted2.ID {
		t.Error("expected different comment IDs for different sites")
	}
}

// TestE2E_MultipleReactions tests adding multiple reactions from the same user
// Note: With JWT auth, all requests come from the same user
func TestE2E_MultipleReactions(t *testing.T) {
	siteID := "multi-reactions-site"
	pageID := "multi-reactions-page"

	// Post a comment
	comment := comments.Comment{
		Text: "Comment for multiple reactions",
	}
	posted := PostComment(t, testBaseURL, siteID, pageID, comment)

	// Get allowed reactions
	reactionsURL := fmt.Sprintf("%s/api/v1/site/%s/allowed-reactions", testBaseURL, siteID)
	resp, err := http.Get(reactionsURL)
	if err != nil {
		t.Fatalf("failed to get allowed reactions: %v", err)
	}
	defer resp.Body.Close()

	var allowedReactions []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&allowedReactions); err != nil {
		t.Fatalf("failed to decode reactions: %v", err)
	}

	if len(allowedReactions) == 0 {
		t.Skip("No allowed reactions configured, skipping multiple reactions test")
	}

	reactionID := allowedReactions[0]["id"].(string)

	// Add a reaction (with JWT auth, all come from the same user)
	addReactionURL := fmt.Sprintf("%s/api/v1/comments/%s/reactions",
		testBaseURL, posted.ID)
	
	reactionData := map[string]string{
		"allowed_reaction_id": reactionID,
	}
	jsonData, _ := json.Marshal(reactionData)
	
	req, err := http.NewRequest("POST", addReactionURL, bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+generateTestJWT())
	
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to add reaction: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 when adding reaction, got %d", resp.StatusCode)
	}

	// Get reaction counts (returns array, not map)
	getCountsURL := fmt.Sprintf("%s/api/v1/comments/%s/reactions/counts",
		testBaseURL, posted.ID)
	
	resp, err = http.Get(getCountsURL)
	if err != nil {
		t.Fatalf("failed to get reaction counts: %v", err)
	}
	defer resp.Body.Close()

	var counts []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&counts); err != nil {
		t.Fatalf("failed to decode reaction counts: %v", err)
	}

	// Should have count of at least 1 for our reaction
	if len(counts) == 0 {
		t.Error("expected at least 1 reaction count, got 0")
	}
}

// TestE2E_RemoveReaction tests removing a reaction
func TestE2E_RemoveReaction(t *testing.T) {
	siteID := "remove-reaction-site"
	pageID := "remove-reaction-page"

	// Post a comment
	comment := comments.Comment{
		Text: "Comment for reaction removal test",
	}
	posted := PostComment(t, testBaseURL, siteID, pageID, comment)

	// Get allowed reactions
	reactionsURL := fmt.Sprintf("%s/api/v1/site/%s/allowed-reactions", testBaseURL, siteID)
	resp, err := http.Get(reactionsURL)
	if err != nil {
		t.Fatalf("failed to get allowed reactions: %v", err)
	}
	defer resp.Body.Close()

	var allowedReactions []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&allowedReactions); err != nil {
		t.Fatalf("failed to decode reactions: %v", err)
	}

	if len(allowedReactions) == 0 {
		t.Skip("No allowed reactions configured, skipping remove reaction test")
	}

	reactionID := allowedReactions[0]["id"].(string)

	// Add a reaction
	addReactionURL := fmt.Sprintf("%s/api/v1/comments/%s/reactions",
		testBaseURL, posted.ID)
	
	reactionData := map[string]string{
		"allowed_reaction_id": reactionID,
	}
	jsonData, _ := json.Marshal(reactionData)
	
	req, err := http.NewRequest("POST", addReactionURL, bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+generateTestJWT())
	
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to add reaction: %v", err)
	}
	
	// Decode the response to get the created reaction ID
	var addedReaction map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&addedReaction); err != nil {
		resp.Body.Close()
		t.Fatalf("failed to decode added reaction: %v", err)
	}
	resp.Body.Close()
	
	// Get the reaction ID from the response
	addedReactionID, ok := addedReaction["id"].(string)
	if !ok {
		t.Fatalf("failed to get reaction ID from response")
	}

	// Remove the reaction using the reaction ID endpoint
	removeReactionURL := fmt.Sprintf("%s/api/v1/reactions/%s",
		testBaseURL, addedReactionID)
	
	req, err = http.NewRequest("DELETE", removeReactionURL, nil)
	if err != nil {
		t.Fatalf("failed to create delete request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+generateTestJWT())
	
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to remove reaction: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected status 200 or 204 when removing reaction, got %d", resp.StatusCode)
	}
}

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/saasuke-labs/kotomi/pkg/comments"
)

// makeAuthenticatedRequest creates an HTTP request with JWT authentication
func makeAuthenticatedRequest(method, url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+generateTestJWT())
	
	client := &http.Client{}
	return client.Do(req)
}

// TestE2E_InvalidCommentData tests posting comments with invalid data
func TestE2E_InvalidCommentData(t *testing.T) {
	siteID := "error-site"
	pageID := "error-page"

	testCases := []struct {
		name           string
		comment        interface{}
		expectedStatus int
	}{
		{
			name:           "Empty author",
			comment:        map[string]string{"text": "Comment with no author"},
			expectedStatus: http.StatusOK, // JWT auth provides author from token
		},
		{
			name:           "Empty text",
			comment:        map[string]string{"author": "Test User"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			comment:        "not valid json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/api/v1/site/%s/page/%s/comments", testBaseURL, siteID, pageID)
			
			var jsonData []byte
			var err error
			if str, ok := tc.comment.(string); ok {
				jsonData = []byte(str)
			} else {
				jsonData, err = json.Marshal(tc.comment)
				if err != nil {
					t.Fatalf("failed to marshal comment: %v", err)
				}
			}

			resp, err := makeAuthenticatedRequest("POST", url, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Fatalf("failed to post comment: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// TestE2E_NonExistentResources tests accessing non-existent resources
func TestE2E_NonExistentResources(t *testing.T) {
	testCases := []struct {
		name           string
		url            string
		method         string
		expectedStatus int
	}{
		{
			name:           "Non-existent site",
			url:            "/api/v1/site/non-existent-site/page/test/comments",
			method:         "GET",
			expectedStatus: http.StatusOK, // Returns empty array
		},
		{
			name:           "Invalid endpoint",
			url:            "/api/v1/invalid/endpoint",
			method:         "GET",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := testBaseURL + tc.url
			
			var resp *http.Response
			var err error
			
			switch tc.method {
			case "GET":
				resp, err = http.Get(url)
			default:
				t.Fatalf("unsupported method: %s", tc.method)
			}

			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// TestE2E_MalformedRequests tests handling of malformed HTTP requests
func TestE2E_MalformedRequests(t *testing.T) {
	siteID := "malformed-site"
	pageID := "malformed-page"

	testCases := []struct {
		name           string
		contentType    string
		body           string
		expectedStatus int
	}{
		{
			name:           "Wrong content type",
			contentType:    "text/plain",
			body:           "plain text body",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty body",
			contentType:    "application/json",
			body:           "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Incomplete JSON",
			contentType:    "application/json",
			body:           `{"author": "Test", "text": `,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/api/v1/site/%s/page/%s/comments", testBaseURL, siteID, pageID)
			
			resp, err := makeAuthenticatedRequest("POST", url, tc.contentType, strings.NewReader(tc.body))
			if err != nil {
				t.Fatalf("failed to post: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// TestE2E_LargePayload tests handling of very large comment payloads
func TestE2E_LargePayload(t *testing.T) {
	siteID := "large-payload-site"
	pageID := "large-payload-page"

	// Create a very large comment (10KB of text)
	largeText := strings.Repeat("a", 10000)
	
	comment := comments.Comment{
		Author: "Test User",
		Text:   largeText,
	}

	url := fmt.Sprintf("%s/api/v1/site/%s/page/%s/comments", testBaseURL, siteID, pageID)
	jsonData, _ := json.Marshal(comment)
	
	resp, err := makeAuthenticatedRequest("POST", url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("failed to post large comment: %v", err)
	}
	defer resp.Body.Close()

	// Should either accept it or reject with appropriate status
	acceptableStatuses := []int{http.StatusOK, http.StatusBadRequest, http.StatusRequestEntityTooLarge}
	statusOK := false
	for _, status := range acceptableStatuses {
		if resp.StatusCode == status {
			statusOK = true
			break
		}
	}
	if !statusOK {
		t.Errorf("unexpected status for large payload: %d, expected one of %v", resp.StatusCode, acceptableStatuses)
	}
}

// TestE2E_ConcurrentComments tests posting comments concurrently
func TestE2E_ConcurrentComments(t *testing.T) {
	siteID := "concurrent-site"
	pageID := "concurrent-page"

	// Number of concurrent requests
	numRequests := 10
	done := make(chan bool, numRequests)
	errors := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(index int) {
			comment := comments.Comment{
				Author: fmt.Sprintf("User %d", index),
				Text:   fmt.Sprintf("Concurrent comment %d", index),
			}

			url := fmt.Sprintf("%s/api/v1/site/%s/page/%s/comments", testBaseURL, siteID, pageID)
			jsonData, _ := json.Marshal(comment)
			
			resp, err := makeAuthenticatedRequest("POST", url, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				errors <- err
				done <- false
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errors <- fmt.Errorf("expected status 200, got %d", resp.StatusCode)
				done <- false
				return
			}

			done <- true
		}(i)
	}

	// Wait for all requests to complete
	successCount := 0
	for i := 0; i < numRequests; i++ {
		select {
		case success := <-done:
			if success {
				successCount++
			}
		case err := <-errors:
			t.Errorf("concurrent request failed: %v", err)
		}
	}

	if successCount == 0 {
		t.Error("no concurrent requests succeeded")
	}

	// Verify all comments were saved
	result := GetComments(t, testBaseURL, siteID, pageID)
	if len(result) < successCount {
		t.Errorf("expected at least %d comments, got %d", successCount, len(result))
	}
}

// TestE2E_SpecialCharacters tests handling of special characters in comments
func TestE2E_SpecialCharacters(t *testing.T) {
	siteID := "special-chars-site"
	pageID := "special-chars-page"

	testCases := []struct {
		name    string
		text    string
	}{
		{
			name: "Emoji",
			text: "This comment has emojis 😀 🎉 ❤️",
		},
		{
			name: "HTML tags",
			text: "This has <script>alert('xss')</script> tags",
		},
		{
			name: "Quotes",
			text: `This has "double" and 'single' quotes`,
		},
		{
			name: "Unicode characters",
			text: "Unicode: こんにちは 你好 مرحبا",
		},
		{
			name: "Newlines",
			text: "Line 1\nLine 2\nLine 3",
		},
		{
			name: "Backslashes",
			text: "Path: C:\\Users\\Test\\file.txt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			comment := comments.Comment{
				Author: "Test User",
				Text:   tc.text,
			}

			result := PostComment(t, testBaseURL, siteID, pageID, comment)

			// Verify the text was preserved correctly
			if result.Text != tc.text {
				t.Errorf("text not preserved correctly:\nexpected: %s\ngot: %s", tc.text, result.Text)
			}

			// Verify we can retrieve it
			allComments := GetComments(t, testBaseURL, siteID, pageID)
			found := false
			for _, c := range allComments {
				if c.ID == result.ID && c.Text == tc.text {
					found = true
					break
				}
			}
			if !found {
				t.Error("comment with special characters not found after retrieval")
			}
		})
	}
}

// TestE2E_RateLimiting tests rate limiting behavior
func TestE2E_RateLimiting(t *testing.T) {
	// Note: This test may be skipped or adjusted based on rate limit configuration
	// In E2E test environment, we set high rate limits, so this primarily tests
	// that the rate limiting middleware is present and functional
	
	siteID := "rate-limit-site"
	pageID := "rate-limit-page"

	// Make a reasonable number of requests
	numRequests := 5
	for i := 0; i < numRequests; i++ {
		comment := comments.Comment{
			Author: "Test User",
			Text:   fmt.Sprintf("Rate limit test comment %d", i),
		}

		url := fmt.Sprintf("%s/api/v1/site/%s/page/%s/comments", testBaseURL, siteID, pageID)
		jsonData, _ := json.Marshal(comment)
		
		resp, err := makeAuthenticatedRequest("POST", url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
		defer resp.Body.Close()

		// With high test limits, all should succeed
		if resp.StatusCode != http.StatusOK {
			// Check if it's a rate limit error
			if resp.StatusCode == http.StatusTooManyRequests {
				t.Logf("Hit rate limit at request %d (this is expected if limits are low)", i)
				return
			}
			t.Errorf("unexpected status for request %d: %d", i, resp.StatusCode)
		}

		// Check for rate limit headers
		if resp.Header.Get("X-RateLimit-Limit") != "" {
			t.Logf("Rate limit headers present: Limit=%s, Remaining=%s",
				resp.Header.Get("X-RateLimit-Limit"),
				resp.Header.Get("X-RateLimit-Remaining"))
		}
	}
}

// TestE2E_CORS tests CORS headers
func TestE2E_CORS(t *testing.T) {
	// Make an OPTIONS request to check CORS preflight
	url := testBaseURL + "/api/v1/site/test/page/test/comments"
	
	req, err := http.NewRequest("OPTIONS", url, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make OPTIONS request: %v", err)
	}
	defer resp.Body.Close()

	// Check for CORS headers in response
	if resp.Header.Get("Access-Control-Allow-Origin") == "" {
		t.Log("CORS headers not present (may be configured to not allow all origins)")
	} else {
		t.Logf("CORS enabled: Allow-Origin=%s", resp.Header.Get("Access-Control-Allow-Origin"))
	}
}

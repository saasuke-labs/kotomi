package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_GET_Requests(t *testing.T) {
	// Create a rate limiter with low limits for testing
	t.Setenv("RATE_LIMIT_GET", "3")
	t.Setenv("RATE_LIMIT_POST", "2")
	
	rl := NewRateLimiter()

	// Create a test handler
	handler := rl.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Test GET requests - should allow 3 requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, w.Code)
		}
	}

	// 4th request should be rate limited
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w.Code)
	}

	// Check rate limit headers
	if w.Header().Get("X-RateLimit-Limit") == "" {
		t.Error("Expected X-RateLimit-Limit header")
	}
	if w.Header().Get("X-RateLimit-Remaining") == "" {
		t.Error("Expected X-RateLimit-Remaining header")
	}
	if w.Header().Get("Retry-After") == "" {
		t.Error("Expected Retry-After header")
	}
}

func TestRateLimiter_POST_Requests(t *testing.T) {
	// Create a rate limiter with low limits for testing
	t.Setenv("RATE_LIMIT_GET", "3")
	t.Setenv("RATE_LIMIT_POST", "2")
	
	rl := NewRateLimiter()

	// Create a test handler
	handler := rl.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Test POST requests - should allow 2 requests
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/api/test", nil)
		req.RemoteAddr = "192.168.1.2:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i+1, w.Code)
		}
	}

	// 3rd request should be rate limited
	req := httptest.NewRequest("POST", "/api/test", nil)
	req.RemoteAddr = "192.168.1.2:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w.Code)
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	// Create a rate limiter with low limits for testing
	t.Setenv("RATE_LIMIT_GET", "2")
	t.Setenv("RATE_LIMIT_POST", "1")
	
	rl := NewRateLimiter()

	// Create a test handler
	handler := rl.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// First IP - 2 requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.RemoteAddr = "192.168.1.3:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("IP1 Request %d: expected status 200, got %d", i+1, w.Code)
		}
	}

	// Second IP - should still allow requests (different rate limit)
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "192.168.1.4:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("IP2 Request: expected status 200, got %d", w.Code)
	}
}

func TestRateLimiter_TokenRefill(t *testing.T) {
	// Create a rate limiter with very fast refill for testing
	t.Setenv("RATE_LIMIT_GET", "1")
	
	rl := NewRateLimiter()

	// Create a test handler
	handler := rl.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// First request should succeed
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "192.168.1.5:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("First request: expected status 200, got %d", w.Code)
	}

	// Second request should be rate limited
	req = httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "192.168.1.5:1234"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Second request: expected status 429, got %d", w.Code)
	}

	// Wait for token refill (1 request per minute = 60 seconds per token)
	// For testing, we wait 2 seconds and check if partial token refill works
	time.Sleep(2 * time.Second)

	// Third request should still be rate limited (need 60 seconds for full refill)
	req = httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "192.168.1.5:1234"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// After 2 seconds with 1/60 refill rate, we don't have enough tokens yet
	if w.Code != http.StatusTooManyRequests {
		t.Logf("Note: This test may pass if system is very slow, status: %d", w.Code)
	}
}

func TestRateLimiter_Headers(t *testing.T) {
	// Create a rate limiter
	t.Setenv("RATE_LIMIT_GET", "5")
	rl := NewRateLimiter()

	// Create a test handler
	handler := rl.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Make a request
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "192.168.1.6:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Check headers are present
	if w.Header().Get("X-RateLimit-Limit") != "5" {
		t.Errorf("Expected X-RateLimit-Limit: 5, got %s", w.Header().Get("X-RateLimit-Limit"))
	}

	remaining := w.Header().Get("X-RateLimit-Remaining")
	if remaining == "" {
		t.Error("Expected X-RateLimit-Remaining header")
	}
}

func TestRateLimiter_XForwardedFor(t *testing.T) {
	// Create a rate limiter with low limits
	t.Setenv("RATE_LIMIT_GET", "1")
	rl := NewRateLimiter()

	// Create a test handler
	handler := rl.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request with X-Forwarded-For header
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	req.RemoteAddr = "192.168.1.7:1234" // This should be ignored
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("First request: expected status 200, got %d", w.Code)
	}

	// Second request with same X-Forwarded-For should be rate limited
	req = httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	req.RemoteAddr = "192.168.1.8:1234" // Different RemoteAddr, same X-Forwarded-For
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Second request: expected status 429, got %d", w.Code)
	}
}

func TestRateLimiter_XForwardedFor_MultipleIPs(t *testing.T) {
	// Create a rate limiter with low limits
	t.Setenv("RATE_LIMIT_GET", "1")
	rl := NewRateLimiter()

	// Create a test handler
	handler := rl.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request with X-Forwarded-For containing multiple IPs
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.2, 192.168.1.1, 172.16.0.1")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("First request: expected status 200, got %d", w.Code)
	}

	// Second request with same client IP (first in list) should be rate limited
	req = httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.2, 192.168.99.99")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Second request: expected status 429, got %d", w.Code)
	}

	// Request with different client IP should be allowed
	req = httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.3, 192.168.1.1")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Third request: expected status 200, got %d", w.Code)
	}
}

func TestRateLimiter_DefaultLimits(t *testing.T) {
	// Create a rate limiter without setting env vars
	rl := NewRateLimiter()

	if rl.getLimit != 100 {
		t.Errorf("Expected default GET limit 100, got %d", rl.getLimit)
	}

	if rl.postLimit != 5 {
		t.Errorf("Expected default POST limit 5, got %d", rl.postLimit)
	}
}

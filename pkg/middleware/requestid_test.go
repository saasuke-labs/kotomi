package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r)
		if requestID == "" {
			t.Error("Request ID not found in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestIDMiddleware(handler)
	
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	
	middleware.ServeHTTP(w, req)
	
	// Check if request ID is in response headers
	requestID := w.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("X-Request-ID header not set in response")
	}
}

func TestRequestIDMiddleware_WithExistingRequestID(t *testing.T) {
	existingID := "existing-request-id-123"
	var capturedID string
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r)
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestIDMiddleware(handler)
	
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", existingID)
	w := httptest.NewRecorder()
	
	middleware.ServeHTTP(w, req)
	
	// Check if existing request ID is preserved
	responseID := w.Header().Get("X-Request-ID")
	if responseID != existingID {
		t.Errorf("Expected request ID %v, got %v", existingID, responseID)
	}
	
	if capturedID != existingID {
		t.Errorf("Expected context request ID %v, got %v", existingID, capturedID)
	}
}

func TestGetRequestID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r)
		if requestID == "" {
			t.Error("GetRequestID returned empty string")
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestIDMiddleware(handler)
	
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	
	middleware.ServeHTTP(w, req)
}

func TestGetRequestID_NoRequestID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	
	requestID := GetRequestID(req)
	if requestID != "" {
		t.Errorf("Expected empty request ID, got %v", requestID)
	}
}

func TestRequestIDMiddleware_UniqueIDs(t *testing.T) {
	var id1, id2 string
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestIDMiddleware(handler)
	
	// First request
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	w1 := httptest.NewRecorder()
	middleware.ServeHTTP(w1, req1)
	id1 = w1.Header().Get("X-Request-ID")
	
	// Second request
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	w2 := httptest.NewRecorder()
	middleware.ServeHTTP(w2, req2)
	id2 = w2.Header().Get("X-Request-ID")
	
	// IDs should be different
	if id1 == id2 {
		t.Errorf("Request IDs should be unique, both are %v", id1)
	}
	
	if id1 == "" || id2 == "" {
		t.Error("Request IDs should not be empty")
	}
}

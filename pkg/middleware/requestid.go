package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// ContextKeyRequestID is the context key for storing request IDs
const ContextKeyRequestID ContextKey = "request_id"

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request ID is already present in headers (e.g., from load balancer)
		requestID := r.Header.Get("X-Request-ID")
		
		// Generate a new request ID if not present
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)
		
		// Add request ID to context
		ctx := context.WithValue(r.Context(), ContextKeyRequestID, requestID)
		
		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from the context
func GetRequestID(r *http.Request) string {
	if requestID, ok := r.Context().Value(ContextKeyRequestID).(string); ok {
		return requestID
	}
	return ""
}

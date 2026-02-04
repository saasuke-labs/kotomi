package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/saasuke-labs/kotomi/pkg/logging"
)

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
		
		// Add request ID to context using the logging package
		ctx := logging.WithRequestID(r.Context(), requestID)
		
		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from the request's context
func GetRequestID(r *http.Request) string {
	return logging.GetRequestID(r.Context())
}

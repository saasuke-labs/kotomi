package middleware

import (
	"os"
	"strings"

	"github.com/rs/cors"
)

// NewCORSMiddleware creates a new CORS middleware with configuration from environment variables
func NewCORSMiddleware() *cors.Cors {
	// Get CORS configuration from environment variables
	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	allowedMethods := os.Getenv("CORS_ALLOWED_METHODS")
	allowedHeaders := os.Getenv("CORS_ALLOWED_HEADERS")
	allowCredentials := os.Getenv("CORS_ALLOW_CREDENTIALS")

	// Set defaults if not provided
	if allowedOrigins == "" {
		// For development, allow all origins by default
		// In production, this should be set to specific origins
		allowedOrigins = "*"
	}

	if allowedMethods == "" {
		allowedMethods = "GET,POST,PUT,DELETE,OPTIONS"
	}

	if allowedHeaders == "" {
		allowedHeaders = "Content-Type,Authorization"
	}

	// Parse allowed origins
	var origins []string
	if allowedOrigins == "*" {
		origins = []string{"*"}
	} else {
		origins = strings.Split(allowedOrigins, ",")
		// Trim whitespace from each origin
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
	}

	// Parse allowed methods
	methods := strings.Split(allowedMethods, ",")
	for i := range methods {
		methods[i] = strings.TrimSpace(methods[i])
	}

	// Parse allowed headers
	headers := strings.Split(allowedHeaders, ",")
	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}

	// Parse allow credentials
	credentials := allowCredentials == "true"

	// Create CORS handler
	c := cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   methods,
		AllowedHeaders:   headers,
		AllowCredentials: credentials,
		// Enable preflight caching for 12 hours
		MaxAge: 43200,
	})

	return c
}

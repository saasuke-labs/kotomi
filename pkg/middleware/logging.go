package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp  string                 `json:"timestamp"`
	Level      LogLevel               `json:"level"`
	Message    string                 `json:"message"`
	RequestID  string                 `json:"request_id,omitempty"`
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	StatusCode int                    `json:"status_code,omitempty"`
	Duration   string                 `json:"duration,omitempty"`
	RemoteAddr string                 `json:"remote_addr,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// Logger is a structured logger
type Logger struct {
	output *log.Logger
}

// NewLogger creates a new structured logger
func NewLogger() *Logger {
	return &Logger{
		output: log.New(os.Stdout, "", 0),
	}
}

// writeLog writes a log entry as JSON
func (l *Logger) writeLog(entry LogEntry) {
	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		// Fallback to standard logging if JSON encoding fails
		l.output.Printf("ERROR: Failed to marshal log entry: %v", err)
		return
	}
	l.output.Println(string(jsonBytes))
}

// Log writes a structured log entry
func (l *Logger) Log(level LogLevel, message string, requestID string, extra map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		RequestID: requestID,
		Extra:     extra,
	}
	l.writeLog(entry)
}

// Debug logs a debug message
func (l *Logger) Debug(message string, requestID string) {
	l.Log(LogLevelDebug, message, requestID, nil)
}

// Info logs an info message
func (l *Logger) Info(message string, requestID string) {
	l.Log(LogLevelInfo, message, requestID, nil)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, requestID string) {
	l.Log(LogLevelWarn, message, requestID, nil)
}

// Error logs an error message
func (l *Logger) Error(message string, requestID string, err error) {
	extra := map[string]interface{}{}
	if err != nil {
		extra["error_detail"] = err.Error()
	}
	l.Log(LogLevelError, message, requestID, extra)
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default status code
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// LoggingMiddleware logs all HTTP requests and responses
func LoggingMiddleware(logger *Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Get request ID from context
			requestID := GetRequestID(r)
			
			// Create response writer wrapper to capture status code
			rw := newResponseWriter(w)
			
			// Get remote address (prefer X-Forwarded-For or X-Real-IP)
			remoteAddr := r.Header.Get("X-Forwarded-For")
			if remoteAddr == "" {
				remoteAddr = r.Header.Get("X-Real-IP")
			}
			if remoteAddr == "" {
				remoteAddr = r.RemoteAddr
			}
			
			// Remove sensitive headers from logging
			userAgent := r.UserAgent()
			
			// Call the next handler
			next.ServeHTTP(rw, r)
			
			// Calculate duration
			duration := time.Since(start)
			
			// Determine log level based on status code
			var level LogLevel
			statusCode := rw.statusCode
			switch {
			case statusCode >= 500:
				level = LogLevelError
			case statusCode >= 400:
				level = LogLevelWarn
			default:
				level = LogLevelInfo
			}
			
			// Log the request/response
			entry := LogEntry{
				Timestamp:  time.Now().UTC().Format(time.RFC3339),
				Level:      level,
				Message:    "HTTP Request",
				RequestID:  requestID,
				Method:     r.Method,
				Path:       sanitizePath(r.URL.Path),
				StatusCode: statusCode,
				Duration:   duration.String(),
				RemoteAddr: sanitizeIP(remoteAddr),
				UserAgent:  userAgent,
			}
			
			logger.writeLog(entry)
		})
	}
}

// sanitizePath removes sensitive information from the path
func sanitizePath(path string) string {
	// Don't log query parameters that might contain sensitive data
	if idx := strings.Index(path, "?"); idx != -1 {
		return path[:idx]
	}
	return path
}

// sanitizeIP removes port from IP address for cleaner logs
func sanitizeIP(addr string) string {
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		// Check if it's an IPv6 address
		if strings.Count(addr, ":") > 1 {
			return addr
		}
		return addr[:idx]
	}
	return addr
}

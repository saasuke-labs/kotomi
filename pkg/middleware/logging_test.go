package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger()
	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}
	if logger.output == nil {
		t.Fatal("Logger output is nil")
	}
}

func TestLogger_Log(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	logger := &Logger{
		output: log.New(&buf, "", 0),
	}
	
	logger.Log(LogLevelInfo, "Test message", "request-123", map[string]interface{}{
		"key": "value",
	})
	
	output := buf.String()
	
	// Parse JSON
	var entry LogEntry
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Failed to parse log output as JSON: %v", err)
	}
	
	if entry.Level != LogLevelInfo {
		t.Errorf("Expected level INFO, got %v", entry.Level)
	}
	if entry.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got %v", entry.Message)
	}
	if entry.RequestID != "request-123" {
		t.Errorf("Expected request ID 'request-123', got %v", entry.RequestID)
	}
	if entry.Extra["key"] != "value" {
		t.Errorf("Expected extra key 'value', got %v", entry.Extra["key"])
	}
}

func TestLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		output: log.New(&buf, "", 0),
	}
	
	logger.Debug("Debug message", "request-123")
	
	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}
	
	if entry.Level != LogLevelDebug {
		t.Errorf("Expected level DEBUG, got %v", entry.Level)
	}
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		output: log.New(&buf, "", 0),
	}
	
	logger.Info("Info message", "request-123")
	
	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}
	
	if entry.Level != LogLevelInfo {
		t.Errorf("Expected level INFO, got %v", entry.Level)
	}
}

func TestLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		output: log.New(&buf, "", 0),
	}
	
	logger.Warn("Warning message", "request-123")
	
	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}
	
	if entry.Level != LogLevelWarn {
		t.Errorf("Expected level WARN, got %v", entry.Level)
	}
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		output: log.New(&buf, "", 0),
	}
	
	testErr := errors.New("test error")
	logger.Error("Error message", "request-123", testErr)
	
	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}
	
	if entry.Level != LogLevelError {
		t.Errorf("Expected level ERROR, got %v", entry.Level)
	}
	if entry.Extra["error_detail"] != "test error" {
		t.Errorf("Expected error_detail 'test error', got %v", entry.Extra["error_detail"])
	}
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	rw := newResponseWriter(w)
	
	rw.WriteHeader(http.StatusNotFound)
	
	if rw.statusCode != http.StatusNotFound {
		t.Errorf("Expected status code %v, got %v", http.StatusNotFound, rw.statusCode)
	}
	
	// Verify it was written to the underlying writer
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected underlying writer status %v, got %v", http.StatusNotFound, w.Code)
	}
}

func TestResponseWriter_Write(t *testing.T) {
	w := httptest.NewRecorder()
	rw := newResponseWriter(w)
	
	data := []byte("test data")
	n, err := rw.Write(data)
	
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %v bytes, wrote %v", len(data), n)
	}
	if rw.statusCode != http.StatusOK {
		t.Errorf("Expected default status code %v, got %v", http.StatusOK, rw.statusCode)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		output: log.New(&buf, "", 0),
	}
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	middleware := LoggingMiddleware(logger)(handler)
	middleware = RequestIDMiddleware(middleware) // Add request ID middleware
	
	req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	w := httptest.NewRecorder()
	
	middleware.ServeHTTP(w, req)
	
	// Parse the log output
	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse log output: %v, output: %s", err, buf.String())
	}
	
	if entry.Message != "HTTP Request" {
		t.Errorf("Expected message 'HTTP Request', got %v", entry.Message)
	}
	if entry.Method != http.MethodGet {
		t.Errorf("Expected method GET, got %v", entry.Method)
	}
	if entry.Path != "/test/path" {
		t.Errorf("Expected path '/test/path', got %v", entry.Path)
	}
	if entry.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %v", entry.StatusCode)
	}
	if entry.Level != LogLevelInfo {
		t.Errorf("Expected level INFO, got %v", entry.Level)
	}
	if entry.RequestID == "" {
		t.Error("Expected request ID to be set")
	}
}

func TestLoggingMiddleware_ErrorStatus(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		output: log.New(&buf, "", 0),
	}
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	
	middleware := LoggingMiddleware(logger)(handler)
	middleware = RequestIDMiddleware(middleware)
	
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	
	middleware.ServeHTTP(w, req)
	
	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}
	
	if entry.Level != LogLevelError {
		t.Errorf("Expected level ERROR for 5xx status, got %v", entry.Level)
	}
	if entry.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code 500, got %v", entry.StatusCode)
	}
}

func TestLoggingMiddleware_WarnStatus(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		output: log.New(&buf, "", 0),
	}
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	
	middleware := LoggingMiddleware(logger)(handler)
	middleware = RequestIDMiddleware(middleware)
	
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	
	middleware.ServeHTTP(w, req)
	
	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}
	
	if entry.Level != LogLevelWarn {
		t.Errorf("Expected level WARN for 4xx status, got %v", entry.Level)
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/api/test", "/api/test"},
		{"/api/test?token=secret", "/api/test"},
		{"/api/test?key1=value1&key2=value2", "/api/test"},
		{"/", "/"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizePath(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizePath(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeIP(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"192.168.1.1:12345", "192.168.1.1"},
		{"192.168.1.1", "192.168.1.1"},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", "2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
		{"::1", "::1"},
		{"[2001:0db8:85a3::8a2e:0370:7334]:8080", "2001:0db8:85a3::8a2e:0370:7334"},
		{"[::1]:8080", "::1"},
		{"[fe80::1]:9090", "fe80::1"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeIP(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeIP(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLoggingMiddleware_WithXForwardedFor(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		output: log.New(&buf, "", 0),
	}
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	middleware := LoggingMiddleware(logger)(handler)
	middleware = RequestIDMiddleware(middleware)
	
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.195")
	w := httptest.NewRecorder()
	
	middleware.ServeHTTP(w, req)
	
	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}
	
	if !strings.Contains(entry.RemoteAddr, "203.0.113.195") {
		t.Errorf("Expected remote addr to contain X-Forwarded-For value, got %v", entry.RemoteAddr)
	}
}

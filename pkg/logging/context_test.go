package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestWithAndGetRequestID(t *testing.T) {
	ctx := context.Background()
	requestID := "test-request-123"
	
	// Add request ID to context
	ctx = WithRequestID(ctx, requestID)
	
	// Retrieve request ID from context
	got := GetRequestID(ctx)
	if got != requestID {
		t.Errorf("GetRequestID() = %v, want %v", got, requestID)
	}
}

func TestGetRequestID_Empty(t *testing.T) {
	ctx := context.Background()
	
	// Should return empty string when not set
	got := GetRequestID(ctx)
	if got != "" {
		t.Errorf("GetRequestID() = %v, want empty string", got)
	}
}

func TestWithAndGetSiteID(t *testing.T) {
	ctx := context.Background()
	siteID := "test-site-456"
	
	ctx = WithSiteID(ctx, siteID)
	got := GetSiteID(ctx)
	
	if got != siteID {
		t.Errorf("GetSiteID() = %v, want %v", got, siteID)
	}
}

func TestWithAndGetPageID(t *testing.T) {
	ctx := context.Background()
	pageID := "test-page-789"
	
	ctx = WithPageID(ctx, pageID)
	got := GetPageID(ctx)
	
	if got != pageID {
		t.Errorf("GetPageID() = %v, want %v", got, pageID)
	}
}

func TestWithAndGetUserID(t *testing.T) {
	ctx := context.Background()
	userID := "test-user-abc"
	
	ctx = WithUserID(ctx, userID)
	got := GetUserID(ctx)
	
	if got != userID {
		t.Errorf("GetUserID() = %v, want %v", got, userID)
	}
}

func TestWithAndGetCommentID(t *testing.T) {
	ctx := context.Background()
	commentID := "test-comment-def"
	
	ctx = WithCommentID(ctx, commentID)
	got := GetCommentID(ctx)
	
	if got != commentID {
		t.Errorf("GetCommentID() = %v, want %v", got, commentID)
	}
}

func TestContextHandler_AutomaticallyAddsContextFields(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	
	// Create a JSON handler that writes to our buffer
	jsonHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	
	// Wrap it with our ContextHandler
	contextHandler := NewContextHandler(jsonHandler)
	logger := slog.New(contextHandler)
	
	// Create context with various fields
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")
	ctx = WithSiteID(ctx, "site-456")
	ctx = WithUserID(ctx, "user-789")
	
	// Log with context - should automatically include context fields
	logger.InfoContext(ctx, "test message", "extra_field", "extra_value")
	
	// Parse the JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v, output: %s", err, buf.String())
	}
	
	// Verify the log contains all context fields
	if logEntry["msg"] != "test message" {
		t.Errorf("Expected message 'test message', got %v", logEntry["msg"])
	}
	if logEntry["request_id"] != "req-123" {
		t.Errorf("Expected request_id 'req-123', got %v", logEntry["request_id"])
	}
	if logEntry["site_id"] != "site-456" {
		t.Errorf("Expected site_id 'site-456', got %v", logEntry["site_id"])
	}
	if logEntry["user_id"] != "user-789" {
		t.Errorf("Expected user_id 'user-789', got %v", logEntry["user_id"])
	}
	if logEntry["extra_field"] != "extra_value" {
		t.Errorf("Expected extra_field 'extra_value', got %v", logEntry["extra_field"])
	}
}

func TestContextHandler_OnlyAddsFieldsThatExist(t *testing.T) {
	var buf bytes.Buffer
	jsonHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	contextHandler := NewContextHandler(jsonHandler)
	logger := slog.New(contextHandler)
	
	// Context with only request_id
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-only")
	
	logger.InfoContext(ctx, "test message")
	
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}
	
	// Should have request_id
	if logEntry["request_id"] != "req-only" {
		t.Errorf("Expected request_id 'req-only', got %v", logEntry["request_id"])
	}
	
	// Should not have other fields
	if _, exists := logEntry["site_id"]; exists {
		t.Error("site_id should not be present when not set in context")
	}
	if _, exists := logEntry["user_id"]; exists {
		t.Error("user_id should not be present when not set in context")
	}
}

func TestContextHandler_WorksWithoutContext(t *testing.T) {
	var buf bytes.Buffer
	jsonHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	contextHandler := NewContextHandler(jsonHandler)
	logger := slog.New(contextHandler)
	
	// Log without context (using plain Info, not InfoContext)
	logger.Info("test message", "field", "value")
	
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}
	
	// Should work fine without context fields
	if logEntry["msg"] != "test message" {
		t.Errorf("Expected message 'test message', got %v", logEntry["msg"])
	}
	if logEntry["field"] != "value" {
		t.Errorf("Expected field 'value', got %v", logEntry["field"])
	}
}

func TestLoggerFromContext_WithAllFields(t *testing.T) {
	var buf bytes.Buffer
	baseLogger := slog.New(slog.NewJSONHandler(&buf, nil))
	
	// Create context with all fields
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")
	ctx = WithSiteID(ctx, "site-456")
	ctx = WithPageID(ctx, "page-789")
	ctx = WithUserID(ctx, "user-abc")
	ctx = WithCommentID(ctx, "comment-def")
	
	// Create logger from context
	logger := LoggerFromContext(ctx, baseLogger)
	
	// Log a message (doesn't need context anymore, fields are pre-filled)
	logger.Info("test message")
	
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}
	
	// Verify all fields are present
	tests := []struct {
		field string
		want  string
	}{
		{"request_id", "req-123"},
		{"site_id", "site-456"},
		{"page_id", "page-789"},
		{"user_id", "user-abc"},
		{"comment_id", "comment-def"},
	}
	
	for _, tt := range tests {
		if logEntry[tt.field] != tt.want {
			t.Errorf("%s = %v, want %v", tt.field, logEntry[tt.field], tt.want)
		}
	}
}

func TestLoggerFromContext_WithNoFields(t *testing.T) {
	var buf bytes.Buffer
	baseLogger := slog.New(slog.NewJSONHandler(&buf, nil))
	
	// Empty context
	ctx := context.Background()
	
	// Should return the same logger when no context fields
	logger := LoggerFromContext(ctx, baseLogger)
	
	// Log a message
	logger.Info("test message")
	
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}
	
	// Should not have context fields
	contextFields := []string{"request_id", "site_id", "page_id", "user_id", "comment_id"}
	for _, field := range contextFields {
		if _, exists := logEntry[field]; exists {
			t.Errorf("%s should not be present when not set in context", field)
		}
	}
}

func TestContextHandler_WithAttrs(t *testing.T) {
	var buf bytes.Buffer
	jsonHandler := slog.NewJSONHandler(&buf, nil)
	contextHandler := NewContextHandler(jsonHandler)
	
	// Create a logger with pre-set attributes
	loggerWithAttrs := slog.New(contextHandler.WithAttrs([]slog.Attr{
		slog.String("service", "test-service"),
	}))
	
	ctx := WithRequestID(context.Background(), "req-123")
	loggerWithAttrs.InfoContext(ctx, "test message")
	
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}
	
	// Should have both the pre-set attribute and context field
	if logEntry["service"] != "test-service" {
		t.Errorf("Expected service 'test-service', got %v", logEntry["service"])
	}
	if logEntry["request_id"] != "req-123" {
		t.Errorf("Expected request_id 'req-123', got %v", logEntry["request_id"])
	}
}

func TestContextHandler_WithGroup(t *testing.T) {
	var buf bytes.Buffer
	jsonHandler := slog.NewJSONHandler(&buf, nil)
	contextHandler := NewContextHandler(jsonHandler)
	
	// Create a logger with a group
	logger := slog.New(contextHandler.WithGroup("mygroup"))
	
	ctx := WithRequestID(context.Background(), "req-123")
	logger.InfoContext(ctx, "test message", "field", "value")
	
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}
	
	// When using groups in slog, context fields added via AddAttrs in Handle()
	// will also be grouped. This is expected behavior - the group wraps everything.
	// So we check that the group exists and contains our field
	mygroup, ok := logEntry["mygroup"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected mygroup to be present and be a map, got %v", logEntry)
	}
	
	// The grouped field should be present
	if mygroup["field"] != "value" {
		t.Errorf("Expected field 'value' in mygroup, got %v", mygroup["field"])
	}
	
	// Context fields added via AddAttrs will also be in the group when using WithGroup
	if mygroup["request_id"] != "req-123" {
		t.Errorf("Expected request_id 'req-123' in mygroup, got %v", mygroup["request_id"])
	}
}

func TestMultipleContextValues(t *testing.T) {
	// Test that multiple context values can be chained
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-1")
	ctx = WithSiteID(ctx, "site-2")
	ctx = WithPageID(ctx, "page-3")
	ctx = WithUserID(ctx, "user-4")
	ctx = WithCommentID(ctx, "comment-5")
	
	// All values should be retrievable
	if GetRequestID(ctx) != "req-1" {
		t.Error("Failed to get request ID")
	}
	if GetSiteID(ctx) != "site-2" {
		t.Error("Failed to get site ID")
	}
	if GetPageID(ctx) != "page-3" {
		t.Error("Failed to get page ID")
	}
	if GetUserID(ctx) != "user-4" {
		t.Error("Failed to get user ID")
	}
	if GetCommentID(ctx) != "comment-5" {
		t.Error("Failed to get comment ID")
	}
}

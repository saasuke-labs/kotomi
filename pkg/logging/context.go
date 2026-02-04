package logging

import (
	"context"
	"log/slog"
)

// ContextKey is the type for context keys to avoid collisions
type ContextKey string

const (
	// ContextKeyRequestID is the context key for request ID
	ContextKeyRequestID ContextKey = "request_id"
	// ContextKeySiteID is the context key for site ID
	ContextKeySiteID ContextKey = "site_id"
	// ContextKeyPageID is the context key for page ID
	ContextKeyPageID ContextKey = "page_id"
	// ContextKeyUserID is the context key for user ID
	ContextKeyUserID ContextKey = "user_id"
	// ContextKeyCommentID is the context key for comment ID
	ContextKeyCommentID ContextKey = "comment_id"
)

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ContextKeyRequestID, requestID)
}

// GetRequestID retrieves the request ID from the context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(ContextKeyRequestID).(string); ok {
		return requestID
	}
	return ""
}

// WithSiteID adds a site ID to the context
func WithSiteID(ctx context.Context, siteID string) context.Context {
	return context.WithValue(ctx, ContextKeySiteID, siteID)
}

// GetSiteID retrieves the site ID from the context
func GetSiteID(ctx context.Context) string {
	if siteID, ok := ctx.Value(ContextKeySiteID).(string); ok {
		return siteID
	}
	return ""
}

// WithPageID adds a page ID to the context
func WithPageID(ctx context.Context, pageID string) context.Context {
	return context.WithValue(ctx, ContextKeyPageID, pageID)
}

// GetPageID retrieves the page ID from the context
func GetPageID(ctx context.Context) string {
	if pageID, ok := ctx.Value(ContextKeyPageID).(string); ok {
		return pageID
	}
	return ""
}

// WithUserID adds a user ID to the context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ContextKeyUserID, userID)
}

// GetUserID retrieves the user ID from the context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(ContextKeyUserID).(string); ok {
		return userID
	}
	return ""
}

// WithCommentID adds a comment ID to the context
func WithCommentID(ctx context.Context, commentID string) context.Context {
	return context.WithValue(ctx, ContextKeyCommentID, commentID)
}

// GetCommentID retrieves the comment ID from the context
func GetCommentID(ctx context.Context) string {
	if commentID, ok := ctx.Value(ContextKeyCommentID).(string); ok {
		return commentID
	}
	return ""
}

// ContextHandler is a slog.Handler that automatically includes contextual fields from context
type ContextHandler struct {
	handler slog.Handler
}

// NewContextHandler creates a new context-aware slog handler
func NewContextHandler(handler slog.Handler) *ContextHandler {
	return &ContextHandler{handler: handler}
}

// Enabled implements slog.Handler
func (h *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle implements slog.Handler - extracts contextual fields and adds them to the log record
func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	// Extract contextual fields from context and add them to the record
	if requestID := GetRequestID(ctx); requestID != "" {
		r.AddAttrs(slog.String("request_id", requestID))
	}
	if siteID := GetSiteID(ctx); siteID != "" {
		r.AddAttrs(slog.String("site_id", siteID))
	}
	if pageID := GetPageID(ctx); pageID != "" {
		r.AddAttrs(slog.String("page_id", pageID))
	}
	if userID := GetUserID(ctx); userID != "" {
		r.AddAttrs(slog.String("user_id", userID))
	}
	if commentID := GetCommentID(ctx); commentID != "" {
		r.AddAttrs(slog.String("comment_id", commentID))
	}
	
	return h.handler.Handle(ctx, r)
}

// WithAttrs implements slog.Handler
func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{handler: h.handler.WithAttrs(attrs)}
}

// WithGroup implements slog.Handler
func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{handler: h.handler.WithGroup(name)}
}

// LoggerFromContext creates a logger that will use context values when logging
// This is a convenience function to get a logger bound to a specific context
func LoggerFromContext(ctx context.Context, baseLogger *slog.Logger) *slog.Logger {
	// Create a child logger with pre-filled context values
	var attrs []any
	
	if requestID := GetRequestID(ctx); requestID != "" {
		attrs = append(attrs, "request_id", requestID)
	}
	if siteID := GetSiteID(ctx); siteID != "" {
		attrs = append(attrs, "site_id", siteID)
	}
	if pageID := GetPageID(ctx); pageID != "" {
		attrs = append(attrs, "page_id", pageID)
	}
	if userID := GetUserID(ctx); userID != "" {
		attrs = append(attrs, "user_id", userID)
	}
	if commentID := GetCommentID(ctx); commentID != "" {
		attrs = append(attrs, "comment_id", commentID)
	}
	
	if len(attrs) > 0 {
		return baseLogger.With(attrs...)
	}
	
	return baseLogger
}

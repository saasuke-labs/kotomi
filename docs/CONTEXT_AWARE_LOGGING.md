# Context-Aware Logging in Kotomi

## Overview

Kotomi uses context-aware structured logging to automatically include contextual fields (like `request_id`, `site_id`, `page_id`, etc.) in log messages. This eliminates the need to manually add these fields to every log call and ensures consistent logging across the application.

## Implementation

### Key Components

1. **`pkg/logging/context.go`**: Provides the core context-aware logging infrastructure
2. **`ContextHandler`**: A custom `slog.Handler` that automatically extracts contextual fields from the context
3. **Context helper functions**: Functions to add and retrieve contextual fields from context

### Contextual Fields

The following contextual fields are automatically propagated through context:

- `request_id`: Unique identifier for each HTTP request
- `site_id`: Site identifier from URL parameters
- `page_id`: Page identifier from URL parameters
- `user_id`: Authenticated user identifier
- `comment_id`: Comment identifier

## Usage

### Setting Up the Logger (main.go)

```go
import (
    "log/slog"
    "github.com/saasuke-labs/kotomi/pkg/logging"
)

// Create a context-aware logger
jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
})
contextHandler := logging.NewContextHandler(jsonHandler)
logger := slog.New(contextHandler)
slog.SetDefault(logger)
```

### Adding Contextual Fields to Context

In handlers, enrich the context with relevant fields early:

```go
func (s *ServerHandlers) PostComments(w http.ResponseWriter, r *http.Request) {
    // Get URL parameters
    siteId := vars["siteId"]
    pageId := vars["pageId"]
    
    // Enrich context with site and page IDs
    ctx := r.Context()
    ctx = logging.WithSiteID(ctx, siteId)
    ctx = logging.WithPageID(ctx, pageId)
    r = r.WithContext(ctx)
    
    // Later, when a comment is created
    comment.ID = uuid.NewString()
    ctx = logging.WithCommentID(ctx, comment.ID)
    r = r.WithContext(ctx)
    
    // All subsequent operations use the enriched context
    // ...
}
```

### Logging with Context

Use the `*Context` variants of logging methods:

```go
// Before (manual field addition)
s.Logger.Error("failed to add comment",
    "error", err,
    "site_id", siteId,
    "page_id", pageId,
    "comment_id", comment.ID,
    "request_id", requestID)

// After (automatic field inclusion)
s.Logger.ErrorContext(ctx, "failed to add comment", "error", err)
```

The fields `request_id`, `site_id`, `page_id`, and `comment_id` are automatically included in the log output because they're in the context.

### Log Output

With context-aware logging, logs automatically include all contextual fields:

```json
{
  "time": "2026-02-04T07:11:00.78Z",
  "level": "ERROR",
  "msg": "failed to add comment",
  "error": "database locked",
  "request_id": "abc123",
  "site_id": "site-456",
  "page_id": "page-789",
  "comment_id": "comment-def"
}
```

## Benefits

1. **Cleaner Code**: No need to manually add `request_id`, `site_id`, etc. to every log call
2. **Consistency**: Contextual fields are always included when present in context
3. **Maintainability**: Adding new contextual fields only requires updating the context package
4. **Debugging**: All related operations automatically include the same contextual identifiers
5. **Tracing**: Request IDs allow tracking a single request through the entire stack

## Middleware Integration

The `RequestIDMiddleware` automatically adds a unique `request_id` to every request's context:

```go
// pkg/middleware/requestid.go
func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }
        
        // Add request ID to context
        ctx := logging.WithRequestID(r.Context(), requestID)
        
        // All handlers down the chain will have this request ID
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Best Practices

1. **Enrich Context Early**: Add contextual fields as soon as they're available in your handler
2. **Use Context Consistently**: Pass the enriched context to all function calls (database operations, etc.)
3. **Use *Context Methods for Logging**: Always use `ErrorContext`, `InfoContext`, `WarnContext`, etc. instead of `Error`, `Info`, `Warn`
4. **Use Middleware for Non-Logging Purposes**: When you need request ID for non-logging purposes (e.g., error responses), use `middleware.GetRequestID(r)` instead of `logging.GetRequestID(ctx)` to avoid coupling non-logging code to the logging package
5. **Add Additional Fields**: You can still add method-specific fields to log calls:
   ```go
   logger.InfoContext(ctx, "moderation completed", 
       "decision", result.Decision,
       "confidence", result.Confidence)
   ```

### Example: Error Handling

```go
// For logging - use context
s.Logger.ErrorContext(ctx, "failed to add comment", "error", err)

// For error responses - use middleware
apierrors.WriteError(w, 
    apierrors.DatabaseError("Failed to add comment").
        WithRequestID(middleware.GetRequestID(r)))
```

This separation keeps logging concerns (automatic field propagation via context) separate from general request handling (explicit request ID retrieval).

## Testing

The logging package includes comprehensive tests:

```bash
go test ./pkg/logging/
```

Tests verify:
- Context field storage and retrieval
- Automatic inclusion of context fields in logs
- Handler compatibility with slog's API
- Behavior with and without context fields

## Migration from Manual Field Addition

When migrating existing code:

1. Add import: `"github.com/saasuke-labs/kotomi/pkg/logging"`
2. Enrich context at handler entry with `logging.WithSiteID()`, `logging.WithPageID()`, etc.
3. Replace `Logger.Error()` with `Logger.ErrorContext(ctx)` for logging
4. Remove manual additions of `"request_id"`, `"site_id"`, `"page_id"`, `"comment_id"` from log calls
5. For non-logging purposes (e.g., error responses), use `middleware.GetRequestID(r)` instead of `logging.GetRequestID(ctx)`

## Related ADRs

This implementation completes **Phase 4 of ADR 002** (Code Structure and Go 1.25 Improvements), which called for:
- Context propagation throughout the stack
- Using context to pass contextual fields for structured logs
- Making logging messages cleaner with automatic field propagation

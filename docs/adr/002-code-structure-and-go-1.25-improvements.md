# ADR 002: Code Structure and Go 1.25 Improvements

**Status:** Proposed  
**Date:** 2026-02-03  
**Authors:** Kotomi Development Team  
**Deciders:** Engineering Team  

## Context and Problem Statement

After analysis of the Kotomi codebase (currently at Go 1.25.0), several opportunities have been identified to improve code maintainability, readability, testability, and adoption of modern Go 1.25 patterns. While the codebase demonstrates good architectural separation of concerns through packages, there are structural issues and missed opportunities for modernization that should be addressed before the v1.0 release.

### Current Codebase State

**Strengths:**
- Clean package structure with clear separation of concerns (auth, comments, models, middleware, etc.)
- Comprehensive test coverage (77 Go files, 150+ tests)
- Good error handling framework with structured API errors
- Modern dependencies (JWT v5, UUID, Gorilla Mux)
- Built-in API documentation with Swagger/OpenAPI

**Areas for Improvement:**
- **Monolithic main.go**: 1169 lines containing initialization, routing, and 40+ HTTP handlers
- **Global variables**: 7 package-level mutable variables in main.go affecting testability
- **No dependency injection**: Handlers depend on global state, making unit testing difficult
- **Inconsistent error handling**: Mix of `log.Printf()`, `http.Error()`, and structured errors
- **Legacy logging**: Using standard `log` package instead of structured `slog` (Go 1.21+)
- **Limited context usage**: Context not propagated to database operations
- **SQL boilerplate**: Repeated `sql.NullString` handling across stores
- **Missing interface assertions**: No compile-time verification of interface implementations
- **Go 1.25 patterns**: Limited use of modern Go features introduced in recent versions

## Decision Drivers

* **Maintainability**: Code should be easy to understand, modify, and extend
* **Testability**: Components should be easily testable in isolation
* **Readability**: Code organization should make it easy to locate and understand functionality
* **Performance**: Improvements should not negatively impact performance
* **Modern Practices**: Adopt Go 1.25 best practices and patterns
* **Developer Experience**: Make it easier for contributors to work with the codebase
* **Production Readiness**: Ensure code meets production-quality standards before v1.0

## Recommended Improvements

### 1. Refactor main.go into Modular Structure

**Problem**: The main.go file is 1169 lines and contains:
- Global variable declarations
- 40+ HTTP handler functions
- Route registration
- Server initialization
- Configuration loading

**Solution**: Split into multiple packages:

```
cmd/
├── main.go                    # Entry point, initialization only (~100 lines)
└── server/
    ├── server.go             # Server struct with dependencies
    ├── routes.go             # Route registration
    ├── config.go             # Configuration loading
    └── handlers/
        ├── comments.go       # Comment-related handlers
        ├── reactions.go      # Reaction handlers
        ├── pages.go          # Page handlers
        ├── auth.go           # Authentication handlers
        └── admin.go          # Admin panel handlers
```

**Benefits**:
- Each file has a single responsibility
- Easy to locate specific functionality
- Better code organization
- Easier to review and test
- Reduced cognitive load

**Implementation**:
```go
// server/server.go
type Server struct {
    commentStore     CommentStore
    db              *sql.DB
    templates       *template.Template
    auth0Config     *auth.Auth0Config
    moderator       moderation.Moderator
    notifications   *notifications.Queue
    logger          *slog.Logger
}

func New(cfg Config) (*Server, error) {
    // Initialize dependencies
    return &Server{...}, nil
}

func (s *Server) Handler() http.Handler {
    router := mux.NewRouter()
    s.registerRoutes(router)
    return router
}
```

**Migration Priority**: HIGH (foundational for other improvements)  
**Effort**: High (requires careful refactoring)  
**Impact**: Very High (enables all other improvements)

---

### 2. Eliminate Global Variables with Dependency Injection

**Problem**: Seven package-level mutable variables in main.go:
```go
var commentStore CommentStore
var db *sql.DB
var templates *template.Template
var auth0Config *auth.Auth0Config
var moderator moderation.Moderator
var moderationConfigStore *moderation.ConfigStore
var notificationQueue *notifications.Queue
```

**Issues**:
- Hard to test (requires global state manipulation)
- Implicit dependencies
- Potential race conditions
- Violates single responsibility principle
- Makes concurrent testing impossible

**Solution**: Use the Server struct pattern (from Improvement #1):

```go
// Before: Handler depends on globals
func postCommentsHandler(w http.ResponseWriter, r *http.Request) {
    // Uses global commentStore, db, moderator, etc.
}

// After: Handler is a method with explicit dependencies
func (s *Server) handlePostComment(w http.ResponseWriter, r *http.Request) {
    // Uses s.commentStore, s.db, s.moderator, etc.
}
```

**Testing Benefits**:
```go
func TestHandlePostComment(t *testing.T) {
    server := &Server{
        commentStore: &mockCommentStore{},
        moderator:    &mockModerator{},
        // ... other mock dependencies
    }
    
    // Test handler in isolation
    req := httptest.NewRequest("POST", "/api/v1/site/test/page/test/comments", body)
    rec := httptest.NewRecorder()
    server.handlePostComment(rec, req)
    
    // Assert results
}
```

**Migration Priority**: HIGH (enables proper testing)  
**Effort**: High (requires refactoring all handlers)  
**Impact**: Very High (dramatically improves testability)

---

### 3. Adopt Structured Logging with slog

**Problem**: Currently using standard library `log` package:
```go
log.Printf("AI moderation failed: %v", err)
log.Printf("Error adding comment: %v", err)
```

**Issues**:
- Unstructured text output
- No log levels
- No structured fields for filtering/parsing
- Difficult to integrate with log aggregation systems
- No context correlation

**Solution**: Migrate to `slog` (standard library since Go 1.21):

```go
// Before
log.Printf("Error adding comment: %v", err)

// After
s.logger.Error("failed to add comment",
    "error", err,
    "site_id", siteId,
    "page_id", pageId,
    "user_id", userID,
    "request_id", requestID,
)
```

**Benefits**:
- Structured output (JSON, text, custom)
- Log levels (Debug, Info, Warn, Error)
- Contextual fields for filtering
- Better integration with observability tools
- Performance improvements over fmt-based logging

**Implementation Pattern**:
```go
// server/server.go
type Server struct {
    logger *slog.Logger
    // ... other fields
}

// Initialize with context-aware logger
func New(cfg Config) (*Server, error) {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    
    return &Server{logger: logger}, nil
}

// Use throughout handlers
func (s *Server) handlePostComment(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    
    s.logger.Info("processing comment",
        "request_id", requestID,
        "site_id", siteId,
        "user", userID,
    )
}
```

**Migration Priority**: MEDIUM (improves observability)  
**Effort**: Medium (systematic replacement of log calls)  
**Impact**: High (better debugging and monitoring)

---

### 4. Propagate Context Throughout the Stack

**Problem**: Limited context usage - only in main function and shutdown:
```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
```

Database operations and handlers don't use context:
```go
func (s *SiteStore) GetByID(id string) (*Site, error) {
    query := `SELECT ... FROM sites WHERE id = ?`
    err := s.db.QueryRow(query, id).Scan(...)
    // No context, no timeout, no cancellation
}
```

**Solution**: Add context to all I/O operations:

```go
// Store methods accept context
func (s *SiteStore) GetByID(ctx context.Context, id string) (*Site, error) {
    query := `SELECT ... FROM sites WHERE id = ?`
    err := s.db.QueryRowContext(ctx, query, id).Scan(...)
    return &site, err
}

// Handlers propagate request context
func (s *Server) handleGetSite(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    site, err := s.siteStore.GetByID(ctx, siteID)
    if err != nil {
        // Handle error
        return
    }
    // ... rest of handler
}

// Add timeouts for long-running operations
func (s *Server) moderateComment(ctx context.Context, comment string) error {
    // Set 5-second timeout for moderation
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    return s.moderator.Moderate(ctx, comment)
}
```

**Benefits**:
- Request cancellation propagates to database
- Timeouts prevent hanging operations
- Request-scoped values (user ID, trace ID)
- Graceful shutdown support
- Better resource management

**Migration Priority**: HIGH (critical for production reliability)  
**Effort**: High (touches many function signatures)  
**Impact**: Very High (prevents resource leaks and timeouts)

---

### 5. Standardize Error Handling

**Problem**: Inconsistent error handling across handlers:
```go
// Method 1: http.Error
http.Error(w, "Invalid request", http.StatusBadRequest)

// Method 2: log.Printf + http.Error
log.Printf("Error: %v", err)
http.Error(w, "Internal server error", http.StatusInternalServerError)

// Method 3: apierrors.WriteError
apierrors.WriteError(w, apierrors.ErrBadRequest.WithDetails("Invalid JSON"))
```

**Solution**: Use apierrors consistently throughout:

```go
// All handlers use structured errors
func (s *Server) handlePostComment(w http.ResponseWriter, r *http.Request) {
    var req CommentRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        apierrors.WriteError(w, 
            apierrors.ErrBadRequest.
                WithDetails("Invalid JSON body").
                WithRequestID(middleware.GetRequestID(r.Context())),
        )
        return
    }
    
    // ... business logic
    
    if err := s.commentStore.Add(r.Context(), comment); err != nil {
        s.logger.Error("failed to add comment", 
            "error", err,
            "request_id", middleware.GetRequestID(r.Context()),
        )
        apierrors.WriteError(w, 
            apierrors.ErrInternal.
                WithDetails("Failed to save comment").
                WithRequestID(middleware.GetRequestID(r.Context())),
        )
        return
    }
}
```

**Error Wrapping Pattern**:
```go
// Wrap errors with context
func (s *SiteStore) GetByID(ctx context.Context, id string) (*Site, error) {
    // ... query execution
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, fmt.Errorf("site %s not found: %w", id, err)
        }
        return nil, fmt.Errorf("failed to query site %s: %w", id, err)
    }
    return site, nil
}
```

**Migration Priority**: MEDIUM (improves API consistency)  
**Effort**: Medium (systematic handler updates)  
**Impact**: High (better error messages and debugging)

---

### 6. Add Compile-Time Interface Verification

**Problem**: No verification that types implement interfaces:
```go
type MockModerator struct { /* ... */ }
// No compile-time check that MockModerator implements Moderator
```

**Solution**: Add interface assertions:

```go
// pkg/moderation/mock.go
var _ Moderator = (*MockModerator)(nil)

// pkg/moderation/openai.go
var _ Moderator = (*OpenAIModerator)(nil)

// pkg/comments/sqlite.go
var _ CommentStore = (*SQLiteStore)(nil)

// pkg/comments/memory.go
var _ CommentStore = (*InMemoryStoreAdapter)(nil)
```

**Benefits**:
- Compile-time verification of interface compliance
- Catches missing methods early
- Documents interface implementation
- Prevents runtime panics from incomplete implementations

**Migration Priority**: LOW (quality improvement)  
**Effort**: Low (add one line per implementation)  
**Impact**: Low (safety net, catches errors early)

---

### 7. Reduce SQL Boilerplate with Helpers

**Problem**: Repeated `sql.NullString` boilerplate in 10+ files:
```go
var domain, description sql.NullString
err := db.QueryRow(query).Scan(&id, &domain, &description, ...)

if domain.Valid {
    site.Domain = domain.String
}
if description.Valid {
    site.Description = description.String
}
```

**Solution Option A**: Helper functions:
```go
// pkg/database/helpers.go
package database

import "database/sql"

// NullString returns the string value or empty string if null
func NullString(ns sql.NullString) string {
    if ns.Valid {
        return ns.String
    }
    return ""
}

// ToNullString converts a string to sql.NullString
func ToNullString(s string) sql.NullString {
    if s == "" {
        return sql.NullString{Valid: false}
    }
    return sql.NullString{String: s, Valid: true}
}

// Usage
site.Domain = database.NullString(domain)
site.Description = database.NullString(description)
```

**Solution Option B**: Custom scanner types:
```go
// pkg/database/types.go
type NullableString string

func (ns *NullableString) Scan(value interface{}) error {
    if value == nil {
        *ns = ""
        return nil
    }
    if str, ok := value.(string); ok {
        *ns = NullableString(str)
        return nil
    }
    return fmt.Errorf("unexpected type for NullableString: %T", value)
}

// Usage in models
type Site struct {
    Domain      NullableString `json:"domain,omitempty"`
    Description NullableString `json:"description,omitempty"`
}

// Direct scanning
err := db.QueryRow(query).Scan(&site.ID, &site.Domain, &site.Description, ...)
```

**Solution Option C**: Use sqlc for code generation:
- Generate type-safe Go code from SQL
- Eliminates manual null handling
- Provides compile-time query validation

**Migration Priority**: LOW (quality of life improvement)  
**Effort**: Low (helpers) to Medium (sqlc)  
**Impact**: Medium (reduces boilerplate, improves readability)

---

### 8. Improve Test Infrastructure

**Problem**: Tests lack structure and helpers:
```go
// No subtests
func TestSiteStore_GetByID(t *testing.T) {
    // Test setup
    // Test case 1
    // Test case 2
    // Test case 3
}

// No test fixtures or factories
db := createTestDB()
site := models.Site{
    ID:      uuid.New().String(),
    OwnerID: uuid.New().String(),
    Name:    "Test Site",
    // ... manual setup
}
```

**Solution**: Add test helpers and patterns:

```go
// pkg/testutil/fixtures.go
package testutil

import (
    "testing"
    "github.com/saasuke-labs/kotomi/pkg/models"
)

// SiteBuilder provides a fluent API for creating test sites
type SiteBuilder struct {
    site models.Site
}

func NewSite() *SiteBuilder {
    return &SiteBuilder{
        site: models.Site{
            ID:      uuid.New().String(),
            OwnerID: uuid.New().String(),
            Name:    "Test Site",
        },
    }
}

func (b *SiteBuilder) WithID(id string) *SiteBuilder {
    b.site.ID = id
    return b
}

func (b *SiteBuilder) WithOwner(ownerID string) *SiteBuilder {
    b.site.OwnerID = ownerID
    return b
}

func (b *SiteBuilder) Build() models.Site {
    return b.site
}

// Usage in tests
func TestSiteStore_GetByID(t *testing.T) {
    db := testutil.NewTestDB(t)
    store := models.NewSiteStore(db)
    
    t.Run("returns site when found", func(t *testing.T) {
        site := testutil.NewSite().
            WithID("test-123").
            WithOwner("owner-456").
            Build()
        
        // ... test logic
    })
    
    t.Run("returns error when not found", func(t *testing.T) {
        // ... test logic
    })
    
    t.Run("returns error on database failure", func(t *testing.T) {
        // ... test logic
    })
}
```

**Migration Priority**: LOW (improves test maintainability)  
**Effort**: Medium (create fixtures and update tests)  
**Impact**: Medium (makes tests easier to write and read)

---

### 9. Adopt Go 1.25 Specific Patterns

#### 9.1 Range Over Integers (Go 1.22+)

**Before**:
```go
for i := 0; i < 10; i++ {
    // ...
}
```

**After**:
```go
for i := range 10 {
    // ...
}
```

#### 9.2 Enhanced Error Wrapping

**Current**: Limited use of `%w`
**Improved**: Consistent error wrapping:
```go
if err != nil {
    return fmt.Errorf("failed to process comment: %w", err)
}
```

#### 9.3 Clear Built-in Function (Go 1.21+)

**Before**:
```go
for k := range m {
    delete(m, k)
}
```

**After**:
```go
clear(m)
```

#### 9.4 Min/Max Built-in Functions (Go 1.21+)

**Before**:
```go
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

**After**:
```go
result := min(a, b)
```

#### 9.5 Structured Logging (Go 1.21+)

Already covered in Improvement #3 (slog)

#### 9.6 Testing Coverage Profile (Go 1.20+)

Add coverage profile to CI:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

**Migration Priority**: LOW (incremental adoption)  
**Effort**: Low (use as opportunities arise)  
**Impact**: Low (minor code improvements)

---

### 10. Database Connection Configuration

**Problem**: No visible connection pool configuration:
```go
db, err := sql.Open("sqlite3", dbPath)
// Default settings used
```

**Solution**: Configure connection pool for production:

```go
db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
if err != nil {
    return nil, fmt.Errorf("failed to open database: %w", err)
}

// Configure connection pool
db.SetMaxOpenConns(25)               // Limit concurrent connections
db.SetMaxIdleConns(5)                // Keep some connections warm
db.SetConnMaxLifetime(5 * time.Minute)  // Recycle old connections
db.SetConnMaxIdleTime(time.Minute)   // Close idle connections

// Verify connection
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
if err := db.PingContext(ctx); err != nil {
    return nil, fmt.Errorf("database not responding: %w", err)
}
```

**Benefits**:
- Better performance under load
- Prevents connection exhaustion
- WAL mode improves SQLite concurrency
- Configurable timeouts

**Migration Priority**: MEDIUM (important for production)  
**Effort**: Low (single location change)  
**Impact**: High (prevents production issues)

---

## Implementation Plan

### Phase 1: Foundation (Weeks 1-2)
1. Create Server struct and move globals (Improvement #2)
2. Refactor main.go into packages (Improvement #1)
3. Update all tests to work with new structure

**Deliverable**: Refactored code structure with dependency injection

### Phase 2: Reliability (Weeks 3-4)
4. Add context propagation (Improvement #4)
5. Configure database connection pool (Improvement #10)
6. Standardize error handling (Improvement #5)

**Deliverable**: Production-ready error handling and resource management

### Phase 3: Observability (Week 5)
7. Migrate to structured logging with slog (Improvement #3)
8. Add interface assertions (Improvement #6)

**Deliverable**: Better debugging and monitoring capabilities

### Phase 4: Polish (Week 6)
9. Add SQL helpers or evaluate sqlc (Improvement #7)
10. Improve test infrastructure (Improvement #8)
11. Adopt Go 1.25 patterns where applicable (Improvement #9)

**Deliverable**: Cleaner code and better test maintainability

---

## Consequences

### Positive

* **Better Testability**: Dependency injection enables comprehensive unit testing
* **Improved Maintainability**: Smaller, focused files are easier to understand and modify
* **Enhanced Reliability**: Context propagation prevents resource leaks and hanging requests
* **Better Observability**: Structured logging improves debugging and monitoring
* **Modern Codebase**: Adopting Go 1.25 patterns aligns with community best practices
* **Production Ready**: Proper configuration and error handling prepare for production use
* **Developer Experience**: Clear structure makes it easier for new contributors

### Negative

* **Breaking Changes**: Handler signatures will change (internal only, no API changes)
* **Test Updates**: All tests will need updates for new structure
* **Learning Curve**: Team needs to adopt new patterns (Server struct, slog)
* **Migration Effort**: Estimated 6 weeks of focused work
* **Temporary Instability**: During migration, code may be in transition state

### Neutral

* **No User Impact**: All changes are internal, API remains stable
* **Backward Compatibility**: External API contracts unchanged
* **Performance**: Negligible impact (slog may be faster than fmt-based logging)

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Breaking existing tests | High | Update tests incrementally, maintain coverage |
| Introducing bugs during refactoring | High | Comprehensive testing after each phase |
| Incomplete migration | Medium | Create tracking issues for each improvement |
| Team resistance to changes | Low | Document benefits, provide examples |
| Extended development time | Medium | Prioritize improvements, Phase 1 is critical |

---

## Alternatives Considered

### Alternative 1: Keep Current Structure
**Rejected**: Technical debt will compound as the project grows. Better to refactor now before v1.0 release.

### Alternative 2: Rewrite from Scratch
**Rejected**: Current codebase is fundamentally sound. Incremental improvements are lower risk and faster.

### Alternative 3: Use Framework (Echo, Gin, Fiber)
**Rejected**: Gorilla Mux is sufficient and well-established. Framework migration is unnecessary churn.

### Alternative 4: Adopt ORM (GORM, sqlx)
**Considered for Future**: Current SQL approach is explicit and performant. ORMs add complexity but may be valuable for complex queries. Revisit after Phase 4.

---

## References

* [Effective Go](https://go.dev/doc/effective_go)
* [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
* [Go 1.21 Release Notes (slog)](https://go.dev/doc/go1.21)
* [Go 1.22 Release Notes (range over int)](https://go.dev/doc/go1.22)
* [Go 1.25 Release Notes](https://go.dev/doc/go1.25)
* [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
* [Go Database/SQL Tutorial](http://go-database-sql.org/)
* [Dependency Injection in Go](https://github.com/google/wire)

---

## Success Metrics

After implementing these improvements, we should see:

* **Test Coverage**: Maintain >80% coverage during migration
* **Handler Unit Tests**: Each handler testable in isolation
* **Code Organization**: No file >500 lines
* **Build Time**: No significant increase
* **Error Consistency**: 100% of handlers use apierrors
* **Logging**: 100% structured logging with slog
* **Context Propagation**: All I/O operations accept context
* **Documentation**: All public APIs documented
* **Interface Verification**: All implementations have compile-time checks

---

## Related Documents

* [ADR 001: User Authentication for Comments and Reactions](001-user-authentication-for-comments-and-reactions.md)
* [Project README](../../README.md)
* [Contributing Guidelines](../../CONTRIBUTING.md)

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-02-03 | Initial proposal |

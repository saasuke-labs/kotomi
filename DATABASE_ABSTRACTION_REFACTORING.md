# Database Abstraction Refactoring

## Problem

The initial Firestore implementation maintained two separate references to the database:
1. `store` - generic `db.Store` interface
2. `sqliteStore` - specific `*comments.SQLiteStore` 

This broke the abstraction because:
- Server configuration expected `*comments.SQLiteStore` directly
- Handlers and admin packages were coupled to SQLite implementation
- The comments package was aware of specific database implementations
- Code needed to know whether it was using SQLite or Firestore

## Solution

Refactored the entire codebase to use a single, unified `db.Store` interface. Now all code is completely database-agnostic.

## Changes Made

### 1. Improved Interface Design

**Before:**
```go
type Store interface {
    CommentStore  // Embedded interface
    GetDB() *sql.DB
    Close() error
}

type CommentStore interface {
    AddPageComment(ctx context.Context, site, page string, comment interface{}) error
    GetPageComments(ctx context.Context, site, page string) ([]interface{}, error)
    // ... more methods with interface{} types
}
```

**After:**
```go
type Store interface {
    // Direct methods with concrete types
    AddPageComment(ctx context.Context, site, page string, comment comments.Comment) error
    GetPageComments(ctx context.Context, site, page string) ([]comments.Comment, error)
    GetCommentByID(ctx context.Context, commentID string) (*comments.Comment, error)
    // ... more methods with concrete types
    GetDB() *sql.DB
    Close() error
}
```

**Benefits:**
- Single, flat interface - no embedded interfaces
- Type safety with concrete `comments.Comment` types
- No type assertions needed in calling code
- Clearer API contract

### 2. Simplified Implementations

**SQLiteAdapter - Before:**
```go
func (a *SQLiteAdapter) AddPageComment(ctx context.Context, site, page string, comment interface{}) error {
    c, ok := comment.(comments.Comment)
    if !ok {
        return fmt.Errorf("invalid comment type: expected comments.Comment, got %T", comment)
    }
    return a.store.AddPageComment(ctx, site, page, c)
}

func (a *SQLiteAdapter) GetPageComments(ctx context.Context, site, page string) ([]interface{}, error) {
    commentsList, err := a.store.GetPageComments(ctx, site, page)
    if err != nil {
        return nil, err
    }
    result := make([]interface{}, len(commentsList))
    for i, c := range commentsList {
        result[i] = c
    }
    return result, nil
}
```

**SQLiteAdapter - After:**
```go
func (a *SQLiteAdapter) AddPageComment(ctx context.Context, site, page string, comment comments.Comment) error {
    return a.store.AddPageComment(ctx, site, page, comment)
}

func (a *SQLiteAdapter) GetPageComments(ctx context.Context, site, page string) ([]comments.Comment, error) {
    return a.store.GetPageComments(ctx, site, page)
}
```

**Benefits:**
- No type conversions needed
- Simple pass-through to underlying store
- Reduced code complexity
- Better performance (no slice allocations)

### 3. Unified Server Configuration

**Before:**
```go
// In cmd/main.go
store, err := db.NewStore(context.Background(), dbConfig)
sqlDB := store.GetDB()

// Backward compatibility hack
var sqliteStore *comments.SQLiteStore
if adapter, ok := store.(*db.SQLiteAdapter); ok {
    sqliteStore = adapter.GetSQLiteStore()
}

// Server config
cfg := server.Config{
    CommentStore: sqliteStore,  // *comments.SQLiteStore (breaks with Firestore!)
    DB: sqlDB,
    // ...
}

// In server/config.go
type Config struct {
    CommentStore *comments.SQLiteStore  // Coupled to SQLite!
    // ...
}
```

**After:**
```go
// In cmd/main.go
store, err := db.NewStore(context.Background(), dbConfig)
sqlDB := store.GetDB()  // May be nil for Firestore

cfg := server.Config{
    CommentStore: store,  // db.Store interface (works with any DB!)
    DB: sqlDB,
    // ...
}

// In server/config.go  
type Config struct {
    CommentStore db.Store  // Database-agnostic!
    // ...
}
```

**Benefits:**
- Single store reference
- No backward compatibility hacks
- Works with SQLite, Firestore, or any future database
- Clean, maintainable code

### 4. Database-Agnostic Handlers

**Before:**
```go
// In handlers/handlers.go
type ServerHandlers struct {
    CommentStore *comments.SQLiteStore  // Coupled to SQLite
    // ...
}

func NewHandlers(commentStore *comments.SQLiteStore, ...) *ServerHandlers {
    return &ServerHandlers{
        CommentStore: commentStore,
        // ...
    }
}
```

**After:**
```go
// In handlers/handlers.go
type ServerHandlers struct {
    CommentStore db.Store  // Database-agnostic
    // ...
}

func NewHandlers(commentStore db.Store, ...) *ServerHandlers {
    return &ServerHandlers{
        CommentStore: commentStore,
        // ...
    }
}
```

### 5. Database-Agnostic Admin Handlers

**Before:**
```go
// In pkg/admin/comments.go
type CommentsHandler struct {
    commentStore *comments.SQLiteStore  // Coupled to SQLite
    // ...
}

func NewCommentsHandler(db *sql.DB, commentStore *comments.SQLiteStore, ...) *CommentsHandler {
    return &CommentsHandler{
        commentStore: commentStore,
        // ...
    }
}
```

**After:**
```go
// In pkg/admin/comments.go
type CommentsHandler struct {
    commentStore db.Store  // Database-agnostic
    // ...
}

func NewCommentsHandler(sqlDB *sql.DB, commentStore db.Store, ...) *CommentsHandler {
    return &CommentsHandler{
        commentStore: commentStore,
        // ...
    }
}
```

## Impact Analysis

### Code Reduction
- **Removed**: ~50 lines of type conversion code
- **Removed**: Backward compatibility hack (`GetSQLiteStore()`)
- **Removed**: Dual store references throughout codebase
- **Simplified**: All adapter methods are now simple pass-throughs

### Type Safety Improvements
- **Before**: Methods used `interface{}` - type assertions required
- **After**: Methods use `comments.Comment` - compile-time type checking
- **Result**: Fewer runtime errors, better IDE support

### Maintainability 
- **Before**: Adding a new database required updating many type assertions
- **After**: Adding a new database only requires implementing the Store interface
- **Result**: Easier to add MongoDB, PostgreSQL, etc. in the future

### Testing
- All existing tests continue to pass
- No changes needed to test infrastructure
- Test coverage remains at same level

## Architecture Benefits

### 1. True Abstraction
The entire application is now truly database-agnostic:
```go
// Handlers don't care about the database
func (s *ServerHandlers) PostComments(w http.ResponseWriter, r *http.Request) {
    comment := comments.Comment{...}
    err := s.CommentStore.AddPageComment(ctx, siteId, pageId, comment)
    // Works with SQLite, Firestore, or any future DB!
}
```

### 2. Single Source of Truth
```go
// Only one store reference
store, err := db.NewStore(ctx, config)

// Used everywhere
cfg := server.Config{CommentStore: store}
handlers := handlers.NewHandlers(store, ...)
adminHandlers := admin.NewCommentsHandler(sqlDB, store, ...)
```

### 3. Clean Dependency Injection
```go
// Clear, explicit dependencies
type Server struct {
    CommentStore db.Store  // What it does
    DB *sql.DB            // Optional for SQL-specific features
}
```

### 4. Easy to Extend
Adding a new database implementation:
```go
// 1. Implement the interface
type PostgresStore struct { ... }
func (s *PostgresStore) AddPageComment(...) error { ... }
// ... implement other methods

// 2. Update factory
case ProviderPostgres:
    return NewPostgresStore(cfg.PostgresURL)

// 3. Done! No changes needed anywhere else
```

## Migration Guide

If you have custom code using the old pattern:

**Before:**
```go
// Don't do this anymore
sqliteStore := getSQLiteStore()
comments, err := sqliteStore.GetPageComments(ctx, site, page)
```

**After:**
```go
// Do this instead
store := getStore()  // Returns db.Store
comments, err := store.GetPageComments(ctx, site, page)
```

The interface is identical, just more generic!

## Conclusion

This refactoring achieves true database abstraction:
- ✅ Single `store` reference throughout codebase
- ✅ All code is database-agnostic
- ✅ Type-safe with concrete types
- ✅ Simpler, more maintainable code
- ✅ Easy to add new database implementations
- ✅ No breaking changes to functionality
- ✅ All tests pass

The codebase is now cleaner, more maintainable, and ready for future database additions!

# ADR 003: SQLite Optimization for Concurrent Access

**Status:** Proposed  
**Date:** 2026-02-03  
**Authors:** Kotomi Development Team  
**Deciders:** Engineering Team  

## Context and Problem Statement

Kotomi uses SQLite as its primary database for storing comments, reactions, users, sites, and related data. As the application scales to support multiple users making concurrent requests, we need to ensure that SQLite is properly configured for optimal performance, particularly for concurrent read and write operations.

SQLite is a powerful embedded database that can handle moderate concurrency well when properly configured. However, its default settings are conservative and optimized for compatibility rather than performance. Without proper tuning, SQLite applications can experience:

- Write operations blocking reads unnecessarily
- Database lock contention under load
- Connection exhaustion
- Slow response times during concurrent access
- "Database is locked" errors under moderate load

### Current State Analysis

#### Configuration

Currently, the SQLite database is initialized with minimal configuration in `pkg/comments/sqlite.go`:

```go
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    // Enable foreign key constraints
    _, err = db.Exec("PRAGMA foreign_keys = ON")
    if err != nil {
        db.Close()
        return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
    }
    
    // ... schema creation
}
```

**Issues Identified:**

1. **No WAL (Write-Ahead Logging) Mode**: The database uses the default ROLLBACK journal mode, which provides poor concurrency. In ROLLBACK mode, writes block all reads.

2. **No Connection Pool Configuration**: The `*sql.DB` instance uses Go's default settings:
   - Unlimited max open connections
   - No connection lifetime management
   - No idle connection limits

3. **Missing Critical PRAGMAs**: Several important SQLite performance settings are not configured:
   - `busy_timeout`: No timeout for lock contention (defaults to 0ms)
   - `cache_size`: Uses default cache size (can be too small)
   - `synchronous`: Uses default FULL mode (slower but safest)
   - `temp_store`: Temporary tables location not optimized

4. **No Connection String Parameters**: The database path is used directly without performance-related connection parameters.

#### Concurrency Patterns

The application demonstrates high concurrent usage:

- **HTTP Handlers**: Multiple simultaneous requests for comments, reactions, and user data
- **Background Workers**: Notification queue processor runs in a goroutine
- **Tests**: Confirm concurrent writes (10 goroutines × 10 comments) work but without optimization
- **Shared Global Database**: All operations share a single `*sql.DB` instance

#### Transaction Usage

- **Present**: Used in import operations with proper rollback handling
- **Missing**: Not used in individual comment/reaction operations
- **Impact**: Each write is a separate transaction, increasing lock contention

## Decision Drivers

* **Scalability**: Support many concurrent users without degradation
* **Performance**: Minimize response time for reads and writes
* **Reliability**: Prevent "database is locked" errors
* **Resource Efficiency**: Manage database connections effectively
* **Data Integrity**: Maintain ACID guarantees while improving performance
* **Backward Compatibility**: Work with existing SQLite database files
* **Production Readiness**: Handle real-world load patterns effectively

## Decision

We will optimize SQLite configuration for concurrent access by implementing the following changes:

### 1. Enable Write-Ahead Logging (WAL) Mode

**Change**: Set `PRAGMA journal_mode = WAL` during database initialization.

**Benefits**:
- Reads don't block writes and vice versa (mostly)
- Multiple concurrent readers with one writer
- Significantly better performance under concurrent load
- Faster commits (no need to write rollback journal)

**Trade-offs**:
- Requires SQLite 3.7.0+ (released 2010, universally available)
- Creates `-wal` and `-shm` files alongside database
- Slightly more disk I/O in some scenarios
- Requires proper cleanup on database close

**Implementation**:
```go
// Enable WAL mode for better concurrency
_, err = db.Exec("PRAGMA journal_mode = WAL")
if err != nil {
    db.Close()
    return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
}
```

### 2. Configure Busy Timeout

**Change**: Set `PRAGMA busy_timeout = 5000` (5 seconds).

**Benefits**:
- Automatically retries on locked database instead of failing immediately
- Handles temporary lock contention gracefully
- Reduces "database is locked" errors under moderate load
- Works with WAL mode for better throughput

**Trade-offs**:
- Requests may take up to 5 seconds if database is heavily contended
- Could mask underlying performance issues if too high

**Implementation**:
```go
// Set busy timeout to handle lock contention gracefully
_, err = db.Exec("PRAGMA busy_timeout = 5000")
if err != nil {
    db.Close()
    return nil, fmt.Errorf("failed to set busy timeout: %w", err)
}
```

### 3. Optimize Cache Size

**Change**: Set `PRAGMA cache_size = -64000` (64MB cache).

**Benefits**:
- Reduces disk I/O by caching more pages in memory
- Faster query execution for frequently accessed data
- Better performance for queries with JOINs and complex indexes
- Particularly beneficial for read-heavy workloads

**Default**: SQLite default is typically 2000 pages (~2MB with 1KB page size)

**Trade-offs**:
- Uses more memory per connection
- May not be necessary for small databases
- Negative value uses KB instead of pages for portability

**Implementation**:
```go
// Increase cache size for better query performance (64MB)
_, err = db.Exec("PRAGMA cache_size = -64000")
if err != nil {
    db.Close()
    return nil, fmt.Errorf("failed to set cache size: %w", err)
}
```

### 4. Configure Synchronous Mode

**Change**: Set `PRAGMA synchronous = NORMAL` (down from FULL).

**Benefits**:
- Faster writes (2-3x improvement in some cases)
- Still safe with WAL mode (WAL has its own sync mechanism)
- Acceptable data integrity guarantee for most applications
- Reduces fsync() system calls

**Safety**: 
- With WAL mode: NORMAL is safe for power loss scenarios
- Without WAL mode: FULL is recommended for critical data

**Trade-offs**:
- Slightly higher risk of database corruption on system crash (mitigated by WAL)
- Not recommended for ROLLBACK journal mode

**Implementation**:
```go
// Set synchronous to NORMAL for better performance (safe with WAL)
_, err = db.Exec("PRAGMA synchronous = NORMAL")
if err != nil {
    db.Close()
    return nil, fmt.Errorf("failed to set synchronous mode: %w", err)
}
```

### 5. Optimize Temporary Storage

**Change**: Set `PRAGMA temp_store = MEMORY`.

**Benefits**:
- Temporary tables and indexes stay in memory
- Faster sorting and temporary data operations
- Reduces disk I/O for complex queries
- Improves performance of queries with ORDER BY, GROUP BY, DISTINCT

**Trade-offs**:
- Uses more memory for temporary operations
- Could be an issue with very large result sets

**Implementation**:
```go
// Use memory for temporary tables and indexes
_, err = db.Exec("PRAGMA temp_store = MEMORY")
if err != nil {
    db.Close()
    return nil, fmt.Errorf("failed to set temp store: %w", err)
}
```

### 6. Configure Connection Pool

**Change**: Set appropriate connection pool limits on the `*sql.DB` instance.

**Benefits**:
- Prevents connection exhaustion
- Better resource management
- Predictable performance under load
- Prevents memory bloat from unlimited connections

**Settings**:
- `SetMaxOpenConns(25)`: Limit to 25 concurrent connections
  - SQLite with WAL supports 1 writer + multiple readers
  - 25 connections is reasonable for moderate load
  - Can be adjusted based on hardware and workload
  
- `SetMaxIdleConns(5)`: Keep 5 connections warm
  - Reduces connection setup overhead
  - Balances resource usage with performance
  
- `SetConnMaxLifetime(5 * time.Minute)`: Recycle connections every 5 minutes
  - Prevents stale connections
  - Helps with long-running applications
  
- `SetConnMaxIdleTime(1 * time.Minute)`: Close idle connections after 1 minute
  - Frees resources when not needed
  - Still maintains pool of warm connections

**Implementation**:
```go
// Configure connection pool for production use
db.SetMaxOpenConns(25)                        // Limit concurrent connections
db.SetMaxIdleConns(5)                         // Keep some connections warm
db.SetConnMaxLifetime(5 * time.Minute)        // Recycle old connections
db.SetConnMaxIdleTime(1 * time.Minute)        // Close idle connections
```

### 7. Set MMAP Size for Better Read Performance

**Change**: Set `PRAGMA mmap_size = 268435456` (256MB memory-mapped I/O).

**Benefits**:
- Faster read operations through memory-mapped I/O
- Reduces system calls for reading data
- OS manages the memory mapping efficiently
- Particularly beneficial for read-heavy workloads

**Trade-offs**:
- Only helps on systems supporting mmap (most modern systems)
- Virtual address space usage (not always physical memory)
- May not help much for small databases

**Implementation**:
```go
// Enable memory-mapped I/O for better read performance (256MB)
_, err = db.Exec("PRAGMA mmap_size = 268435456")
if err != nil {
    // Log warning but don't fail - mmap not supported on all systems
    log.Printf("Warning: Could not enable mmap: %v", err)
}
```

### 8. Add Connection Health Check

**Change**: Verify database connectivity with ping and timeout.

**Benefits**:
- Fails fast if database is not accessible
- Prevents silent failures during initialization
- Provides clear error messages for configuration issues

**Implementation**:
```go
// Verify database connection with timeout
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
if err := db.PingContext(ctx); err != nil {
    db.Close()
    return nil, fmt.Errorf("database not responding: %w", err)
}
```

### 9. Connection String Optimization

**Change**: Pass performance parameters via connection string when possible.

**Benefits**:
- Some settings can be applied per connection
- Clearer configuration in one place
- Better for certain SQLite builds

**Implementation**:
```go
// Build connection string with performance parameters
connectionString := dbPath + "?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL&_cache_size=-64000&_temp_store=MEMORY"
db, err := sql.Open("sqlite3", connectionString)
```

**Note**: This is an alternative to PRAGMA statements. We'll use PRAGMA statements for clarity and compatibility, but connection string parameters are documented here as an option.

## Implementation Plan

### Phase 1: Core Optimizations (Immediate)

1. **Update `pkg/comments/sqlite.go::NewSQLiteStore()`**:
   - Add WAL mode
   - Add busy_timeout
   - Add synchronous = NORMAL
   - Add connection pool configuration
   - Add connection health check

2. **Test Compatibility**:
   - Run full test suite
   - Verify WAL mode works with existing tests
   - Check for any test failures

### Phase 2: Performance Tuning (After Phase 1)

3. **Add Additional PRAGMAs**:
   - cache_size
   - temp_store
   - mmap_size

4. **Benchmark Performance**:
   - Create benchmark tests for concurrent reads/writes
   - Measure before/after performance
   - Document improvements

### Phase 3: Monitoring and Tuning (Production)

5. **Add Observability**:
   - Log PRAGMA settings on startup
   - Add database stats endpoint
   - Monitor connection pool usage

6. **Tune Settings**:
   - Adjust connection pool size based on actual load
   - Monitor WAL checkpoint performance
   - Tune cache size based on working set

## Expected Performance Improvements

Based on SQLite documentation and industry experience:

| Metric | Current | Optimized | Improvement |
|--------|---------|-----------|-------------|
| Concurrent read throughput | Low (blocked by writes) | High (not blocked) | 5-10x |
| Write latency (no contention) | 1x | 0.5x | 2x faster |
| Write latency (with contention) | High (locks immediately) | Lower (retries) | 3-5x |
| "Database locked" errors | Frequent under load | Rare | 95% reduction |
| Memory usage per connection | ~2MB | ~64MB | 32x (acceptable) |
| Queries with large result sets | Disk-bound | Memory-bound | 2-3x |

**Real-World Testing**: These improvements should be validated with realistic workload testing.

## Consequences

### Positive

* **Better Concurrency**: WAL mode allows reads and writes to proceed concurrently
* **Fewer Lock Errors**: Busy timeout handles transient lock contention gracefully
* **Improved Performance**: Cache size and synchronous mode reduce I/O overhead
* **Resource Management**: Connection pool prevents resource exhaustion
* **Production Ready**: Configuration handles real-world concurrent workloads
* **Minimal Code Changes**: Changes localized to initialization code
* **Backward Compatible**: Existing database files work with new settings
* **Industry Standard**: Follows SQLite best practices for high-concurrency applications

### Negative

* **More Memory Usage**: Larger cache and connection pool use more RAM (~320MB for 5 idle connections)
* **Additional Files**: WAL mode creates `-wal` and `-shm` files
* **Slightly Complex Initialization**: More PRAGMA statements to maintain
* **Testing Burden**: Need to verify settings work correctly across platforms
* **Migration Considerations**: Existing databases transition to WAL on first write
* **Checkpoint Management**: WAL files need periodic checkpointing (automatic by SQLite)

### Neutral

* **No API Changes**: External API remains unchanged
* **No Schema Changes**: Database schema stays the same
* **Platform Compatibility**: Works on all modern systems (Linux, macOS, Windows)
* **SQLite Version**: Requires SQLite 3.7.0+ (released 2010, universally available)

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| WAL files not cleaned up | Low | Low | SQLite handles automatically; document manual cleanup if needed |
| Memory exhaustion from large cache | Low | Medium | Monitor memory usage; reduce cache_size if needed |
| Connection pool too small | Medium | Medium | Make configurable via environment variable; monitor pool saturation |
| Platform incompatibilities | Low | Medium | Test on Linux, macOS, Windows; gracefully handle PRAGMA failures |
| WAL checkpoint delays | Low | Low | SQLite handles automatically; can tune with PRAGMA wal_autocheckpoint |
| Existing code assumes ROLLBACK mode | Very Low | High | Comprehensive testing to verify no dependencies on journal mode |

## Alternatives Considered

### Alternative 1: Use PostgreSQL or MySQL Instead

**Pros**:
- Better native concurrency (true multi-writer)
- More robust for very high scale
- Better tooling and monitoring

**Cons**:
- External dependency (deployment complexity)
- Overkill for embedded/single-server deployments
- Increases operational burden
- Not suitable for edge deployments

**Decision**: Rejected. SQLite with proper configuration is sufficient for Kotomi's use case (embedded comments system). Most deployments won't need separate database server.

### Alternative 2: Use Connection String Parameters Only

**Pros**:
- Simpler code (fewer PRAGMA statements)
- All configuration in one place

**Cons**:
- Not all SQLite drivers support all parameters
- Less portable across SQLite builds
- Harder to debug (parameters not visible in logs)

**Decision**: Rejected. PRAGMA statements are more explicit, portable, and can be logged for debugging.

### Alternative 3: Implement Application-Level Write Queue

**Pros**:
- Serialize writes at application level
- Could batch multiple writes
- More control over write ordering

**Cons**:
- Complex implementation
- Adds latency to writes
- Not necessary with proper SQLite configuration
- Doesn't solve read concurrency

**Decision**: Rejected. Proper SQLite configuration is sufficient. Write queue would add unnecessary complexity.

### Alternative 4: Use Read Replicas

**Pros**:
- Better read scaling
- Can distribute read load

**Cons**:
- SQLite doesn't support true replication
- Would need third-party tools (Litestream, etc.)
- Adds deployment complexity
- Eventual consistency issues

**Decision**: Rejected for now. Single SQLite instance with WAL is sufficient for target scale. Revisit if read load becomes bottleneck.

### Alternative 5: Keep Current Configuration

**Pros**:
- No changes needed
- Simpler code

**Cons**:
- Poor performance under load
- "Database locked" errors likely
- Not production-ready for multi-user scenarios
- Doesn't meet scalability requirements

**Decision**: Rejected. Current configuration inadequate for production use with multiple concurrent users.

## Validation and Testing

### Test Cases

1. **Concurrent Read Test**: 50 goroutines reading comments simultaneously
2. **Concurrent Write Test**: 10 goroutines writing comments simultaneously
3. **Mixed Read/Write Test**: 40 readers + 10 writers simultaneously
4. **Lock Timeout Test**: Verify busy_timeout works under contention
5. **WAL Mode Test**: Verify WAL files are created and managed correctly
6. **Connection Pool Test**: Verify pool limits are respected
7. **Performance Benchmark**: Measure throughput before/after optimizations

### Success Criteria

- All existing tests pass with new configuration
- No increase in test flakiness
- Concurrent write test completes without "database locked" errors
- Benchmark shows measurable improvement in concurrent scenarios
- No memory leaks observed in long-running tests

### Monitoring

Post-deployment, monitor:
- Database connection pool utilization
- Query latency (p50, p95, p99)
- "Database locked" error rate (should be near zero)
- WAL file size (should checkpoint automatically)
- Memory usage per connection

## Configuration Reference

### Recommended Settings Summary

```go
// Enable WAL mode
PRAGMA journal_mode = WAL

// Lock handling
PRAGMA busy_timeout = 5000

// Performance tuning
PRAGMA synchronous = NORMAL
PRAGMA cache_size = -64000
PRAGMA temp_store = MEMORY
PRAGMA mmap_size = 268435456

// Connection pool
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
db.SetConnMaxIdleTime(1 * time.Minute)
```

### Tuning Guidelines

**For High-Read Workloads**:
- Increase `cache_size` to `-128000` (128MB)
- Increase `mmap_size` to `536870912` (512MB)
- Increase `MaxOpenConns` to 50

**For High-Write Workloads**:
- Decrease `synchronous` to OFF (only if data loss is acceptable)
- Keep `MaxOpenConns` at 25 or lower (SQLite is single-writer)
- Consider batching writes in transactions

**For Low-Memory Environments**:
- Decrease `cache_size` to `-32000` (32MB)
- Decrease `mmap_size` to `134217728` (128MB)
- Decrease `MaxIdleConns` to 2

**For Edge/Embedded Deployments**:
- Use `:memory:` or temp directory for database
- Disable `mmap_size` (set to 0)
- Use smaller `cache_size` (`-16000` = 16MB)

## References

* [SQLite Write-Ahead Logging](https://www.sqlite.org/wal.html)
* [SQLite PRAGMA Statements](https://www.sqlite.org/pragma.html)
* [SQLite Performance Tuning](https://www.sqlite.org/performance.html)
* [Go database/sql Package](https://pkg.go.dev/database/sql)
* [SQLite and Multiple Processes](https://www.sqlite.org/lockingv3.html)
* [Optimizing SQLite for Performance](https://phiresky.github.io/blog/2020/sqlite-performance-tuning/)
* [SQLite Concurrency](https://www.sqlite.org/isolation.html)

## Related ADRs

* [ADR 002: Code Structure and Go 1.25 Improvements](002-code-structure-and-go-1.25-improvements.md) - Connection pool configuration mentioned

## Success Metrics

After implementing these optimizations, we should observe:

* **Zero "database locked" errors** under normal load (< 25 concurrent requests)
* **5-10x improvement** in concurrent read throughput
* **2-3x improvement** in write latency
* **99th percentile latency** under 100ms for simple queries with concurrency
* **Connection pool utilization** stays below 80% under normal load
* **WAL file size** stays under 100MB (automatic checkpointing working)
* **All tests pass** with new configuration without flakiness

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-02-03 | Initial proposal with comprehensive SQLite optimization recommendations |

# Gotchas and Pitfalls

This document covers non-obvious behaviors, common mistakes, and important considerations when working with Pivot.

## Build Requirements

### SQLite JSON Support

SQLite queries using JSON functions require the `json1` build tag:

```bash
# Correct
go test --tags json1 ./...
go build --tags json1 -o bin/pivot cmd/pivot/*.go

# Incorrect - will fail on JSON operations
go test ./...
go build -o bin/pivot cmd/pivot/*.go
```

**Why**: SQLite's JSON1 extension is conditionally compiled. Without the tag, JSON field queries will fail silently or return unexpected results.

## v3 vs v4 Differences

### Import Paths

```go
// v3 (current production)
import "github.com/PerformLine/pivot/v3"
import "github.com/PerformLine/pivot/v3/dal"
import "github.com/PerformLine/pivot/v3/backends"

// v4 (in development)
import "github.com/PerformLine/pivot/v4"
import "github.com/PerformLine/pivot/v4/dal"
import "github.com/PerformLine/pivot/v4/backends"
```

### Context Propagation

v4 adds `context.Context` to all backend methods:

```go
// v3
backend.Retrieve("users", id)
mapper.Get(id, &user)

// v4
backend.Retrieve(ctx, "users", id)
mapper.Get(ctx, id, &user)
```

**Migration note**: If you need cancellation, timeouts, or deadline propagation, use v4. Otherwise, v3 is stable for production use.

### Logging Library

```go
// v3 uses go-stockutil/log
import "github.com/PerformLine/go-stockutil/log"

// v4 uses go-clog
import "github.com/PerformLine/go-clog/clog"
```

## Backend-Specific Quirks

### MongoDB: ID Field Mapping

MongoDB uses `_id` internally, but Pivot exposes it as `id`:

```go
// In your Go code
record.Get("id")      // Works
record.Set("id", "x") // Works

// In MongoDB
db.collection.find({_id: "x"})  // Actual storage
```

**Gotcha**: If you query MongoDB directly (bypassing Pivot), remember to use `_id`.

### DynamoDB: Limited Query Support

DynamoDB indexer only supports queries on:
- Partition key (exact match)
- Sort key (range queries)

```go
// Works
filter.Parse("partition_key/is:value")
filter.Parse("sort_key/gt:100")

// Does NOT work (falls back to scan)
filter.Parse("other_field/contains:test")
```

**Gotcha**: Non-key queries trigger expensive table scans. Design your schema accordingly.

### Redis: No Query Support

Redis backend has no real indexer - it's key-value only:

```go
// Works
backend.Retrieve("users", "user-123")

// Does NOT work meaningfully
indexer.Query(collection, filter)  // Returns empty or error
```

**Use case**: Redis is best for caching or simple key lookups, not complex queries.

### PostgreSQL: Case Sensitivity

PostgreSQL identifiers are case-sensitive when quoted:

```go
// Collection name "Users" (capital U)
// Pivot may generate: SELECT * FROM "Users"
// This differs from: SELECT * FROM users
```

**Recommendation**: Use lowercase collection and field names consistently.

### SQLite: Concurrent Writes

SQLite allows only one writer at a time:

```go
// This can cause "database is locked" errors
go func() { backend.Insert(...) }()
go func() { backend.Update(...) }()
```

**Solution**: Use a connection pool with `max_open_conns=1` for writes, or use WAL mode:

```
sqlite:///./data.db?_journal_mode=WAL
```

## Filter Syntax Gotchas

### Operator Default

No operator means `is` (exact match):

```go
// These are equivalent
filter.Parse("name/john")
filter.Parse("name/is:john")
```

### Pipe for OR Values

Pipe `|` separates OR values within a field:

```go
// name is "john" OR "jane"
filter.Parse("name/john|jane")

// NOT: name is "john|jane" literally
```

### Slash Escaping

Field values containing `/` need URL encoding:

```go
// Searching for "a/b"
filter.Parse("path/is:a%2Fb")
```

### Type Prefixes

Type prefixes help with ambiguous values:

```go
// Without prefix, "123" might be parsed as int
filter.Parse("code/123")

// Force string comparison
filter.Parse("str:code/123")

// Force integer comparison
filter.Parse("int:count/gt:100")
```

## Composite Key Handling

### Joiner Character

Default joiner is `:` - be careful with values containing colons:

```go
// If ID parts are ["tenant:1", "user:2"]
// Joined ID becomes "tenant:1:user:2"
// Decomposition may fail or be wrong
```

**Solution**: Use a different joiner:

```go
collection.IndexCompoundFieldJoiner = "|"
```

### Key Order Matters

Composite keys are positional:

```go
// Schema defines keys as [tenant_id, user_id]
record.SetKeys("tenant1", "user1")  // Correct order

// Wrong order produces wrong ID
record.SetKeys("user1", "tenant1")  // ID = "user1:tenant1"
```

## Relationship Resolution

### AllowMissingEmbeddedRecords

When false (default), missing related records cause errors:

```go
collection.AllowMissingEmbeddedRecords = false  // Default

// If post.author_id references non-existent user
// Query will fail or return partial data
```

**Set to true** for optional relationships:

```go
collection.AllowMissingEmbeddedRecords = true
// Missing author returns null instead of error
```

### Circular References

Be careful with bidirectional relationships:

```go
// Users embed Groups, Groups embed Users
// This can cause infinite loops or stack overflow
```

**Solution**: Use `NoEmbed: true` on one side of the relationship.

### Expansion Performance

Auto-expansion triggers N+1 queries:

```go
// Query returns 100 users
// Each user has author relationship
// = 1 query + 100 relationship queries
```

**Solutions**:
1. Use `?noexpand=true` when you don't need related data
2. Batch load relationships manually
3. Use database-level joins (SQL backends)

## Migration Pitfalls

### Infinite Loop Risk

There's a potential infinite loop in migration when a collection references itself. From mapper comments:

```go
// TODO: potential infinite loop if collection A references B and B references A
```

**Workaround**: Migrate collections in dependency order manually.

### Type Changes

Changing field types after data exists can lose data:

```go
// Original: age as string "25"
// Changed to: age as int
// Migration may fail or truncate
```

**Recommendation**: Test migrations on a copy of production data first.

## Connection String Gotchas

### Trailing Slashes

Some backends are sensitive to trailing slashes:

```go
// May behave differently
"mysql://localhost/db"
"mysql://localhost/db/"
```

### Special Characters in Passwords

URL-encode special characters:

```go
// Password is "p@ss/word"
"mysql://user:p%40ss%2Fword@localhost/db"
```

### localhost vs 127.0.0.1

Some systems treat these differently:

```go
// May use Unix socket
"mysql://localhost/db"

// Forces TCP connection
"mysql://127.0.0.1/db"
```

## Testing Considerations

### Test Database Cleanup

Always clean up after tests:

```go
func TestSomething(t *testing.T) {
    backend, _ := pivot.NewDatabase("sqlite:///:memory:")
    defer backend.Flush()  // Cleanup
    // ...
}
```

### Parallel Test Safety

Don't share backends across parallel tests:

```go
// Dangerous
var sharedBackend Backend

func TestA(t *testing.T) {
    t.Parallel()
    sharedBackend.Insert(...)  // Race condition
}
```

### Docker Container Timing

Container tests may fail on slow systems:

```go
// Increase retry timeout
pool.MaxWait = 2 * time.Minute
```

## Performance Considerations

### Large RecordSets

Query without limits can return huge datasets:

```go
// Bad: could return millions of records
indexer.Query(collection, filter.All())

// Good: always paginate
f := filter.All()
f.Limit = 100
f.Offset = 0
```

### Index Usage

Ensure your queries use indexes:

```go
// If 'status' isn't indexed, this scans entire table
filter.Parse("status/is:active")

// Add index in schema
{
    "name": "status",
    "type": "str",
    "indexed": true
}
```

### Bleve Index Size

Bleve indexes grow on disk and can become large:

```go
// Monitor index directory size
// Consider periodic compaction
indexer.FlushIndex()
```

## Security Considerations

### SQL Injection

Pivot uses parameterized queries, but be careful with raw queries:

```go
// Safe - uses placeholders
filter.Parse("name/is:" + userInput)

// Dangerous - if you bypass Pivot
db.Query("SELECT * FROM users WHERE name = '" + userInput + "'")
```

### Credential Exposure

Connection strings in logs can expose credentials:

```go
// Logged connection string shows password
log.Info("Connecting to: ", connectionString)

// Better: redact credentials
log.Info("Connecting to: ", connectionString.Host())
```

### Netrc File Permissions

If using `.netrc` for credentials:

```bash
# Must be readable only by owner
chmod 600 ~/.netrc
```

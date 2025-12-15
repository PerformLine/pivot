# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test Commands

```bash
make all          # Run deps, fmt, test, build, docs
make test         # Run tests: go test -count=1 --tags json1 ./...
make fmt          # Format code, run go vet, tidy modules
make build        # Build CLI to bin/pivot
make deps         # Install dependencies
```

Single test file: `go test -count=1 --tags json1 ./dal/`

The `--tags json1` flag is required for SQLite JSON extension support.

## Architecture

Pivot is a multi-database abstraction library providing a unified interface for diverse database systems.

### Package Structure

- **pivot (root)**: Entry point - `NewDatabase()`, `LoadSchemata()`, `LoadFixtures()`
- **dal/**: Data Abstraction Layer - Collection, Record, RecordSet, Field definitions
- **backends/**: Database adapters (SQL, MongoDB, Redis, DynamoDB, Filesystem, Elasticsearch, Bleve)
- **filter/**: Database-agnostic query representation with URL-friendly syntax
- **mapper/**: High-level ORM/ODM layer with struct tag support
- **v4/**: Next-generation version (in development, mirrors main structure)

### Key Interfaces

- **Backend** (`backends/backends.go`): Core CRUD operations - Initialize, Insert, Retrieve, Update, Delete, CreateCollection
- **Indexer** (`backends/indexers.go`): Query/search operations - Index, Query, QueryFunc, DeleteQuery
- **Mapper** (`mapper/model.go`): ORM operations - Create, Get, All, Update, Delete, Migrate

### Connection Strings

Format: `scheme://[user[:password]@]host[:port][/path][?options]`

Examples:
- `sqlite:///./test.db`
- `mysql://user:pass@localhost:3306/database`
- `postgres://localhost/dbname`
- `mongodb://localhost/dbname`

### Struct Tags

```go
type Example struct {
    ID   string `pivot:"id,identity"`
    Name string `pivot:"name"`
}
```

### Filter Syntax

URL-friendly query syntax: `field/[operator:]value`

Operators: `is` (default), `not`, `contains`, `like`, `unlike`, `prefix`, `suffix`, `gt`, `gte`, `lt`, `lte`, `range`

Examples:
- `id/123` - exact match
- `name/contains:test` - substring match
- `price/range:10|20` - between 10 (inclusive) and 20 (exclusive)

See `filter/README.md` for full documentation.

### Test Infrastructure

- Schema files: `test/schema/*.json`
- Fixtures: `test/fixtures/*.json`
- Main test entry: `db_test.go` with `TestAll()`
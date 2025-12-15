# Pivot Architecture Overview

## Purpose

Pivot is a multi-database abstraction library that provides a unified interface for accessing, querying, and aggregating data across diverse database systems and search engines. It decouples application code from specific database implementations while providing both low-level backend operations and high-level ORM/ODM semantics.

## Tech Stack

| Category | Technology | Notes |
|----------|------------|-------|
| **Language** | Go 1.22 | Toolchain go1.24.1 |
| **SQL Databases** | MySQL, PostgreSQL, SQLite | Full backend + indexer support |
| **NoSQL Databases** | MongoDB, Redis, DynamoDB | Varying indexer support |
| **Search Engines** | Elasticsearch, Bleve | Indexer-only (search/query) |
| **Filesystem** | File/FS backends | JSON, YAML, CSV storage |
| **HTTP Router** | husobee/vestigo | REST API routing |
| **HTTP Middleware** | urfave/negroni | Request handling |
| **Templates** | PerformLine/diecast | Web UI rendering |
| **CLI Framework** | urfave/cli | Command-line interface |
| **Struct Utilities** | fatih/structs | Reflection helpers |

## Architecture Layers

```mermaid
flowchart TB
    subgraph "Application Layer"
        APP[Application Code]
        CLI[CLI Commands]
        HTTP[REST API Server]
    end

    subgraph "ORM Layer"
        MAPPER[Mapper Interface]
        MODEL[Model Operations]
    end

    subgraph "Abstraction Layer"
        BACKEND[Backend Interface]
        INDEXER[Indexer Interface]
        AGGREGATOR[Aggregator Interface]
    end

    subgraph "Filter Layer"
        FILTER[Filter Parser]
        GEN[Generators]
        SQL_GEN[SQL Generator]
        MONGO_GEN[MongoDB Generator]
        ES_GEN[Elasticsearch Generator]
    end

    subgraph "Backend Implementations"
        SQL[SQL Backend]
        MONGO[MongoDB Backend]
        REDIS[Redis Backend]
        DYNAMO[DynamoDB Backend]
        FS[Filesystem Backend]
    end

    subgraph "Data Storage"
        MYSQL[(MySQL)]
        PG[(PostgreSQL)]
        SQLITE[(SQLite)]
        MONGODB[(MongoDB)]
        REDISDB[(Redis)]
        DYNAMODB[(DynamoDB)]
        FILES[(Files)]
        ES[(Elasticsearch)]
        BLEVE[(Bleve)]
    end

    APP --> MAPPER
    CLI --> MAPPER
    HTTP --> MAPPER

    MAPPER --> MODEL
    MODEL --> BACKEND
    MODEL --> INDEXER

    BACKEND --> SQL
    BACKEND --> MONGO
    BACKEND --> REDIS
    BACKEND --> DYNAMO
    BACKEND --> FS

    INDEXER --> FILTER
    FILTER --> GEN
    GEN --> SQL_GEN
    GEN --> MONGO_GEN
    GEN --> ES_GEN

    SQL --> MYSQL
    SQL --> PG
    SQL --> SQLITE
    MONGO --> MONGODB
    REDIS --> REDISDB
    DYNAMO --> DYNAMODB
    FS --> FILES

    INDEXER -.-> ES
    INDEXER -.-> BLEVE

    style APP fill:#e1f5fe
    style CLI fill:#e1f5fe
    style HTTP fill:#e1f5fe
    style MAPPER fill:#fff3e0
    style BACKEND fill:#f3e5f5
    style INDEXER fill:#f3e5f5
```

## Package Structure

| Package | Import Path | Description |
|---------|-------------|-------------|
| **pivot** | `github.com/PerformLine/pivot/v3` | Entry point: `NewDatabase()`, `LoadSchemata()`, `LoadFixtures()` |
| **dal** | `github.com/PerformLine/pivot/v3/dal` | Data Abstraction Layer: Collection, Record, Field, RecordSet |
| **backends** | `github.com/PerformLine/pivot/v3/backends` | Backend/Indexer implementations and interfaces |
| **filter** | `github.com/PerformLine/pivot/v3/filter` | Query filter parsing and representation |
| **filter/generators** | `github.com/PerformLine/pivot/v3/filter/generators` | Backend-specific query generators |
| **mapper** | `github.com/PerformLine/pivot/v3/mapper` | High-level ORM/ODM operations |
| **client** | `github.com/PerformLine/pivot/v3/client` | HTTP client for REST API |
| **util** | `github.com/PerformLine/pivot/v3/util` | HTTP utilities, feature flags |
| **v4** | `github.com/PerformLine/pivot/v4` | Next-generation with context support |

## Key Entry Points

### Library Usage

```go
// Connect to database
backend, err := pivot.NewDatabase("sqlite:///./data.db")
backend.Initialize()

// Load schema definitions
pivot.LoadSchemata(backend, "path/to/schemas")

// Create mapper for ORM operations
widgets := mapper.NewModel(backend, WidgetsSchema)
widgets.Migrate()
```

### CLI Commands

```bash
pivot web CONNECTION [INDEXER]  # Start REST API server
pivot diff CONNECTION SCHEMA    # Compare schema vs database
pivot copy SRC DST              # Copy data between sources
pivot filter FILTER_STRING      # Convert filter to native query
```

### REST API

```
GET  /api/status                           # Server status
GET  /api/collections                      # List collections
GET  /api/schema/:collection               # Get collection schema
GET  /api/collections/:collection/where/*  # Query with filter
POST /api/collections/:collection/records  # Create records
```

## Supported Backends

| System | Backend | Indexer | Aggregator | Notes |
|--------|:-------:|:-------:|:----------:|-------|
| MySQL/MariaDB | Yes | Yes | Yes | Full support |
| PostgreSQL | Yes | Yes | Yes | Full support |
| SQLite 3.x | Yes | Yes | Yes | Requires `--tags json1` |
| MongoDB | Yes | Yes | Yes | Full support |
| Redis | Yes | Partial | No | Key-based retrieval only |
| DynamoDB | Yes | Partial | No | Range/Sort key queries only |
| Filesystem | Yes | Yes | Yes | JSON/YAML/CSV storage |
| Elasticsearch | No | Yes | Yes | Search indexing only |
| Bleve | No | Yes | No | Embedded full-text search |

## Configuration

### Connection Strings

Format: `scheme://[user:pass@]host[:port]/dataset[?options]`

```
sqlite:///./data.db
mysql://user:pass@localhost:3306/database
postgres://localhost/dbname?sslmode=disable
mongodb://localhost/dbname
dynamodb://us-east-1/tablename
redis://localhost:6379/0
file:///path/to/data.json
```

### Environment Support

```go
config := pivot.Configuration{
    Backend:  "sqlite:///./data.db",
    Indexer:  "bleve:///./search.idx",
    Environments: map[string]Configuration{
        "production": {
            Backend: "postgres://prod-host/db",
        },
    },
}
config.ForEnv("production")  // Returns production config
```

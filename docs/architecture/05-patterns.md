# Coding Patterns

This document describes the key architectural patterns and conventions used in Pivot.

## Backend Interface Pattern

The Backend interface provides a pluggable storage abstraction. All database-specific code implements this interface.

### Interface Definition

```go
// backends/backends.go
type Backend interface {
    // Lifecycle
    Initialize() error
    SetIndexer(dal.ConnectionString) error
    Ping(time.Duration) error
    Flush() error

    // Schema Operations
    RegisterCollection(*dal.Collection)
    GetConnectionString() *dal.ConnectionString
    CreateCollection(definition *dal.Collection) error
    DeleteCollection(collection string) error
    ListCollections() ([]string, error)
    GetCollection(collection string) (*dal.Collection, error)

    // Record CRUD
    Exists(collection string, id interface{}) bool
    Retrieve(collection string, id interface{}, fields ...string) (*dal.Record, error)
    Insert(collection string, records *dal.RecordSet) error
    Update(collection string, records *dal.RecordSet, target ...string) error
    Delete(collection string, ids ...interface{}) error

    // Search & Aggregation
    WithSearch(collection *dal.Collection, filters ...*filter.Filter) Indexer
    WithAggregator(collection *dal.Collection) Aggregator

    // Feature Support
    Supports(feature ...BackendFeature) bool
    String() string
}
```

### Implementation Pattern

```go
type SqlBackend struct {
    Backend                         // Embed interface (not required, just documentation)
    conn    *dal.ConnectionString
    db      *sql.DB
    indexer Indexer
    // ... backend-specific fields
}

func NewSqlBackend(conn dal.ConnectionString) *SqlBackend {
    return &SqlBackend{
        conn: &conn,
    }
}

func (self *SqlBackend) Initialize() error {
    // Connect to database
    db, err := sql.Open(self.conn.Backend(), self.conn.String())
    if err != nil {
        return err
    }
    self.db = db
    return nil
}
```

### Registration Pattern

Backends auto-register with the factory:

```go
// backends/backends.go
func init() {
    RegisterBackend("mysql", NewSqlBackend)
    RegisterBackend("postgres", NewSqlBackend)
    RegisterBackend("sqlite", NewSqlBackend)
    RegisterBackend("mongodb", NewMongoBackend)
    // ...
}

// Usage
backend := MakeBackend(connectionString)
```

## Indexer Interface Pattern

Separates search/query capability from storage operations.

### Interface Definition

```go
// backends/indexers.go
type Indexer interface {
    IndexConnectionString() *dal.ConnectionString
    IndexInitialize(Backend) error
    IndexExists(collection *dal.Collection, id interface{}) bool
    IndexRetrieve(collection *dal.Collection, id interface{}) (*dal.Record, error)
    IndexRemove(collection *dal.Collection, ids []interface{}) error
    Index(collection *dal.Collection, records *dal.RecordSet) error

    // Query Capabilities
    QueryFunc(collection *dal.Collection, filter *filter.Filter, resultFn IndexResultFunc) error
    Query(collection *dal.Collection, filter *filter.Filter, resultFns ...IndexResultFunc) (*dal.RecordSet, error)
    ListValues(collection *dal.Collection, fields []string, filter *filter.Filter) (map[string][]interface{}, error)
    DeleteQuery(collection *dal.Collection, f *filter.Filter) error

    FlushIndex() error
    GetBackend() Backend
}
```

### Dual Index Pattern

A backend can serve as its own indexer (SQL, MongoDB) or use a separate indexer (Backend + Elasticsearch):

```go
// Same system for storage and query
backend, _ := pivot.NewDatabase("mysql://localhost/db")
// SqlBackend implements both Backend and Indexer

// Separate systems
backend, _ := pivot.NewDatabaseWithOptions("mongodb://localhost/db", backends.ConnectOptions{
    Indexer: "elasticsearch://localhost:9200/index",
})
// MongoBackend for storage, ElasticsearchIndexer for queries
```

## Generator Pattern

Compiles abstract filters to backend-specific queries.

### Interface Definition

```go
// filter/generator.go
type IGenerator interface {
    Initialize(string) error
    Finalize(*Filter) error
    Push([]byte)
    Set([]byte)
    Payload() []byte
    WithCriterion(Criterion) error
    OrCriterion(Criterion) error
    WithField(string) error
    GroupByField(string) error
    AggregateByField(Aggregation, string) error
    SetOption(string, interface{}) error
    GetValues() []interface{}
    Reset()
}
```

### SQL Generator Example

```go
// filter/generators/sql.go
type Sql struct {
    buf          bytes.Buffer
    values       []interface{}
    TypeMapping  SqlTypeMapping
    // ...
}

func (self *Sql) WithCriterion(criterion Criterion) error {
    switch criterion.Operator {
    case "is":
        self.buf.WriteString(fmt.Sprintf("%s = ?", criterion.Field))
        self.values = append(self.values, criterion.Values[0])
    case "contains":
        self.buf.WriteString(fmt.Sprintf("%s LIKE ?", criterion.Field))
        self.values = append(self.values, "%"+criterion.Values[0].(string)+"%")
    // ... other operators
    }
    return nil
}
```

### Generator Selection

```go
func GetGenerator(name string) IGenerator {
    switch name {
    case "sql":
        return new(generators.Sql)
    case "mongodb":
        return new(generators.MongoDB)
    case "elasticsearch":
        return new(generators.Elasticsearch)
    }
    return nil
}
```

## Mapper/ORM Pattern

High-level object mapping over backends.

### Model Creation

```go
// mapper/model.go
func NewModel(backend Backend, collection *dal.Collection) *Model {
    return &Model{
        backend:    backend,
        collection: collection,
    }
}
```

### Struct Tag Convention

```go
type User struct {
    ID        string    `pivot:"id,identity"`      // Primary key
    Email     string    `pivot:"email"`            // Regular field
    Name      string    `pivot:"name"`
    CreatedAt time.Time `pivot:"created_at"`
    UpdatedAt time.Time `pivot:"updated_at,omitempty"`  // Skip if empty
}
```

### Tag Options

| Option | Description |
|--------|-------------|
| `identity` | Marks field as primary key |
| `omitempty` | Skip field if zero value |
| `-` | Ignore field entirely |

### CRUD Operations

```go
users := mapper.NewModel(backend, UsersSchema)

// Create
user := User{Email: "test@example.com", Name: "Test"}
users.Create(&user)  // user.ID now populated

// Read
var found User
users.Get(user.ID, &found)

// Update
found.Name = "Updated"
users.Update(&found)

// Delete
users.Delete(found.ID)

// Query
var results []User
users.All(filter.Parse("name/contains:test"), &results)
```

## Embedded Record Backend Pattern

Wraps any backend to automatically expand relationships.

```go
// backends/embedded-record-backend.go
type EmbeddedRecordBackend struct {
    Backend
    wrapped Backend
}

func NewEmbeddedRecordBackend(backend Backend) *EmbeddedRecordBackend {
    return &EmbeddedRecordBackend{
        wrapped: backend,
    }
}

func (self *EmbeddedRecordBackend) Retrieve(collection string, id interface{}, fields ...string) (*dal.Record, error) {
    record, err := self.wrapped.Retrieve(collection, id, fields...)
    if err != nil {
        return nil, err
    }

    // Expand embedded collections
    PopulateRelationships(self.wrapped, collection, record)

    return record, nil
}
```

## Multi-Indexer Pattern

Strategy-based indexer selection for distributed setups.

```go
// backends/indexer-multi.go
type MultiIndex struct {
    RetrievalStrategy  IndexSelectionStrategy  // For queries
    PersistStrategy    IndexSelectionStrategy  // For indexing
    DeleteStrategy     IndexSelectionStrategy  // For deletion
    indexers           []Indexer
}

type IndexSelectionStrategy int

const (
    Sequential     IndexSelectionStrategy = iota  // Try each until success
    All                                           // Execute on all
    First                                         // First indexer only
    AllExceptFirst                                // Skip first, use all others
    Random                                        // Random selection
)
```

### Usage

```go
multi := NewMultiIndex(
    sqlIndexer,
    elasticsearchIndexer,
    bleveIndexer,
)
multi.RetrievalStrategy = Sequential  // Try SQL, then ES, then Bleve
multi.PersistStrategy = All           // Index to all
```

## Error Handling Pattern

### Collection Not Found

```go
if dal.IsCollectionNotFoundErr(err) {
    // Handle missing collection
}
```

### Record Errors

```go
record, err := backend.Retrieve("users", id)
if err != nil {
    return err
}
if record == nil {
    return fmt.Errorf("record not found")
}
if record.Error != nil {
    return record.Error
}
```

## Fixture Loading Pattern

### JSON Fixture Format

```json
[
    {
        "id": "user-1",
        "fields": {
            "name": "John Doe",
            "email": "john@example.com"
        }
    },
    {
        "id": "user-2",
        "fields": {
            "name": "Jane Doe",
            "email": "jane@example.com"
        },
        "optional": true
    }
]
```

### Loading Fixtures

```go
pivot.LoadFixtures(backend, "path/to/fixtures")
// Loads all .json and .yaml files
// Files named <collection>.json load into that collection
```

## Schema Definition Pattern

### Go Code

```go
var WidgetsSchema = &dal.Collection{
    Name:                   "widgets",
    IdentityFieldType:      dal.StringType,
    IdentityFieldFormatter: dal.GenerateUUID,
    Fields: []dal.Field{
        {
            Name:      "type",
            Type:      dal.StringType,
            Required:  true,
            Validator: dal.ValidateIsOneOf("foo", "bar", "baz"),
        },
    },
}
```

### JSON File

```json
{
    "name": "widgets",
    "identityFieldType": "str",
    "fields": [
        {
            "name": "type",
            "type": "str",
            "required": true
        }
    ]
}
```

### Loading Schemas

```go
pivot.LoadSchemata(backend, "path/to/schemas")
// Loads all .json and .yaml files as collection definitions
```

## Connection Options Pattern

```go
backend, err := pivot.NewDatabaseWithOptions(
    "mysql://localhost/db",
    backends.ConnectOptions{
        Indexer:               "elasticsearch://localhost:9200",
        SkipInitialize:        false,
        AutocreateCollections: true,
    },
)
```

## Testing Pattern

### Main Test Entry

```go
// db_test.go
func TestAll(t *testing.T) {
    // Setup backend
    backend, _ := pivot.NewDatabase("sqlite:///:memory:")
    backend.Initialize()

    // Load test schemas
    pivot.LoadSchemata(backend, "test/schema")

    // Load test fixtures
    pivot.LoadFixtures(backend, "test/fixtures")

    // Run subtests
    t.Run("BasicCRUD", func(t *testing.T) {
        testBasicCRUD(t, backend)
    })
}
```

### Docker Test Containers

```go
import "github.com/ory/dockertest"

func setupMySQLContainer() (*sql.DB, func()) {
    pool, _ := dockertest.NewPool("")
    resource, _ := pool.Run("mysql", "5.7", []string{
        "MYSQL_ROOT_PASSWORD=secret",
    })

    cleanup := func() {
        pool.Purge(resource)
    }

    // Wait for container
    pool.Retry(func() error {
        db, err := sql.Open("mysql", "root:secret@tcp(localhost)/test")
        return db.Ping()
    })

    return db, cleanup
}
```

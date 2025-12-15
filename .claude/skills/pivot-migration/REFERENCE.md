# Pivot Migration Technical Reference

Detailed technical reference for migrating from Pivot to standard Go database libraries.

## Pivot Import Patterns to Search For

```go
// Main imports to find
"github.com/PerformLine/pivot/v3"
"github.com/PerformLine/pivot/v3/dal"
"github.com/PerformLine/pivot/v3/backends"
"github.com/PerformLine/pivot/v3/filter"
"github.com/PerformLine/pivot/v3/mapper"

// v4 imports (if used)
"github.com/PerformLine/pivot/v4"
"github.com/PerformLine/pivot/v4/dal"
"github.com/PerformLine/pivot/v4/backends"
```

## Grep Patterns for Analysis

```bash
# Find all Pivot imports
grep -r "PerformLine/pivot" --include="*.go"

# Find pivot struct tags
grep -r 'pivot:"' --include="*.go"

# Find filter.Parse usage
grep -r "filter\.Parse" --include="*.go"

# Find mapper usage
grep -r "mapper\." --include="*.go"

# Find backend initialization
grep -r "pivot\.NewDatabase" --include="*.go"
grep -r "backends\.New" --include="*.go"

# Find schema files
find . -name "*.json" -path "*/schema/*"
```

## Complete Backend Interface Migration

### Backend Methods

```go
// PIVOT
type Backend interface {
    Initialize() error
    SetIndexer(dal.ConnectionString) error
    Ping(time.Duration) error
    Flush() error
    RegisterCollection(*dal.Collection)
    GetConnectionString() *dal.ConnectionString
    CreateCollection(definition *dal.Collection) error
    DeleteCollection(collection string) error
    ListCollections() ([]string, error)
    GetCollection(collection string) (*dal.Collection, error)
    Exists(collection string, id interface{}) bool
    Retrieve(collection string, id interface{}, fields ...string) (*dal.Record, error)
    Insert(collection string, records *dal.RecordSet) error
    Update(collection string, records *dal.RecordSet, target ...string) error
    Delete(collection string, ids ...interface{}) error
    WithSearch(collection *dal.Collection, filters ...*filter.Filter) Indexer
    WithAggregator(collection *dal.Collection) Aggregator
    Supports(feature ...BackendFeature) bool
}

// GORM EQUIVALENT PATTERNS
// Initialize() →
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

// Ping() →
sqlDB, _ := db.DB()
sqlDB.Ping()

// CreateCollection() →
db.AutoMigrate(&Model{})

// DeleteCollection() →
db.Migrator().DropTable("table_name")

// ListCollections() →
db.Migrator().GetTables()

// Exists() →
var count int64
db.Model(&Model{}).Where("id = ?", id).Count(&count)
exists := count > 0

// Retrieve() →
var record Model
db.First(&record, "id = ?", id)

// Insert() →
db.Create(&records)

// Update() →
db.Save(&record)

// Delete() →
db.Delete(&Model{}, ids)
```

### Indexer Methods

```go
// PIVOT
type Indexer interface {
    Query(collection *dal.Collection, filter *filter.Filter, resultFns ...IndexResultFunc) (*dal.RecordSet, error)
    QueryFunc(collection *dal.Collection, filter *filter.Filter, resultFn IndexResultFunc) error
    ListValues(collection *dal.Collection, fields []string, filter *filter.Filter) (map[string][]interface{}, error)
    DeleteQuery(collection *dal.Collection, f *filter.Filter) error
}

// GORM EQUIVALENT PATTERNS
// Query() →
var results []Model
db.Where("status = ?", "active").Find(&results)

// QueryFunc() →
rows, _ := db.Model(&Model{}).Where("status = ?", "active").Rows()
defer rows.Close()
for rows.Next() {
    var m Model
    db.ScanRows(rows, &m)
    // process m
}

// ListValues() →
var values []string
db.Model(&Model{}).Distinct("field").Pluck("field", &values)

// DeleteQuery() →
db.Where("status = ?", "inactive").Delete(&Model{})
```

### Aggregator Methods

```go
// PIVOT
type Aggregator interface {
    Sum(collection *dal.Collection, field string, f ...*filter.Filter) (float64, error)
    Count(collection *dal.Collection, f ...*filter.Filter) (uint64, error)
    Minimum(collection *dal.Collection, field string, f ...*filter.Filter) (float64, error)
    Maximum(collection *dal.Collection, field string, f ...*filter.Filter) (float64, error)
    Average(collection *dal.Collection, field string, f ...*filter.Filter) (float64, error)
    GroupBy(collection *dal.Collection, fields []string, aggregates []filter.Aggregate, f ...*filter.Filter) (*dal.RecordSet, error)
}

// GORM EQUIVALENT PATTERNS
// Count() →
var count int64
db.Model(&Model{}).Where("status = ?", "active").Count(&count)

// Sum() →
var sum float64
db.Model(&Model{}).Where("status = ?", "active").Select("COALESCE(SUM(amount), 0)").Scan(&sum)

// Minimum() →
var min float64
db.Model(&Model{}).Where("status = ?", "active").Select("MIN(amount)").Scan(&min)

// Maximum() →
var max float64
db.Model(&Model{}).Where("status = ?", "active").Select("MAX(amount)").Scan(&max)

// Average() →
var avg float64
db.Model(&Model{}).Where("status = ?", "active").Select("AVG(amount)").Scan(&avg)

// GroupBy() →
type Result struct {
    Status string
    Total  float64
    Count  int64
}
var results []Result
db.Model(&Model{}).Select("status, SUM(amount) as total, COUNT(*) as count").Group("status").Scan(&results)
```

## Complete Filter Operator Translation

### Comparison Operators

| Pivot | Operator | SQL | GORM |
|-------|----------|-----|------|
| `field/value` | `is` (default) | `field = 'value'` | `Where("field = ?", value)` |
| `field/is:value` | `is` | `field = 'value'` | `Where("field = ?", value)` |
| `field/not:value` | `not` | `field != 'value'` | `Where("field != ?", value)` |
| `field/gt:N` | `gt` | `field > N` | `Where("field > ?", N)` |
| `field/gte:N` | `gte` | `field >= N` | `Where("field >= ?", N)` |
| `field/lt:N` | `lt` | `field < N` | `Where("field < ?", N)` |
| `field/lte:N` | `lte` | `field <= N` | `Where("field <= ?", N)` |
| `field/range:A\|B` | `range` | `field >= A AND field < B` | `Where("field >= ? AND field < ?", A, B)` |

### String Operators

| Pivot | Operator | SQL | GORM |
|-------|----------|-----|------|
| `field/contains:x` | `contains` | `field LIKE '%x%'` | `Where("field LIKE ?", "%x%")` |
| `field/like:x` | `like` | `LOWER(field) LIKE '%x%'` | `Where("LOWER(field) LIKE ?", "%x%")` |
| `field/unlike:x` | `unlike` | `LOWER(field) NOT LIKE '%x%'` | `Where("LOWER(field) NOT LIKE ?", "%x%")` |
| `field/prefix:x` | `prefix` | `field LIKE 'x%'` | `Where("field LIKE ?", "x%")` |
| `field/suffix:x` | `suffix` | `field LIKE '%x'` | `Where("field LIKE ?", "%x")` |

### Multi-Value (OR within field)

```go
// Pivot: field/a|b|c
// SQL: field IN ('a', 'b', 'c')
// GORM:
db.Where("field IN ?", []string{"a", "b", "c"})
```

### Multi-Field (AND between fields)

```go
// Pivot: name/john/age/gt:18/status/active
// SQL: name = 'john' AND age > 18 AND status = 'active'
// GORM:
db.Where("name = ?", "john").
    Where("age > ?", 18).
    Where("status = ?", "active")
```

### Type Prefixes

```go
// Pivot supports type hints: str:, int:, float:, bool:, time:
// str:code/123   → treats 123 as string
// int:count/gt:5 → treats 5 as integer

// In GORM, types are inferred from Go variables
db.Where("code = ?", "123")      // string
db.Where("count > ?", 5)         // int
db.Where("price > ?", 19.99)     // float
db.Where("active = ?", true)     // bool
db.Where("created_at > ?", time) // time.Time
```

## Connection String Conversions

### MySQL

```go
// Pivot
"mysql://user:pass@localhost:3306/dbname"
"mysql://user:pass@localhost:3306/dbname?timeout=30s"

// GORM
"user:pass@tcp(localhost:3306)/dbname?parseTime=true"
"user:pass@tcp(localhost:3306)/dbname?parseTime=true&timeout=30s"

// Code
import "gorm.io/driver/mysql"
dsn := "user:pass@tcp(localhost:3306)/dbname?parseTime=true"
db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})
```

### PostgreSQL

```go
// Pivot
"postgres://user:pass@localhost:5432/dbname"
"postgresql://user:pass@localhost/dbname?sslmode=disable"

// GORM
"host=localhost user=user password=pass dbname=dbname port=5432 sslmode=disable"
// Or URL format:
"postgres://user:pass@localhost:5432/dbname?sslmode=disable"

// Code
import "gorm.io/driver/postgres"
dsn := "host=localhost user=user password=pass dbname=dbname port=5432 sslmode=disable"
db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
```

### SQLite

```go
// Pivot
"sqlite:///./data.db"
"sqlite:///path/to/data.db"

// GORM
"./data.db"
"file:./data.db?cache=shared&mode=rwc"

// Code
import "gorm.io/driver/sqlite"
db, _ := gorm.Open(sqlite.Open("./data.db"), &gorm.Config{})
```

### MongoDB

```go
// Pivot
"mongodb://localhost/dbname"
"mongodb://user:pass@localhost:27017/dbname"

// mongo-go-driver
"mongodb://localhost:27017"
"mongodb://user:pass@localhost:27017"

// Code
import "go.mongodb.org/mongo-driver/mongo"
client, _ := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
db := client.Database("dbname")
```

### Redis

```go
// Pivot
"redis://localhost:6379/0"
"redis://:password@localhost:6379/0"

// go-redis
// Code
import "github.com/redis/go-redis/v9"
rdb := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "password",
    DB:       0,
})
```

### DynamoDB

```go
// Pivot
"dynamodb://us-east-1/tablename"

// aws-sdk-go-v2
// Code
import (
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
)
cfg, _ := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
client := dynamodb.NewFromConfig(cfg)
```

### OpenSearch/Elasticsearch

```go
// Pivot
"elasticsearch://localhost:9200/indexname"

// opensearch-go v4
// Code
import "github.com/opensearch-project/opensearch-go/v4"
client, _ := opensearchapi.NewClient(opensearchapi.Config{
    Addresses: []string{"http://localhost:9200"},
})
```

## Schema JSON to GORM Model

### Input: Pivot Schema

```json
{
    "name": "users",
    "identityFieldType": "str",
    "fields": [
        {"name": "email", "type": "str", "required": true, "unique": true},
        {"name": "name", "type": "str", "required": true},
        {"name": "age", "type": "int"},
        {"name": "balance", "type": "float", "default": 0},
        {"name": "active", "type": "bool", "default": true},
        {"name": "metadata", "type": "object"},
        {"name": "tags", "type": "array"},
        {"name": "created_at", "type": "time"},
        {"name": "updated_at", "type": "time"}
    ]
}
```

### Output: GORM Model

```go
import (
    "time"
    "gorm.io/datatypes"
)

type User struct {
    ID        string         `gorm:"primaryKey;type:varchar(36)"`
    Email     string         `gorm:"uniqueIndex;not null"`
    Name      string         `gorm:"not null"`
    Age       int64          `gorm:"default:null"`
    Balance   float64        `gorm:"default:0"`
    Active    bool           `gorm:"default:true"`
    Metadata  datatypes.JSON `gorm:"type:json"`
    Tags      datatypes.JSON `gorm:"type:json"`
    CreatedAt time.Time      `gorm:"autoCreateTime"`
    UpdatedAt time.Time      `gorm:"autoUpdateTime"`
}

// TableName overrides the table name
func (User) TableName() string {
    return "users"
}

// BeforeCreate hook for UUID generation (replaces IdentityFieldFormatter)
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.ID == "" {
        u.ID = uuid.New().String()
    }
    return nil
}
```

## Validator Migration

### Pivot Validators → Go Validation

```go
// Pivot
dal.ValidateIsOneOf("a", "b", "c")
dal.ValidateNotEmpty
dal.ValidatePositiveInteger
dal.ValidatePositiveOrZeroInteger

// go-playground/validator (with gin)
type Model struct {
    Status  string `validate:"oneof=a b c"`
    Name    string `validate:"required"`
    Count   int    `validate:"gt=0"`
    Amount  int    `validate:"gte=0"`
}

// Or manual validation
func (m *Model) Validate() error {
    validStatuses := map[string]bool{"a": true, "b": true, "c": true}
    if !validStatuses[m.Status] {
        return errors.New("invalid status")
    }
    if m.Name == "" {
        return errors.New("name required")
    }
    return nil
}
```

## Formatter Migration

### Pivot Formatters → GORM Hooks

```go
// Pivot
dal.GenerateUUID           // Auto-generate UUID
dal.CurrentTime            // Set current time always
dal.CurrentTimeIfUnset     // Set current time if empty
dal.TrimSpace              // Trim whitespace
dal.FormatLowerCase        // Lowercase

// GORM Hooks
func (m *Model) BeforeCreate(tx *gorm.DB) error {
    // GenerateUUID
    if m.ID == "" {
        m.ID = uuid.New().String()
    }

    // TrimSpace
    m.Name = strings.TrimSpace(m.Name)

    // FormatLowerCase
    m.Email = strings.ToLower(m.Email)

    // CurrentTimeIfUnset
    if m.CreatedAt.IsZero() {
        m.CreatedAt = time.Now()
    }

    return nil
}

func (m *Model) BeforeSave(tx *gorm.DB) error {
    // CurrentTime (always update)
    m.UpdatedAt = time.Now()
    return nil
}
```

## Relationship Migration

### Pivot Embedded Collections → GORM Associations

```go
// Pivot Schema
{
    "name": "posts",
    "embeddedCollections": [
        {
            "collectionName": "users",
            "keys": "author_id",
            "fields": ["name", "email"]
        }
    ]
}

// GORM Model
type Post struct {
    ID       string `gorm:"primaryKey"`
    Title    string
    AuthorID string
    Author   User   `gorm:"foreignKey:AuthorID"`
}

type User struct {
    ID    string `gorm:"primaryKey"`
    Name  string
    Email string
}

// Query with preload (replaces auto-expand)
var post Post
db.Preload("Author").First(&post, "id = ?", postID)

// Or with specific fields
db.Preload("Author", func(db *gorm.DB) *gorm.DB {
    return db.Select("id", "name", "email")
}).First(&post, "id = ?", postID)
```

## Transaction Migration

```go
// Pivot (implicit transactions per operation)
backend.Insert(collection, recordset)

// GORM explicit transactions
err := db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&user).Error; err != nil {
        return err // rollback
    }
    if err := tx.Create(&profile).Error; err != nil {
        return err // rollback
    }
    return nil // commit
})
```

## Error Handling Migration

```go
// Pivot
if dal.IsCollectionNotFoundErr(err) {
    // handle
}

// GORM
if errors.Is(err, gorm.ErrRecordNotFound) {
    // handle
}

// Check for specific errors
switch {
case errors.Is(err, gorm.ErrRecordNotFound):
    // not found
case errors.Is(err, gorm.ErrDuplicatedKey):
    // duplicate
default:
    // other error
}
```

## Testing Migration

### Pivot Test Setup

```go
// Pivot
backend, _ := pivot.NewDatabase("sqlite:///:memory:")
backend.Initialize()
pivot.LoadSchemata(backend, "test/schema")
pivot.LoadFixtures(backend, "test/fixtures")
```

### GORM Test Setup

```go
// GORM with SQLite in-memory
db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
db.AutoMigrate(&User{}, &Post{})

// Load fixtures manually or with a library
users := []User{
    {ID: "1", Name: "Test User", Email: "test@example.com"},
}
db.Create(&users)

// Or use a fixture library like testfixtures
import "github.com/go-testfixtures/testfixtures/v3"
```
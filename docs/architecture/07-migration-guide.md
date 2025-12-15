# Migration Guide

This document provides guidance for migrating from Pivot to industry-standard Go database libraries.

## Standard Library Equivalents

| Pivot Backend | Recommended Replacement | Notes |
|---------------|------------------------|-------|
| SQL (MySQL/PostgreSQL/SQLite) | [GORM](https://gorm.io) or [sqlx](https://github.com/jmoiron/sqlx) | GORM for ORM, sqlx for lightweight SQL |
| MongoDB | [mongo-go-driver](https://github.com/mongodb/mongo-go-driver) | Official MongoDB driver |
| Redis | [go-redis](https://github.com/redis/go-redis) | Modern Redis client |
| DynamoDB | [aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2) | AWS SDK v2 (not v1) |
| Elasticsearch | [go-elasticsearch](https://github.com/elastic/go-elasticsearch) | Official Elastic client |
| Filesystem | Standard `os`/`io` + `encoding/json` | No library needed |

## Interface Method Mapping

### Backend Interface → Standard Patterns

| Pivot Method | GORM Equivalent | sqlx Equivalent |
|--------------|-----------------|-----------------|
| `Initialize()` | `gorm.Open(dialect, dsn)` | `sqlx.Connect(driver, dsn)` |
| `Ping(duration)` | `db.Raw("SELECT 1")` | `db.Ping()` |
| `CreateCollection(schema)` | `db.AutoMigrate(&Model{})` | Manual `CREATE TABLE` |
| `DeleteCollection(name)` | `db.Migrator().DropTable(&Model{})` | `db.Exec("DROP TABLE")` |
| `ListCollections()` | `db.Migrator().GetTables()` | Query `information_schema` |
| `Insert(collection, recordset)` | `db.Create(&records)` | `db.NamedExec(sql, records)` |
| `Retrieve(collection, id)` | `db.First(&record, id)` | `db.Get(&record, sql, id)` |
| `Update(collection, recordset)` | `db.Save(&record)` | `db.NamedExec(sql, record)` |
| `Delete(collection, ids)` | `db.Delete(&Model{}, ids)` | `db.Exec(sql, ids)` |
| `Exists(collection, id)` | `db.First(&r, id).Error == nil` | `db.Get(&r, sql, id)` |

### Indexer Interface → Standard Patterns

| Pivot Method | GORM Equivalent | sqlx Equivalent |
|--------------|-----------------|-----------------|
| `Query(collection, filter)` | `db.Where(...).Find(&results)` | `db.Select(&results, sql)` |
| `QueryFunc(collection, filter, fn)` | `db.Where(...).Rows()` + iterate | `db.Queryx(sql)` + iterate |
| `DeleteQuery(collection, filter)` | `db.Where(...).Delete(&Model{})` | `db.Exec(sql)` |
| `ListValues(collection, fields)` | `db.Distinct(fields).Find(...)` | `db.Select(&vals, sql)` |

### Mapper Interface → GORM

| Pivot Method | GORM Equivalent |
|--------------|-----------------|
| `mapper.NewModel(backend, schema)` | Define struct, call `db.AutoMigrate()` |
| `Create(&record)` | `db.Create(&record)` |
| `Get(id, &dest)` | `db.First(&dest, id)` |
| `Update(&record)` | `db.Save(&record)` |
| `Delete(id)` | `db.Delete(&Model{}, id)` |
| `All(filter, &results)` | `db.Where(...).Find(&results)` |
| `Migrate()` | `db.AutoMigrate(&Model{})` |
| `Count(filter)` | `db.Where(...).Count(&count)` |
| `Sum(field, filter)` | `db.Where(...).Select("SUM(field)")` |

## Struct Tag Migration

### Pivot → GORM

```go
// Pivot
type User struct {
    ID        string    `pivot:"id,identity"`
    Email     string    `pivot:"email"`
    Name      string    `pivot:"name"`
    Age       int       `pivot:"age,omitempty"`
    CreatedAt time.Time `pivot:"created_at"`
}

// GORM
type User struct {
    ID        string    `gorm:"primaryKey;column:id"`
    Email     string    `gorm:"column:email;uniqueIndex"`
    Name      string    `gorm:"column:name"`
    Age       int       `gorm:"column:age"`
    CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}
```

### Pivot → sqlx

```go
// Pivot
type User struct {
    ID    string `pivot:"id,identity"`
    Email string `pivot:"email"`
    Name  string `pivot:"name"`
}

// sqlx
type User struct {
    ID    string `db:"id"`
    Email string `db:"email"`
    Name  string `db:"name"`
}
```

### Tag Option Mapping

| Pivot Tag | GORM Tag | sqlx Tag |
|-----------|----------|----------|
| `pivot:"field_name"` | `gorm:"column:field_name"` | `db:"field_name"` |
| `pivot:"id,identity"` | `gorm:"primaryKey"` | `db:"id"` (manual PK handling) |
| `pivot:"field,omitempty"` | `gorm:"default:null"` | N/A (handle in code) |
| `pivot:"-"` | `gorm:"-"` | `db:"-"` |

## Filter Syntax Translation

### Pivot Filter → SQL WHERE

| Pivot Filter | SQL Equivalent | GORM |
|--------------|----------------|------|
| `name/john` | `WHERE name = 'john'` | `Where("name = ?", "john")` |
| `name/is:john` | `WHERE name = 'john'` | `Where("name = ?", "john")` |
| `name/not:john` | `WHERE name != 'john'` | `Where("name != ?", "john")` |
| `name/john\|jane` | `WHERE name IN ('john','jane')` | `Where("name IN ?", []string{"john","jane"})` |
| `age/gt:18` | `WHERE age > 18` | `Where("age > ?", 18)` |
| `age/gte:18` | `WHERE age >= 18` | `Where("age >= ?", 18)` |
| `age/lt:65` | `WHERE age < 65` | `Where("age < ?", 65)` |
| `age/lte:65` | `WHERE age <= 65` | `Where("age <= ?", 65)` |
| `age/range:18\|65` | `WHERE age >= 18 AND age < 65` | `Where("age >= ? AND age < ?", 18, 65)` |
| `name/contains:john` | `WHERE name LIKE '%john%'` | `Where("name LIKE ?", "%john%")` |
| `name/prefix:john` | `WHERE name LIKE 'john%'` | `Where("name LIKE ?", "john%")` |
| `name/suffix:son` | `WHERE name LIKE '%son'` | `Where("name LIKE ?", "%son")` |
| `name/like:john` | `WHERE LOWER(name) LIKE '%john%'` | `Where("LOWER(name) LIKE ?", "%john%")` |

### Multi-Field Filters

```go
// Pivot
filter.Parse("name/contains:john/age/gt:18")

// GORM
db.Where("name LIKE ?", "%john%").Where("age > ?", 18)

// sqlx
db.Select(&users, "SELECT * FROM users WHERE name LIKE $1 AND age > $2", "%john%", 18)
```

### Filter with Pagination

```go
// Pivot
f := filter.Parse("status/active")
f.Limit = 20
f.Offset = 40
f.Sort = []SortBy{{Field: "created_at", Descending: true}}

// GORM
db.Where("status = ?", "active").
    Order("created_at DESC").
    Limit(20).
    Offset(40).
    Find(&results)

// sqlx
db.Select(&results, `
    SELECT * FROM users
    WHERE status = $1
    ORDER BY created_at DESC
    LIMIT 20 OFFSET 40`, "active")
```

## Connection String Migration

### Pivot → Standard DSN

| Pivot | Standard |
|-------|----------|
| `mysql://user:pass@localhost:3306/db` | `user:pass@tcp(localhost:3306)/db?parseTime=true` |
| `postgres://user:pass@localhost/db` | `postgres://user:pass@localhost/db?sslmode=disable` |
| `sqlite:///./data.db` | `./data.db` (or `file:./data.db?cache=shared`) |
| `mongodb://localhost/db` | `mongodb://localhost:27017/db` |
| `redis://localhost:6379/0` | `localhost:6379` + `DB: 0` option |

### Example Connection Setup

```go
// Pivot
backend, _ := pivot.NewDatabase("mysql://user:pass@localhost:3306/mydb")
backend.Initialize()

// GORM
dsn := "user:pass@tcp(localhost:3306)/mydb?parseTime=true"
db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})

// sqlx
dsn := "user:pass@tcp(localhost:3306)/mydb?parseTime=true"
db, _ := sqlx.Connect("mysql", dsn)
```

## Schema Definition Migration

### Pivot Schema → GORM Model

```go
// Pivot schema (JSON)
{
    "name": "widgets",
    "identityFieldType": "str",
    "fields": [
        {"name": "type", "type": "str", "required": true},
        {"name": "price", "type": "float"},
        {"name": "created_at", "type": "time"}
    ]
}

// GORM model
type Widget struct {
    ID        string    `gorm:"primaryKey;type:varchar(36)"`
    Type      string    `gorm:"not null"`
    Price     float64
    CreatedAt time.Time `gorm:"autoCreateTime"`
}

// Migration
db.AutoMigrate(&Widget{})
```

### Field Type Mapping

| Pivot Type | Go Type | GORM Type | SQL Type |
|------------|---------|-----------|----------|
| `str` | `string` | `string` | `VARCHAR(255)` |
| `int` | `int64` | `int64` | `BIGINT` |
| `float` | `float64` | `float64` | `DOUBLE` |
| `bool` | `bool` | `bool` | `BOOLEAN` |
| `time` | `time.Time` | `time.Time` | `DATETIME` |
| `object` | `map[string]interface{}` | `datatypes.JSON` | `JSON` |
| `array` | `[]interface{}` | `datatypes.JSON` | `JSON` |

## Common Migration Patterns

### Pattern 1: Basic CRUD Service

```go
// BEFORE: Pivot
type UserService struct {
    users mapper.Mapper
}

func NewUserService(backend backends.Backend) *UserService {
    return &UserService{
        users: mapper.NewModel(backend, UsersSchema),
    }
}

func (s *UserService) Create(u *User) error {
    return s.users.Create(u)
}

func (s *UserService) Get(id string) (*User, error) {
    var u User
    err := s.users.Get(id, &u)
    return &u, err
}

// AFTER: GORM
type UserService struct {
    db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
    return &UserService{db: db}
}

func (s *UserService) Create(u *User) error {
    return s.db.Create(u).Error
}

func (s *UserService) Get(id string) (*User, error) {
    var u User
    err := s.db.First(&u, "id = ?", id).Error
    return &u, err
}
```

### Pattern 2: Query with Filter

```go
// BEFORE: Pivot
func (s *UserService) FindActive(minAge int) ([]User, error) {
    var users []User
    f := filter.Parse(fmt.Sprintf("status/active/age/gte:%d", minAge))
    err := s.users.All(f, &users)
    return users, err
}

// AFTER: GORM
func (s *UserService) FindActive(minAge int) ([]User, error) {
    var users []User
    err := s.db.Where("status = ?", "active").
        Where("age >= ?", minAge).
        Find(&users).Error
    return users, err
}
```

### Pattern 3: Paginated Query

```go
// BEFORE: Pivot
func (s *UserService) List(page, perPage int) (*dal.RecordSet, error) {
    f := filter.All()
    f.Limit = perPage
    f.Offset = (page - 1) * perPage
    f.Sort = []filter.SortBy{{Field: "created_at", Descending: true}}
    return s.users.Find(f)
}

// AFTER: GORM
type PaginatedResult struct {
    Users      []User
    Total      int64
    Page       int
    PerPage    int
    TotalPages int
}

func (s *UserService) List(page, perPage int) (*PaginatedResult, error) {
    var users []User
    var total int64

    s.db.Model(&User{}).Count(&total)

    err := s.db.Order("created_at DESC").
        Limit(perPage).
        Offset((page - 1) * perPage).
        Find(&users).Error

    return &PaginatedResult{
        Users:      users,
        Total:      total,
        Page:       page,
        PerPage:    perPage,
        TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
    }, err
}
```

### Pattern 4: Relationship Loading

```go
// BEFORE: Pivot (auto-expand via schema)
// Collection has EmbeddedCollections defined
// Records automatically include related data

// AFTER: GORM (explicit preload)
type Post struct {
    ID       string `gorm:"primaryKey"`
    Title    string
    AuthorID string
    Author   User `gorm:"foreignKey:AuthorID"`
}

func (s *PostService) GetWithAuthor(id string) (*Post, error) {
    var post Post
    err := s.db.Preload("Author").First(&post, "id = ?", id).Error
    return &post, err
}
```

### Pattern 5: Aggregations

```go
// BEFORE: Pivot
count, _ := s.users.Count(filter.Parse("status/active"))
sum, _ := s.users.Sum("balance", filter.Parse("status/active"))

// AFTER: GORM
var count int64
s.db.Model(&User{}).Where("status = ?", "active").Count(&count)

var sum float64
s.db.Model(&User{}).Where("status = ?", "active").
    Select("COALESCE(SUM(balance), 0)").Scan(&sum)
```

## Migration Checklist

### Phase 1: Inventory
- [ ] List all services using Pivot
- [ ] Document which Pivot features each service uses
- [ ] Identify custom validators/formatters in use
- [ ] Map connection strings for all environments

### Phase 2: Preparation
- [ ] Choose replacement libraries (GORM vs sqlx per service)
- [ ] Create new model structs with appropriate tags
- [ ] Write adapter layer if gradual migration needed
- [ ] Set up new connection configuration

### Phase 3: Implementation (per service)
- [ ] Add new database library dependency
- [ ] Create new repository/DAO layer
- [ ] Migrate CRUD operations
- [ ] Migrate query/filter operations
- [ ] Migrate aggregation operations
- [ ] Update relationship handling
- [ ] Remove Pivot imports and dependencies

### Phase 4: Validation
- [ ] Run existing tests against new implementation
- [ ] Verify query performance
- [ ] Check connection pooling behavior
- [ ] Validate transaction handling
- [ ] Test error handling paths

### Phase 5: Cleanup
- [ ] Remove Pivot dependency from go.mod
- [ ] Delete unused schema JSON files
- [ ] Update documentation
- [ ] Archive or deprecate Pivot-specific code

## Potential Challenges

### 1. Filter Syntax Usage
If filter strings are constructed dynamically or passed from external sources (APIs, configs), you'll need to build a translation layer or replace with query builders.

### 2. Schema-Driven Validation
Pivot validates against schema definitions. With GORM/sqlx, validation moves to:
- Struct tags (`binding:"required"` with gin)
- Validation libraries (go-playground/validator)
- Database constraints

### 3. Auto-Generated IDs
Pivot's `IdentityFieldFormatter` (like `GenerateUUID`) needs replacement:
```go
// GORM hook
func (u *User) BeforeCreate(tx *gorm.DB) error {
    u.ID = uuid.New().String()
    return nil
}
```

### 4. Embedded Collections
Pivot auto-embeds related records. GORM requires explicit `Preload()` calls or `Joins()`.

### 5. Multi-Backend Support
If a service uses multiple backends (e.g., SQL + Elasticsearch), you'll need separate clients and potentially a facade layer.

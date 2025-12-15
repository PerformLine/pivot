# Data Model

This document describes the core data structures in Pivot's Data Abstraction Layer (DAL).

## Core DAL Structures

```mermaid
erDiagram
    Collection ||--o{ Field : contains
    Collection ||--o{ Constraint : has
    Collection ||--o{ Relationship : embeds
    RecordSet ||--o{ Record : contains
    Record ||--o{ FieldValue : has
    Field ||--o| Validator : uses
    Field ||--o| Formatter : uses

    Collection {
        string Name
        string IdentityField
        Type IdentityFieldType
        func IdentityFieldFormatter
        string IndexName
        string[] IndexCompoundFields
        int64 TotalRecords
        bool View
    }

    Field {
        string Name
        string Description
        Type Type
        Type KeyType
        Type Subtype
        int Length
        int Precision
        bool Required
        bool Unique
        string UniqueGroup
        interface DefaultValue
        bool Identity
        bool Key
    }

    Constraint {
        interface On
        string Collection
        interface Field
        string Options
        string Into
        bool NoEmbed
    }

    Relationship {
        interface Keys
        ptr Collection
        string CollectionName
        string[] Fields
        bool Force
    }

    Record {
        interface ID
        map Fields
        bytes Data
        error Error
        string CollectionName
        string Operation
        bool Optional
    }

    RecordSet {
        ptr[] Records
        int64 ResultCount
        int Page
        int TotalPages
        int RecordsPerPage
        bool KnownSize
    }
```

## Collection Schema

A `Collection` represents a database table, MongoDB collection, or data container.

### Key Fields

| Field | Type | Description |
|-------|------|-------------|
| `Name` | string | Collection/table name |
| `IdentityField` | string | Primary key field name (default: "id") |
| `IdentityFieldType` | dal.Type | Type of primary key |
| `IdentityFieldFormatter` | func | Auto-generation function (e.g., `dal.GenerateUUID`) |
| `Fields` | []Field | Field definitions |
| `IndexCompoundFields` | []string | Fields for composite indexing |
| `IndexCompoundFieldJoiner` | string | Separator for compound keys (default: ":") |
| `EmbeddedCollections` | []Relationship | Related collections to auto-embed |
| `Constraints` | []Constraint | Foreign key constraints |

### Example Schema Definition

```go
var UsersSchema = &dal.Collection{
    Name:                   "users",
    IdentityFieldType:      dal.StringType,
    IdentityFieldFormatter: dal.GenerateUUID,
    Fields: []dal.Field{
        {
            Name:        "email",
            Type:        dal.StringType,
            Required:    true,
            Unique:      true,
        },
        {
            Name:        "name",
            Type:        dal.StringType,
            Required:    true,
        },
        {
            Name:        "age",
            Type:        dal.IntType,
            Validator:   dal.ValidatePositiveInteger,
        },
        {
            Name:        "created_at",
            Type:        dal.TimeType,
            Formatter:   dal.CurrentTimeIfUnset,
        },
    },
}
```

### JSON Schema Definition

```json
{
  "name": "users",
  "identityFieldType": "str",
  "fields": [
    {
      "name": "email",
      "type": "str",
      "required": true,
      "unique": true
    },
    {
      "name": "name",
      "type": "str",
      "required": true
    },
    {
      "name": "age",
      "type": "int"
    }
  ]
}
```

## Field Types

| Type Constant | String | Go Equivalent | Description |
|--------------|--------|---------------|-------------|
| `StringType` | `str` | `string` | Text data |
| `BooleanType` | `bool` | `bool` | True/false values |
| `IntType` | `int` | `int64` | Integer numbers |
| `FloatType` | `float` | `float64` | Decimal numbers |
| `TimeType` | `time` | `time.Time` | Date/time values |
| `ObjectType` | `object` | `map[string]interface{}` | Nested objects |
| `ArrayType` | `array` | `[]interface{}` | List of values |
| `RawType` | `raw` | `[]byte` | Binary data |
| `AutoType` | `auto` | varies | Auto-detected type |

## Field Definition

### Key Properties

| Property | Type | Description |
|----------|------|-------------|
| `Name` | string | Field name |
| `Description` | string | Documentation |
| `Type` | dal.Type | Data type |
| `KeyType` | dal.Type | Key type for maps |
| `Subtype` | dal.Type | Element type for arrays |
| `Length` | int | Max length (strings) |
| `Precision` | int | Decimal precision |
| `Required` | bool | Cannot be null/empty |
| `Unique` | bool | Unique constraint |
| `UniqueGroup` | string | Multi-field unique constraint |
| `DefaultValue` | interface{} | Default if not provided |
| `Identity` | bool | Is primary key |
| `Key` | bool | Part of composite key |
| `Validator` | func | Validation function |
| `Formatter` | func | Transform function |

### Built-in Validators

```go
dal.ValidateIsOneOf("a", "b", "c")      // Value must be one of listed values
dal.ValidateNotEmpty                     // String cannot be empty
dal.ValidatePositiveInteger              // Integer must be > 0
dal.ValidatePositiveOrZeroInteger        // Integer must be >= 0
```

### Built-in Formatters

```go
dal.GenerateUUID           // Generate UUID v4
dal.CurrentTime            // Set to current time (always)
dal.CurrentTimeIfUnset     // Set to current time (if empty)
dal.TrimSpace              // Trim whitespace
dal.FormatLowerCase        // Convert to lowercase
dal.FormatUpperCase        // Convert to uppercase
```

## Record Structure

A `Record` represents a single row/document.

```go
type Record struct {
    ID             interface{}            // Primary key value
    Fields         map[string]interface{} // Field data
    Data           []byte                 // Raw binary data
    Error          error                  // Associated error
    CollectionName string                 // Target collection
    Operation      string                 // "create", "update", "delete"
    Optional       bool                   // For fixtures: non-fatal on error
}
```

### Record Operations

```go
// Create record
record := dal.NewRecord(id)
record.Set("name", "John")
record.Set("email", "john@example.com")

// Access fields
name := record.Get("name")                    // Get field value
nested := record.Get("address.city")          // Dot notation for nested
keys := record.Keys()                         // Get composite key parts

// Composite keys
record.SetKeys("tenant123", "user456")        // Set composite key
id := record.ID                               // Returns "tenant123:user456"
```

## RecordSet Structure

A `RecordSet` contains query results with pagination metadata.

```go
type RecordSet struct {
    Records        []*Record              // Result records
    ResultCount    int64                  // Total matching records
    Page           int                    // Current page number
    TotalPages     int                    // Total pages available
    RecordsPerPage int                    // Records per page
    Options        map[string]interface{} // Additional metadata
    KnownSize      bool                   // Whether count is exact
}
```

## ConnectionString

Parses database connection URIs.

```go
type ConnectionString struct {
    URI     *url.URL
    Options map[string]interface{}
}
```

### Format

```
backend+protocol://[user:pass@]host[:port]/dataset[?opt1=val1&opt2=val2]
```

### Methods

```go
cs, _ := dal.ParseConnectionString("mysql://user:pass@localhost:3306/mydb?timeout=30s")

cs.Backend()                    // "mysql"
cs.Protocol()                   // "" (or "tcp", "unix", etc.)
cs.Host()                       // "localhost"
cs.Port()                       // 3306
cs.Dataset()                    // "mydb"
cs.Credentials()                // "user", "pass", true

cs.OptString("timeout")         // "30s"
cs.OptBool("ssl")               // false (default)
cs.OptDuration("timeout")       // 30 * time.Second

cs.LoadCredentialsFromNetrc("~/.netrc")  // Load from netrc file
```

### Scheme Aliases

| Alias | Canonical |
|-------|-----------|
| `psql` | `postgresql` |
| `postgres` | `postgresql` |

## Constraint (Foreign Key)

Defines relationships between collections.

```go
type Constraint struct {
    On         interface{} // Local field(s)
    Collection string      // Referenced collection
    Field      interface{} // Referenced field(s)
    Options    string      // Backend-specific (e.g., "ON DELETE CASCADE")
    Into       string      // Target field for embedded data
    NoEmbed    bool        // Skip when expanding relationships
}
```

### Example

```go
Constraints: []dal.Constraint{
    {
        On:         "author_id",
        Collection: "users",
        Field:      "id",
        Options:    "ON DELETE SET NULL",
        Into:       "author",
    },
}
```

## Relationship (Embedded Collections)

Defines collections to auto-embed in query results.

```go
type Relationship struct {
    Keys           interface{}  // Join key value(s)
    Collection     *Collection  // Referenced collection definition
    CollectionName string       // Collection name if not resolved
    Fields         []string     // Fields to include from related collection
    Force          bool         // Force expansion even if missing
}
```

## Schema Delta Detection

Used for migration comparison.

```go
type SchemaDelta struct {
    Type           DeltaType      // CollectionDelta, FieldDelta
    Issue          DeltaIssue     // What's different
    Message        string         // Human-readable description
    Collection     string         // Affected collection
    Name           string         // Field name
    Parameter      string         // Specific parameter
    Desired        interface{}    // Expected value
    Actual         interface{}    // Current value
    ReferenceField *Field         // Original field definition
}
```

### Delta Issues

| Issue | Description |
|-------|-------------|
| `CollectionNameIssue` | Collection name mismatch |
| `CollectionKeyNameIssue` | Primary key name differs |
| `CollectionKeyTypeIssue` | Primary key type differs |
| `FieldMissingIssue` | Field exists in schema but not database |
| `FieldNameIssue` | Field name mismatch |
| `FieldTypeIssue` | Field type mismatch |
| `FieldLengthIssue` | Field length differs |
| `FieldPropertyIssue` | Other property mismatch |

---
name: pivot-migration
description: Assists with migrating from Pivot database abstraction library to standard Go libraries (GORM, sqlx, mongo-go-driver, go-redis, opensearch-go). Use when analyzing Pivot usage in a service, planning migration strategy, converting code patterns, or translating filter syntax.
allowed-tools: Read, Grep, Glob, Task
---

# Pivot Migration Assistant

This skill helps migrate services from Pivot to industry-standard Go database libraries.

## Quick Reference

### Replacement Libraries

| Pivot Backend | Replacement |
|---------------|-------------|
| SQL (MySQL/PostgreSQL/SQLite) | GORM or sqlx |
| MongoDB | mongo-go-driver |
| Redis | go-redis |
| DynamoDB | aws-sdk-go-v2 |
| Elasticsearch | opensearch-go v4 |

### Common Migration Tasks

1. **Analyze Pivot usage** - Find all Pivot imports and usage patterns
2. **Convert struct tags** - `pivot:"field"` → `gorm:"column:field"` or `db:"field"`
3. **Translate filters** - Pivot filter syntax → SQL WHERE / GORM Where()
4. **Migrate CRUD** - Mapper methods → GORM/sqlx equivalents
5. **Update connections** - Pivot connection strings → standard DSN format

### Key Files to Check

When analyzing a service for Pivot usage:
```
go.mod                    # Pivot dependency version
**/*repository*.go        # Repository/DAO layers
**/*service*.go           # Service layers using Pivot
**/*model*.go             # Struct definitions with pivot tags
**/schema/*.json          # Pivot schema definitions
**/fixtures/*.json        # Test fixtures
```

### Struct Tag Conversion

```go
// Pivot
`pivot:"id,identity"`     → `gorm:"primaryKey"`
`pivot:"field_name"`      → `gorm:"column:field_name"` or `db:"field_name"`
`pivot:"field,omitempty"` → `gorm:"default:null"`
`pivot:"-"`               → `gorm:"-"` or `db:"-"`
```

### Filter Translation Quick Reference

| Pivot | SQL | GORM |
|-------|-----|------|
| `field/value` | `field = 'value'` | `Where("field = ?", value)` |
| `field/gt:N` | `field > N` | `Where("field > ?", N)` |
| `field/gte:N` | `field >= N` | `Where("field >= ?", N)` |
| `field/lt:N` | `field < N` | `Where("field < ?", N)` |
| `field/lte:N` | `field <= N` | `Where("field <= ?", N)` |
| `field/contains:x` | `field LIKE '%x%'` | `Where("field LIKE ?", "%x%")` |
| `field/prefix:x` | `field LIKE 'x%'` | `Where("field LIKE ?", "x%")` |
| `field/a\|b` | `field IN ('a','b')` | `Where("field IN ?", []string{"a","b"})` |

### Method Mapping

| Pivot Mapper | GORM |
|--------------|------|
| `Create(&r)` | `db.Create(&r)` |
| `Get(id, &r)` | `db.First(&r, id)` |
| `Update(&r)` | `db.Save(&r)` |
| `Delete(id)` | `db.Delete(&Model{}, id)` |
| `All(filter, &results)` | `db.Where(...).Find(&results)` |

## Usage Examples

### Analyze a service
```
"Analyze Pivot usage in the user-service repository and create a migration plan"
```

### Convert a file
```
"Convert this repository file from Pivot to GORM, preserving the same functionality"
```

### Translate a filter
```
"Translate this Pivot filter to GORM: status/active/created_at/gt:2024-01-01"
```

## Related Documentation

- [Full Migration Guide](../../../docs/architecture/07-migration-guide.md)
- [Pivot Patterns](../../../docs/architecture/05-patterns.md)
- [Data Model Reference](../../../docs/architecture/03-data-model.md)
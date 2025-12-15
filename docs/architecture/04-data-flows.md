# Data Flows

This document describes the key data flows in Pivot using sequence diagrams.

## Query Execution Flow

Shows how a query flows from HTTP request to database and back.

```mermaid
sequenceDiagram
    participant Client
    participant Server
    participant Filter
    participant Generator
    participant Indexer
    participant Backend
    participant Database

    Client->>Server: GET /api/collections/users/where/name/contains:john
    Server->>Filter: Parse("name/contains:john")
    Filter->>Filter: Create Criterion[]

    Server->>Indexer: Query(collection, filter)
    Indexer->>Generator: Select generator (SQL/Mongo/ES)
    Indexer->>Generator: Render(filter)
    Generator->>Generator: Build native query
    Generator-->>Indexer: WHERE name LIKE '%john%'

    Indexer->>Database: Execute query
    Database-->>Indexer: Raw results
    Indexer->>Indexer: Build RecordSet

    alt Has relationships & autoexpand
        Indexer->>Backend: PopulateRelationships()
        Backend->>Database: Fetch related records
        Database-->>Backend: Related data
        Backend-->>Indexer: Embedded records
    end

    Indexer-->>Server: RecordSet
    Server->>Server: JSON serialize
    Server-->>Client: HTTP 200 + JSON
```

## CRUD Operations

### Create Record

```mermaid
sequenceDiagram
    participant App
    participant Mapper
    participant Backend
    participant Validator
    participant Formatter
    participant Database

    App->>Mapper: Create(&record)
    Mapper->>Mapper: Reflect struct to Record

    loop For each field
        Mapper->>Validator: Validate(value)
        alt Invalid
            Validator-->>Mapper: Error
            Mapper-->>App: Validation error
        end
        Mapper->>Formatter: Format(value)
        Formatter-->>Mapper: Formatted value
    end

    alt Has identity formatter
        Mapper->>Formatter: GenerateUUID()
        Formatter-->>Mapper: New ID
    end

    Mapper->>Backend: Insert(collection, recordset)
    Backend->>Database: INSERT INTO...
    Database-->>Backend: Success
    Backend-->>Mapper: nil

    Mapper->>Mapper: Populate ID back to struct
    Mapper-->>App: nil (success)
```

### Read Record

```mermaid
sequenceDiagram
    participant App
    participant Mapper
    participant Backend
    participant Database

    App->>Mapper: Get(id, &dest)
    Mapper->>Backend: Retrieve(collection, id)
    Backend->>Database: SELECT * WHERE id = ?

    alt Record exists
        Database-->>Backend: Row data
        Backend->>Backend: Build Record
        Backend-->>Mapper: *Record
        Mapper->>Mapper: Map to struct
        Mapper-->>App: nil (success)
    else Not found
        Database-->>Backend: No rows
        Backend-->>Mapper: nil Record
        Mapper-->>App: NotFoundError
    end
```

### Update Record

```mermaid
sequenceDiagram
    participant App
    participant Mapper
    participant Backend
    participant Database

    App->>Mapper: Update(&record)
    Mapper->>Mapper: Extract ID from struct
    Mapper->>Mapper: Reflect struct to Record

    loop For each field
        Mapper->>Mapper: Validate & Format
    end

    Mapper->>Backend: Update(collection, recordset)
    Backend->>Database: UPDATE ... WHERE id = ?
    Database-->>Backend: Rows affected

    alt Rows affected > 0
        Backend-->>Mapper: nil
        Mapper-->>App: nil (success)
    else No rows affected
        Backend-->>Mapper: nil
        Mapper-->>App: nil (no error, but nothing changed)
    end
```

### Delete Record

```mermaid
sequenceDiagram
    participant App
    participant Mapper
    participant Backend
    participant Database

    App->>Mapper: Delete(id)
    Mapper->>Backend: Delete(collection, id)
    Backend->>Database: DELETE FROM ... WHERE id = ?
    Database-->>Backend: Rows affected
    Backend-->>Mapper: nil
    Mapper-->>App: nil (success)
```

## Filter Compilation Flow

Shows how URL filter syntax compiles to backend-specific queries.

```mermaid
sequenceDiagram
    participant Input
    participant Filter
    participant Criterion
    participant Generator
    participant Output

    Input->>Filter: "type:name/contains:john/int:age/gt:18"
    Filter->>Filter: Split by "/"

    loop For each pair
        Filter->>Criterion: Parse field/value
        Criterion->>Criterion: Extract type prefix
        Criterion->>Criterion: Extract operator
        Criterion->>Criterion: Parse value(s)
        Criterion-->>Filter: Criterion object
    end

    Filter->>Generator: Select by backend type

    alt SQL Backend
        Generator->>Generator: SQL Generator
        loop For each criterion
            Generator->>Generator: Build WHERE clause
            Generator->>Generator: Add placeholder
            Generator->>Generator: Collect value
        end
        Generator-->>Output: "WHERE name LIKE ? AND age > ?" + ["%john%", 18]
    else MongoDB Backend
        Generator->>Generator: MongoDB Generator
        loop For each criterion
            Generator->>Generator: Build query object
        end
        Generator-->>Output: {"$and": [{"name": {"$regex": "john"}}, {"age": {"$gt": 18}}]}
    else Elasticsearch
        Generator->>Generator: ES Generator
        loop For each criterion
            Generator->>Generator: Build query DSL
        end
        Generator-->>Output: {"bool": {"must": [...]}}
    end
```

## Relationship Resolution Flow

Shows how embedded collections are resolved.

```mermaid
sequenceDiagram
    participant Query
    participant Backend
    participant Relationships
    participant RelatedBackend
    participant Database

    Query->>Backend: Query with autoexpand=true
    Backend->>Database: Execute main query
    Database-->>Backend: Primary records

    Backend->>Relationships: PopulateRelationships(records, collection)

    loop For each embedded collection
        Relationships->>Relationships: Extract foreign key values
        Relationships->>RelatedBackend: Retrieve related records
        RelatedBackend->>Database: SELECT WHERE id IN (...)
        Database-->>RelatedBackend: Related records
        RelatedBackend-->>Relationships: Related RecordSet

        loop For each primary record
            Relationships->>Relationships: Match by key
            Relationships->>Relationships: Embed in record.Fields
        end
    end

    Relationships-->>Backend: Records with embedded data
    Backend-->>Query: Complete RecordSet
```

## REST API Request Flow

```mermaid
sequenceDiagram
    participant Client
    participant Vestigo
    participant Handler
    participant Pivot
    participant Backend
    participant Response

    Client->>Vestigo: HTTP Request
    Vestigo->>Vestigo: Route matching
    Vestigo->>Handler: Call endpoint handler

    Handler->>Handler: Parse query params
    Handler->>Handler: Extract path params

    alt Query endpoint
        Handler->>Pivot: Parse filter string
        Pivot-->>Handler: Filter object
        Handler->>Backend: WithSearch().Query()
    else Record endpoint
        Handler->>Backend: Retrieve/Insert/Update/Delete
    end

    Backend-->>Handler: Result / Error

    alt Success
        Handler->>Response: JSON encode result
        Response-->>Client: HTTP 200 + JSON body
    else Error
        Handler->>Response: Error response
        Response-->>Client: HTTP 4xx/5xx + error message
    end
```

## Multi-Indexer Strategy Flow

Shows how MultiIndex selects and executes across multiple indexers.

```mermaid
sequenceDiagram
    participant Client
    participant MultiIndex
    participant Strategy
    participant Indexer1
    participant Indexer2
    participant Indexer3

    Client->>MultiIndex: Query(filter)
    MultiIndex->>Strategy: Select strategy (Sequential/All/First/Random)

    alt Sequential Strategy
        Strategy->>Indexer1: Query(filter)
        alt Success
            Indexer1-->>Strategy: Results
            Strategy-->>Client: Results
        else Failure
            Indexer1-->>Strategy: Error
            Strategy->>Indexer2: Query(filter)
            alt Success
                Indexer2-->>Strategy: Results
                Strategy-->>Client: Results
            else Failure
                Strategy->>Indexer3: Query(filter)
                Indexer3-->>Strategy: Results/Error
                Strategy-->>Client: Results/Error
            end
        end
    else All Strategy
        par Execute on all
            Strategy->>Indexer1: Query(filter)
            Strategy->>Indexer2: Query(filter)
            Strategy->>Indexer3: Query(filter)
        end
        Indexer1-->>Strategy: Results1
        Indexer2-->>Strategy: Results2
        Indexer3-->>Strategy: Results3
        Strategy->>Strategy: Merge results
        Strategy-->>Client: Merged results
    end
```

## Schema Migration Flow

```mermaid
sequenceDiagram
    participant App
    participant Mapper
    participant Collection
    participant Backend
    participant Database

    App->>Mapper: Migrate()
    Mapper->>Backend: GetCollection(name)

    alt Collection exists
        Backend->>Database: Query schema
        Database-->>Backend: Current schema
        Backend-->>Mapper: *Collection (actual)

        Mapper->>Collection: Diff(desired, actual)
        Collection-->>Mapper: []SchemaDelta

        loop For each delta
            alt Field missing
                Mapper->>Backend: ALTER TABLE ADD COLUMN
            else Type mismatch
                Mapper->>Backend: ALTER TABLE MODIFY COLUMN
            end
        end
    else Collection doesn't exist
        Backend-->>Mapper: nil
        Mapper->>Backend: CreateCollection(schema)
        Backend->>Database: CREATE TABLE...
        Database-->>Backend: Success
    end

    Backend-->>Mapper: nil
    Mapper-->>App: nil (success)
```

## Composite Key Handling

```mermaid
sequenceDiagram
    participant Client
    participant Indexer
    participant Backend
    participant Database

    Note over Client,Database: Storing with composite key
    Client->>Backend: Insert record with keys ["tenant1", "user123"]
    Backend->>Backend: Join keys with joiner ":"
    Backend->>Database: INSERT ... id = "tenant1:user123"

    Note over Client,Database: Querying with composite key
    Client->>Indexer: Query with compound fields
    Indexer->>Database: SELECT ...
    Database-->>Indexer: Results with joined IDs
    Indexer->>Indexer: Split ID by joiner
    Indexer->>Indexer: Set record.Keys = ["tenant1", "user123"]
    Indexer-->>Client: Records with decomposed keys
```

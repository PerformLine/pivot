# C4 Architecture Diagrams

This document provides C4 model diagrams showing Pivot's architecture at different levels of abstraction.

## C4 Context Diagram

Shows Pivot in the broader ecosystem with external actors and systems.

```mermaid
flowchart TB
    subgraph "External Actors"
        DEV[Developer<br/>Application Code]
        ADMIN[Administrator<br/>CLI User]
        CLIENT[HTTP Client<br/>API Consumer]
    end

    PIVOT[Pivot Library<br/>Multi-Database Abstraction]

    subgraph "SQL Databases"
        MYSQL[(MySQL/<br/>MariaDB)]
        PG[(PostgreSQL)]
        SQLITE[(SQLite)]
    end

    subgraph "NoSQL Databases"
        MONGO[(MongoDB)]
        REDIS[(Redis)]
        DYNAMO[(DynamoDB)]
    end

    subgraph "Search Engines"
        ES[(Elasticsearch)]
        BLEVE[(Bleve Index)]
    end

    subgraph "File Storage"
        FS[(Filesystem<br/>JSON/YAML/CSV)]
    end

    DEV -->|"Import & Use"| PIVOT
    ADMIN -->|"CLI Commands"| PIVOT
    CLIENT -->|"REST API"| PIVOT

    PIVOT -->|"SQL Queries"| MYSQL
    PIVOT -->|"SQL Queries"| PG
    PIVOT -->|"SQL Queries"| SQLITE
    PIVOT -->|"BSON/Wire Protocol"| MONGO
    PIVOT -->|"Redis Protocol"| REDIS
    PIVOT -->|"AWS SDK"| DYNAMO
    PIVOT -->|"HTTP/REST"| ES
    PIVOT -->|"Embedded"| BLEVE
    PIVOT -->|"File I/O"| FS

    style PIVOT fill:#1168bd,color:#fff
    style DEV fill:#08427b,color:#fff
    style ADMIN fill:#08427b,color:#fff
    style CLIENT fill:#08427b,color:#fff
```

## C4 Container Diagram

Shows the runtime containers and their interactions.

```mermaid
flowchart TB
    subgraph "Pivot System"
        subgraph "CLI Container"
            CLI_BIN[pivot CLI Binary<br/>Go Executable]
        end

        subgraph "REST API Container"
            SERVER[HTTP Server<br/>vestigo + negroni]
            UI[Web UI<br/>diecast templates]
        end

        subgraph "Core Library"
            PIVOT_LIB[Pivot Library<br/>pivot.NewDatabase]
            MAPPER_LIB[Mapper Layer<br/>ORM Operations]
        end

        subgraph "Backend Adapters"
            SQL_BE[SQL Backend<br/>MySQL/PG/SQLite]
            MONGO_BE[MongoDB Backend<br/>mgo.v2 driver]
            REDIS_BE[Redis Backend<br/>redigo driver]
            DYNAMO_BE[DynamoDB Backend<br/>AWS SDK]
            FS_BE[Filesystem Backend<br/>JSON/YAML/CSV]
        end

        subgraph "Indexer Adapters"
            SQL_IDX[SQL Indexer]
            MONGO_IDX[MongoDB Indexer]
            ES_IDX[Elasticsearch Indexer<br/>HTTP Client]
            BLEVE_IDX[Bleve Indexer<br/>Embedded]
            MULTI_IDX[Multi-Indexer<br/>Strategy Selection]
        end

        subgraph "Filter System"
            FILTER[Filter Parser]
            SQL_GEN[SQL Generator]
            MONGO_GEN[MongoDB Generator]
            ES_GEN[ES Generator]
        end
    end

    APP[Application Code] -->|"Import"| PIVOT_LIB
    CLI_BIN --> PIVOT_LIB
    SERVER --> PIVOT_LIB
    SERVER --> UI

    PIVOT_LIB --> MAPPER_LIB
    MAPPER_LIB --> SQL_BE
    MAPPER_LIB --> MONGO_BE
    MAPPER_LIB --> REDIS_BE
    MAPPER_LIB --> DYNAMO_BE
    MAPPER_LIB --> FS_BE

    SQL_BE --> SQL_IDX
    MONGO_BE --> MONGO_IDX
    SQL_IDX --> FILTER
    MONGO_IDX --> FILTER
    ES_IDX --> FILTER
    BLEVE_IDX --> FILTER
    MULTI_IDX --> SQL_IDX
    MULTI_IDX --> ES_IDX

    FILTER --> SQL_GEN
    FILTER --> MONGO_GEN
    FILTER --> ES_GEN

    style PIVOT_LIB fill:#1168bd,color:#fff
    style MAPPER_LIB fill:#1168bd,color:#fff
    style SERVER fill:#438dd5,color:#fff
    style CLI_BIN fill:#438dd5,color:#fff
```

## C4 Component Diagram

Shows the internal structure of the core library.

```mermaid
flowchart TB
    subgraph "pivot package"
        NEW_DB[NewDatabase<br/>Factory Function]
        LOAD_SCHEMA[LoadSchemata<br/>Schema Loader]
        LOAD_FIX[LoadFixtures<br/>Fixture Loader]
        CONFIG[Configuration<br/>Env Support]
    end

    subgraph "dal package"
        COLLECTION[Collection<br/>Schema Definition]
        RECORD[Record<br/>Data Instance]
        RECORDSET[RecordSet<br/>Result Set]
        FIELD[Field<br/>Column Definition]
        CONN_STR[ConnectionString<br/>URI Parser]
        CONSTRAINT[Constraint<br/>Foreign Keys]
        RELATION[Relationship<br/>Embedded Refs]
    end

    subgraph "backends package"
        BE_IFACE[Backend Interface<br/>CRUD Operations]
        IDX_IFACE[Indexer Interface<br/>Query Operations]
        AGG_IFACE[Aggregator Interface<br/>Aggregations]

        SQL_IMPL[SqlBackend<br/>MySQL/PG/SQLite]
        MONGO_IMPL[MongoBackend<br/>MongoDB]
        REDIS_IMPL[RedisBackend<br/>Redis]
        DYNAMO_IMPL[DynamoBackend<br/>DynamoDB]
        FS_IMPL[FilesystemBackend<br/>File Storage]

        SQL_IDX_IMPL[SqlIndexer]
        MONGO_IDX_IMPL[MongoIndexer]
        ES_IDX_IMPL[ElasticsearchIndexer]
        BLEVE_IDX_IMPL[BleveIndexer]
        MULTI_IDX_IMPL[MultiIndex]

        EMBED_BE[EmbeddedRecordBackend<br/>Relationship Wrapper]
        MAPPER_IMPL[Mapper Interface<br/>ORM Operations]
    end

    subgraph "filter package"
        FILTER_STRUCT[Filter<br/>Query Representation]
        CRITERION[Criterion<br/>Single Condition]
        GENERATOR[IGenerator Interface]
    end

    subgraph "filter/generators package"
        SQL_GEN_IMPL[SQL Generator<br/>WHERE Clauses]
        MONGO_GEN_IMPL[MongoDB Generator<br/>Query DSL]
        ES_GEN_IMPL[ES Generator<br/>Query DSL]
    end

    subgraph "mapper package"
        MODEL[Model<br/>ORM Implementation]
    end

    NEW_DB --> CONN_STR
    NEW_DB --> BE_IFACE
    LOAD_SCHEMA --> COLLECTION
    LOAD_FIX --> RECORD

    COLLECTION --> FIELD
    COLLECTION --> CONSTRAINT
    COLLECTION --> RELATION
    RECORDSET --> RECORD

    BE_IFACE --> SQL_IMPL
    BE_IFACE --> MONGO_IMPL
    BE_IFACE --> REDIS_IMPL
    BE_IFACE --> DYNAMO_IMPL
    BE_IFACE --> FS_IMPL

    IDX_IFACE --> SQL_IDX_IMPL
    IDX_IFACE --> MONGO_IDX_IMPL
    IDX_IFACE --> ES_IDX_IMPL
    IDX_IFACE --> BLEVE_IDX_IMPL
    IDX_IFACE --> MULTI_IDX_IMPL

    SQL_IDX_IMPL --> FILTER_STRUCT
    MONGO_IDX_IMPL --> FILTER_STRUCT
    ES_IDX_IMPL --> FILTER_STRUCT

    FILTER_STRUCT --> CRITERION
    FILTER_STRUCT --> GENERATOR
    GENERATOR --> SQL_GEN_IMPL
    GENERATOR --> MONGO_GEN_IMPL
    GENERATOR --> ES_GEN_IMPL

    MODEL --> MAPPER_IMPL
    MAPPER_IMPL --> BE_IFACE
    MAPPER_IMPL --> IDX_IFACE

    EMBED_BE --> BE_IFACE
    EMBED_BE --> RELATION

    style BE_IFACE fill:#9c27b0,color:#fff
    style IDX_IFACE fill:#9c27b0,color:#fff
    style AGG_IFACE fill:#9c27b0,color:#fff
    style GENERATOR fill:#9c27b0,color:#fff
    style MAPPER_IMPL fill:#9c27b0,color:#fff
```

## Backend Interface Hierarchy

```mermaid
classDiagram
    class Backend {
        <<interface>>
        +Initialize() error
        +SetIndexer(ConnectionString) error
        +Ping(Duration) error
        +RegisterCollection(Collection)
        +GetCollection(string) Collection
        +CreateCollection(Collection) error
        +DeleteCollection(string) error
        +ListCollections() []string
        +Exists(collection, id) bool
        +Retrieve(collection, id, fields) Record
        +Insert(collection, RecordSet) error
        +Update(collection, RecordSet) error
        +Delete(collection, ids) error
        +WithSearch(Collection, Filter) Indexer
        +WithAggregator(Collection) Aggregator
        +Supports(features) bool
    }

    class Indexer {
        <<interface>>
        +IndexConnectionString() ConnectionString
        +IndexInitialize(Backend) error
        +IndexExists(Collection, id) bool
        +IndexRetrieve(Collection, id) Record
        +Index(Collection, RecordSet) error
        +IndexRemove(Collection, ids) error
        +Query(Collection, Filter) RecordSet
        +QueryFunc(Collection, Filter, ResultFunc) error
        +ListValues(Collection, fields, Filter) map
        +DeleteQuery(Collection, Filter) error
        +FlushIndex() error
        +GetBackend() Backend
    }

    class Aggregator {
        <<interface>>
        +AggregatorConnectionString() ConnectionString
        +AggregatorInitialize(Backend) error
        +Sum(Collection, field, Filter) float64
        +Count(Collection, Filter) uint64
        +Minimum(Collection, field, Filter) float64
        +Maximum(Collection, field, Filter) float64
        +Average(Collection, field, Filter) float64
        +GroupBy(Collection, fields, aggregates, Filter) RecordSet
    }

    class SqlBackend {
        +db *sql.DB
        +queryGenTypeMapping SqlTypeMapping
    }

    class MongoBackend {
        +session *mgo.Session
        +db *mgo.Database
    }

    class RedisBackend {
        +pool *redis.Pool
        +keyPrefix string
    }

    class DynamoBackend {
        +db *dynamodb.DynamoDB
        +region string
    }

    class FilesystemBackend {
        +root string
        +format SerializationFormat
    }

    Backend <|.. SqlBackend
    Backend <|.. MongoBackend
    Backend <|.. RedisBackend
    Backend <|.. DynamoBackend
    Backend <|.. FilesystemBackend

    Indexer <|.. SqlBackend
    Indexer <|.. MongoBackend
    Aggregator <|.. SqlBackend
    Aggregator <|.. MongoBackend
```

## REST API Endpoint Structure

```mermaid
flowchart LR
    subgraph "Status & Metadata"
        S1[GET /api/status]
        S2[GET /api/collections]
        S3[GET /api/schema]
        S4[POST /api/schema]
        S5[GET /api/schema/:collection]
        S6[DELETE /api/schema/:collection]
    end

    subgraph "Query Operations"
        Q1[GET .../query/]
        Q2[POST .../query/]
        Q3[GET .../where/*filter]
        Q4[DELETE .../where/*filter]
        Q5[GET .../aggregate/:fields]
        Q6[GET .../list/*fields]
    end

    subgraph "Record CRUD"
        R1[POST .../records]
        R2[PUT .../records]
        R3[GET .../records/:id]
        R4[POST .../records/:id]
        R5[DELETE .../records/*id]
    end

    API[/api/collections/:collection] --> Q1
    API --> Q2
    API --> Q3
    API --> Q4
    API --> Q5
    API --> Q6
    API --> R1
    API --> R2
    API --> R3
    API --> R4
    API --> R5
```

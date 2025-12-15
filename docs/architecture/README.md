# Pivot Architecture Documentation

This directory contains comprehensive architecture documentation for the Pivot multi-database abstraction library.

## Quick Navigation

| Document | Description |
|----------|-------------|
| [01-overview.md](01-overview.md) | Service purpose, tech stack, architecture layers |
| [02-c4-diagrams.md](02-c4-diagrams.md) | C4 Context, Container, and Component diagrams |
| [03-data-model.md](03-data-model.md) | DAL structures, Collection/Record/Field schemas |
| [04-data-flows.md](04-data-flows.md) | Query execution, CRUD operations, relationship resolution |
| [05-patterns.md](05-patterns.md) | Backend/Indexer patterns, Generator pattern, Mapper ORM |
| [06-gotchas.md](06-gotchas.md) | Known oddities, v3 vs v4 differences, backend quirks |

## Suggested Reading Order

**For new developers:**
1. Start with [01-overview.md](01-overview.md) to understand what Pivot does and its tech stack
2. Review [02-c4-diagrams.md](02-c4-diagrams.md) for visual architecture understanding
3. Study [03-data-model.md](03-data-model.md) to understand core data structures
4. Read [05-patterns.md](05-patterns.md) before writing code

**For debugging/troubleshooting:**
1. Check [06-gotchas.md](06-gotchas.md) for common pitfalls
2. Review [04-data-flows.md](04-data-flows.md) to trace request paths

## Related Documentation

- [CLAUDE.md](../../CLAUDE.md) - Quick reference for Claude Code
- [filter/README.md](../../filter/README.md) - Filter syntax documentation
- [README.md](../../README.md) - Project overview and examples

# Issue 003 - PostgreSQL schema, migrations, seed strategy, and query conventions

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** 002  
**Owner:** Main integrator

## Objective

Create the durable relational model and database conventions required by the file vault before backend services are implemented.

## Acceptance criteria

- [x] Versioned initial migration exists.
- [x] Users, blobs, files, folders, tags, sharing, download events, and audit events are modeled.
- [x] Constraints protect ownership, references, counters, MIME metadata, and share target exclusivity.
- [x] Indexes support ownership, combined filters, filename search, folders, shares, downloads, and audit queries.
- [x] Seed strategy is documented and keeps password hashing in the application layer.
- [x] Query and statistics conventions are documented.
- [x] Migration files pass static structural inspection and are cross-referenced with the API contract.

## Verification evidence

Static migration verification passed: the migrations contain all required tables, extensions, enums, constraints, and index families, begin/end with transactions, and contain no destructive reset operations or credentials. Live verification is now also complete: both migrations applied successfully to the local `balkanid` database; a read-only structural query confirmed 11 public tables, 37 indexes, and 4 custom enums. The migration README and database overview define the seed boundary, reference-count deletion lifecycle, ownership model, statistics formulas, and query conventions.

## Files created

- `migrations/000001_initial_schema.sql`
- `migrations/README.md`
- `docs/database.md`

## Closure note

Issue 003 is closed. Issue 004 is now active: Go service skeleton, configuration loading, health checks, and database wiring.

# ADR 0001: Use a modular monolith

## Decision

Implement the backend as one Go deployable with explicit domain modules and interfaces.

## Rationale

The task needs strong transactional behavior across metadata and blob references. A modular monolith keeps those transactions and local development simple while still separating auth, files, storage, sharing, search, statistics, admin, and audit responsibilities.

## Consequences

- One deployment and one local process to operate.
- Clear package boundaries and service interfaces allow later extraction.
- Cross-module database transactions remain straightforward.
- Horizontal scaling is supported through stateless API instances plus PostgreSQL/Redis/blob storage.

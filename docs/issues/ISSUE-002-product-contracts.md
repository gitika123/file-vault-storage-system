# Issue 002 - Product contracts, configuration, API outline, and ADRs

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** 001  
**Owner:** Main integrator

## Objective

Freeze the shared decisions that backend, frontend, database, infrastructure, and future sub-agents must follow.

## Acceptance criteria

- [x] Configuration variables and defaults are documented.
- [x] Stable domain error codes and transport mapping are documented.
- [x] GraphQL schema outline exists for core metadata and actions.
- [x] REST contract exists for binary upload/download/preview and health endpoints.
- [x] Architecture decisions explain the modular monolith and GraphQL/REST boundary.
- [x] Quota/statistics formulas are explicit and testable.
- [x] `.env.example` exists without real secrets.
- [x] Contract verification passes and this issue is marked `CLOSED`.

## Files created

- `docs/contracts/configuration.md`
- `docs/contracts/error-catalog.md`
- `docs/contracts/graphql-schema.graphql`
- `docs/contracts/rest-endpoints.md`
- `docs/adr/0001-modular-monolith.md`
- `docs/adr/0002-graphql-and-rest-boundary.md`
- `docs/adr/0003-quota-and-statistics-semantics.md`
- `.env.example`

## Verification evidence

Contract verification passed with PowerShell: all eight required files exist and the check found the required GraphQL query/mutation markers, upload/download paths, stable error codes, and core environment variables. Issue 003 is now active.

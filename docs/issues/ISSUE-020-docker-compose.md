# Issue 020 - Docker Compose and multi-stage images

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** 004, 013  
**Owner:** Main integrator

## Acceptance criteria

- [x] Compose starts PostgreSQL, Redis, API, and frontend with health dependencies.
- [x] Backend and frontend use multi-stage production images.
- [x] Database migrations are mounted into the first-run PostgreSQL initialization boundary.
- [x] Secrets are supplied through environment interpolation and are not committed.
- [x] Compose smoke test passes on Docker Desktop.

## Verification

Docker Desktop 4.88.1 was installed and started. `docker compose config` passed; both multi-stage images built successfully; PostgreSQL, Redis, API, and frontend started with healthy dependency checks; the frontend returned HTTP 200; and authenticated login, current-user, and storage-statistics requests returned HTTP 200 through the frontend proxy. Demo users are seeded idempotently by the API container at startup using ignored local environment variables.

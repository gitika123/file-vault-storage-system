# Issue 004 - Go service skeleton, health checks, configuration, and database wiring

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** 002, 003  
**Owner:** Main integrator

## Objective

Create the first executable backend boundary with typed configuration, production validation, PostgreSQL connection wiring, graceful shutdown, and liveness/readiness endpoints.

## Acceptance criteria

- [x] Go module and API entrypoint exist in `backend/`.
- [x] Configuration names and defaults match the contract.
- [x] Production validation requires database, Redis, and sufficiently strong secrets.
- [x] PostgreSQL is opened through a dedicated platform package.
- [x] Liveness and readiness handlers exist with safe status responses.
- [x] Server has timeouts and graceful SIGINT/SIGTERM shutdown.
- [x] Configuration tests cover valid defaults and invalid production secrets.
- [x] Static structure verification passes.
- [x] Go compile/test execution is explicitly deferred until the toolchain is available.

## Verification evidence

`SERVICE_SKELETON_STATIC_CHECK=PASS`. The current environment has no `go` executable, so `go test ./...` and `go vet ./...` could not be run here. They are required in the CI/infrastructure verification issue before final delivery.

## Files created

- `backend/go.mod`
- `backend/cmd/api/main.go`
- `backend/internal/platform/config/config.go`
- `backend/internal/platform/config/config_test.go`
- `backend/internal/platform/database/database.go`
- `backend/internal/platform/health/health.go`

## Closure note

Issue 004 is closed for implementation and static verification. Issue 005 is active: authentication, sessions, CSRF protection, and RBAC.

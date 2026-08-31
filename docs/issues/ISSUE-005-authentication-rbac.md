# Issue 005 - Authentication, sessions, CSRF protection, and RBAC

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** 003, 004  
**Owner:** Main integrator

## Objective

Implement secure identity primitives and a database-backed session/policy boundary that all GraphQL and REST handlers will use.

## Completed in this issue so far

- Argon2id password hashing with random salts and constant-time verification.
- Opaque 256-bit session and CSRF tokens generated from the OS cryptographic random source.
- SHA-256 hashes stored for session/CSRF lookup; raw tokens are never stored in the database.
- HttpOnly, SameSite=Lax session cookie helpers with production Secure support.
- CSRF validation for state-changing HTTP methods.
- Central policy helpers for authenticated users, admins, and the strict uploader-only delete rule.
- PostgreSQL `sessions` migration with expiry, revocation, last-seen timestamps, and indexes.
- Database-backed authentication service for credential verification, session creation, principal loading, and revocation.

## Remaining acceptance criteria

- [x] Password and token primitives are implemented.
- [x] Session schema and service are implemented.
- [x] Central policy helpers are implemented.
- [x] Login/logout/current-user transport handlers are wired into the API.
- [x] CSRF token issuance is wired into the API.
- [x] Authentication and policy tests run with the Go toolchain; database-backed login is wired to the migrated schema.
- [x] The transport middleware exposes one principal/policy boundary for future GraphQL and REST handlers.
- [x] Issue is marked `CLOSED` after seeded-account integration verification.

## Verification status

Auth file checks pass. `gofmt` was run on all backend source, and `go test ./...` passes with Go 1.27. The live PostgreSQL sessions migration is applied and the service is wired to it. A live API smoke test seeded Alice, authenticated successfully, confirmed session/CSRF cookie issuance, confirmed `/me` returned 200, confirmed CSRF-protected logout returned 200, and confirmed `/me` returned 401 after logout. Issue 006 is now active.

# Issue 011 - Private, public, and direct sharing

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** 005, 009  
**Owner:** Main integrator

## Acceptance criteria

- [x] Only the file owner can create or revoke shares.
- [x] Public share tokens are cryptographically random and only SHA-256 hashes are persisted.
- [x] Public shares support optional expiration metadata and revocation.
- [x] Direct shares resolve active recipients by email and enforce `view`/`download` permissions.
- [x] Self-sharing and disabled/nonexistent recipients are rejected.
- [x] Share mutations require session authentication and CSRF validation.
- [x] Go tests pass and live API checks verify public creation, direct creation, revocation, and hashed token storage.

## Verification evidence

`gofmt` and `go test ./...` pass. Live API checks created a public share and a direct download share for Alice's file, revoked the public share, and confirmed a 32-byte token hash in PostgreSQL. The raw token was returned only in the creation response and is not stored in plaintext. Binary access and public-token resolution are intentionally completed in Issue 012.

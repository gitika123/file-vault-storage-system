# Issue 012 - Secure downloads, previews, counters, and audit events

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** 011, 008  
**Owner:** Main integrator

## Acceptance criteria

- [x] Authenticated owners/admins can download files they are authorized to access.
- [x] Direct recipients require the `download` permission for binary downloads.
- [x] Public downloads require an active, unexpired, non-revoked token.
- [x] Public token hashes are compared in the database; raw tokens are never persisted.
- [x] Previews are limited to allowlisted PDF/image MIME types and use inline disposition.
- [x] Successful content access increments the file counter and records a download event.
- [x] Successful content access writes an audit event with actor/share context.
- [x] Go tests pass and live owner, preview, public-link, counter, and event checks pass.

## Verification evidence

`gofmt` and `go test ./...` pass. Live API checks returned HTTP 200 for an authenticated private download, an authenticated PDF preview, and a public-token download. PostgreSQL showed the file counter and download-event count incrementing to three for the three successful accesses. Public share resolution uses a SHA-256 token hash and rejects revoked/expired rows.

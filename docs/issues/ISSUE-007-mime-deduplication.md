# Issue 007 - MIME validation and SHA-256 deduplication

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** 003, 006  
**Owner:** Main integrator

## Objective

Turn the safe staging layer into a complete upload workflow: validate content against the declared type, persist one physical blob per digest, and create logical file references for users.

## Planned scope

- Content-signature MIME detection for supported file types.
- Compatibility policy for declared MIME, detected MIME, and filename extension.
- Multipart single/multi-upload handler with authenticated ownership.
- Atomic blob insert/reuse using the unique SHA-256 constraint.
- Storage key generation from digest, never filename/path.
- New-blob commit and existing-blob temporary cleanup.
- Per-file result payloads for partial multi-upload success.
- Race-safe concurrent identical-upload handling.
- Tests for forged extensions, same-user duplicates, cross-user duplicates, and concurrent uploads.

## Acceptance criteria

- [x] Renamed/mismatched content is rejected with `INVALID_MIME` / HTTP 415.
- [x] Valid single and multi-file uploads create file references.
- [x] Identical content reuses one blob and exposes deduplication status without cross-user ownership disclosure.
- [x] Concurrent identical uploads do not create duplicate blobs or incorrect reference counts.
- [x] Temporary objects are removed after both duplicate and failure paths.
- [x] Upload results preserve per-file success/failure.
- [x] Go tests pass against the migrated PostgreSQL database configuration.
- [x] Issue is marked `CLOSED` after live API verification.

## Verification evidence

`gofmt` and `go test ./...` pass. Live API checks returned one created valid PDF and one `INVALID_MIME` rejection for a forged DOCX. A repeat upload returned `deduplicated: true`. Four concurrent same-content uploads all succeeded as deduplicated, and a read-only database query confirmed exactly one matching physical blob with seven logical references.

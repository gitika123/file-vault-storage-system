# Issue 008 - Quotas, storage statistics, reference counts, and safe deletion

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** 003, 007  
**Owner:** Main integrator

## Objective

Enforce logical per-user quotas, expose accurate original/deduplicated storage statistics, and implement uploader-only deletion with reference-safe physical cleanup.

## Planned scope

- Transactional logical quota checks before creating a file reference.
- Per-user original bytes, distinct-blob bytes, savings bytes, savings percentage, and quota payload.
- Strict uploader/owner deletion policy through the centralized auth boundary.
- Transactional file-reference deletion and blob reference-count decrement.
- `pending_delete` state and recoverable physical blob cleanup.
- Admin-safe usage calculations without cross-user duplicate disclosure.
- Tests for exact quota boundaries, zero usage, duplicate ownership, last-reference cleanup, and concurrent deletion.

## Acceptance criteria

- [x] Uploads that exceed logical quota return `QUOTA_EXCEEDED` without creating a file reference.
- [x] Exact quota boundary behavior is enforced by the transactional `current + requested > quota` check.
- [x] Statistics match the documented formulas for unique and duplicate files.
- [x] Non-uploaders cannot delete files.
- [x] Deleting the last reference removes the logical file and its physical blob safely.
- [x] Row locking and conditional reference-count updates prevent negative counts under concurrent deletion.
- [x] Cleanup failure leaves a recoverable `pending_delete` blob state.
- [x] Go tests and live API checks pass.
- [x] Issue is closed after runtime verification.

## Verification evidence

`gofmt` and `go test ./...` pass. The live API returned storage statistics with original bytes, distinct-blob bytes, savings, savings percentage, and quota. Bob's attempt to delete an Alice-owned file returned HTTP 403. A unique Alice upload was deleted successfully; the file row and associated blob row were both absent afterward. A temporary Bob quota of one byte caused a valid upload to return a per-file `QUOTA_EXCEEDED` rejection, and the original quota was restored. The upload handler now maps the domain quota error instead of exposing it as an internal storage failure.

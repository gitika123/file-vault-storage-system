# Issue 006 - Blob storage abstraction and streaming upload limits

**Status:** IN PROGRESS  
**Priority:** P0  
**Depends on:** 004, 005  
**Owner:** Main integrator

## Objective

Create the storage boundary and upload transport needed to stream files safely without loading whole files into memory. Deduplication and MIME validation are intentionally separate issues so each invariant can be tested independently.

## Planned scope

- `BlobStore` interface for create/open/delete/exists operations.
- Local filesystem implementation using generated keys and no user-controlled paths.
- Temporary upload files with cleanup on success/failure.
- Multipart request limits: per-file, aggregate request, filename length, and file count.
- Streaming copy with byte counting and SHA-256 calculation hooks.
- `POST /api/uploads` endpoint returning per-file results and partial-success errors.
- Unit tests for size limits, path safety, cleanup, and stream behavior.
- Integration test against the local PostgreSQL/blob configuration.

## Acceptance criteria

- [x] Upload bytes are streamed to temporary storage rather than buffered in memory.
- [x] File and request limits are configurable and enforced.
- [x] Storage keys cannot contain user-controlled paths or filenames.
- [x] Temporary files are cleaned after success and failure.
- [x] Storage failures have typed errors for stable API mapping.
- [x] Go storage and upload-limit tests pass.
- [x] Runtime verification confirms the backend test suite passes with Go 1.27.
- [x] The user-facing multipart endpoint is explicitly deferred to Issue 007, where it can persist validated deduplicated references.

## Notes

SHA-256 deduplication, content MIME validation, quota accounting, and reference-count lifecycle belong to Issues 007 and 008. This issue exposes the staged object size/digest and storage hooks without duplicating those business rules.

## Verification evidence

`gofmt` was run on backend source and `go test ./...` passed. Storage tests cover stage/commit/open/delete, oversized stream rejection, invalid-key rejection, filename normalization, and request-size boundaries.

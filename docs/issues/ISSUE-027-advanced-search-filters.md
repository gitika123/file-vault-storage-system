# Issue 027 - Complete advanced search and filtering

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** 010, 026  
**Owner:** Main integrator

## Requirement source

The hiring task requires search by filename, filtering by MIME type, size range, date range, tags, and uploader name. Multiple filters must be combinable, and queries should be optimized for performance at scale.

## Scope

- Add uploaded-after and uploaded-before filters to the REST file-list API.
- Add uploader-name filtering with authorization-safe joins.
- Preserve existing filename, MIME, size, folder, tag, and cursor filters.
- Add indexes/query-plan evidence for the combined search path.
- Add API and database-backed tests for boundaries and combinations.

## Acceptance criteria

- [x] Date-range filters work independently and together.
- [x] Uploader-name filtering works with owner/admin scope rules.
- [x] All required filters can be combined.
- [x] Invalid ranges return stable validation errors.
- [x] Query remains parameterized, paginated, indexed, and accompanied by performance evidence.

## Verification

Added `uploadedAfter`, `uploadedBefore`, and `uploaderName` filters to the owner-scoped REST listing. Invalid timestamps and inverted ranges return `INVALID_INPUT`/HTTP 400. The uploader-name trigram index is included in migration `000003_search_indexes.sql`, while the existing owner/date, MIME, size, filename, folder, and tag indexes support the remaining filters. Live Compose verification returned HTTP 200 with one result for a combined uploader/date query and HTTP 400 for an inverted range.

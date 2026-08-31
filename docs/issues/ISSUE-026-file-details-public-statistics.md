# Issue 026 - Complete file details and public statistics UX

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** 012, 014, 016  
**Owner:** Main integrator

## Requirement source

The hiring task requires users to view detailed metadata for owned files, including uploader, size, upload date, and deduplication information. It also requires owners to track download counters for publicly shared files.

## Scope

- Add an owner-scoped file detail endpoint containing uploader, size, declared/detected MIME, upload date, deduplication status, logical/physical storage information, visibility, and download count.
- Display the complete metadata in the frontend detail view.
- Display public download count and sharing status to the owner.
- Preserve owner authorization and avoid disclosing data from other users.

## Acceptance criteria

- [x] File detail API returns the required metadata for an authorized owner.
- [x] Frontend detail view displays uploader, size, upload date, deduplication, visibility, and public download count.
- [x] Unauthorized users cannot retrieve another user's private metadata.
- [x] Backend and frontend tests pass.
- [x] Live smoke test verifies the detail and statistics flow.

## Verification

Added `GET /api/files/{id}` with owner-scoped authorization and blob/uploader/public-share metadata. The React detail modal now displays the required fields. Live Compose verification authenticated Alice, uploaded a PDF, and retrieved detail data with uploader, detected MIME, deduplication state, and download count. The first immediate detail request received the expected HTTP 429 limiter response; the delayed retry returned HTTP 200.

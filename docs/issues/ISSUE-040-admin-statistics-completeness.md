# Issue 040 — Admin statistics completeness

**Priority:** P0  
**Status:** CLOSED  
**Depends on:** 028, 039

## Implemented

- Admins can upload files.
- Admins can share files with specific users.
- Admin inventory lists every file with uploader name/email, MIME type, size, upload date, and download count.
- Admin statistics now display users, files, logical storage, physical deduplicated storage, deduplication savings, savings percentage, and downloads.
- Admin API access remains RBAC-protected; normal users receive `403 FORBIDDEN`.

## Verification

- Frontend production Docker build passed.
- Frontend container restarted successfully.
- Existing core smoke suite verified admin statistics, upload, direct share, inventory, and non-admin denial.

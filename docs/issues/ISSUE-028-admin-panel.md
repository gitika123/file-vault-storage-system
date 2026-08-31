# Issue 028 - Admin panel and cross-user analytics

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** 017, 026, 027  
**Owner:** Main integrator

## Requirement source

The hiring task requires an admin panel where admins can upload files and share them with other users, list all files with uploader details, and view download counts and usage statistics. The deliverables call graphs, sharing, and analytics bonus features; the core admin panel itself is mandatory.

## Scope

- Add an admin-only file listing endpoint with uploader details and pagination.
- Add an admin-only frontend panel with file inventory, uploader, size, date, deduplication, visibility, and download count.
- Add admin usage/statistics cards and an upload/share workflow using existing secure APIs.
- Enforce admin RBAC on every admin endpoint and UI route.
- Add tests for admin access and ordinary-user denial.

## Acceptance criteria

- [x] Admin can open a dedicated admin panel.
- [x] Admin panel lists all files with uploader details and download counts.
- [x] Admin can view aggregate usage statistics.
- [x] Admin can upload and share files through the panel.
- [x] Ordinary users receive 403 from admin APIs and cannot access the panel data.

## Verification

Added the admin-only `GET /api/admin/files` inventory endpoint and a React administrator panel with workspace usage cards, uploader/file metadata, download counts, upload, and direct sharing controls. Live Compose verification returned HTTP 200 for the seeded admin on both admin endpoints and HTTP 403 for the seeded normal user. Go tests and the frontend production build pass.

# Issue 047 — Separate admin files and statistics tabs

**Priority:** P0  
**Status:** CLOSED

## Implemented

- Admin Statistics & Oversight now opens on the uploaded-files inventory by default.
- Added separate `Files & uploads` and `Statistics` tabs.
- Files & uploads contains the complete admin inventory, uploader details, download counts, upload, and sharing controls.
- Statistics contains only the exact metric cards and visual analytics.
- The regular user dashboard is not rendered underneath the admin screen.

## Verification

- Frontend TypeScript/Vite production build passed.
- Frontend container rebuilt and restarted successfully.

# Issue 034 - Core smoke coverage and public-download regression

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** 012, 019, 020, 033

## Problem

The project needed one deterministic end-to-end check covering the mandatory hiring-task journeys. That check exposed a production-path HTTP 500 when an anonymous user downloaded a public share.

## Resolution

- Added `scripts/core-smoke.ps1`, a repeatable local-proxy smoke suite covering authentication, CSRF, upload, MIME rejection, deduplication, listing, metadata, folders, moves, statistics, public sharing/download, preview, authenticated download, and admin authorization.
- Fixed PostgreSQL parameter typing for nullable UUID values and the public-share audit boolean in `backend/internal/files/content.go`.
- Restored the authenticated dashboard search control and connected it to the owner-scoped filename filter after the browser UAT caught its absence.
- Kept the smoke helper paced so it validates the configured rate-limit behavior without creating artificial burst failures.

## Acceptance evidence

Command:

```powershell
pwsh -NoProfile -File scripts/core-smoke.ps1
```

Result: **17 core checks passed** against `http://localhost:5173`, including public-token download (HTTP 200), PDF preview (HTTP 200), authenticated download (HTTP 200), and admin/non-admin authorization boundaries.

Additional verification:

```powershell
cd frontend
npm run build
```

Result: strict TypeScript/Vite production build passed.

Browser UAT result: both signed-out and seeded-user Chromium journeys passed.

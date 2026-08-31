# Issue 038 — Drive-style folder user sharing and inherited access

**Priority:** P0  
**Status:** CLOSED  
**Depends on:** 035, 037

## Objective

Allow a folder owner to share a folder with a specific registered user, while preserving the existing public-link workflow. Files inside the shared folder inherit the recipient's view/download permission.

## Implemented

- Added folder-aware direct-share validation on `POST /api/shares/direct`.
- Added folder-level share persistence using the existing `user_shares.folder_id` relation.
- Added inherited authorization for files directly contained by a shared folder.
- Added a user-facing folder sharing action that accepts an email address; leaving the prompt blank continues to the public-link flow.
- Hardened the authorization query with `NULLIF(... )::uuid` so root-level files do not produce invalid UUID casts.
- Retained clear permission semantics: `view` permits preview, while `download` permits download access.

## Verification

- API and frontend Docker images rebuilt successfully.
- PostgreSQL and Redis healthy; API and frontend restarted successfully.
- `scripts/core-smoke.ps1`: 27 checks passed.
- `frontend/npm run test:uat`: 3 Chromium tests passed.
- Frontend production build passed inside the Docker image.

## Notes

Folder deletion remains intentionally safe-only-empty. The public folder landing page lists immediate child files and provides preview/download links. Render deployment remains tracked separately in Issue 024.

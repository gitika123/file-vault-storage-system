# Issue 037 - Folder lifecycle, public preview, filters, statistics, and admin acceptance

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** 009, 011, 012, 027, 028, 036

## Resolution

- Added owner-scoped folder rename and deletion APIs with safe protection against deleting non-empty folders.
- Added folder rename/share/delete controls and folder-level public links in the Folders view.
- Public file links now open a preview/download landing page; shared-folder links list child files with preview/download actions.
- Added complete combined dashboard filters: filename, MIME, size range, date range, tag, and uploader.
- Corrected storage cards to distinguish physical deduplicated usage, logical original usage, savings bytes, and savings percentage.
- Added subtle statistics/file-panel motion with `prefers-reduced-motion` support.
- Expanded the smoke suite to cover folder lifecycle, public folder access, combined filters, and admin upload/direct-share/inventory flows.

## Verification

```powershell
pwsh -NoProfile -File scripts/core-smoke.ps1
```

Result: **27 core checks passed**, including explicit `FOLDER_NOT_EMPTY` behavior, folder rename/delete, public folder landing/preview, combined filters, storage statistics, admin upload, admin sharing, admin inventory, and normal-user admin denial.

```powershell
Push-Location frontend
npm run build
npm run test:uat
Pop-Location
```

Result: frontend build passed and all 3 Chromium UAT tests passed.

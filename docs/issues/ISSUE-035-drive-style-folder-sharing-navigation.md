# Issue 035 - Drive-style folder, sharing, and navigation workflows

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** 009, 011, 014, 016, 034

## Problem

The file-details dialog changed the selected destination immediately but did not provide an explicit action to commit the move. Public links had no adjacent copy control, and the sidebar's Folders and Uploads entries were visual-only.

## Resolution

- Added a staged folder selector with an explicit disabled-until-changed `Move` button.
- Added a Google Drive-style copy-link control with copied feedback beside generated public links.
- Converted sidebar entries to accessible buttons with active states.
- Added functional All files, Folders, folder-contents, and Recent uploads views.
- Preserved owner-wide All files listing while making the folder view an explicit folder navigator.

## Acceptance evidence

```powershell
Push-Location frontend
npm run build
$env:UAT_EMAIL='alice@example.com'
$env:UAT_PASSWORD='<local seeded value>'
npm run test:uat
Pop-Location
```

Result: production build passed and all 3 Chromium UAT tests passed, including authenticated sidebar navigation. The existing `scripts/core-smoke.ps1` suite also passed all 17 core API checks after the change.

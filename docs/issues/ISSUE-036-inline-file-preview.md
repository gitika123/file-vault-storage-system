# Issue 036 - Inline file preview workflow

**Status:** CLOSED  
**Priority:** P1 (bonus)

## Decision

Preview is not required for the mandatory core feature list; secure download is the required content-access path. It is retained because the hiring task explicitly encourages file previews as a bonus and it improves the Google Drive-style experience.

## Resolution

- Preview now opens inside the file-details dialog instead of navigating away from the vault.
- PDFs render in an inline frame and images render inline.
- Unsupported types receive a clear explanation and can still be downloaded.

## Verification

- Frontend TypeScript/Vite production build passed.
- Existing core smoke suite passed 17/17, including authenticated PDF preview HTTP 200.
- Chromium UAT passed 3/3.

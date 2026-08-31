# Issue 033 - Upload MIME and drag-and-drop regression

**Status:** CLOSED  
**Priority:** P0  

## Root cause

The process-local rate limiter ignored its configured burst, so the dashboard's parallel startup requests triggered false `429` responses. The upload drop target also relied on a label-only drop handler, allowing the browser's default file-navigation behavior in some drag paths. MP4 was not included in the MIME signature policy.

## Resolution

- The limiter now honors the configured burst allowance.
- Added an explicit drag/drop capture target that prevents browser navigation.
- Added MP4 `ftyp` signature detection and extension validation.
- Frontend TypeScript/Vite build and Docker image rebuild passed.

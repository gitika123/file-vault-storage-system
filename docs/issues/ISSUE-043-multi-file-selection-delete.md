# Issue 043 — Multi-file selection and deletion

**Priority:** P1  
**Status:** CLOSED

## Implemented

- Added a checkbox to each file row for multi-selection.
- Added selected-count, Clear selection, and Delete selected controls.
- Bulk deletion confirms the operation and executes the existing protected delete endpoint per file, preserving owner authorization and deduplication reference-count safety.
- Selection is cleared after completion and partial failures are shown to the user.

## Verification

- Frontend TypeScript/Vite production build passed in Docker.
- Frontend container restarted successfully.

# Issue 015 - Drag-and-drop multi-upload UX

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** 006, 007, 013  
**Owner:** Main integrator

## Acceptance criteria

- [x] File picker accepts multiple files.
- [x] Upload area accepts drag-and-drop and prevents browser navigation.
- [x] Upload state is visible while a batch is being submitted.
- [x] Per-file backend rejection messages are surfaced to the user.
- [x] Successful upload refreshes stats and the file list.
- [x] Production build passes.

## Verification evidence

The React upload control now supports both picker and drop events, shows `Uploading…` during the batch, preserves backend per-file errors, and refreshes the vault afterward. `npm run build` passes strict TypeScript and Vite compilation.

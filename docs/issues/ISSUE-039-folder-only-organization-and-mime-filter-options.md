# Issue 039 — Folder-only organization view and selectable MIME filters

**Priority:** P0  
**Status:** CLOSED  
**Depends on:** 014, 027, 035

## Objective

Keep upload actions in the All files/Uploads workflows, make Folders a focused organization and sharing workspace, and provide selectable MIME filter values rather than requiring users to know MIME syntax.

## Implemented

- The upload/drop zone is hidden whenever the Folders view is active, including folder navigation.
- Folder view remains responsible for creating, opening, renaming, sharing, and safely deleting folders.
- The MIME filter exposes selectable common values through a native datalist, including image, text, audio, video, PDF, archive, JSON, CSV, and binary MIME types.
- Filename search and the existing size/date/tag/uploader filters remain combinable.

## Verification

- Frontend Docker production build passed.
- Frontend container restarted successfully.
- Existing 3-test Chromium UAT suite and 27-check core API smoke suite remain passing from the immediately preceding rebuild.

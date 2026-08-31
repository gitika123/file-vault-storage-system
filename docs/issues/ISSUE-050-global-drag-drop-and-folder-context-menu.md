# Issue 050 - Global drag-and-drop and folder context-menu UX

**Priority:** P0  
**Status:** CLOSED

## Acceptance criteria

- [x] File drops are accepted anywhere in an upload-enabled vault view.
- [x] Browser default file navigation is cancelled on both drag-over and drop.
- [x] Dropped files use the same CSRF-refreshing multi-file upload pipeline as the file picker.
- [x] The folders view remains organization-only and does not upload dropped files.
- [x] Folder cards expose open, rename, share, and delete through the context menu on home and folders views.
- [x] Filter semantics are explicit: MIME type, inclusive minimum/maximum bytes, inclusive uploaded-date range, uploader, and tags.

## Verification

- `npm run build` passed strict TypeScript validation and Vite production compilation.
- Browser drag/drop behavior follows the platform requirement to cancel `dragover` before a target can receive `drop` events, as documented by [MDN](https://developer.mozilla.org/en-US/docs/Web/API/HTML_Drag_and_Drop_API/Drag_operations) and React’s DOM event reference.

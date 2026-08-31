# Issue 014 - Vault dashboard, file list, filters, and details UI

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** 009, 010, 013  
**Owner:** Main integrator

## Acceptance criteria

- [x] Dashboard presents storage, deduplicated usage, quota, and visible-file context.
- [x] File list displays name, type, size, tags, date, and destructive action affordance.
- [x] Search field is wired to owner-scoped backend filtering.
- [x] Empty, error, signed-out, and responsive states are present.
- [x] File delete action uses the CSRF-protected API.
- [x] Production frontend build passes.

## Verification evidence

`npm run build` passed after the dashboard shell was wired to the live API contracts. The implementation includes responsive navigation, storage cards, search-driven file refresh, file metadata rows, upload entry point, and delete confirmation.

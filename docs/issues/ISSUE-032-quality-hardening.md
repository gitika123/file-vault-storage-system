# Issue 032 - Quality hardening and scale evidence

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** 027, 028  
**Owner:** Main integrator

## Scope

- Expose all required advanced search controls in the frontend.
- Add pagination to the admin inventory.
- Make direct `view` permission usable for safe previews while retaining download restrictions.
- Replace the process-local limiter with a documented distributed implementation or explicitly isolate the fallback.
- Add repeatable query-plan/benchmark evidence and regression tests.
- Add pagination to the admin inventory and expose the required advanced filters in the frontend.

## Completion evidence

- Added MIME, size, date, uploader, tag, and filename controls to the frontend search surface.
- Added cursor pagination to the admin inventory and preserved owner/RBAC boundaries.
- Direct `view` shares are usable for safe previews while download permission remains separate.
- Rate-limit responses include `Retry-After`; the current limiter is explicitly documented as process-local and must be replaced with a shared store before multi-replica production.
- Added PostgreSQL query-plan baseline evidence in `docs/performance/search-plan.md`.
- Frontend production build and containerized backend build passed after the changes.

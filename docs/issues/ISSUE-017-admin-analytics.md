# Issue 017 - Admin analytics and protected dashboard API

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** 005, 012, 013  
**Owner:** Main integrator

## Acceptance criteria

- [x] Admin analytics are exposed through a dedicated endpoint.
- [x] Non-admin users receive `FORBIDDEN`.
- [x] Metrics include active users, logical files/bytes, ready physical bytes, and downloads.
- [x] The query avoids exposing individual user/file metadata.

## Verification evidence

The endpoint is mounted behind the existing session middleware and checks the principal role before running aggregate-only SQL. Live checks returned HTTP 200 for the seeded admin account and HTTP 403 for the seeded normal user; backend tests pass.

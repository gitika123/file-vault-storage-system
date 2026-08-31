# Issue 030 - Real-time download-count updates

**Status:** CLOSED  
**Priority:** P1 bonus  
**Depends on:** 012, 026, 028  
**Owner:** Main integrator

## Scope

- Add a lightweight Server-Sent Events stream for authorized owner/admin download-count updates.
- Publish an event after a successful download counter update.
- Keep polling/API behavior as the reliable fallback.
- Add lifecycle, authorization, and frontend subscription tests.

## Completion evidence

- Added authorized SSE endpoint `/api/events/downloads`.
- Download counter updates publish owner-scoped events; administrators receive all events.
- The React vault subscribes with `EventSource` and refreshes its data, while normal API refresh remains the fallback.
- Docker integration build and frontend production build passed after integration.

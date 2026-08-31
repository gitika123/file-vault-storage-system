# Issue 019 - Playwright UAT

**Status:** CLOSED

The manual UAT checklist is recorded in `docs/uat-checklist.md`. `frontend/tests/vault.spec.ts` covers the signed-out login surface and a seeded authenticated journey. Run `npm run test:uat` from `frontend` with the API and Vite server running; set `UAT_EMAIL` and `UAT_PASSWORD` for the authenticated case.

## Verification evidence

With the local API and Vite server running, `npm run test:uat` completed with 2 passing Chromium tests. The signed-out login surface and seeded authenticated vault journey both passed.

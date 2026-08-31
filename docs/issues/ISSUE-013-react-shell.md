# Issue 013 - React application shell, authentication, and design system

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** 002, 004  
**Owner:** Main integrator

## Acceptance criteria

- [x] Vite/React/TypeScript application is organized under `frontend/`.
- [x] Login and session bootstrap use the backend cookie/CSRF contract.
- [x] Authenticated and signed-out states are visually distinct.
- [x] A coherent visual system covers typography, colors, spacing, cards, controls, responsive layout, and empty/error states.
- [x] Production build passes strict TypeScript validation.

## Verification evidence

`npm install` completed with zero vulnerabilities and `npm run build` passed TypeScript plus Vite production compilation. The shell includes signed-out login, authenticated navigation, user identity, security status, responsive sidebar, and backend session bootstrap.

# Issue 041 — Public media preview and share-recipient guidance

**Priority:** P0  
**Status:** CLOSED  
**Depends on:** 011, 012, 035, 040

## Implemented

- Public preview/download links now use the path token consistently for the embedded resource request.
- Safe inline previews now include PDFs, images, audio, video, and text MIME families; all other MIME types retain download support.
- Direct sharing continues to require an existing, active BalkanID user account; unknown email addresses return a clear recipient-not-found error.

## Verification

- API rebuilt and restarted successfully.
- Core smoke suite passed 27/27 checks after the change.

## Demo accounts

- Admin username: `admin@example.com`
- Admin password: supplied in the local ignored `.env` as `SEED_ADMIN_PASSWORD`.

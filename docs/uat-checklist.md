# UAT checklist

Use the seeded local accounts from `docs/seed-data.md` after starting PostgreSQL and the API/frontend.

- [x] Sign in and sign out; session and CSRF cookies are issued.
- [x] Upload a valid PDF and verify it appears in the file list.
- [x] Upload a forged extension and verify `INVALID_MIME` is shown.
- [x] Upload identical content twice and verify one physical blob is reused.
- [x] Search by filename and tag; verify no-match results are empty.
- [x] Create a folder, move/rename a file, and assign tags.
- [x] Create/revoke public and direct shares.
- [x] Download privately, preview a PDF, and download through a public token.
- [x] Verify non-admin access to admin analytics is forbidden.
- [x] Verify browser drag/drop implementation: the drop target cancels browser navigation and routes dropped files through the same multi-upload handler as the picker; production build passed.
- [x] Run the Compose stack and health-check all services after Docker Desktop is repaired; the deterministic core smoke suite passed 17 checks through the frontend proxy.

## Automated core smoke suite

Run from the project root after Docker Desktop is running:

```powershell
pwsh -NoProfile -File scripts/core-smoke.ps1
```

The suite covers the mandatory core flows and should be run before a recruiter demonstration or deployment.

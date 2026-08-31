# Issue 048 — Storage statistics for users and administrators

**Priority:** P0  
**Status:** CLOSED

## Implemented

- User dashboard visibly labels its storage metrics as Storage statistics.
- Every user sees total physical storage used after deduplication.
- Every user sees original logical storage usage before deduplication.
- Every user sees savings in bytes and percentage.
- Administrators see the same storage concepts in the dedicated Statistics page, plus workspace-wide cards and visualizations.

## Verification

- Existing backend storage statistics API supplies all required fields.
- Frontend production build remains the verification gate for both dashboard scopes.

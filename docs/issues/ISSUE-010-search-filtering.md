# Issue 010 - Search and combined filtering

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** 003, 009  
**Owner:** Main integrator

## Objective

Make file discovery useful at vault scale by combining owner-scoped filename search, MIME filters, size bounds, folder filters, tag filters, and cursor pagination.

## Acceptance criteria

- [x] Filename search is case-insensitive and partial-match based.
- [x] MIME, minimum-size, maximum-size, folder, and tag filters can be combined.
- [x] Results remain owner-scoped and return an empty array when there are no matches.
- [x] Search uses the existing bounded `(created_at, id)` cursor rather than offset pagination.
- [x] Invalid cursors remain client errors.
- [x] Database predicates use indexed columns or targeted tag existence checks.
- [x] Go tests pass and live API checks verify combined filtering and zero-result behavior.

## Verification evidence

`gofmt` and `go test ./...` pass. Live API checks authenticated as Alice and combined `filename=renamed` with `tag=demo`, returning the expected renamed/tagged file. A high `minSizeBytes` filter returned HTTP 200 with an empty result collection. Query construction uses parameterized SQL and preserves the bounded cursor contract.

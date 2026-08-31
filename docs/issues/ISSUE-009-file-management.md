# Issue 009 - File management, folders, tags, and cursor pagination

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** 005, 008  
**Owner:** Main integrator

## Objective

Provide authenticated metadata management for the vault: stable file listing, folder organization, rename/move operations, tags, and bounded cursor pagination.

## Acceptance criteria

- [x] Authenticated users can list only their own files.
- [x] File listing is bounded to a maximum page size and uses a stable `(created_at, id)` cursor.
- [x] Invalid cursors are rejected as client input.
- [x] Users can create and list folders with owner-scoped parent relationships.
- [x] Users can rename and move only their own files.
- [x] Users can replace file tags, with owner-scoped tag reuse and validation.
- [x] Mutating endpoints require the existing CSRF protection.
- [x] Go tests pass and live API checks verify the workflow.

## Verification evidence

`gofmt` and `go test ./...` pass. Live API checks authenticated as Alice, created the `Recruiter Demo` folder, listed files with `first=2` and a `nextCursor`, renamed a file, assigned `demo` and `review` tags, moved the file into the new folder, and confirmed the updated folder/tag fields in the subsequent listing. Folder/file queries are owner-scoped through the SQL predicates; mutations are mounted behind session plus CSRF middleware.

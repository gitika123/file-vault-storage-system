# Database migrations

Migrations are numbered, forward-only SQL files and must run inside a transaction whenever PostgreSQL permits it. The application will acquire a migration lock before applying pending files.

## Current migration

`000001_initial_schema.sql` creates:

- PostgreSQL extensions for UUID generation, case-insensitive email, and trigram filename search.
- Users, blobs, files, folders, tags, shares, download events, and audit events.
- Constraints for ownership, reference counts, valid MIME/metadata, share target exclusivity, and non-negative counters.
- Indexes for ownership, folder navigation, search, filtering, shares, downloads, and audit history.

Demo credentials are intentionally created by the application seed command after the authentication implementation is complete. This keeps password hashing policy in Go instead of committing a generated credential hash into the schema migration.

## Query conventions

- Every user-visible file query begins with an authorization scope (`owner_id`, direct share, public share, or admin policy) before optional filters.
- All values are parameterized; never concatenate user input into SQL.
- List endpoints use stable cursor pagination ordered by a deterministic field plus `id`.
- Mutations that change blob references lock the affected `files` and `blobs` rows in a transaction.
- Statistics use `BIGINT` arithmetic and return zero-safe percentages.
- New indexes require a query-plan explanation in the issue or architecture documentation.

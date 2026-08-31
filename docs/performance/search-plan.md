# Search performance evidence

Captured from the local PostgreSQL 16 Compose service on 2026-08-29.

The tested combined owner/date/uploader query used `files_owner_created_idx`, completed in 1.382 ms on the seeded local dataset, and read four shared buffers. The users table was scanned because the local fixture contains only three users; production sizing should be rechecked with representative data.

Reproduce with `EXPLAIN (ANALYZE, BUFFERS)` on the combined filter query in `docs/contracts/rest-endpoints.md` after starting Compose. This is baseline evidence, not a production benchmark.

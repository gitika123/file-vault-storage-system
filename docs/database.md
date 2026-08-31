# Database schema overview

## Ownership model

`files` are logical user-owned references. `blobs` are physical content objects identified by a unique SHA-256 digest. Multiple file rows can reference one blob. `owner_id` is the user whose quota and vault contain the file; `uploaded_by` records the actor that performed the upload, allowing a future admin/on-behalf-of flow without losing provenance.

## Deletion model

The application deletes a file reference and decrements `blobs.reference_count` in one transaction. When the count reaches zero, the blob becomes eligible for garbage collection. The physical object is deleted only after the transaction commits and only while the reference count remains zero. `ON DELETE RESTRICT` on the blob relationship prevents accidental database-level removal while references exist.

## Statistics formulas

- Original usage: `SUM(files.size_bytes)` for active files owned by the user.
- Deduplicated usage: `SUM(DISTINCT blob_id size)` for distinct blobs referenced by the user's active files.
- Savings bytes: original usage minus deduplicated usage.
- Savings percentage: zero for zero original usage; otherwise savings divided by original usage multiplied by 100.
- Quota: logical original usage compared with `users.quota_bytes`.

## Search indexes

The owner/time, owner/size, owner/MIME, and folder indexes support the common filter paths. The trigram index supports partial filename search. Tag joins have indexes in both directions. Search implementation must still inspect query plans after the access-scope predicate is added.

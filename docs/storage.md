# Blob storage design

The storage package stages each upload into a temporary file while streaming bytes through a bounded reader and SHA-256 hasher. Once metadata validation and the database transaction are ready, the caller commits the staged object under a generated key. User-controlled filenames never become storage paths.

The local implementation is used for development and Docker volumes. The `LocalStore` boundary is intentionally small so an S3-compatible implementation can be added without changing upload or deduplication business logic.

Temporary objects are deleted on every failed stage. A successful stage remains until the caller commits it; higher-level upload code must discard it if database or blob commit fails. Cleanup/reconciliation is part of the deduplication issue.

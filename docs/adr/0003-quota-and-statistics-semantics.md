# ADR 0003: Define quota and storage statistics using logical ownership

## Decision

Quotas use the sum of active logical file references owned by a user. Original usage sums every owned file size. Per-user deduplicated usage counts each distinct blob referenced by that user once. Savings is the difference.

## Rationale

Logical quotas prevent deduplication from becoming a quota bypass. Explicit formulas keep the UI, API, SQL, tests, and reviewer explanation consistent.

## Consequences

- A duplicate upload still consumes logical quota even when it consumes no new physical bytes.
- Cross-user physical sharing is reflected in global storage efficiency but does not reveal another user's ownership.
- All calculations need regression tests, especially zero-byte and exact-boundary cases.

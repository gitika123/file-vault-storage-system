# ADR 0002: Use GraphQL for metadata and REST for binary streams

## Decision

Use GraphQL for typed metadata queries and mutations. Use REST multipart/streaming endpoints for uploads, downloads, and previews.

## Rationale

GraphQL matches the preferred API requirement and is well suited to combined metadata filters and dashboard views. Browser upload progress, streaming responses, range support, and large payload limits are clearer with conventional HTTP endpoints.

## Consequences

- The authorization policy must be shared by both transport layers.
- API documentation must describe both contracts.
- The frontend has one Apollo client plus a small authenticated HTTP upload/download client.

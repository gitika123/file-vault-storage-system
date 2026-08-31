# BalkanID Secure File Vault - Implementation Plan

## 1. Goal and submission strategy

Build a secure, polished file vault that satisfies every mandatory requirement, demonstrates production-minded engineering, and can be evaluated in minutes with one command.

The submission should optimize for the evaluation criteria in this order:

1. Correct, demonstrable core behavior.
2. Clear architecture and security decisions.
3. Polished, responsive user experience.
4. Automated verification and easy setup.
5. A small set of complete, useful bonus features.

The project should not chase every bonus before the core is reliable. The best bonus set is: RBAC, folders and user-specific sharing, upload progress, previews, audit logs, admin analytics, CI/UAT automation, Kubernetes manifests, and a live deployment. Helm and real-time subscriptions come last.

## 2. Recommended technical decisions

| Area | Decision | Why it is a strong hiring-task choice |
|---|---|---|
| Backend | Go, `gqlgen`, `pgx`, and `sqlc` | Typed GraphQL contract, explicit SQL, good performance, and compile-time query safety. |
| API | GraphQL for metadata and actions; REST for multipart upload and binary download | Uses the preferred API style without forcing large binary streams through GraphQL. Document this as a deliberate production choice. |
| Database | PostgreSQL with versioned migrations | Meets the brief and supports transactional deduplication, filtering, and statistics well. |
| Frontend | React, TypeScript, Vite, Apollo Client, React Router, React Hook Form, and a restrained component system | Fast to build, typed, testable, and polished without unnecessary framework complexity. |
| File storage | A `BlobStore` interface with local-volume implementation; optional S3-compatible adapter | Docker Compose works with no cloud credentials while the boundary remains production-ready. |
| Authentication | Email/password, Argon2id password hashes, short-lived access token in an HttpOnly cookie, CSRF protection, and role-based authorization | Private files and admin behavior are impossible to secure credibly without authentication. |
| Rate limiting | Configurable per-user token bucket backed by Redis, with a documented in-memory development fallback | Correct across multiple backend replicas and suitable for Kubernetes. |
| Testing | Go unit/integration tests, PostgreSQL integration tests, frontend component tests, and Playwright UAT | Covers business invariants and visible user journeys. |
| Delivery | Multi-stage Dockerfiles, Docker Compose, GitHub Actions, Kubernetes manifests, and optional cloud deployment | Directly maps to required and recommended deliverables. |

### Scope decisions that must be documented

- A user quota is based on logical owned bytes, not physical deduplicated bytes. This prevents users from bypassing quotas by uploading already-known content and keeps quota behavior fair and predictable.
- Per-user original usage is the sum of all of that user's active file sizes.
- Per-user deduplicated usage is the sum of each distinct blob referenced by that user, counted once.
- Per-user savings is `original usage - deduplicated usage`; percentage is zero when original usage is zero, otherwise `savings / original usage * 100`.
- Global physical usage is the sum of stored blob sizes, counted once globally.
- Deduplication is global at the storage layer, but the API must never reveal whether another user owns identical content. This avoids a cross-user information leak.
- A filename is metadata only. Physical storage keys are generated from blob identity and never contain user-controlled paths.
- Soft deletion is not enough by itself. A file reference is removed transactionally, and an unreferenced blob is marked for garbage collection. Bytes are removed only when the reference count is zero.

## 3. Target architecture

```text
React/TypeScript SPA
  |-- GraphQL: login state, file metadata, search, folders, shares, stats, admin
  |-- REST multipart: uploads with progress
  |-- REST stream: authenticated/public downloads and previews
  v
Go API
  |-- auth and centralized authorization policies
  |-- upload/deduplication service
  |-- file/folder/share service
  |-- search/statistics/admin service
  |-- audit and garbage-collection service
  |
  |-- PostgreSQL: durable metadata and transactional state
  |-- Redis: distributed rate-limit counters
  `-- BlobStore: local Docker volume or S3-compatible storage
```

The Go application should use a modular monolith. That is easier to review and operate than premature microservices, while domain boundaries make future extraction possible.

Suggested backend modules:

- `auth`: identities, sessions, roles, CSRF, and password handling.
- `files`: file metadata, ownership, tags, and deletion.
- `storage`: blob streaming, hashing, deduplication, and garbage collection.
- `folders`: hierarchy and inherited access checks.
- `sharing`: private, public-link, and direct-user grants.
- `search`: access-scoped filters and cursor pagination.
- `stats`: per-user and global storage/download analytics.
- `admin`: privileged queries and actions.
- `audit`: immutable activity events.
- `platform`: configuration, database, Redis, logging, health, and telemetry.

Resolvers and HTTP handlers must stay thin. They validate transport input, call a service, and map typed domain errors to stable API errors. SQL and blob operations belong behind repositories/interfaces.

## 4. Repository layout

```text
/
  backend/
    cmd/api/
    internal/{auth,files,storage,folders,sharing,search,stats,admin,audit,platform}/
    graph/
    migrations/
    queries/
    Dockerfile
  frontend/
    src/{app,components,features,graphql,hooks,lib,pages,test}/
    public/
    Dockerfile
  docs/
    architecture.md
    database.md
    security.md
    api.md
    deployment.md
    testing.md
    uat-checklist.md
    ai-usage.md
  k8s/
  .github/workflows/ci.yml
  docker-compose.yml
  .env.example
  Makefile
  README.md
```

## 5. Database design

### Core tables

**`users`**

- `id UUID PK`, `email CITEXT UNIQUE`, `display_name`, `password_hash`, `role`, `quota_bytes`, timestamps.
- Role is constrained to `user` or `admin` initially.

**`blobs`**

- `id UUID PK`, `sha256 BYTEA UNIQUE`, `size_bytes BIGINT`, `detected_mime`, `storage_key UNIQUE`, `reference_count`, `state`, timestamps.
- `state` supports `pending`, `ready`, and `pending_delete` to make upload and cleanup failures recoverable.
- Constraints prevent negative size and reference counts.

**`files`**

- `id UUID PK`, `owner_id FK`, `blob_id FK`, nullable `folder_id FK`, `name`, `declared_mime`, `detected_mime`, `size_bytes`, `was_deduplicated`, timestamps.
- The file is the user-owned logical reference; the blob is the shared physical object.
- Keep owner metadata even if the uploader is later disabled. Decide whether uploader and owner are separate; for this task they can be the same except an admin uploading on behalf of a user, where both fields should be retained.

**`folders`**

- `id UUID PK`, `owner_id FK`, nullable `parent_id FK`, `name`, timestamps.
- Unique active name per owner and parent.
- Prevent cycles in the service layer; cap nesting depth.

**`tags` and `file_tags`**

- User-scoped normalized tags and a many-to-many join with a composite primary key.

**`public_shares`**

- `id`, resource type/id or separate nullable file/folder foreign keys with a check constraint, `token_hash UNIQUE`, creator, optional expiry/revocation, timestamps.
- Store only a hash of the high-entropy public token.

**`user_shares`**

- Resource, recipient user, permission (`view` or `download`), grantor, timestamps, unique grant constraint.

**`download_events`**

- `id`, file, blob, actor when authenticated, public share when applicable, timestamp.
- Keep an aggregate counter on files for fast reads and events for analytics/audit. Increment atomically in the same database transaction.

**`audit_events`**

- Actor, action, resource type/id, safe JSON metadata, IP hash or carefully minimized network metadata, timestamp.
- Never log credentials, cookies, share tokens, or file contents.

### Important indexes

- B-tree indexes starting with `owner_id` or other access-scope columns, followed by `created_at`, `size_bytes`, and MIME where useful.
- `pg_trgm` GIN index on normalized filename for contains/partial search.
- Composite indexes on folder ownership/parent, share recipient/resource, and download file/timestamp.
- Tag lookup and join indexes in both traversal directions.
- Partial indexes for active/non-deleted rows if soft-delete metadata is retained.

Search queries must apply authorization scope before filters, use parameterized SQL, return cursor pagination, and have `EXPLAIN (ANALYZE, BUFFERS)` evidence against seeded volume data in the architecture notes.

## 6. Critical backend workflows

### Upload, validation, and race-safe deduplication

1. Authenticate the user and apply the per-user rate limiter.
2. Enforce request count, per-file size, aggregate request size, and filename length limits before expensive work.
3. Stream each upload to a temporary object while computing SHA-256 and byte count. Never load the whole file into memory.
4. Detect MIME from content signatures. Compare detected type with declared MIME and extension using a documented compatibility policy. Reject mismatches with HTTP `415` and a safe, actionable message.
5. Lock/check the user's logical usage and reject the upload if it would exceed the configured quota. Return a stable `QUOTA_EXCEEDED` error with limit and current usage.
6. Insert or locate the blob using the unique SHA-256 constraint. Handle concurrent identical uploads with `INSERT ... ON CONFLICT` plus row locking; application-level check-then-insert is insufficient.
7. For a new blob, atomically promote the temporary object to its final storage key and mark the row ready. For an existing blob, discard the temporary object.
8. Create the user's file reference and increment the blob reference count in a transaction.
9. Return per-file success/failure data so a multi-upload can partially succeed without hiding failures.
10. Emit a structured audit event without exposing the hash to other users.

Add a recovery job for stale `pending` blobs and abandoned temporary objects. This turns crash handling into a designed behavior instead of an implicit leak.

### Deletion and garbage collection

1. Fetch the file under row lock and authorize against the immutable owner/uploader rule. Admin status must not silently override the explicit “only uploader can delete” requirement unless a separately documented compliance action is added.
2. Delete the logical file reference and decrement the blob reference count in one transaction.
3. When count reaches zero, mark the blob `pending_delete` and enqueue or record cleanup work.
4. After commit, remove bytes from the blob store, then remove/finalize the blob row. Retry safely on storage failure.
5. Test two users deleting references to the same blob concurrently.

### Sharing and download

- Default every file/folder to private.
- Route every metadata, preview, and download request through one centralized policy: owner, explicit recipient, valid public share, or permitted admin read.
- Use at least 128 bits of randomness for public tokens, show the token once, store its hash, and support revoke/expiry.
- Sanitize `Content-Disposition`, set `X-Content-Type-Options: nosniff`, use a restrictive content security policy for inline previews, and force attachment for unsafe types.
- Increment download counts atomically only after a valid download is authorized. Define whether canceled/partial streams count; a simple documented policy is to count when streaming successfully begins.
- Folder shares should be evaluated consistently for descendants, with tests for moving items into and out of shared folders.

### Rate limiting and errors

- Default to 2 requests/second/user with a configurable burst. Apply stricter limits to login and public-token endpoints by IP plus identity where available.
- Return HTTP `429`, GraphQL extension code `RATE_LIMITED`, and `Retry-After`.
- Define typed errors for unauthenticated, forbidden, not found, invalid MIME, quota exceeded, rate limited, conflict, and internal failure.
- Avoid resource-existence leaks: unauthorized private resources should generally appear not found.

## 7. API contract

### GraphQL queries

- `me`
- `files(filter, sort, first, after)` with combined filename, MIME, size, date, tags, uploader, and folder filters.
- `file(id)` and `folder(id)`.
- `storageStats`.
- `adminFiles`, `adminUsageStats`, and `adminDownloadStats`, protected by RBAC.

### GraphQL mutations

- `login`, `logout`, and optional development-only registration.
- `renameFile`, `deleteFile`, `setFileTags`, and `moveFile`.
- Folder create/rename/move/delete.
- Create/revoke public share and grant/revoke user share.
- Admin upload metadata/on-behalf-of flow if implemented.

### GraphQL design standards

- Cursor pagination, not unbounded arrays.
- Purpose-built input types and payloads with field-level errors.
- DataLoader/batched repository calls to avoid N+1 queries.
- Depth/complexity and request-size limits.
- Descriptions in SDL for fields, inputs, errors, and authorization expectations.
- No raw database or storage errors exposed to clients.

### REST endpoints

- `POST /api/uploads` for multipart single/multi-upload with progress support.
- `GET /api/files/{id}/content` for authorized download.
- `GET /api/files/{id}/preview` for supported safe previews.
- `GET /public/{token}` or a public metadata query plus `GET /public/{token}/download`.
- `GET /health/live` and `GET /health/ready`.

## 8. Frontend experience

### Required screens

1. **Sign in** with seeded demo credentials clearly documented for local review.
2. **Vault dashboard** with used/quota progress, original usage, deduplicated usage, savings bytes/percentage, recent activity, and a prominent upload action.
3. **Files view** with table/grid toggle, folder navigation, filename search, filter drawer, active filter chips, sorting, cursor pagination, loading/empty/error states, and responsive behavior.
4. **Upload center** with drag-and-drop, file picker, multi-file queue, per-file progress, MIME/quota errors, retry/cancel, and a deduplicated badge on success.
5. **File details** with safe preview, uploader, size, MIME, upload date, tags, folder, dedup status, share state, download count, download, and owner-only delete confirmation.
6. **Sharing dialog** for private/public/direct-user access, link copy, expiry, and revoke.
7. **Public share page** that works without authentication and discloses only intended metadata.
8. **Admin dashboard** with all files/uploader details, aggregate storage cards, download trends, top files, and upload/share actions.

### UX quality bar

- Keyboard-accessible dropzone and dialogs, visible focus, semantic labels, adequate contrast, and screen-reader announcements for upload progress.
- Skeletons rather than layout jumps; specific empty states with next actions.
- Optimistic updates only where rollback is unambiguous.
- Confirmation for destructive actions, with exact filename and consequence.
- Responsive at phone, tablet, and desktop widths.
- Never rely on color alone for status.
- Use consistent byte/date formatting and show the timezone.

## 9. Security and reliability checklist

- Argon2id password hashing with sensible parameters and constant-time verification.
- HttpOnly, Secure-in-production, SameSite cookies; CSRF token for state-changing requests.
- Strict CORS allowlist and production TLS termination.
- Centralized authorization for GraphQL, REST, previews, and storage access.
- Parameterized SQL, GraphQL complexity limits, upload limits, and server timeouts.
- Content signature validation; extension and user MIME are not trusted.
- Random storage keys or digest-derived internal keys; no path traversal.
- Safe response headers and no inline execution of active content.
- Secrets from environment/Kubernetes Secrets, never committed.
- Structured JSON logs with request IDs and redaction.
- Graceful shutdown, readiness checks for dependencies, and database connection limits.
- Idempotent migrations and startup behavior.
- Dependency, static-analysis, and container vulnerability scanning in CI where time allows.
- README must distinguish implemented controls from production follow-ups such as antivirus scanning, object-store encryption/KMS, backups, and retention policies.

## 10. Testing and UAT plan

### Backend unit tests

- MIME compatibility policy, quota calculations, storage statistics, permission matrix, share-token hashing, cursor encoding, and error mapping.
- Reference count transitions and zero/empty percentage cases.

### Integration tests with real PostgreSQL and storage

- Same user uploads identical content twice: one blob, two file references, correct savings.
- Two users upload identical content: one physical blob, no cross-user metadata leak, independent logical stats.
- Two concurrent identical uploads: one ready blob and correct reference count.
- Delete one of multiple references: bytes remain. Delete last reference: cleanup occurs.
- Unauthorized user cannot list, view, preview, download, share, or delete a private file.
- Owner rule remains enforced for delete.
- Public token grants only intended access; revoked/expired token fails.
- Combined search filters produce correct access-scoped results and stable pagination.
- Quota boundary at exactly the limit and one byte over.
- Rate limit returns `429` and `Retry-After`.
- Download counter increments atomically under concurrent requests.
- Failed blob write or database transaction leaves recoverable state and no live reference to missing bytes.

### Frontend tests

- Components for upload queue, filter composition, stats, share dialog, and destructive confirmation.
- API error mapping to readable messages.
- Accessibility checks for core screens.

### Playwright UAT automation

1. Sign in, drag/drop several files, observe progress and completed rows.
2. Upload a duplicate and verify savings/dedup indicators.
3. Combine filename, MIME, size, date, tag, and uploader filters.
4. Create a public share, open it in an anonymous context, download, and verify the count.
5. Attempt delete as a non-owner and verify denial; delete references as owner and verify correct lifecycle.
6. Open admin dashboard and verify all-file/uploader analytics.

Create a manual `docs/uat-checklist.md` for browser/device checks, error-state exploration, keyboard navigation, responsive layout, and recovery behavior. Wire a reliable subset into CI.

## 11. Infrastructure and delivery

### Docker Compose acceptance target

`docker compose up --build` must start frontend, backend, PostgreSQL, and Redis with health checks, named volumes, migrations, and optional deterministic seed data. The frontend should be reachable from one documented URL without manual package installation.

Use multi-stage images, a non-root runtime user, pinned base-image versions, small build contexts, and health checks. Persist PostgreSQL and blob data in separate named volumes.

### Kubernetes

Provide `k8s/` manifests for namespace, ConfigMaps, example Secret references, backend/frontend/PostgreSQL/Redis Deployments or StatefulSets as appropriate, Services, Ingress, PVCs, probes, resource requests/limits, and migration execution. Do not commit real secrets. Document that a managed database/object store is preferred in real production.

### CI pipeline

Jobs should run in this order, with caching:

1. Backend formatting, linting, vet/static analysis, unit tests, and race tests where practical.
2. Frontend typecheck, lint, unit tests, and production build.
3. Integration tests with PostgreSQL/Redis services.
4. Docker image builds.
5. Playwright smoke/UAT tests against the composed stack.
6. Optional image publish and staging deploy on protected branches/tags.

## 12. Documentation package

The root README should be optimized for a recruiter who has ten minutes:

- One-paragraph product pitch and screenshots/GIF.
- Feature checklist separating core and bonus work.
- Architecture diagram and the GraphQL-plus-REST rationale.
- One-command setup, seeded accounts, URLs, and common troubleshooting.
- Five-minute demo script.
- Test commands and CI badge.
- Security highlights and known tradeoffs.
- Public deployment link when available.

Supporting docs:

- `architecture.md`: boundaries, workflows, failure handling, and tradeoffs.
- `database.md`: ER diagram, table purpose, indexes, statistics definitions, and deletion lifecycle.
- `security.md`: threat model, authorization matrix, upload validation, tokens, and remaining production controls.
- `api.md`: GraphQL and REST responsibilities, auth/error conventions, and links to SDL.
- `deployment.md`: Compose, Kubernetes, cloud, migrations, volumes, backups, and configuration.
- `testing.md` and `uat-checklist.md`: strategy, commands, evidence, and exploratory methods.
- `ai-usage.md`: prompts/skills used, methodology, tasks delegated to agents, outputs accepted/rejected, human verification, and the implementation plan. Record actual prompts and outcomes as work happens; do not fabricate history or present private chain-of-thought.

## 13. Execution phases and exit criteria

### Phase 0 - Contract and skeleton

- Confirm the decisions in this plan; create repository structure, ADRs, GraphQL SDL draft, error contract, database ERD, configuration schema, and basic design tokens/wireframes.
- Add Compose dependencies, health endpoints, migration runner, CI skeleton, and seed strategy.
- **Exit:** all services boot; contracts compile; every requirement is mapped to an owner and test.

### Phase 1 - Identity and data foundation

- Implement migrations, repositories, seed users, login/logout/current-user, RBAC middleware, CSRF, and centralized policy interfaces.
- **Exit:** authenticated session works end-to-end; private resources are denied by default; migrations round-trip in a clean database.

### Phase 2 - Upload, deduplication, quota, and deletion

- Implement streaming upload, MIME validation, storage abstraction, race-safe deduplication, multi-upload response, logical quota, statistics, reference-safe deletion, cleanup recovery, and audit events.
- **Exit:** concurrency and failure-path integration tests pass; frontend can upload with progress and show accurate stats.

### Phase 3 - Management, search, sharing, and download

- Implement file list/details/tags/folders, combined filters, pagination/indexes, private/public/user shares, downloads, previews, and counters.
- **Exit:** permission matrix and search tests pass; public share works anonymously; query plans are documented.

### Phase 4 - Product UI and admin

- Complete responsive vault UX, upload queue, detail/preview/share flows, polished errors/empty states, admin all-files view, charts, analytics, and accessibility pass.
- **Exit:** all core user journeys are usable on phone and desktop; no placeholder UI or raw API errors remain.

### Phase 5 - Hardening and evidence

- Complete unit/integration/E2E suites, race checks, security review, performance seed/query-plan checks, structured logs, recovery checks, and UAT documentation.
- **Exit:** CI is green from a clean clone; mandatory acceptance matrix is fully checked with automated evidence where possible.

### Phase 6 - Deployment and submission polish

- Finish Docker images, Kubernetes manifests, cloud deployment, screenshots/GIF, architecture/security/deployment docs, AI usage log, and demo script.
- **Exit:** a reviewer can clone and run locally with one command; public demo is healthy; README accurately reflects reality.

### Phase 7 - Optional stretch work

- GraphQL subscriptions or server-sent events for live download counts.
- Helm chart and staging deployment automation.
- Malware scanning adapter, resumable uploads, or expiring download links.
- Only start after the core definition of done remains green.

## 14. Sub-agent execution plan

The main agent should own architecture, shared contracts, integration, and final quality. Parallel agents should work only after Phase 0 freezes the initial schema/API boundaries.

### Wave 1 - foundations (limited parallelism)

**Agent A: database and query layer**

- Own migrations, SQL queries, indexes, generated query code, seed volume data, and query-plan evidence.
- Deliver tests for constraints, search combinations, pagination, and statistics queries.
- Must not change GraphQL names without an integration proposal.

**Agent B: frontend foundation and UX system**

- Own routing, layout, auth state, design tokens/components, responsive shell, error/loading patterns, and mocked screen states.
- Work against frozen generated GraphQL types/mocks.

**Agent C: infrastructure foundation**

- Own Dockerfiles, Compose, health checks, configuration examples, Make targets, and CI skeleton.
- Must keep one-command local setup as its acceptance test.

Main integrator owns backend module skeleton, auth policy, GraphQL/REST contracts, and merge order.

### Wave 2 - feature slices (parallel after contracts)

**Agent D: upload/deduplication/storage**

- Streaming multipart handler, hash/MIME checks, blob store, quota, reference counts, garbage collection, recovery, and concurrency tests.

**Agent E: file management/search**

- File/folder/tag services, GraphQL resolvers, combined filters, cursor pagination, and details/delete metadata flows.

**Agent F: sharing/download/security**

- Central permission matrix, public/direct shares, token security, preview/download handlers, counters, and audit events.

**Agent G: vault frontend**

- Upload queue/progress, files/folders, search/filter UI, details, previews, stats, sharing, and delete experience.

Each agent must return: changed files, assumptions, commands run, test results, unresolved risks, and any contract change request. No agent should “fix” another domain silently.

### Wave 3 - admin, QA, and delivery

**Agent H: admin and analytics**

- Admin API queries, authorization tests, dashboard UI, graphs, uploader details, and download/storage analytics.

**Agent I: QA/UAT**

- Requirement traceability matrix, Playwright journeys, accessibility checks, UAT checklist, edge-case exploration, and reproducible defect reports. This agent should primarily test, not rewrite feature implementations.

**Agent J: deployment and documentation**

- Kubernetes, optional cloud configuration, deployment docs, diagrams, API docs, README structure, and AI-use log consolidation. It must verify commands rather than copy assumptions.

Main integrator resolves cross-domain changes, runs the full stack, reviews security, performs visual QA, verifies documentation from a clean setup, and makes the final submission commit.

## 15. Requirement traceability and acceptance matrix

| Requirement | Demonstration | Automated evidence |
|---|---|---|
| SHA-256 deduplication/references | Upload identical content under different names/users; show one physical blob and correct stats | Concurrent integration tests and DB assertions |
| Single/multi drag-and-drop upload | Upload queue with progress and mixed outcomes | Playwright upload flow |
| MIME mismatch rejection | Rename an image as a document and upload | Backend integration + UI error test |
| File metadata/list/details | Files view and detail drawer | GraphQL and component tests |
| Strict owner deletion | Attempt as recipient/admin/other user, then owner | Permission integration + Playwright |
| Reference-safe cleanup | Delete first then final reference | Storage lifecycle integration test |
| Combined search filters | Apply all supported filter types together | Query integration tests + query-plan evidence |
| 2 calls/sec configurable limit | Burst requests and inspect `429`/`Retry-After` | Rate-limit integration test |
| 10 MB configurable quota | Upload at boundary and over limit | Quota unit/integration tests |
| Storage savings | Dashboard cards after duplicate uploads | Stats query and E2E assertions |
| Admin panel | List all users' files and analytics | RBAC API tests + admin E2E |
| Docker Compose | Fresh `docker compose up --build` | CI smoke job |
| Documentation/AI methodology | Follow README from clean clone and inspect docs | Link/command checks where practical |
| Kubernetes bonus | Apply/validate manifests and inspect probes/PVCs | Schema validation in CI |

## 16. Five-minute recruiter demo

1. Start with architecture and the GraphQL/REST decision in under 30 seconds.
2. Sign in as Alice and drag/drop multiple files, including one with a forged extension. Show progress, partial success, and safe MIME rejection.
3. Upload the same content again under a different name. Show dedup status and storage savings changing.
4. Combine search filters, open details/preview, add a tag, and move the file into a folder.
5. Create a public link, open it anonymously, download, and show the counter update.
6. Sign in as Bob and demonstrate explicit sharing plus inability to delete Alice's file.
7. Delete duplicate references in order and explain reference-safe garbage collection.
8. Open the admin dashboard and show uploader details and analytics.
9. End on CI, automated UAT, Kubernetes manifests, and the architecture/security docs.

Keep seeded demo data deterministic, maintain a scripted reset command, and record a short GIF/video as a fallback if the live demo environment fails.

## 17. Risks and controls

| Risk | Control |
|---|---|
| Bonus work crowds out correctness | Phase gates; no stretch work before all mandatory acceptance tests pass. |
| Dedup race creates duplicate/corrupt state | Unique hash constraint, transactions, row locking, pending states, and concurrency tests. |
| Database commit and blob write diverge | Staged upload, recoverable states, idempotent cleanup/reconciliation. |
| Authorization differs across APIs | One centralized policy used by GraphQL, REST, preview, share, and admin paths. |
| Search is correct but slow | Access-first SQL, proper indexes, cursor pagination, seeded benchmarks, query-plan evidence. |
| In-memory limiter fails with replicas | Redis-backed limiter and explicit dependency health checks. |
| Public preview becomes an XSS vector | Safe type allowlist, sandboxing/headers, attachment fallback, and no sniffing. |
| Stats are ambiguous | Fix definitions in Section 2 and expose the same calculations in UI/API/docs/tests. |
| Agents create incompatible changes | Freeze contracts, assign directory ownership, require change proposals, integrate in waves. |
| README overclaims production readiness | Separate implemented controls, demonstrated evidence, and future production work. |

## 18. Final definition of done

- Every mandatory requirement has a checked acceptance row and no known critical defect.
- `docker compose up --build` works from a clean clone and preserves database/blob data correctly.
- Core tests, linters, type checks, builds, and selected Playwright UAT pass in CI.
- Uploads stream rather than buffer, MIME is content-validated, quotas and rate limits are configurable, and all file access is authorized.
- Deduplication and deletion remain correct under concurrent requests and failures.
- UI is responsive, accessible, polished, and handles loading, empty, error, and partial-success states.
- GraphQL SDL, migrations, architecture, security, testing, deployment, UAT, and AI methodology are documented.
- Kubernetes files validate; any cloud URL in the README is live and reproducible.
- Demo accounts/data, reset instructions, screenshots, and the five-minute walkthrough are ready.
- A final reviewer follows the README exactly from a clean environment before submission.

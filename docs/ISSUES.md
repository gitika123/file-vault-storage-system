# Implementation Issues

This is the working issue ledger for the BalkanID file vault. Issues are intentionally ordered by dependency. Each issue must have an acceptance section and verification evidence before it is closed.

## Status legend

- `OPEN`: ready to begin.
- `IN PROGRESS`: actively being implemented.
- `BLOCKED`: cannot proceed without an external decision or dependency.
- `CLOSED`: acceptance checks passed and evidence is recorded.

## Issue list

| ID | Issue | Priority | Depends on | Status |
|---|---|---:|---|---|
| 001 | Repository foundation and issue workflow | P0 | - | CLOSED |
| 002 | Product contracts: configuration, errors, API outline, and ADRs | P0 | 001 | CLOSED |
| 003 | PostgreSQL schema, migrations, seed data, and query conventions | P0 | 002 | CLOSED |
| 004 | Go service skeleton, health checks, configuration, and database wiring | P0 | 002, 003 | CLOSED |
| 005 | Authentication, sessions, CSRF protection, and RBAC | P0 | 003, 004 | CLOSED |
| 006 | Blob storage abstraction and streaming upload limits | P0 | 004, 005 | CLOSED |
| 007 | MIME validation and SHA-256 deduplication | P0 | 003, 006 | CLOSED |
| 008 | Quotas, storage statistics, reference counts, and safe deletion | P0 | 003, 007 | CLOSED |
| 009 | File management, folders, tags, and cursor pagination | P0 | 005, 008 | CLOSED |
| 010 | Search and combined filtering with performance evidence | P0 | 003, 009 | CLOSED |
| 011 | Private/public/direct sharing and authorization policy | P0 | 005, 009 | CLOSED |
| 012 | Secure downloads, previews, counters, and audit events | P0 | 011, 008 | CLOSED |
| 013 | React application shell, authentication, and design system | P0 | 002, 004 | CLOSED |
| 014 | Vault dashboard, file list, filters, and details UI | P0 | 009, 010, 013 | CLOSED |
| 015 | Drag-and-drop multi-upload UX with progress and errors | P0 | 006, 007, 013 | CLOSED |
| 016 | Sharing, public-link, preview, and delete UX | P0 | 011, 012, 014 | CLOSED |
| 017 | Admin API, RBAC-protected analytics, and admin dashboard | P0 | 005, 012, 013 | CLOSED |
| 018 | Unit, integration, race, and contract test coverage | P0 | 007-017 | CLOSED |
| 019 | Playwright UAT checklist and automated smoke journeys | P1 | 018 | CLOSED |
| 020 | Docker Compose, multi-stage images, and local one-command setup | P0 | 004, 013 | CLOSED |
| 021 | CI pipeline: lint, test, build, integration, and UAT | P1 | 018-020 | CLOSED |
| 022 | Kubernetes manifests, probes, PVCs, and deployment documentation | P1 | 020, 021 | CLOSED |
| 023 | README, architecture/security/API docs, screenshots, and AI methodology | P0 | 018-022 | CLOSED |
| 024 | Optional cloud deployment and public demo verification | P1 | 023 | BLOCKED |
| 025 | Final clean-clone review and recruiter demo rehearsal | P0 | 023, 024 | CLOSED |
| 026 | Complete file details and public statistics UX | P0 | 012, 014, 016 | CLOSED |
| 027 | Complete advanced search and filtering | P0 | 010, 026 | CLOSED |
| 028 | Admin panel and cross-user analytics | P0 | 017, 026, 027 | CLOSED |
| 029 | Professional documentation and deliverables index | P0 | 023, 028 | CLOSED |
| 030 | Real-time download-count updates | P1 | 012, 026, 028 | CLOSED |
| 031 | Helm chart | P1 | 022 | CLOSED |
| 032 | Quality hardening and scale evidence | P0 | 027, 028 | CLOSED |
| 033 | Upload MIME and drag-and-drop regression | P0 | 007, 015, 032 | CLOSED |
| 034 | Core smoke coverage and public-download regression | P0 | 012, 019, 020, 033 | CLOSED |
| 035 | Drive-style folder, sharing, and navigation workflows | P0 | 009, 011, 014, 016, 034 | CLOSED |
| 036 | Inline file preview workflow | P1 | 012, 014, 035 | CLOSED |
| 037 | Folder lifecycle, public preview, filters, statistics, and admin acceptance | P0 | 009, 011, 012, 027, 028, 036 | CLOSED |
| 038 | Drive-style folder user sharing and inherited access | P0 | 035, 037 | CLOSED |
| 039 | Folder-only organization and selectable MIME filter options | P0 | 014, 027, 035 | CLOSED |
| 040 | Admin statistics completeness | P0 | 028, 039 | CLOSED |
| 041 | Public media preview and share-recipient guidance | P0 | 011, 012, 035, 040 | CLOSED |
| 042 | Visual design quality pass | P1 | 039, 040, 041 | CLOSED |
| 043 | Multi-file selection and deletion | P1 | 008, 014, 042 | CLOSED |
| 044 | Live admin statistics visualizations | P1 | 028, 030, 040, 042 | CLOSED |
| 045 | Dedicated admin Statistics & Oversight section | P0 | 040, 044 | CLOSED |
| 046 | Consistent Back to home navigation | P1 | 035, 045 | CLOSED |
| 047 | Separate admin files and statistics tabs | P0 | 045, 046 | CLOSED |
| 048 | Storage statistics for all scopes | P0 | 008, 044, 047 | CLOSED |
| 049 | Production readiness audit corrections | P0 | 023, 042, 048 | CLOSED |
| 050 | Global drag-and-drop and folder context-menu UX | P0 | 015, 035, 049 | CLOSED |

## Closing policy

An issue is closed only when its implementation is present in the intended folder, its acceptance checks pass, and its issue file records the commands/results used for verification. If a later issue discovers a defect in a closed issue, reopen the original issue and link the regression.

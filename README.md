# BalkanID Secure File Vault

A production-minded secure file vault built for the BalkanID Full Stack Engineering hiring task. It provides authenticated upload, content-addressed deduplication, owner-scoped management, controlled sharing, search, quotas, storage analytics, and administrator oversight.

## Current status

All required core features are implemented and locally verified. Bonus work includes Docker Compose, CI/UAT automation, previews, RBAC, audit events, upload progress state, live admin visualizations, and Helm packaging. Render deployment remains an external release step requiring provider resources and production secrets; the process-local rate limiter and local blob store should be replaced with shared production services before running multiple API replicas. See the evidence and status in [`docs/DELIVERABLES.md`](docs/DELIVERABLES.md) and [`docs/ISSUES.md`](docs/ISSUES.md).

Start with the [consolidated engineering documentation](docs/PROJECT_DOCUMENTATION.md), which brings together setup, architecture, schema, API behavior, requirements coverage, verification, deployment, and AI methodology. The [documentation hub](docs/README.md), [deliverables matrix](docs/DELIVERABLES.md), and [living progress report](docs/progress-report.tex) remain available as supporting records.

## Runtime architecture

- Go REST API; GraphQL SDL is retained as a documented alternative contract because the brief accepts REST and marks GraphQL as preferred.
- PostgreSQL for relational metadata, sessions, sharing, audit events, and transactional blob references.
- Local content-addressed blob storage for development and Docker; production should use a persistent disk or object storage adapter.
- React, TypeScript, and Vite frontend served by Nginx.
- Configurable per-user in-memory limiter; Redis is included in Compose and distributed limiter hardening is tracked separately.

## Quick start

1. Copy `.env.example` to `.env` and set strong local values.
2. Start Docker Desktop.
3. Run:

```powershell
docker compose up --build -d
```

4. Open <http://localhost:5173>.

The API container runs the idempotent demo seed command at startup. See [`docs/seed-data.md`](docs/seed-data.md) for local accounts. Stop services with `docker compose down`; use `-v` only when intentionally discarding local volumes.

## Native verification

```powershell
.\scripts\verify.ps1
```

This runs backend tests and the frontend production build. CI additionally runs formatting checks and Playwright UAT.

## Repository map

```text
backend/       Go API, domain services, auth, storage, and seed command
frontend/      React/TypeScript application and Playwright tests
docs/          Documentation hub, contracts, ADRs, issues, UAT, and report
migrations/    Forward-only PostgreSQL migrations
k8s/           Kubernetes manifests
scripts/       Local verification helpers
tests/         Shared upload and API fixtures
```

## Engineering workflow

Use the issue-driven workflow in [`docs/ISSUES.md`](docs/ISSUES.md): define acceptance criteria, implement in the intended module, verify with automated/live checks, record evidence, and then close the issue.

# File Vault

Secure multi-user file storage with folders, sharing, deduplication, search, quotas, and administrator oversight.

## Features

- User registration, sign-in, sessions, CSRF protection, and role-based access.
- Single and multiple uploads, drag-and-drop, MIME validation, and upload limits.
- SHA-256 content deduplication with reference counting and storage statistics.
- File metadata, previews, downloads, tags, folders, rename, move, and deletion rules.
- Private files and folders, public links, direct user sharing, permissions, and revocation.
- Filename search and combined MIME, size, date, tag, and uploader filters.
- Per-user API rate limiting and storage quotas with structured error responses.
- User storage analytics and a protected administrator inventory and statistics view.

## Stack

- Go REST API
- React, TypeScript, and Vite frontend
- PostgreSQL for metadata and relationships
- Redis for shared rate-limit state
- Local or S3-compatible blob storage
- Docker Compose, Kubernetes manifests, Helm chart, and GitHub Actions

## Run locally

Requirements: Docker Desktop with the engine running.

```powershell
Copy-Item .env.example .env
docker compose up --build -d
```

Open <http://localhost:5173>.

The API applies database migrations automatically on startup. Optional local seed data can be enabled with the seed variables documented in [docs/seed-data.md](docs/seed-data.md).

Stop the stack:

```powershell
docker compose down
```

Use `docker compose down -v` only when you intend to remove local database and file-storage volumes.

## Verify

```powershell
.\scripts\verify.ps1
.\scripts\core-smoke.ps1
```

The CI workflow also runs `go vet`, Go tests, frontend linting, TypeScript checks, the production build, Playwright UAT, and Docker build validation.

## Repository layout

```text
backend/       Go API and domain services
frontend/      React application and browser tests
migrations/    PostgreSQL schema migrations
tests/         Upload fixtures
scripts/       Local verification scripts
docs/          Hiring-task documentation and technical contracts
k8s/           Kubernetes manifests
helm/          Helm chart
```

## Documentation

The complete hiring-task document is [docs/hiring-task-documentation.tex](docs/hiring-task-documentation.tex). It covers the product, technology, core features, deliverables, infrastructure, bonus work, verification, and AI-assisted methodology.

Supporting API, schema, storage, deployment, testing, and configuration references are under [docs/](docs/).

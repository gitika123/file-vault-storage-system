# Deployment and Operations Guide

## Docker Compose

Compose starts PostgreSQL, Redis, the Go API, and the Nginx frontend. Migrations run during first database initialization and the API seeds demo users idempotently from environment variables.

```powershell
docker compose config
docker compose up --build -d
docker compose ps
docker compose logs --tail=100 api
```

The API readiness endpoint is `/health/ready`; the frontend is available at `http://localhost:5173`.

## Kubernetes

The `k8s/` directory contains the namespace, API and frontend Deployments/Services, PostgreSQL StatefulSet/Service, Ingress, probes, ConfigMap, Secret template, and PVCs. The API supports S3-compatible shared object storage and Redis-backed distributed rate limiting when configured. Use managed PostgreSQL/Redis/object storage for a real multi-replica production environment; the included PostgreSQL StatefulSet and ReadWriteOnce PVCs are intended for evaluation or a single-node cluster.

## Helm

The `helm/balkanid` chart templates the API and frontend Deployments/Services, PostgreSQL StatefulSet/Service, Ingress, ConfigMap, probes, service account, and PVCs. It expects an existing secret containing database, Redis, PostgreSQL, S3, session, and CSRF settings.

```powershell
helm lint helm/balkanid
helm template balkanid helm/balkanid --namespace balkanid
helm upgrade --install balkanid helm/balkanid --namespace balkanid --create-namespace
```

## Render plan

The recommended Render split is a Static Site for the frontend, a Docker Web Service for the API, Render Postgres, Render Key Value, and an S3-compatible object-storage bucket. Configure URLs, strong secrets, origins, and seed values through provider environment settings only.

## Operational safeguards

- Keep `.env` and Kubernetes secret values outside version control.
- Use managed Postgres/Key Value and object storage for multi-instance production.
- Back up relational data and test restore procedures.
- Set a provider spending limit before enabling paid resources.
## Kubernetes and CI/CD

The repository includes complete raw manifests under `k8s/` and a Helm chart under `helm/balkanid/` for the API, frontend, PostgreSQL, Ingress, ConfigMap, Secret reference, and PVCs. Kubernetes production deployments should provide `balkanid-secrets` from an external secret manager with `DATABASE_URL`, `REDIS_URL`, `SESSION_SECRET`, `CSRF_SECRET`, `POSTGRES_PASSWORD`, `BLOB_STORE_BUCKET`, `BLOB_STORE_ACCESS_KEY`, and `BLOB_STORE_SECRET_KEY`.

The GitHub Actions workflow runs Go formatting, `go vet`, backend tests, frontend ESLint, frontend builds, Playwright UAT, and Linux Docker builds. Pushes to `main` publish images to GHCR using the repository `GITHUB_TOKEN`. A manual workflow dispatch from `main` can deploy to a Kubernetes cluster when `KUBE_CONFIG_B64` is configured as a repository secret.

For multi-replica operation, configure `REDIS_URL` so the API uses the shared Redis rate limiter and configure `BLOB_STORE_DRIVER=s3` with an S3-compatible object store. Local filesystem storage and the in-memory limiter remain intended only for local evaluation.
## Kubernetes and CI/CD status

The repository includes complete raw manifests under `k8s/` and a full Helm chart under `helm/balkanid/` for the API, frontend, PostgreSQL, Ingress, ConfigMap, Secret reference, and PVCs. Before a real deployment, provision `balkanid-secrets` through an external secret manager with `DATABASE_URL`, `REDIS_URL`, `SESSION_SECRET`, `CSRF_SECRET`, `POSTGRES_PASSWORD`, `BLOB_STORE_BUCKET`, `BLOB_STORE_ACCESS_KEY`, and `BLOB_STORE_SECRET_KEY`.

GitHub Actions runs Go formatting, `go vet`, backend tests, frontend ESLint, frontend builds, Playwright UAT, and Linux Docker builds. Pushes to `main` publish images to GHCR with `GITHUB_TOKEN`. A manual workflow dispatch from `main` deploys the manifests when `KUBE_CONFIG_B64` is configured as a repository secret.

For multi-replica production, configure `REDIS_URL` for the shared rate limiter and `BLOB_STORE_DRIVER=s3` for shared S3-compatible object storage. Local filesystem storage and the in-memory limiter remain available for local evaluation only.

# Hiring Task Deliverables Matrix

| Deliverable | Level | Status | Evidence |
|---|---|---|---|
| Go backend | Required | Complete | `backend/`, tests, health endpoints |
| REST or GraphQL API | Required; GraphQL preferred | Complete | REST runtime; GraphQL SDL alternative contract |
| PostgreSQL schema and migrations | Required | Complete | `migrations/`, `docs/database.md` |
| React/TypeScript frontend | Required | Complete | `frontend/` production build |
| Deduplication, uploads, MIME validation | Required | Complete | SHA-256 blobs, multipart upload, content validation |
| File details and sharing | Required | Complete | Owner detail API/UI and public/direct file sharing |
| Folder organization | Optional/bonus | Complete | Folder APIs and file moves |
| Search and combined filtering | Required | Complete | Filename, MIME, size, date, tags, uploader filters |
| Rate limits and quotas | Required | Complete with scale caveat | Configurable local limiter and quota errors |
| Storage statistics | Required | Complete | User/admin statistics APIs and UI |
| Admin panel | Required; graphs bonus | Complete | RBAC inventory, uploader details, admin upload/direct sharing, download counts, usage cards, and live visualizations |
| Docker Compose | Required | Complete | Multi-stage images and live four-service stack |
| Kubernetes | Bonus | Complete with operational caveat | API, frontend, PostgreSQL, Services, Ingress, ConfigMap/Secret templates, probes, and PVCs in `k8s/`; managed PostgreSQL/Redis and shared object storage remain preferred for multi-replica production |
| CI/CD | Bonus | Complete | GitHub Actions tests, builds, and UAT |
| Cloud deployment | Bonus | Pending external step | Render deployment plan is documented; it requires provider resources, secrets, and a public-domain decision |
| UAT automation | Bonus | Complete | Checklist and Playwright smoke tests |
| Real-time updates | Bonus | Complete | SSE download events with owner/admin authorization and frontend refresh fallback; Issue 030 |
| Helm chart | Bonus | Complete | Configurable API/PVC chart; Helm lint and template validation passed in Helm 3.16 container |
| Query-plan evidence | Bonus-quality | Complete | `docs/performance/search-plan.md` and indexed owner/date query |
| AI methodology/prompts | Required when AI is used | Complete after Issue 029 | `docs/ai-prompts.md` |

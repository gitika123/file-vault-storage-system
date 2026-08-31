# Issue 049 - Production readiness audit corrections

**Priority:** P0  
**Status:** CLOSED

## Acceptance criteria

- [x] Remove local evaluation language from the end-user login experience.
- [x] Keep per-user storage visualizations visible in the React vault experience.
- [x] Make folder cards use the requested context-menu workflow instead of persistent inline actions.
- [x] Ensure preview requests do not increment download counters; only actual downloads do.
- [x] Complete the Kubernetes deliverable set with frontend, ingress, configuration, secret templates, and PostgreSQL resources.
- [x] Document the remaining external cloud-deployment and multi-replica storage prerequisites accurately.

## Verification

- `go test ./...` passed with a project-local Go cache.
- `npm run build` passed strict TypeScript validation and Vite production compilation.
- `rg` confirmed evaluation-only login copy is absent from `frontend/src`.
- The hiring-task PDF was reviewed page by page and the deliverables matrix was reconciled against the repository.

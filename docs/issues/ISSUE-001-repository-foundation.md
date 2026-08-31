# Issue 001 - Repository foundation and issue workflow

**Status:** CLOSED  
**Priority:** P0  
**Depends on:** None  
**Owner:** Main integrator

## Objective

Create a clean implementation root with predictable folders, preserve the approved implementation plan, and establish an issue-by-issue workflow that can be verified by a reviewer or sub-agent.

## Scope

- Create the implementation root `balkanid-file-vault/`.
- Add dedicated backend, frontend, documentation, migrations, Kubernetes, scripts, workflows, and test directories.
- Copy the approved plan into `docs/IMPLEMENTATION_PLAN.md`.
- Create the issue ledger and this issue record.
- Add a root README explaining project status, structure, and workflow.
- Do not implement application behavior in this issue.

## Acceptance criteria

- [x] All implementation files are contained under `balkanid-file-vault/`.
- [x] The directory layout matches the documented repository map.
- [x] The complete implementation plan is available under `docs/`.
- [x] Issues are dependency-ordered and have explicit statuses.
- [x] The next issue is unambiguous: Issue 002.
- [x] The implementation root is initialized as its own Git repository.
- [x] An expected-tree verification is recorded after initialization.

## Verification evidence

Initial structure was created from the workspace root. Git was initialized inside the implementation root and `git status --short --branch` confirmed the expected new project files on the repository's initial branch. A recursive file inspection confirmed that all implementation folders are nested under `balkanid-file-vault/`. The repository is intentionally uncommitted at this stage so the initial implementation can be reviewed as one coherent change.

## Files created

- `README.md`
- `docs/IMPLEMENTATION_PLAN.md`
- `docs/ISSUES.md`
- `docs/issues/ISSUE-001-repository-foundation.md`

## Closure note

Issue 001 is closed. Issue 002 is now the active issue and owns the shared product contracts, configuration, error model, API outline, and architecture decision records.

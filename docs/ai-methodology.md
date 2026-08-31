# AI-assisted implementation methodology

The project was developed issue-by-issue from the hiring-task brief. Each issue has an objective, dependencies, acceptance criteria, implementation evidence, and explicit blockers. Design decisions are recorded in ADRs; security-sensitive choices are centralized in policy, session, storage, upload, sharing, and content services.

AI was used for requirement decomposition, scaffolding proposals, test-case generation, code review, security review, and documentation drafting. The workflow kept human control over scope and acceptance decisions. Runtime behavior was checked through Go tests, `go vet`, race testing, PostgreSQL queries, live HTTP calls, Docker Compose health checks, frontend builds, and Playwright UAT. Secrets and demo passwords remain local-only and are never committed.

Reusable prompt patterns and the evidence standard are documented in [`ai-prompts.md`](ai-prompts.md).

# BalkanID File Vault Documentation

Documentation hub for the BalkanID secure file-vault implementation.

## Primary project document

Read the [consolidated engineering documentation](PROJECT_DOCUMENTATION.md) for the complete reviewer-facing guide: setup, architecture, database schema, API contract, feature coverage, testing, deployment, production notes, and AI-assisted methodology.

## Product and architecture

- [Implementation plan](IMPLEMENTATION_PLAN.md) - scope, decisions, milestones, and risks.
- [Database overview](database.md) - ownership, deduplication, deletion, statistics, and indexes.
- [Storage design](storage.md) - blob lifecycle and local storage behavior.
- [Architecture decisions](adr/0001-modular-monolith.md), [API boundary](adr/0002-graphql-and-rest-boundary.md), and [quota semantics](adr/0003-quota-and-statistics-semantics.md).

## API and configuration

- [REST endpoint contract](contracts/rest-endpoints.md) - implemented HTTP routes.
- [GraphQL SDL](contracts/graphql-schema.graphql) - preferred-contract draft; no GraphQL runtime is currently shipped.
- [Error catalog](contracts/error-catalog.md) and [configuration reference](contracts/configuration.md).

## Operations and delivery

- [Deliverables matrix](DELIVERABLES.md) - requirements mapped to evidence and status.
- [Deployment guide](deployment.md) - Docker, Kubernetes, Helm, and Render plan.
- [Search performance evidence](performance/search-plan.md) - reproducible PostgreSQL query-plan baseline.
- [Seed data](seed-data.md) and [UAT checklist](uat-checklist.md).
- [AI methodology and prompts](ai-prompts.md).
- [LaTeX hiring-task documentation](hiring-task-documentation.tex) - ordered product, core features, deliverables, infrastructure, bonus work, and screenshot placeholders.
- [Living progress report](progress-report.tex).

## Change control

Every material change should link to an issue in [ISSUES.md](ISSUES.md), define acceptance criteria, and record reproducible verification evidence before closure.

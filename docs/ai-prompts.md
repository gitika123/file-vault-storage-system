# AI Prompts and Responsible Usage Guide

This document records reusable prompt patterns used for the project. It summarizes the method without reproducing private conversation history or credential-bearing instructions.

## Working principles

1. Give the model the product requirement, repository constraints, and one issue at a time.
2. Request small, reviewable changes with explicit acceptance criteria.
3. Require tests, security implications, and documentation updates with every change.
4. Verify claims through commands, database queries, live HTTP checks, and UI/UAT.
5. Never include secrets, tokens, or personal credentials in prompts, source, or logs.
6. Treat model output as a proposal until human review and reproducible checks pass.

## Reusable prompt patterns

### Requirement decomposition

> Review the product brief and convert it into dependency-ordered issues. Classify each item as required, optional, bonus, or blocked. Add acceptance criteria and a reproducible verification plan.

### Secure backend implementation

> Implement the current issue in the existing Go module. Preserve authorization boundaries, stable errors, parameterized SQL, transaction semantics, and configuration conventions. Add focused tests for success, failure, authorization, boundary, and concurrency behavior. Do not change unrelated modules.

### Database review

> Review this schema/query for ownership leaks, race conditions, reference-count correctness, rollback behavior, and index usage. Propose the smallest forward-only migration and include query-plan or benchmark evidence.

### Frontend feature

> Implement this journey in the existing React/TypeScript design system. Cover loading, empty, error, success, keyboard, responsive, and authorization states. Keep API calls typed and add a Playwright journey for the highest-value behavior.

### Security review

> Act as a security reviewer. Trace authentication, CSRF, authorization, file-path handling, MIME validation, token storage, headers, rate limiting, and error disclosure. Report severity, evidence, exploit scenario, and minimal remediation.

### Final verification

> Review the issue against its acceptance criteria. Run focused tests, then the full verification suite. Report exact commands, results, caveats, and documentation claims that exceed implementation evidence.

## Evidence standard

AI-assisted work is complete only when implementation, tests, runtime behavior, and documentation agree. A compile is not proof of authorization correctness, a plan is not proof of deployment, and a generated screenshot is not proof of an interaction without a UI test.

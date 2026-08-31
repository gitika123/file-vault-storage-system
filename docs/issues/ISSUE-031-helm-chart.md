# Issue 031 - Helm chart

**Status:** CLOSED  
**Priority:** P1 bonus  
**Depends on:** 022  
**Owner:** Main integrator

## Scope

- Package the Kubernetes deployment into a configurable Helm chart.
- Template namespace, API deployment/service, blob PVC, secrets references, probes, image, replica count, and storage size.
- Document install, upgrade, rollback, and secret requirements.

## Current evidence

- Added configurable chart templates for the API Deployment/Service, service account, probes, blob PVC, replica count, image, storage size, and existing-secret references.
- Added install, upgrade, rollback, and secret guidance to `docs/deployment.md`.
- `helm lint` passed with zero failures using Helm 3.16.4, and `helm template` rendered successfully in an ephemeral validation container.

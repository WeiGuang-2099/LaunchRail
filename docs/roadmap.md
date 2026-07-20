# LaunchRail Roadmap

The roadmap distinguishes code that exists from intended platform expansion. Milestones are completed as working vertical slices rather than infrastructure demonstrations.

## Implemented

### Service catalog data foundation

- Go control-plane skeleton, configuration, structured logging, request IDs, and graceful shutdown.
- Service domain behavior and validation.
- PostgreSQL schema migrations and repository implementation.
- Repository and migration integration test foundations.

### Full-stack service catalog

- Register, list, inspect, update, and archive REST endpoints.
- Stable JSON API errors and OpenAPI 3.1 contract.
- Database-backed readiness.
- React service catalog with full lifecycle routes and user states.
- Full-stack Docker Compose delivery with a dedicated migration job.
- Isolated PostgreSQL test profile and frontend test toolchain.

## In progress

### Deterministic engineering baseline

- Seven least-privilege CI jobs for backend, database integration, frontend, contracts, Compose, E2E, and security.
- Dockerized Playwright browser testing with an idempotent catalog seed and direct exit-code propagation.
- Pinned Redocly and Trivy execution without host installations or an unnecessary Java vulnerability database.
- Source, secret, Dockerfile, and runtime-image scanning with CycloneDX SBOM artifacts.
- Non-root backend and frontend runtime images with health checks.
- Dependabot coverage for Go modules, npm, Actions, Dockerfiles, and Compose.

The repository implementation is present, but this milestone remains **In progress**. It moves to **Implemented** only after a real pull request passes all seven checks, the `main` ruleset is active, the Security job has no unhandled High or Critical findings, and all three SBOM artifacts can be downloaded and parsed.

## Designed, not implemented

### Identity security baseline

- Generic OIDC Authorization Code Flow through a Go backend for frontend.
- PostgreSQL-backed opaque sessions with explicit CSRF controls.
- Controlled team directory, team-scoped service permissions, and administrator-only ownership transfer.
- Transactional append-only audit requirements.
- A DFD and STRIDE threat model covering the planned identity boundary.

The decisions are recorded in [ADR 0003](adr/0003-oidc-bff-sessions.md), [ADR 0004](adr/0004-team-scoped-authorization-and-audit.md), and the [identity threat model](security/threat-model.md). No authentication or authorization runtime behavior has been implemented yet.

## Next

### Identity and ownership controls

- Implement generic OIDC sign-in with Keycloak as the local reference provider.
- Add backend-managed sessions, logout, CSRF protection, and current-user state.
- Enforce the accepted platform-admin and team-owner authorization matrix.
- Resolve owner teams from the controlled directory and restrict ownership transfer.
- Write catalog mutations and audit events in one transaction.

The service catalog is the accepted product baseline. Identity implementation follows completion of the engineering baseline and remains the next product milestone.

## Later

1. GitHub repository discovery and webhook-driven synchronization.
2. Kubernetes workload and environment inventory on kind or k3d.
3. GitOps application delivery through Argo CD.
4. Progressive delivery through Argo Rollouts.
5. OpenTelemetry instrumentation and Prometheus, Grafana, Tempo, and Loki deployment.
6. AWS and Terraform infrastructure delivery.
7. Reliability objectives, error budgets, and automated operational evidence.

Later milestones are plans only and have no implied implementation in the current repository.

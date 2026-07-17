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

The service catalog is the accepted product baseline. The security design is complete; implementation and verification are the next milestone.

## Later

1. GitHub repository discovery and webhook-driven synchronization.
2. Kubernetes workload and environment inventory on kind or k3d.
3. GitOps application delivery through Argo CD.
4. Progressive delivery through Argo Rollouts.
5. OpenTelemetry instrumentation and Prometheus, Grafana, Tempo, and Loki deployment.
6. AWS and Terraform infrastructure delivery.
7. Reliability objectives, error budgets, and automated operational evidence.

Later milestones are plans only and have no implied implementation in the current repository.

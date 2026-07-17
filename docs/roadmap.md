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

## Next

### Identity and ownership controls

- Authenticate users through an external identity provider.
- Define platform-admin and service-owner permissions.
- Resolve owner teams from a controlled team directory.
- Add audit events for catalog mutations.

This milestone should be planned only after the current service catalog is accepted as the stable product baseline.

## Later

1. GitHub repository discovery and webhook-driven synchronization.
2. Kubernetes workload and environment inventory on kind or k3d.
3. GitOps application delivery through Argo CD.
4. Progressive delivery through Argo Rollouts.
5. OpenTelemetry instrumentation and Prometheus, Grafana, Tempo, and Loki deployment.
6. AWS and Terraform infrastructure delivery.
7. Reliability objectives, error budgets, and automated operational evidence.

Later milestones are plans only and have no implied implementation in the current repository.

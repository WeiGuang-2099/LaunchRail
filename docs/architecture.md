# LaunchRail Architecture

## Current system

LaunchRail is a modular monolith with a React client and PostgreSQL persistence. The control plane is one Go deployable with domain, application, transport, and infrastructure boundaries. This keeps the first production-style vertical slice reviewable while leaving external integrations behind narrow interfaces.

```mermaid
flowchart LR
    User["Platform user"] --> Browser["React service catalog"]
    Browser -->|"REST JSON /api/v1"| Nginx["Nginx frontend gateway"]
    Nginx --> API["Go control plane"]
    API --> Domain["Catalog application and domain"]
    Domain --> Repository["ServiceRepository interface"]
    Repository --> PostgresAdapter["PostgreSQL adapter"]
    PostgresAdapter --> DB[("PostgreSQL")]
    Migrator["Migration job"] --> DB
    API -->|"Ping"| DB
```

## Backend boundaries

- **Domain:** service identity, mutable metadata, lifecycle validation, timestamps, and archive rules. It does not depend on HTTP or PostgreSQL.
- **Application:** Register, Get, List, Update, and Archive use cases. It coordinates the domain, repository interface, and injected clock.
- **Transport:** versioned HTTP routes, JSON decoding, request limits, DTO mapping, and safe error responses.
- **Infrastructure:** PostgreSQL repository, migration runner, configuration, structured logging, and graceful HTTP server lifecycle.

The API process does not apply migrations. A separate migration command owns schema transitions, and Compose enforces migration completion before API startup.

## Request and failure flow

1. Nginx serves the single-page application and proxies `/api` without enabling broad cross-origin access.
2. HTTP middleware accepts or generates an `X-Request-ID`, then records structured request completion logs.
3. Handlers enforce JSON media type, a 1 MiB request limit, strict fields, path IDs, and list query bounds.
4. The application service invokes domain behavior and persists through `ServiceRepository`.
5. Known domain and repository errors become stable client error codes. Unknown errors are logged with the request ID and returned as a generic `500` response.

## Data model

The `services` table stores UUID identity, an immutable unique service key, mutable catalog metadata, lifecycle, audit timestamps, and a nullable archive timestamp. Database constraints mirror important domain invariants. Archived records remain queryable and cannot be edited.

## Engineering verification boundary

Development and CI use the same repository-owned commands and pinned container images. Host Go and Node.js execute compilation and unit tests; Redocly, Playwright with Chromium, PostgreSQL, and Trivy execute in containers.

```mermaid
flowchart LR
    Change["Source change"] --> Fast["Host Go and npm checks"]
    Change --> Contract["Pinned Redocly container"]
    Change --> Compose["Compose configuration validation"]
    Compose --> Stack["PostgreSQL, migration, API, frontend, deterministic seed"]
    Stack --> Browser["Pinned Playwright and Chromium container"]
    Change --> Security["Trivy source, secret, and Dockerfile scan"]
    Change --> Images["Backend and frontend runtime images"]
    Images --> Security
    Security --> SBOM["Three CycloneDX SBOM artifacts"]
    Fast --> Required["Seven GitHub required checks"]
    Contract --> Required
    Browser --> Required
    Security --> Required
```

The E2E profile reaches services only by Compose DNS names. Its fixed SQL seed is idempotent, so repeated runs verify the same list and detail path without relying on user data. The browser, traces, screenshots, and video remain inside the E2E container or ignored report directories.

Trivy reuses one vulnerability cache, skips the unrelated Java database, and fails on source or runtime-image High and Critical findings and on detected secrets. Dockerfile misconfiguration scanning covers the Dockerfiles currently present. Compose is syntax-validated; it is not described as Kubernetes or Terraform policy coverage.

The Go runtime image runs as `launchrail`. The Nginx runtime image runs as `nginx`, with writable PID, cache, and temporary paths granted only to that account. Both images retain health checks.

## Accepted identity design

Authentication and authorization are designed but not implemented. [ADR 0003](adr/0003-oidc-bff-sessions.md) selects generic OIDC through a Go backend for frontend with PostgreSQL-backed opaque sessions. [ADR 0004](adr/0004-team-scoped-authorization-and-audit.md) defines controlled team ownership, application-layer authorization, and transactional audit events. The [identity threat model](security/threat-model.md) records trust boundaries, required controls, validation, and residual risks.

Keycloak is the planned local reference provider, not an application dependency. In a deployed environment, Nginx remains the only browser-facing LaunchRail component and proxies both `/auth` and `/api` to the control plane.

```mermaid
sequenceDiagram
    actor User as Platform user
    participant Browser as React browser
    participant Gateway as Nginx gateway
    participant API as Go control plane
    participant IdP as OIDC provider
    participant DB as PostgreSQL

    User->>Browser: Start sign-in
    Browser->>Gateway: GET /auth/login
    Gateway->>API: Start bound OIDC transaction
    API-->>Browser: Redirect with state, nonce, and PKCE
    Browser->>IdP: Authenticate
    IdP-->>Browser: Authorization code
    Browser->>Gateway: GET /auth/callback
    Gateway->>API: Forward callback
    API->>IdP: Exchange code and validate claims
    API->>DB: Store hashed opaque session
    API-->>Browser: Set HttpOnly cookie and redirect
    Browser->>Gateway: API request with session and CSRF header
    Gateway->>API: Same-origin request
    API->>DB: Resolve session, team, service, and policy
    API->>DB: Commit mutation and audit event together
```

This flow is an accepted target architecture. The current Nginx configuration, API routes, database schema, React client, and OpenAPI document do not contain these identity capabilities.

## Planned architecture

The following are not part of the current implementation:

- Identity provider integration and role-based access control; their design is accepted above.
- GitHub repository discovery and metadata synchronization.
- Kubernetes workload and environment inventory.
- GitOps delivery with Argo CD and progressive delivery with Argo Rollouts.
- OpenTelemetry collection and Prometheus, Grafana, Tempo, and Loki backends.
- AWS and Terraform delivery layers.

These capabilities should extend the modular control plane until an independently scalable or isolated deployment boundary is demonstrated.

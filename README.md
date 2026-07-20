# LaunchRail

LaunchRail is a portfolio-grade internal developer platform built as working, testable software. The current milestone delivers a full-stack service catalog where platform users can register, inspect, update, list, and archive service records.

## Current capabilities

- Versioned REST API with consistent JSON errors and request IDs.
- Service catalog domain validation and immutable service keys.
- PostgreSQL persistence with explicit SQL migrations.
- React and TypeScript interface for the full service record lifecycle.
- Liveness and database-backed readiness checks.
- Docker Compose delivery for PostgreSQL, migrations, API, and frontend.
- Unit, HTTP, frontend, PostgreSQL integration, and containerized browser test suites.
- Reproducible contract linting, security scanning, image scanning, and CycloneDX SBOM generation.

The repository now contains the seven CI jobs intended to protect `main`: `Backend`, `Backend Integration`, `Frontend`, `Contracts`, `Compose`, `End-to-End`, and `Security`. Repository rules are an external GitHub setting and are not considered active until the instructions in [docs/github-settings.md](docs/github-settings.md) have been applied and a real pull request has passed all seven checks.

Authentication and authorization are not implemented. Their accepted design uses generic OIDC, backend-managed sessions, controlled team ownership, and append-only audit events. See [ADR 0003](docs/adr/0003-oidc-bff-sessions.md), [ADR 0004](docs/adr/0004-team-scoped-authorization-and-audit.md), and the [identity threat model](docs/security/threat-model.md). These documents define the next milestone; they do not describe current runtime behavior.

GitHub synchronization, Kubernetes delivery, GitOps, deployment orchestration, and the wider observability stack are also planned work. See [docs/roadmap.md](docs/roadmap.md).

## Quick start

Requirements:

- Docker with the Compose plugin.
- Ports `3000`, `8080`, and `5432` available on localhost, or overrides in a local `.env` file.

Start the complete stack:

```text
docker compose up --build
```

Open:

- Service catalog UI: `http://localhost:3000`
- API readiness: `http://localhost:8080/readyz`
- API specification: [api/openapi.yaml](api/openapi.yaml)

Compose runs the migration job after PostgreSQL becomes healthy and starts the API only after migrations complete. To stop the running services without deleting data:

```text
docker compose stop
```

## API summary

| Method   | Path                    | Behavior                               |
| -------- | ----------------------- | -------------------------------------- |
| `POST`   | `/api/v1/services`      | Register a service                     |
| `GET`    | `/api/v1/services`      | List active, archived, or all services |
| `GET`    | `/api/v1/services/{id}` | Read a service record                  |
| `PUT`    | `/api/v1/services/{id}` | Replace mutable metadata               |
| `DELETE` | `/api/v1/services/{id}` | Archive a service                      |
| `GET`    | `/healthz`              | Process liveness                       |
| `GET`    | `/readyz`               | Database-backed readiness              |

Service keys are permanent. Archiving is idempotent and terminal; records are retained for operational history.

The API summary and [OpenAPI specification](api/openapi.yaml) describe the currently implemented unauthenticated API. Planned identity routes and permission fields will be added to OpenAPI only when their runtime implementation is delivered.

## Development

Backend checks:

```text
cd backend
go fmt ./...
go vet ./...
go test ./...
```

Frontend checks:

```text
cd frontend
npm ci
npm run format:check
npm run lint
npm run test
npm run build
```

PostgreSQL integration tests use a dedicated Compose profile and database:

```text
docker compose --profile test up -d db-test
cd backend
set TEST_DATABASE_URL=postgres://launchrail_test:launchrail_test@localhost:5433/launchrail_test?sslmode=disable
go test -tags=integration ./...
```

PowerShell users can set the variable with `$env:TEST_DATABASE_URL=...`.

Detailed setup and engineering workflows are in [docs/development.md](docs/development.md). The system design is documented in [docs/architecture.md](docs/architecture.md).

## Full-stack verification

Run the deterministic browser smoke test with one Compose command:

```text
docker compose --profile e2e up --build --abort-on-container-exit --exit-code-from e2e e2e
```

The profile starts PostgreSQL, applies migrations, starts the API and frontend, seeds one fixed catalog service, and runs Playwright from its own Linux container. It does not install a browser on the host. The test verifies readiness, the service list, the fixed service detail, and its repository link. The Playwright exit code is the Compose exit code.

## Verification matrix

| Check                 | Evidence                                                            |
| --------------------- | ------------------------------------------------------------------- |
| `Backend`             | Go formatting, vet, unit tests, and HTTP tests                      |
| `Backend Integration` | Repository tests against PostgreSQL 17                              |
| `Frontend`            | Locked install, formatting, ESLint, Vitest, and production build    |
| `Contracts`           | OpenAPI lint with the pinned Redocly container                      |
| `Compose`             | Default, `test`, and `e2e` profile parsing                          |
| `End-to-End`          | Containerized Playwright full-stack smoke test                      |
| `Security`            | Dependencies, secrets, Dockerfiles, runtime images, and three SBOMs |

Redocly, Trivy, and the Playwright browser environment run from pinned containers. Only the Playwright npm test runner was added to the frontend lock file; no Redocly, Trivy, direct YAML-parser dependency, or host browser package was added.

## Security engineering

The current security boundary is documented in the [engineering threat model](docs/security/engineering-threat-model.md). Exact scan scope, failure thresholds, and exceptions are defined in the [scan policy](docs/security/scan-policy.md). [ADR 0005](docs/adr/0005-deterministic-engineering-verification.md) records why the verification tools are containerized and pinned.

The source tree and runtime images must have no unhandled High or Critical findings, and secret detections fail immediately. Source, backend image, and frontend image CycloneDX SBOMs are generated by the `Security` job. These controls cover the software currently present; they do not claim Terraform or Kubernetes coverage before those assets exist.

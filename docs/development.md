# Development Guide

## Toolchain

- Go 1.26.5, as declared by `backend/go.mod`.
- Node.js 20.20.2 or a compatible release supported by the locked frontend dependencies.
- npm using the committed `package-lock.json`.
- Docker with the Compose plugin.

The browser and repository-wide verification tools are containers rather than host installations:

| Tool               | Pinned image                                                                                                         | Host impact                             |
| ------------------ | -------------------------------------------------------------------------------------------------------------------- | --------------------------------------- |
| Playwright 1.61.0  | `mcr.microsoft.com/playwright:v1.61.0-noble@sha256:57b65fdc9ceabe0ef613124c7bbe2babcf9362c4d85e382fe3b03604e84b428a` | No host browser download                |
| Redocly CLI 2.27.0 | `redocly/cli:2.27.0@sha256:119913c77da63685f2e76be82470231823278a1658edd3806ea4c7c48d21b891`                         | No npm dependency                       |
| Trivy 0.72.0       | `aquasec/trivy:0.72.0@sha256:cffe3f5161a47a6823fbd23d985795b3ed72a4c806da4c4df16266c02accdd6f`                       | No host binary; persistent Docker cache |

The frontend lock file contains `@playwright/test@1.61.0` because the E2E image executes that test runner. The image supplies its matching Chromium binary and Linux system dependencies.

Do not install or modify global tools as part of normal repository development.

## Environment

`.env.example` contains safe local defaults. The API requires `LAUNCHRAIL_DATABASE_URL` when run outside Compose. Compose injects an internal database URL automatically.

Important variables:

| Variable                      | Purpose                                 | Default |
| ----------------------------- | --------------------------------------- | ------- |
| `POSTGRES_PORT`               | Development database host port          | `5432`  |
| `TEST_POSTGRES_PORT`          | Isolated integration database host port | `5433`  |
| `API_PORT`                    | API host port                           | `8080`  |
| `FRONTEND_PORT`               | UI host port                            | `3000`  |
| `LAUNCHRAIL_LOG_LEVEL`        | Structured log threshold                | `info`  |
| `LAUNCHRAIL_SHUTDOWN_TIMEOUT` | API graceful shutdown window            | `10s`   |

Never commit real credentials or connection strings for shared environments.

## Backend workflow

Start only PostgreSQL when running the API from the host:

```text
docker compose up -d db
cd backend
go run ./cmd/migrate up
go run ./cmd/api
```

Set `LAUNCHRAIL_DATABASE_URL` before the migration and API commands.

Required checks after backend changes:

```text
go fmt ./...
go vet ./...
go test ./...
```

Integration tests have the `integration` build tag. Start the isolated database through the Compose test profile, set `TEST_DATABASE_URL` to port 5433, and run:

```text
go test -tags=integration ./...
```

Repository integration tests run inside transactions and roll back test records. The test URL must never target a shared or production database.

## Frontend workflow

With PostgreSQL, migrations, and API running:

```text
cd frontend
npm ci
npm run dev
```

Vite serves the UI on `http://127.0.0.1:5173` and proxies `/api` to `http://127.0.0.1:8080`.

Required checks after frontend changes:

```text
npm run format:check
npm run lint
npm run test
npm run build
```

Keep API access in `src/api`, use strict TypeScript, and preserve accessible loading, empty, validation, success, and error states.

## Compose and end-to-end workflow

Validate every committed Compose shape before starting services:

```text
docker compose config -q
docker compose --profile test config -q
docker compose --profile e2e config -q
```

Build and start the normal stack with `docker compose up --build`. The expected startup order is database, migration job, API, then frontend. Check `/healthz`, `/readyz`, and the UI after Docker changes.

Run the complete deterministic E2E path with:

```text
docker compose --profile e2e up --build --abort-on-container-exit --exit-code-from e2e e2e
```

This starts the normal stack, applies `backend/seeds/e2e.sql`, and runs one Playwright worker from the pinned E2E image. It uses the image's Chromium and does not run `playwright install` on the host. The test container reaches `http://frontend:8080` and `http://api:8080` through Compose DNS. Traces, screenshots, and video are retained only when a test fails.

Do not remove volumes, reset databases, or run destructive Docker commands without explicit approval.

## OpenAPI contract

From the repository root in PowerShell, run the pinned Redocly image without installing a package:

```powershell
$apiPath = (Resolve-Path .\api).Path
docker run --rm --volume "${apiPath}:/spec:ro" redocly/cli:2.27.0@sha256:119913c77da63685f2e76be82470231823278a1658edd3806ea4c7c48d21b891 lint --config /spec/redocly.yaml /spec/openapi.yaml
```

The OpenAPI document version is `0.2.0`. That version identifies this documented engineering baseline; it does not introduce new runtime routes.

## Security workflow

The `Security` job is the canonical executable policy. It:

1. Runs `go list -m -json all` and `npm audit --audit-level=high`.
2. Reuses one Trivy cache while scanning source dependencies, repository secrets, and Dockerfile misconfiguration.
3. Builds the backend and frontend runtime images, then fails either image scan on High or Critical findings.
4. Generates source, backend-image, and frontend-image CycloneDX JSON SBOMs and parses all three before upload.

Set `TRIVY_SKIP_JAVA_DB_UPDATE=true`; this repository has no Java artifacts and must not download the Java vulnerability database. Local image scans use `docker save` into explicit temporary tar files because the scanner does not receive the Docker socket. Delete only those tar files after the scans complete. Keep the Trivy cache so later scans reuse the downloaded databases.

The authoritative thresholds, exclusions, and exception process are in [security/scan-policy.md](security/scan-policy.md). The pinned image and exact CI invocations are in [.github/workflows/security.yml](../.github/workflows/security.yml).

## Continuous integration

The repository defines seven status-check job names:

| Job                   | Required result                                      |
| --------------------- | ---------------------------------------------------- |
| `Backend`             | Go formatting, vet, and tests pass                   |
| `Backend Integration` | PostgreSQL repository integration tests pass         |
| `Frontend`            | Locked install, format, lint, Vitest, and build pass |
| `Contracts`           | Redocly accepts the OpenAPI document                 |
| `Compose`             | All three Compose configurations parse               |
| `End-to-End`          | The E2E container exits successfully                 |
| `Security`            | Scans and SBOM validation succeed                    |

Actions and tool containers are pinned, job permissions are read-only, jobs have timeouts, and superseded branch runs are cancelled. A solo-maintainer ruleset is configured only after a real pull request has produced all seven check names; follow [github-settings.md](github-settings.md).

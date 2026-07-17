# Development Guide

## Toolchain

- Go 1.26.5, as declared by `backend/go.mod`.
- Node.js 20.20.2 or a compatible release supported by the locked frontend dependencies.
- npm using the committed `package-lock.json`.
- Docker with the Compose plugin.

Do not install or modify global tools as part of normal repository development.

## Environment

`.env.example` contains safe local defaults. The API requires `LAUNCHRAIL_DATABASE_URL` when run outside Compose. Compose injects an internal database URL automatically.

Important variables:

| Variable | Purpose | Default |
| --- | --- | --- |
| `POSTGRES_PORT` | Development database host port | `5432` |
| `TEST_POSTGRES_PORT` | Isolated integration database host port | `5433` |
| `API_PORT` | API host port | `8080` |
| `FRONTEND_PORT` | UI host port | `3000` |
| `LAUNCHRAIL_LOG_LEVEL` | Structured log threshold | `info` |
| `LAUNCHRAIL_SHUTDOWN_TIMEOUT` | API graceful shutdown window | `10s` |

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

## Compose workflow

Validate configuration before starting services:

```text
docker compose config -q
docker compose --profile test config -q
```

Build and start the full stack with `docker compose up --build`. The expected startup order is database, migration job, API, then frontend. Check `/healthz`, `/readyz`, and the UI after Docker changes.

Do not remove volumes, reset databases, or run destructive Docker commands without explicit approval.

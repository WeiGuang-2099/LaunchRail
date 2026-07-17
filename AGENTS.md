# LaunchRail Engineering Instructions

## Project mission

LaunchRail is a portfolio-grade internal developer platform that demonstrates production software engineering, platform engineering, DevSecOps, Kubernetes, GitOps, observability, and reliability engineering.

The repository must contain working, testable software rather than disconnected demos or placeholder implementations.

## Communication

* Write source code, identifiers, comments, commit-ready documentation, API descriptions, and error messages in English.
* Explain completed work and important technical decisions to the user in Chinese.
* Be precise about what was implemented, tested, or left incomplete.
* Never claim that a command passed unless it was actually executed successfully.

## Development approach

* Work incrementally, one milestone at a time.
* Complete a working vertical slice before adding another infrastructure layer.
* Do not implement future milestones unless the current task explicitly requests them.
* Prefer a modular monolith for the control plane unless there is a demonstrated reason to introduce another deployable service.
* Avoid speculative abstractions, unnecessary design patterns, premature microservices, and placeholder code.
* Do not leave core behavior as TODO comments.
* Keep domain logic separate from HTTP, database, and infrastructure adapters.
* Design integrations behind narrow interfaces so Kubernetes, GitHub, and GitOps implementations can be added later.

## Technology direction

* Backend: Go.
* Frontend: React, TypeScript, and Vite.
* Database: PostgreSQL.
* Local infrastructure: Docker Compose initially, then kind or k3d in later milestones.
* API style: versioned REST JSON API with an OpenAPI specification.
* Cloud infrastructure in later milestones: AWS and Terraform.
* Kubernetes delivery in later milestones: Argo CD and Argo Rollouts.
* Observability in later milestones: OpenTelemetry, Prometheus, Grafana, Tempo, and Loki.

Use stable dependency versions compatible with the installed development environment. Do not modify global development tools unless explicitly approved.

## Backend requirements

* Use idiomatic Go and explicit error handling.
* Use dependency injection through constructors rather than global mutable state.
* Use context propagation for request-scoped operations.
* Implement graceful shutdown.
* Use structured logging.
* Add request and correlation IDs.
* Validate all externally supplied data.
* Return consistent machine-readable API errors.
* Use database migrations.
* Keep SQL explicit and reviewable.
* Do not expose internal errors or secrets through HTTP responses.

## Frontend requirements

* Use strict TypeScript.
* Do not use `any` unless it is unavoidable and documented.
* Provide loading, empty, validation, success, and error states.
* Build accessible forms and navigation.
* Keep API access in a dedicated client layer.
* Keep visual components separate from server-state logic.
* Do not introduce a large UI framework without a clear need.

## Testing requirements

* Add unit tests for domain behavior and validation.
* Add integration tests for database-backed repositories and HTTP handlers where practical.
* Add regression tests whenever fixing a defect.
* Tests must cover success cases, validation failures, duplicate resources, missing resources, and persistence behavior.
* Prefer deterministic tests over timing-dependent tests.
* Do not weaken or delete tests merely to make a build pass.

## Verification

After changing backend code, run:

```text
go fmt ./...
go vet ./...
go test ./...
```

After changing frontend code, run the repository's configured formatting, linting, testing, and production build commands.

After changing Docker configuration, validate the Compose file and run the affected services where the local environment permits.

Before finishing a task:

1. Run the smallest relevant tests.
2. Run the broader repository checks when practical.
3. Fix failures caused by the changes.
4. Report the exact commands executed and their results.
5. State any checks that could not be run and explain why.

## Safety and repository rules

* Do not push to a remote repository.
* Do not create commits unless explicitly requested.
* Do not delete user-authored work.
* Do not run destructive database, Docker, Git, Kubernetes, or filesystem commands without explicit approval.
* Do not place secrets, tokens, passwords, or real credentials in source files.
* Provide `.env.example` files containing safe placeholders.
* Keep generated files and build artifacts out of Git when appropriate.

## Documentation requirements

Maintain documentation as implementation evolves:

* `README.md`
* `docs/architecture.md`
* `docs/development.md`
* `docs/roadmap.md`
* `docs/adr/`
* OpenAPI specification
* Relevant diagrams using Mermaid where appropriate

Documentation must distinguish implemented functionality from planned functionality.

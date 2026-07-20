# ADR 0005: Establish deterministic engineering verification

- Status: Accepted
- Date: 2026-07-20

## Context

The service catalog is a working vertical slice, but repository quality was enforced only through developer-run commands. LaunchRail needs reproducible pull-request checks, a browser-level path through the complete stack, supply-chain evidence, and an explicit blocking policy before identity or platform integrations expand the attack surface.

The verification tooling must not become part of the runtime control plane or pollute the frontend dependency graph with unrelated contract and security tools.

## Decision

GitHub Actions is the pull-request verification system. Required jobs separately cover Go, PostgreSQL integration, frontend quality, OpenAPI, Compose, full-stack browser behavior, and security. Third-party actions use full commit SHAs, checkout credentials are not persisted, permissions are read-only, and workflows neither publish images nor deploy software.

Playwright 1.61 runs inside the matching Microsoft Playwright Noble image. Compose starts PostgreSQL, migrations, the Go API, Nginx, an idempotent SQL seed, and the browser test. The browser uses Docker service names, one worker, no retries, and retains evidence only on failure.

Redocly CLI 2 validates the OpenAPI 3.1 document from its official container image. It is not an npm dependency of the frontend.

Trivy 0.72 scans locked source dependencies, repository secrets, current Dockerfile misconfigurations, and both runtime images. High and Critical findings and detected secrets are blocking. CycloneDX source and image SBOMs are retained as workflow artifacts. The Java vulnerability database is disabled because LaunchRail has no Java components.

Docker Compose is syntax-validated but is not described as policy-scanned. Terraform, Kubernetes, and Helm scanning will be added only when those files exist.

## Consequences

The repository gains deterministic and reviewable evidence before later milestones begin. Browser and scanner images are large, but are versioned, downloaded once per local Docker cache, and isolated from host tooling.

The engineering milestone remains incomplete until the workflows run successfully on GitHub and the required checks are enforced for `main`. Repository rules are external state and require an explicit owner action.

## Alternatives considered

- Component tests only were rejected because they cannot validate Nginx, migrations, API routing, and PostgreSQL together.
- Host-installed browsers and scanners were rejected to avoid machine-specific setup.
- Adding Redocly to frontend npm dependencies was rejected because contract tooling is unrelated to the shipped UI.
- Placeholder Terraform or Kubernetes files were rejected because scanning an artificial scope would overstate the implemented platform.

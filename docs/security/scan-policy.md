# Security Scan Policy

## Purpose

This policy defines the automated security baseline for the implemented LaunchRail service catalog. It applies to pull requests, pushes to `main`, weekly scheduled scans, and manual workflow runs.

## Automated scope

| Area                | Tool and coverage                                                                  | Enforcement                        |
| ------------------- | ---------------------------------------------------------------------------------- | ---------------------------------- |
| npm dependencies    | `npm audit` against the committed lock file                                        | High and Critical block            |
| Source dependencies | Trivy filesystem scan of Go and npm lock files, including development dependencies | High and Critical block            |
| Secrets             | Trivy repository secret scan excluding generated output and Git metadata           | Any detection blocks               |
| Configuration       | Trivy misconfiguration scan of the current Dockerfiles                             | High and Critical block            |
| Runtime images      | Trivy vulnerability and secret scans of built backend and frontend images          | High and Critical block            |
| Inventory           | CycloneDX source and runtime image SBOMs                                           | Missing or invalid documents block |

Compose files are validated with `docker compose config`; they are not claimed as policy-scanned. Terraform, Kubernetes, and Helm do not exist in the current repository and are outside the implemented scan scope.

## Tooling and cache

- Playwright uses `mcr.microsoft.com/playwright:v1.61.0-noble` pinned by digest.
- OpenAPI validation uses `redocly/cli:2.27.0` pinned by digest.
- Local security validation uses `aquasec/trivy:0.72.0` pinned by digest.
- GitHub-hosted actions use full commit SHAs.
- Trivy reuses one cache and skips the Java vulnerability database.
- GitHub restores the Security job's date-keyed Trivy database cache between runs; all scans in a run share that directory.
- Local image archives are temporary inputs and are removed after validation.
- Generated reports, caches, archives, and SBOMs are not committed.

## Exceptions

There are no standing exceptions.

A proposed exception must be reviewed in a pull request and document:

1. The exact advisory or rule identifier and affected component.
2. Why remediation is not currently practical.
3. Reachability and impact in LaunchRail.
4. Compensating controls.
5. An accountable owner and an expiry date.
6. A follow-up issue or milestone for removal.

Do not add blanket severity exclusions. An ignore entry is allowed only after the exception record is accepted and must expire.

## Evidence

The `Security` job log records blocking scan results. The `launchrail-sbom` artifact contains:

- `source.cdx.json`
- `backend-image.cdx.json`
- `frontend-image.cdx.json`

Artifacts are retained for 30 days. The workflow does not publish images, send SBOMs to a third party, or deploy software.

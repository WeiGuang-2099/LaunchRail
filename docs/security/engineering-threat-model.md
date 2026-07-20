# Engineering Pipeline Threat Model

- Status: Implemented repository baseline; GitHub enforcement pending
- Date: 2026-07-20
- Scope version: LaunchRail v0.2 engineering verification
- Review trigger: Any workflow permission, dependency source, artifact, container build, scanner, or release change

## Scope

This model covers repository automation from a proposed change through build, test, scanning, and evidence retention. It does not replace the separate identity threat model and does not claim controls for planned Kubernetes, Terraform, GitOps, or cloud delivery.

## Assets and trust boundaries

| Asset                             | Required property                                 |
| --------------------------------- | ------------------------------------------------- |
| Source, workflows, and lock files | Integrity and reviewability                       |
| GitHub Actions token              | Least privilege and non-persistence               |
| Dependency and container inputs   | Authenticity and pinned provenance                |
| Test database                     | Isolation from developer and production data      |
| Scan results and SBOMs            | Integrity, reproducibility, and bounded retention |
| Docker daemon on the CI runner    | Isolation to the ephemeral job                    |

Untrusted pull-request code crosses into a GitHub-hosted runner. The workflows build and execute that code but receive only a read-only repository token, do not persist checkout credentials, do not publish images, and do not deploy.

## Threats and controls

| Threat                                                   | Control                                                                             |
| -------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| A mutable third-party action changes behavior            | Every external action is pinned to a full commit SHA                                |
| A mutable container input changes behavior               | Tool and current application base images are pinned by digest                       |
| A pull request exfiltrates a credential                  | Read-only permissions, no deployment credentials, no persisted checkout credentials |
| A dependency introduces a known vulnerability            | npm audit, Trivy source scan, Dependabot, and blocking High/Critical policy         |
| A secret is committed                                    | Trivy secret scanning blocks the workflow                                           |
| A runtime base image contains a known vulnerability      | Both runtime images are built and scanned before acceptance                         |
| A Dockerfile weakens runtime isolation                   | Misconfiguration scanning and explicit non-root runtime users                       |
| Browser tests pass against stale developer data          | Idempotent fixed seed and clean CI database                                         |
| A test passes locally but fails through the real gateway | Playwright runs through Nginx, API, migrations, and PostgreSQL                      |
| Evidence is missing or malformed                         | Three CycloneDX documents are parsed before artifact upload                         |
| A cache poisons a later run                              | GitHub cache scope, pinned tools, and authoritative rescan results                  |
| A scanner outage creates a false pass                    | Scan or SBOM command failure fails the job                                          |

## Residual risks

- GitHub-hosted runner and registry compromise remain external systemic risks.
- A newly published vulnerability may not exist in the scanner database at build time.
- A full-SHA action can still contain malicious code; review and Dependabot updates remain required.
- Local Docker caches are not authoritative evidence. GitHub job results and retained artifacts are the milestone evidence.
- Branch rules are external repository state and must be verified after the first successful workflow run.

## Acceptance checklist

- [ ] All seven required checks have reported successfully on a pull request.
- [ ] `main` requires those checks and blocks force pushes and deletion.
- [ ] Runtime images have no unhandled High or Critical findings.
- [ ] The three SBOM documents parse as CycloneDX.
- [ ] No workflow publishes an image or deploys software.
- [ ] Documentation describes only implemented scan targets.

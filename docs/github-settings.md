# GitHub Repository Settings

Workflow files define check behavior, but repository rules are external GitHub state. Apply these settings after both workflows have reported successfully on the default branch.

## Solo-maintainer ruleset for `main`

Create an active branch ruleset targeting the default branch:

- Require a pull request before merging.
- Require zero approving reviews.
- Require all review conversations to be resolved.
- Require branches to be up to date before merging.
- Block force pushes.
- Block deletion of `main`.
- Limit bypass to an explicitly approved emergency administrator role.

Require these exact status checks:

| Workflow                 | Required check        |
| ------------------------ | --------------------- |
| `Continuous Integration` | `Backend`             |
| `Continuous Integration` | `Backend Integration` |
| `Continuous Integration` | `Frontend`            |
| `Continuous Integration` | `Contracts`           |
| `Continuous Integration` | `Compose`             |
| `Continuous Integration` | `End-to-End`          |
| `Security Baseline`      | `Security`            |

A check can be selected only after it has reported for the repository.

## Security settings

Enable the dependency graph, Dependabot alerts, and Dependabot security updates. Enable secret scanning and push protection when supported by the repository plan.

## Verification

1. Push the workflow changes through the approved bootstrap process.
2. Wait for all seven checks to succeed.
3. Apply the ruleset above.
4. Open a documentation-only pull request.
5. Confirm merge is blocked while a required check is pending or failing.
6. Confirm a non-bypass user cannot force-push or delete `main`.
7. Record the ruleset identifier in the project operations record.

Do not mark the engineering credibility milestone complete until this verification succeeds.

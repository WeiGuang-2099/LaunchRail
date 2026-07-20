# ADR 0004: Use a controlled team directory for authorization and audit

- Status: Accepted
- Date: 2026-07-17

## Context

Service records currently accept `ownerTeam` as arbitrary text, and the application has no concept of an actor. Authentication alone would identify callers but would not prevent one authenticated user from modifying another team's service, assigning an invalid team, or denying a completed mutation.

LaunchRail needs an authorization model that preserves internal service discovery, keeps business rules out of React and HTTP handlers, and can later apply to GitHub and deployment resources without introducing provider-specific roles throughout the domain.

## Decision

### Identity and roles

The immutable external identity key is the pair of normalized OIDC issuer and `sub`. Email and display name are attributes, not identifiers.

Every authenticated user can read the service catalog. The `platform-admin` role is granted only when an exact group from the validated OIDC claim matches the configured administrator group. The name of the groups claim and administrator group path are deployment configuration; request bodies and headers cannot grant either role.

A user is a service owner for a team when a validated OIDC group exactly matches that team's configured external group. Team membership is captured in the server-side session and passed to application use cases as trusted actor context. React may use returned permissions to present the interface, but the API remains authoritative.

### Controlled team directory

PostgreSQL will hold the controlled team directory. A team has a UUID, immutable unique slug, display name, exact OIDC group mapping, created and updated timestamps, and an optional disabled timestamp. Group matching is exact and case-sensitive after claim normalization; prefix and substring matches are forbidden.

The existing public `ownerTeam` field remains the stable team slug so the service API does not require a breaking identifier change. Registrations and updates must resolve that slug to an enabled directory record.

Team records are provisioned through an explicit, idempotent operator bootstrap process rather than database migrations or a public management UI. Local Compose will eventually provide non-secret reference mappings consistent with its Keycloak realm. The initial authentication milestone exposes only a read interface:

```text
GET /api/v1/teams
```

For a non-administrator it returns enabled teams the caller belongs to. For a platform administrator it returns every enabled team. Disabled teams remain resolvable for historical service records but cannot receive new or transferred services.

### Authorization policy

| Operation                                          | Authenticated user            | Team member                     | Platform administrator     |
| -------------------------------------------------- | ----------------------------- | ------------------------------- | -------------------------- |
| Read active or archived services                   | Allow                         | Allow                           | Allow                      |
| List selectable teams                              | Own enabled teams             | Own enabled teams               | All enabled teams          |
| Register a service                                 | Deny without an eligible team | Allow for an owned enabled team | Allow for any enabled team |
| Edit mutable metadata without changing `ownerTeam` | Deny                          | Allow for the owning team       | Allow                      |
| Change `ownerTeam`                                 | Deny                          | Deny                            | Allow to an enabled team   |
| Archive a service                                  | Deny                          | Allow for the owning team       | Allow                      |
| Repeat an authorized archive                       | Deny                          | Return `204`                    | Return `204`               |

Authorization is evaluated in the application layer after loading the target resource when a decision depends on its current owner. HTTP middleware authenticates the session and creates actor context, but handlers do not implement resource policy. Repository methods do not trust actor, role, team, or owner data supplied by the client.

An authenticated caller who lacks permission receives the existing JSON error envelope with HTTP `403` and code `permission_denied`. Because every authenticated user can read service existence and ownership, returning `403` for a denied mutation does not disclose information that the caller could not already read.

Service responses will eventually include advisory capability hints calculated by the backend:

```json
{
  "permissions": {
    "canEdit": true,
    "canArchive": true,
    "canTransferOwnership": false
  }
}
```

### Audit requirements

Every successful service registration, metadata update, ownership transfer, and archive creates an append-only audit event in the same PostgreSQL transaction as the catalog mutation. A successful mutation must roll back if its audit event cannot be stored.

The event records:

- event UUID and UTC timestamp;
- actor issuer, subject, and display-name snapshot;
- action, resource type, and resource UUID;
- request ID and outcome;
- owning team before and after the action;
- the minimum before and after metadata required to explain the change.

Secrets, provider tokens, session identifiers, CSRF tokens, authorization codes, and raw HTTP bodies are prohibited. Denied authorization attempts are written to structured security logs with actor, action, target, request ID, and reason; they do not need to participate in a catalog transaction.

The runtime database role will have insert and select access to audit events but no update or delete access. The migration role retains schema-management responsibility under ADR 0002. An administrator-only audit query can be added with the implementation milestone; audit data is never exposed through the general service response.

## Alternatives considered

### Use OIDC groups directly as the team directory

Rejected because a user's claim cannot reliably enumerate valid teams they do not belong to, and provider group names would become ungoverned service metadata.

### Manage users, roles, and memberships entirely in LaunchRail

Rejected because it would duplicate identity lifecycle management and create a second source of truth for membership.

### Let source-team members transfer ownership

Rejected because a user could move a resource into another team's control without that team's approval. Ownership transfer is an explicit platform-administrator operation.

### Restrict catalog reads by team

Rejected because internal service discovery is a core catalog purpose. This decision must be revisited if service records later contain secrets or tenant-isolated data.

## Consequences

- Service ownership becomes a validated authorization boundary rather than display metadata.
- Existing `ownerTeam` clients remain structurally compatible, but values must resolve to an enabled team.
- Application use cases require trusted actor context and transactional audit support.
- Administrator-group or team-membership changes can remain effective until an existing session expires or is revoked.
- Database privileges must separate runtime mutation access from migration and audit-administration capabilities.
- The authorization matrix and proposed response additions remain design contracts until the identity milestone is implemented and OpenAPI is updated.

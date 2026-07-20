# ADR 0003: Use OIDC with backend-managed sessions

- Status: Accepted
- Date: 2026-07-17

## Context

LaunchRail is an internal developer platform whose React application and Go API are presented through one Nginx origin. The service catalog currently has no authentication, so every API caller can read and mutate catalog records. The next milestone needs browser sign-in through an external identity provider without placing reusable provider tokens in browser storage or coupling the control plane to one identity vendor.

The design must work with the modular monolith, support local development through Docker Compose, preserve consistent API errors, and provide a secure basis for team-scoped authorization.

## Decision

### Provider and protocol

Use a vendor-neutral OpenID Connect provider discovered from one explicitly configured issuer. Keycloak will be the local Compose reference provider, but Keycloak-specific roles, administration APIs, and token formats will not become application contracts.

The Go control plane will act as a confidential OIDC client and backend for frontend (BFF). It will use Authorization Code Flow with all of the following controls:

- PKCE with the `S256` challenge method for every authorization request.
- Cryptographically random, transaction-specific, single-use `state` and `nonce` values bound to the initiating browser.
- Exact, pre-registered redirect URI matching.
- Validation of the ID token signature, issuer, audience, expiry, issued-at time, and nonce before creating a session.
- A fixed configured issuer; request data must never select the discovery, authorization, token, or JWKS endpoint.
- A `returnTo` value limited to a normalized same-origin path beginning with one `/`; absolute and protocol-relative URLs are rejected.

These requirements follow the redirect-flow and authorization-code guidance in [OAuth 2.0 Security Best Current Practice](https://www.rfc-editor.org/info/rfc9700/).

### Session model

After a successful callback, the control plane will create a PostgreSQL-backed session. The browser receives only a cryptographically random opaque identifier with at least 256 bits of entropy. PostgreSQL stores a one-way hash of that identifier together with:

- the OIDC issuer and immutable `sub` claim;
- display name and optional email for presentation only;
- the mapped team membership and platform-administrator decision captured at sign-in;
- a CSRF secret hash;
- issued, last-seen, absolute-expiry, and revoked timestamps.

The control plane will discard provider access, refresh, and ID tokens after it has validated and normalized the required claims. Provider tokens, client secrets, session identifiers, authorization codes, and CSRF tokens must never be logged or included in audit events.

Sessions have a 30-minute idle timeout and an eight-hour absolute lifetime. Both values are configuration, but deployments must not disable either timeout. A new successful login rotates the session identifier, logout revokes the server-side record, and an expired or revoked session is rejected even if its browser cookie remains present. Session cleanup must be bounded and safe to run concurrently.

### Browser and request security

Production uses this cookie:

```text
__Host-launchrail_session=<opaque>; Secure; HttpOnly; SameSite=Lax; Path=/
```

The cookie has no `Domain` attribute. Production authentication requires HTTPS. Local HTTP development uses a differently named cookie without `Secure` only when an explicit development mode is enabled and the public host is a loopback address. The API must refuse that exception for a non-loopback host.

SameSite is defense in depth, not the complete CSRF control. Every authenticated `POST`, `PUT`, `PATCH`, and `DELETE` request, including logout, must provide a session-bound synchronizer token in `X-CSRF-Token` and pass an exact Origin check. Safe methods must not mutate state. LaunchRail will remain same-origin and will not enable permissive credentialed CORS. These controls follow the [OWASP Session Management](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html) and [CSRF Prevention](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html) guidance.

Authentication callback failures are logged with the request ID and presented to the browser as a generic authentication failure. Internal provider, token, and claim-validation errors are not exposed.

### Planned HTTP contract

The authentication milestone will add the following routes:

| Method | Path                             | Authentication        | Behavior                                                                                      |
| ------ | -------------------------------- | --------------------- | --------------------------------------------------------------------------------------------- |
| `GET`  | `/auth/login?returnTo=/services` | Anonymous             | Start a bound OIDC authorization transaction.                                                 |
| `GET`  | `/auth/callback`                 | Anonymous transaction | Validate the response, create a session, and redirect to the validated path.                  |
| `POST` | `/auth/logout`                   | Session and CSRF      | Revoke the session, clear browser state, and return `204`.                                    |
| `GET`  | `/api/v1/me`                     | Session               | Return the normalized user, mapped teams, administrator flag, CSRF token, and session expiry. |

`/healthz`, `/readyz`, and the two routes required to complete the OIDC redirect flow remain anonymous. Every other `/api/v1` route requires an authenticated session. An absent, invalid, expired, or revoked session returns the existing JSON error envelope with HTTP `401` and code `authentication_required`.

Nginx will proxy `/auth/` and `/api/` to the Go control plane. The API port must not be exposed outside the trusted reverse-proxy boundary in a deployed environment.

## Alternatives considered

### React Authorization Code Flow with PKCE

Rejected because the browser would own provider tokens and authentication lifecycle logic. It would enlarge the impact of cross-site scripting and distribute security-critical behavior between React and Go.

### Bearer-token validation only

Rejected because it would not provide the required interactive browser sign-in and logout lifecycle.

### Encrypted stateless session cookies

Rejected because immediate logout, administrator revocation, membership-change response, and key rotation are harder to enforce consistently.

### In-memory sessions

Rejected because sessions would be lost on restart and would not work across API replicas.

## Consequences

- Authentication and provider tokens remain outside React, reducing their browser exposure.
- PostgreSQL becomes a runtime dependency for authenticated requests as well as service persistence.
- Cookie authentication requires explicit CSRF defenses and a trusted same-origin reverse proxy.
- Group and administrator claims remain valid until the session expires or is revoked. The threat model records this bounded stale-authorization window as residual risk.
- Local HTTP is deliberately less secure than production and is acceptable only on loopback with an explicit configuration switch.
- The OpenAPI specification remains unchanged until these routes and behaviors are implemented.

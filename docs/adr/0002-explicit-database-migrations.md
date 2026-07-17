# ADR 0002: Run database migrations as an explicit release step

- Status: Accepted
- Date: 2026-07-17

## Context

Automatically changing the schema from every API instance creates race conditions, complicates rollbacks, and couples application availability to migration behavior.

## Decision

Keep the migration runner in a dedicated command. The API validates database connectivity but never applies schema changes. Local Compose runs a one-shot migration service after PostgreSQL is healthy and starts the API only after that job succeeds.

## Consequences

- Migration execution is visible and independently testable.
- Application replicas can start without competing to change schema state.
- Delivery systems must run the migration command before deploying an API version that requires the new schema.
- Down migrations remain explicit operations and must not be run without reviewing their data impact.

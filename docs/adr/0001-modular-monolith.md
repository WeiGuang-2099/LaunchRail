# ADR 0001: Use a modular monolith for the control plane

- Status: Accepted
- Date: 2026-07-17

## Context

LaunchRail needs clear domain and integration boundaries but does not yet have scaling, ownership, or isolation evidence that justifies multiple backend deployables.

## Decision

Build the Go control plane as one deployable modular monolith. Keep catalog domain behavior independent from HTTP and PostgreSQL, coordinate use cases in an application layer, and place integrations behind narrow interfaces and adapters.

## Consequences

- Local operation and testing remain simple.
- Transactions and error behavior are easier to reason about.
- Future integrations can be added without defining premature network contracts.
- A module may become a separate service only when operational evidence demonstrates an independent deployment boundary.

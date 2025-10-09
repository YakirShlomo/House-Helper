# Architecture Decision Records (ADRs)

This directory contains records of architectural decisions made in the House Helper project.

## What is an ADR?

An Architecture Decision Record (ADR) captures an important architectural decision made along with its context and consequences. ADRs help teams:

- **Understand past decisions**: Why was this approach chosen?
- **Onboard new team members**: Quickly learn the architectural rationale
- **Avoid revisiting settled decisions**: Don't rehash old debates
- **Document trade-offs**: Understand what was gained and what was sacrificed

## Format

Each ADR follows this template:

```markdown
# ADR-XXXX: Title

**Status**: Accepted | Proposed | Deprecated | Superseded by ADR-YYYY

**Date**: YYYY-MM-DD

**Deciders**: List of people involved

## Context

What is the issue we're trying to solve? What factors are at play?

## Decision

What did we decide to do?

## Consequences

### Positive
- What improves?

### Negative
- What are the downsides?

### Neutral
- What else changes?

## Alternatives Considered

What other options did we evaluate? Why were they rejected?
```

## Index of ADRs

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [001](./ADR-001-go-for-backend.md) | Use Go for Backend Services | Accepted | 2024-01-10 |
| [002](./ADR-002-flutter-for-mobile.md) | Use Flutter for Mobile App | Accepted | 2024-01-10 |
| [003](./ADR-003-postgresql-over-nosql.md) | Use PostgreSQL over NoSQL | Accepted | 2024-01-11 |
| [004](./ADR-004-kafka-for-events.md) | Use Kafka for Event Streaming | Accepted | 2024-01-12 |
| [005](./ADR-005-temporal-for-workflows.md) | Use Temporal for Workflow Orchestration | Accepted | 2024-01-12 |
| [006](./ADR-006-eks-over-ecs.md) | Use Amazon EKS over ECS | Accepted | 2024-01-13 |
| [007](./ADR-007-helm-for-deployments.md) | Use Helm for Kubernetes Deployments | Accepted | 2024-01-13 |
| [008](./ADR-008-monorepo-structure.md) | Use Monorepo Structure | Accepted | 2024-01-14 |
| [009](./ADR-009-jwt-authentication.md) | Use JWT for Authentication | Accepted | 2024-01-14 |
| [010](./ADR-010-microservices-architecture.md) | Use Microservices Architecture | Accepted | 2024-01-15 |

---

## License

Copyright © 2024 House Helper. All rights reserved.

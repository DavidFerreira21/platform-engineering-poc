# ADR-0006: Identity choice and deferred platform capabilities

- **Status:** Proposed / deferred
- **Date:** 2026-11-16

## Context
The platform will eventually need user authentication, groups, roles, self-service permissions and an integration path for a developer portal. The spoken discussion contained an ambiguous product name ("ClickHouse") in the identity context; it must not be treated as an approved IAM choice.

## Decision
- **Identity provider: not yet confirmed.** Keycloak is a candidate for self-hosted OIDC, groups and roles; AWS-based identity options can also be evaluated. ClickHouse is an analytical database, not an IAM replacement.
- Initial API can use a documented development-only authentication mode; it must not be represented as production security.
- **Backstage** is deferred until the Control Plane API, product contracts and lifecycle are stable.
- **Crossplane** is deferred until the Terraform executor and its interface are proven; it will be a separate reconciliation-based executor, not a drop-in replacement for Argo Workflows.
- AI-assisted SRE, advanced tracing, artifact signing and enterprise-grade multi-tenancy are deferred.

## Follow-up decisions
1. Select the identity provider and OIDC/RBAC model.
2. Define catalog schema, product lifecycle and request/approval contracts.
3. Select GitOps bootstrap ownership and an AWS identity mechanism for local kind.

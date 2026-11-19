# ADR-0001: Kubernetes API Server as the platform API

- **Status:** Accepted as POC target
- **Date:** 2026-11-18

## Context
The former POC proposed a separate FastAPI REST control API. The chosen platform architecture can instead expose declarative Kubernetes-native contracts.

## Decision
Use the Hub Kubernetes API Server, CRDs, admission/RBAC and Platform Controllers as the public platform control interface. Controllers reconcile spec to status; client integrations may authenticate directly or through scoped adapters, without mandatory custom REST. GitOps owns persisted desired-state writes.

## Consequences and trade-offs
We reuse native discovery, watches, RBAC and resource lifecycle, but must design versioned OpenAPI schemas, status conditions, admission policies, conversion and upgrade strategy. Kubernetes API accessibility/security needs planning.

## Validation
Validate through phased labs before real AWS provisioning; update design when a constraint is disproven.

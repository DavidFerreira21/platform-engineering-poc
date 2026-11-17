# ADR-0002: Separate CI, orchestration and GitOps responsibilities

- **Status:** Accepted for POC
- **Date:** 2026-11-16

## Context
Tekton, Argo Workflows and Argo CD overlap if treated as interchangeable pipeline engines. We need distinct contracts.

## Decision
- **GitHub**: source control, pull requests and declarative source of truth.
- **Tekton Pipelines** (and Triggers when needed): CI for linting, unit/integration tests, container builds, vulnerability scanning and publishing artifacts.
- **Argo Workflows**: imperative infrastructure workflows, especially Terraform plan/apply, execution dependencies, status and controlled approval gates.
- **Argo CD**: continuous GitOps reconciliation of Kubernetes configuration and application releases, including the remote workload cluster.

A code delivery path can be: GitHub -> Tekton -> image registry -> reviewed GitOps manifest change -> Argo CD. A self-service infrastructure request follows: consumer -> Control Plane API -> validated contract/authorization -> Argo Workflows -> Terraform -> AWS. These are **different paths**, not one mandatory chain.

## Consequences
- Avoid two equivalent Terraform executors (Tekton and Argo Workflows) in the MVP.
- Argo CD reconciles declared Kubernetes state; it is not a generic Terraform job scheduler.
- Use explicit approvals before real infrastructure apply and record actor, target and result.
- CI builds and GitOps promotions require explicit artifact provenance/version references.
- Additional Kubernetes controllers increase local resource consumption and operational overhead.

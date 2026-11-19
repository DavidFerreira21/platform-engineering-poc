# ADR-0004: Distinct declarative GitOps, orchestration and execution

- **Status:** Accepted as POC target
- **Date:** 2026-11-18

## Context
Consumers can enter through Backstage, CLI, GitOps or AI/MCP while Git remains authoritative for durable desired state.

## Decision
Git is declarative desired state; Argo CD synchronizes Hub CRs. CLI/Backstage/MCP persistent changes should produce reviewed PRs. Platform Controllers reconcile lifecycle and invoke Argo Workflows for steps, Terraform for plans/applies and optional Crossplane for continuous infrastructure reconciliation. Each local Environment Management EKS may run Argo CD to manage its PCI and Non-PCI clusters, conditional on security approval. Argo CD is not Terraform's step scheduler.

## Consequences and trade-offs
Keep a single spec writer, distinguish hub and spoke Argo instances, prevent Terraform/Crossplane double ownership, and separate Argo GitOps from Argo Workflows. CI tooling is selected independently later.

## Validation
Validate through phased labs before real AWS provisioning; update design when a constraint is disproven.

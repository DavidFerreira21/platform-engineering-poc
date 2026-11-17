# ADR-0001: Two kind clusters and logical architectural planes

- **Status:** Accepted for POC
- **Date:** 2026-11-16

## Context
We need an inexpensive, reproducible learning environment with realistic management/workload boundaries, without creating one Kubernetes cluster per architectural plane.

## Decision
Use **kind** with two independent clusters:
- `platform-management`: hosts the Control Plane, Workflow/GitOps Plane and Observability Plane.
- `platform-workload`: hosts sample consumer applications and Kubernetes resources, representing the Kubernetes portion of the Data Plane.
- **AWS** is an additional external Data Plane for infrastructure products; Data Plane is not synonymous with a Kubernetes cluster.

Four **logical** planes are retained:
1. Control: product API, contracts, catalog, validation, authorization and execution status.
2. Workflow/GitOps: CI, imperative workflows, declarative Kubernetes delivery.
3. Data: provisioned AWS infrastructure and consumer workloads.
4. Observability: cross-cutting instrumentation, metrics, logs, audit and eventual tracing.

Suggested management namespaces: `platform-system`, `platform-workflows`, `argocd`, `platform-observability`, `ingress-system`; add `tekton-pipelines` and security namespaces as needed. Namespaces are **not** independent security boundaries.

## Consequences
- Argo CD in management can register/reconcile the workload cluster; access must be narrowly scoped.
- RBAC, dedicated ServiceAccounts and NetworkPolicies are required for meaningful isolation.
- The two clusters share a local host and are not highly available or production-equivalent.
- Bootstrap ownership (Helm-first versus Argo CD takeover) remains an implementation decision.

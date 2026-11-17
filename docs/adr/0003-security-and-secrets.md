# ADR-0003: Kubernetes policy, scanning and AWS-backed secrets

- **Status:** Accepted for POC
- **Date:** 2026-11-16

## Decision
- **Kyverno** for Kubernetes admission and policy-as-code guardrails (e.g., labels, allowed registries, security contexts).
- **Trivy** for vulnerability and misconfiguration scanning in Tekton CI.
- **AWS Secrets Manager** as the sole external secrets vault.
- **External Secrets Operator (ESO)** in each cluster that needs synchronization from Secrets Manager into Kubernetes Secrets.
- **AWS IAM** for least-privilege access; Terraform may provision secret containers, policies and integration resources.

**Not selected:** OpenBao; Cosign is out of the initial scope. Signing and attestations can be reconsidered later.

## Authentication and operational constraints
- Local kind does **not** provide EKS IRSA or EKS Pod Identity automatically. Select and document a short-lived/federated AWS identity strategy before enabling live cloud access; never assume AWS permissions on local pods.
- ESO is useful but optional: applications that support it may read Secrets Manager directly through AWS SDKs.
- Kubernetes Secrets synchronized by ESO remain sensitive and require RBAC, appropriate encryption-at-rest considerations and restricted logging.
- Avoid writing secret **values** through Terraform because they may be persisted in state; provision metadata and IAM separately from secure value injection.
- Keep Terraform state outside Git and restrict its backend access.

## Consequences
A managed AWS dependency is intentional even for the local lab. Cloud usage may incur costs and must be explicitly enabled. Secret rotation and refresh semantics need integration tests.

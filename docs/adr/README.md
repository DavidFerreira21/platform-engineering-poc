# Architecture Decision Records

These ADRs are the **new authoritative POC baseline (2026-11-18)**, replacing the obsolete pre-Kubernetes-native records. They describe target design, not deployed infrastructure.

| ADR | Decision |
| --- | --- |
| [0001](0001-kubernetes-native-platform-api.md) | Kubernetes API Server and CRDs as the platform API |
| [0002](0002-contract-families.md) | AWSEnvironment, Domain, specialized Capability families; no generic Claim |
| [0003](0003-hub-spoke-topology.md) | Platform Management Cluster and UAT/PROD account spokes |
| [0004](0004-workflow-and-gitops.md) | GitOps ownership, Argo Workflows, Terraform, Argo CD and Crossplane |
| [0005](0005-long-running-environment-lifecycle.md) | Manual gates, technical checks and lifecycle status |
| [0006](0006-observability-and-local-operators.md) | Cross-cutting telemetry and local management components |
| [0007](0007-security-and-isolation.md) | Identities, PCI boundary, account and deletion protections |
| [0008](0008-monorepo-and-learning.md) | Monorepo placement and incremental learning implementation |

## ADR conventions
All records have Status: **Accepted as POC target**, Date, Context, Decision and Consequences. A future change supersedes rather than silently rewrites a previously implemented choice; implementation status is tracked separately. New names remain organization-neutral.

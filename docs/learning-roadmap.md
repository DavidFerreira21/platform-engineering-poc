# Progressive POC learning roadmap

**Principle:** show components working in **small, observable increments**. This is an implementation sequence for the [target architecture](architecture.md), not a commitment to create seven real EKS clusters.

| Phase | Learn and build | Observable acceptance | Safe failure exercise |
| --- | --- | --- | --- |
| 00 | Repository baseline, ADRs, ownership, architecture checks | Docs and names agree; structure is navigable | Detect a deliberately invalid reference |
| 01 | Local `kind` Hub; Kubernetes API, RBAC, events and reconciliation | Inspect custom resource discovery and events | Invalid manifest / authorization denied |
| 02 | `AWSEnvironment` CRD and controller (mock AWS account lifecycle) | Create `uat`; status conditions evolve; no AWS access | Invalid stage and controller retry |
| 03 | Argo Workflows, versioned Environment workflow, Terraform `plan` dry-run | Observe workflow + execution reference in CR status | Step fails, resumes safely |
| 04 | Manual gates and technical checks for IP, routes, certificates | Suspended workflow resumes only after authorized approval and verifiable checks | Wrong approver / failed route validation |
| 05 | GitOps through Argo CD with PR-driven desired state | Git change -> CR -> controller -> status, drift correction | Manual spec drift restored |
| 06 | Simulate two isolated spokes UAT and PROD; management + PCI/Non-PCI roles | Distinct accounts/clusters represented in topology; local Argo CD targeting demonstrated | Wrong cluster target blocked |
| 07 | `Domain` lifecycle and namespace/resource associations | Two domains in one Environment; references validated | Unauthorized cross-domain access |
| 08 | Specialized `Database` and `ObjectStorage` Capability CRDs | Resource creation/status with mock implementation; later controlled AWS | Invalid parameters / failed reconciliation |
| 09 | Observability: OTel, metrics, logs, dashboard, audit across Hub and Spokes | Trace one request through controller/workflow/spoke | Diagnose failed operation |
| 10 | Backstage, CLI, AI/MCP experience adapters | Same resource contracts via all channels; GitOps PR for writes | Unauthorized actor rejected |
| 11 | Crossplane alternative engine and real AWS opt-in | Compare Terraform vs reconcile ownership without double managing resources | Drift and ownership conflict simulation |
| 12 | Security/recovery: PCI isolation, policy, deletion protection, backup/bootstrap | Explicit boundary review and recovery runbook | Management cluster loss simulation |

## Lab contract
For every lab: **why/concept**, target diagram before/after, prerequisites, step-by-step actions, commands and expected output, inspect events/conditions/logs, deliberately break and recover, acceptance checks, troubleshooting, tests and cleanup. Each lab points to canonical source files; never copy production code into tutorial folders.

## Scope discipline
- `kind` models roles first. Target EKS physical topology: one Hub and two spokes with Management EKS, PCI, Non-PCI per spoke; **do not create seven real clusters up front**.
- Model remote dependencies and manual ticket processes without fake integration claims.
- Initial AWS calls mocked; `terraform plan` must not run `apply`; real accounts and costs always require explicit authorization.
- Backstage, AI/MCP, Crossplane, and complete telemetry are target capabilities, not prerequisites for phase 01.
- Kubernetes-native API replaces the former FastAPI-first lab sequence; Tekton/Keycloak are not baseline requirements.

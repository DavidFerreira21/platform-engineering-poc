# AGENTS.md — Platform Engineering Lab

## Mission and source of truth
Build the personal Kubernetes-native Platform Engineering POC described in [architecture](docs/architecture.md) and [ADRs](docs/adr/README.md). These records supersede the repository's previous FastAPI-first, two-kind-cluster design. Do not reference any employer, real landing zone, account IDs or proprietary names.

## Non-negotiable architectural contract
- **Experience Layer:** Backstage, CLI, GitOps, AI/MCP; Git is the declarative desired-state source. Persistent changes from Backstage/CLI/MCP should be PR-driven; read/status operations may query the Kubernetes API with restricted RBAC.
- **Hub:** `Platform Management Cluster`. Kubernetes API Server is the platform API. CRD families: `AWSEnvironment` (one account and foundations), `Domain` (business/service ecosystem), specialized Capabilities (e.g. `Database`, `ObjectStorage`, `ComputeInstance`). `Capability` is a **category**, not necessarily a single CRD; no generic Claim in v1.
- **Control Plane:** CRD schemas, validation/policy, Platform Controllers, reconciled status and lifecycle.
- **Workflow Plane:** Argo Workflows, Terraform Executor, Argo CD, future/optional Crossplane, authorized manual gates; do not make Argo CD an infrastructure task engine.
- **Observability Plane:** telemetry for hub/spokes; collectors local as needed, backend location decided later.
- **Spokes:** UAT and PROD only, each a distinct AWS account with Environment Management EKS, EKS PCI, EKS Non-PCI and AWS Runtime. No country/entity hierarchy in this personal POC.
- **Hub-and-spoke describes management topology**, not presumed AWS networking. Logical planes are not dedicated Kubernetes clusters.
- Operators may require cluster-local installation; do not assume all controllers can operate remotely. PCI administrative boundaries must be assessed before shared Argo CD controls both PCI and Non-PCI.

## Repo ownership
- `control-plane/`: CRD schemas, API contracts, admission/reconciliation controller source and tests.
- `workflow-plane/`: Argo WorkflowTemplates, Terraform runner definitions, approval gates, GitOps ApplicationSets/config, optional Crossplane integration.
- `observability-plane/`: OpenTelemetry, metrics/logs/traces and audit config.
- `data-plane/uat/` and `data-plane/prod/`: desired-state spoke-specific configuration, management EKS, PCI/Non-PCI clusters, AWS Runtime declarations.
- `platform/`: hub bootstrap and kind simulation; deployment/installation wiring.
- `docs/`, `labs/`, `tests/`: decisions, learning, cross-plane validation.
Do not duplicate manifests across folders: use reusable definitions and thin environment overlays. Do not create speculative tools/services or implement everything in one PR.

## Delivery rules
1. Read architecture and applicable ADRs before changing behavior; update them if a decision changes.
2. Work in feature branches, submit a reviewable PR, and keep main untouched until review.
3. Favor small incremental labs and mock/dry-run by default; explicit permission before cloud-side effects.
4. Never commit secrets, kubeconfigs, Terraform state, private certificates, real AWS account IDs or organization-specific information.
5. Separate desired state (`spec`), observed status (`status`), and workflow execution logs/artifacts.
6. Reconcile idempotently; use retries with bounded timeouts, approval authorization, audit records, drift detection and safe deletion policies.
7. Clearly distinguish **target architecture**, **implemented**, **mocked**, and **deferred** components. Keep documentation links and naming synchronized.

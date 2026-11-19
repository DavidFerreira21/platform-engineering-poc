# Kubernetes-native Platform Engineering POC — Architecture

**Status:** Target architecture (not deployed). **Pattern:** hierarchical management Hub-and-Spoke. **Scope:** personal UAT + PROD only.

## Topology at a glance

```text
EXPERIENCE LAYER
  Backstage | CLI | GitOps (Git/PR) | AI/MCP
                     |
            PLATFORM MANAGEMENT CLUSTER (HUB)
              CONTROL PLANE
                Kubernetes API Server + RBAC/admission
                CRDs: AWSEnvironment | Domain | Capability resource kinds
                Platform Controllers (reconciliation/lifecycle/status)
              WORKFLOW PLANE (logical)
                Argo CD | Argo Workflows | Terraform Executor
                Manual gates | Crossplane (progressive)
              OBSERVABILITY PLANE (logical, cross-cutting)
                OTel | metrics | logs/traces | dashboards | audit/alerts
                     |
                   DATA PLANE
                   /        \
             SPOKE UAT     SPOKE PROD
             AWS Account   AWS Account
               |               |
        Environment       Environment
        Management EKS    Management EKS
         Argo CD           Argo CD
         Operators         Operators
         OTel gateway      OTel gateway
               |               |
          EKS PCI          EKS PCI
          EKS Non-PCI      EKS Non-PCI
          AWS Runtime      AWS Runtime
        (RDS/S3/EC2/Net) (RDS/S3/EC2/Net)
```

**Logical vs physical:** Control, Workflow and Observability are responsibilities co-located in the Hub, not three required clusters. Each Spoke's Management EKS is distinct from its PCI and Non-PCI workload EKS clusters. Target physical topology totals **1 hub + 3 UAT + 3 PROD = 7 clusters**; the local simulation should start smaller and incrementally represent the roles. Operators deployed in PCI/Non-PCI when cluster-scoped access or security boundaries require them. Nothing here mandates a network transit hub.

## The platform API and resource taxonomy
The **Kubernetes API Server** is the platform API. The three contract **families** are:
1. **`AWSEnvironment`**: a first-class resource whose desired state maps to exactly one AWS account, its bootstrap/foundation, security and optional management/workload EKS configuration. Support adoption of an existing account; deletion must not destroy the account by default. Long-running provisioning emits explicit `status.conditions`.
2. **`Domain`**: a business/application ecosystem. Multiple domains may exist in one environment. Model references or per-environment projections explicitly; do not assume a domain is physically nested inside another CR.
3. **Capabilities**: specialized kinds such as `Database`, `ObjectStorage`, `ComputeInstance`, `Messaging`. Each CR instance requests and represents consumption of that service; **no generic `CapabilityClaim` for v1**. A Capability is a **family**, not necessarily a `kind: Capability`.

Schemas, reference validation, authorization and reconciliation are owned by controllers and admission policy; Kubernetes does not automatically enforce cross-CRD referential integrity.

Illustrative, not installed:
```yaml
apiVersion: platform.lab.example/v1alpha1
kind: AWSEnvironment
metadata:
  name: uat
spec:
  stage: uat
  profile: standard
  clusters:
    management: true
    pci: true
    nonPCI: true
```
A workload capability may declare `environmentRef: uat` and `domainRef: payments`. API group/version and exact schemas will be finalized in implementation, **not** copied from an employer.

## End-to-end flows

### Desired-state ingestion
- Git holds reviewed manifests. Argo CD synchronizes Hub resources to the Kubernetes API.
- Backstage/CLI/AI-MCP should produce reviewed PRs for durable GitOps-managed mutations; RBAC-gated direct API reads support diagnostics and status.
- Avoid competing spec ownership between Argo CD and direct write clients. Operational state remains in Kubernetes `status`, not written back to Git.

### AWSEnvironment lifecycle
1. Platform Controller notices creation or spec change and validates identity, references, policy and account ownership.
2. It starts/reconciles a **versioned Argo Workflow**, without embedding all procedural steps in the public CR.
3. Workflow creates/adopts account; requests IP allocation (manual initially); runs Terraform network/bootstrap; verifies routes/connectivity; awaits security certificates/authorized approval; configures management and workload clusters; validates readiness.
4. Use suspend/resume for human gates with independent actor identity, evidence, decision and audit. Opening a ticket is not proof of completion. Technical checks must verify technical conditions.
5. Controller records conditions such as `AccountReady`, `NetworkReady`, `SecurityReady`, `ClustersReady`, `Ready`, and references the active workflow. Keep Terraform state/artifacts outside etcd and Git.
6. Never silently delete an AWS account if a CR is removed; require separate explicit decommission lifecycle and protective policy.

### Spoke operations
- Hub provisions and bootstraps each Environment Management EKS.
- Each spoke-local Argo CD reconciles its environment's EKS PCI and EKS Non-PCI clusters if security approval permits sharing of administrative scope; otherwise separate control.
- Cluster-local operators/agents stay in workload clusters where required. Environment Management EKS hosts **candidate** remote operators, local Argo CD, telemetry gateways and auxiliary services.
- AWS Runtime (RDS, S3, EC2, networking, managed services) belongs inside the spoke's AWS account, alongside—not below—the EKS clusters.
- Local telemetry can be forwarded to a suitable backend; PCI boundaries, retention and routing remain validation items.

## Architecture boundaries and caveats
- Git = source of declarative desired state; Kubernetes API = platform API + observed operational status; Argo CD = GitOps; Argo Workflows = procedural orchestrator; Terraform = task executor; Crossplane = continuous AWS reconciliation candidate.
- Hub controller owns platform lifecycle, not the internal step-by-step scheduling of every workflow.
- A Management EKS is not necessarily a security boundary for PCI; a single controller with credentials for PCI and Non-PCI is in both administrative scopes.
- A management-cluster outage must not be assumed to terminate running workloads; bootstrap, identity, disaster recovery and local controllers need independent recovery paths.
- No corporate references and no actual AWS operations without explicit approval.

See [decisions](adr/README.md), [repository boundaries](repository-structure.md), [roadmap](learning-roadmap.md).

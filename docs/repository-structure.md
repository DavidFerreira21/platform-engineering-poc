# Monorepo boundaries and target tree

The repository mirrors **responsibility** and **environment deployment ownership**, not one folder per Kubernetes cluster. The folders with README files are **scaffolds only**, not deployed components.

```text
platform-engineering-poc/
├── AGENTS.md
├── README.md
├── control-plane/
│   ├── crds/                 # AWSEnvironment, Domain, specialized capabilities
│   └── controllers/          # reconciliation, lifecycle, status, reference checks
├── workflow-plane/
│   ├── argo-workflows/       # long-running workflows, validations, human gates
│   ├── terraform-executor/   # runner contracts and plans
│   ├── argocd/               # hub GitOps config, spoke bootstrap integration
│   └── crossplane/           # optional future adapter/compositions
├── observability-plane/      # OTel, metrics, logging, traces, audit, alerts
├── data-plane/
│   ├── uat/
│   │   ├── management-eks/   # local Argo CD/operators/telemetry gateway
│   │   ├── eks-pci/          # PCI Kubernetes workload target
│   │   ├── eks-non-pci/      # non-PCI Kubernetes workload target
│   │   └── aws-runtime/      # AWS services in UAT account
│   └── prod/
│       ├── management-eks/
│       ├── eks-pci/
│       ├── eks-non-pci/
│       └── aws-runtime/
├── platform/
│   └── hub/                  # Platform Management Cluster bootstrap/kind
├── docs/
│   ├── adr/
│   ├── architecture.md
│   ├── learning-roadmap.md
│   └── repository-structure.md
├── labs/                     # incremental learning instructions
└── tests/                    # architecture, contracts, integration/E2E
```

## Ownership rules
- Controller code/CRD definitions: `control-plane/`. Helm/bootstrap to install them: `platform/hub/`.
- Workflow **definitions** and shared runners: `workflow-plane/`. Hub Argo CD source only here; spoke-local configuration belongs in `data-plane/{uat,prod}/management-eks/` with reusable shared components referenced instead of copied.
- Observability **common definitions**: `observability-plane/`. Spoke-specific routing/overlays under their environment's management EKS configuration.
- Infra/resource **intention** for UAT and PROD, including workload Kubernetes target settings: `data-plane/`. Reusable definitions should live in their owning domain and be referenced by environment overlays.
- The repository does not require all directories to exist from day one. Bootstrap first, then populate incremental slices.
- No `apps/control-api` or mandatory FastAPI server: Kubernetes API is authoritative. A Backstage plugin, CLI or MCP adapter can be added after the platform API contracts mature.
- No CI engine is mandated by this architecture. Choose and document one independently when implementing build/test/supply-chain automation.
- Avoid duplicate source-of-truth manifests and avoid embedding secrets, real account IDs, actual routes or certificates in YAML.

## Review checklist
Path ownership clear; UAT/PROD isolation clear; API group and schema version declared; Git spec ownership controlled; `status` reconciliation idempotent; least privilege and PCI boundaries reviewed; no live AWS change without approval; lab has validation and cleanup instructions.

# Repository structure and placement guide

This is the **target monorepo structure**. Directories are created incrementally as implementation begins.

```text
platform-engineering-poc/
├── apps/                    # Executable API, orchestrator and future interfaces
├── control-plane/           # Product contracts, catalog, schemas, domain policy
├── workflow-plane/          # Tekton, Argo Workflows, executors, GitOps integration
├── data-plane/              # Terraform modules, blueprints, target workloads
├── observability-plane/     # Instrumentation config, dashboards, alerts, audit
├── platform/                # kind, bootstrap, Helm, ingress, IDP deployments
├── docs/                    # Architecture, ADRs, guides
├── labs/                    # Incremental hands-on exercises
└── tests/                   # Cross-component integration and end-to-end tests
```

## What belongs where?

| Example | Directory | Why |
| --- | --- | --- |
| FastAPI request handler, Pydantic runtime code | `apps/control-api/` | Executable application |
| Versioned infrastructure product schema and catalog metadata | `control-plane/` | Consumer-facing contract |
| Tekton CI pipeline and reusable Tasks | `workflow-plane/tekton/` | Execution definition |
| Argo Workflows Terraform plan template | `workflow-plane/argo-workflows/` | Execution definition |
| Argo CD Application/ApplicationSet definitions | `workflow-plane/argocd/` | GitOps configuration |
| Reusable Terraform module for a consumer product | `data-plane/modules/` | Delivered infrastructure |
| Example AWS environment or consumer Kubernetes workload | `data-plane/environments/` or `data-plane/workloads/` | Data Plane target |
| kind cluster configuration and cluster lifecycle scripts | `platform/kind/` | Platform runtime/bootstrap |
| Helm values for Traefik, Argo CD, Keycloak or platform API | `platform/helm/` | Installing IDP components |
| OpenTelemetry configuration, dashboards and alerts | `observability-plane/` | Cross-cutting observability definitions |
| Lab instructions and validation steps | `labs/` | Learning path |
| End-to-end self-service request test | `tests/e2e/` | Cross-component verification |

These paths are illustrative and may be refined in implementation PRs. A single component can have **source** in one directory and **deployment** in another without duplicating the same artifact.

## Boundaries

### `apps/` vs `control-plane/`
The API server and orchestration code live in `apps/`. The versioned schemas, product definitions and declarative contracts consumed by that code live in `control-plane/`. The API should load or validate against these contracts through an explicit versioned interface.

### `platform/` vs `data-plane/`
`platform/` builds and runs the IDP itself (management and local workload cluster bootstrap, controllers, ingress and shared services). `data-plane/` represents what the IDP delivers to consumers: AWS infrastructure products and sample workloads. **Data Plane is not a synonym for the workload kind cluster**.

### `workflow-plane/` vs `platform/`
Pipeline/workflow/GitOps **definitions** go in `workflow-plane/`. Installing the Tekton/Argo controllers, their Helm charts and base cluster configuration goes in `platform/`. Argo CD's application definitions remain in `workflow-plane/` even when they deploy components defined elsewhere.

### `observability-plane/` vs `platform/`
Dashboards, instrumentation and alert definitions belong to `observability-plane/`; the runtime installation of monitoring operators/servers belongs to `platform/`.

## Planned implementation approach
1. Start with `platform/kind/` and `labs/01-bootstrap/` for the two kind clusters.
2. Add a minimal `apps/control-api/` and first `control-plane/` contract.
3. Add `workflow-plane/` components incrementally, keeping Terraform apply gated.
4. Add `data-plane/` examples, policy, secrets and telemetry as each lab needs them.

## Guardrails
- Git is the source of truth for declarative definitions.
- Never commit credentials, kubeconfigs, tokens, Terraform state or real organization identifiers.
- Use mock execution first; require explicit authorization for real AWS provisioning.
- Follow [ADR-0007](adr/0007-monorepo-structure-and-boundaries.md) and [architecture](architecture.md).

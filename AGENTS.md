# Agent instructions

## Objective
Build a reproducible Platform Engineering proof of concept, with Git as the source of truth. Follow [ADR-0007](docs/adr/0007-monorepo-structure-and-boundaries.md) and the [repository structure guide](docs/repository-structure.md).

## Architecture
- Consumer: CLI/API first; Backstage portal later.
- Control Plane: API, versioned product contracts, validation, authorization and execution tracking.
- Workflow/GitOps Plane: Tekton for CI; Argo Workflows for Terraform and imperative processes; Argo CD for Kubernetes reconciliation.
- Data Plane: AWS infrastructure products and consumer Kubernetes workloads.
- Observability Plane: cross-cutting instrumentation, metrics, logs and audit.
- Crossplane: future executor behind an interface; not required for MVP.

## Directory ownership
- `apps/`: executable services and interfaces, including FastAPI source.
- `control-plane/`: product catalog, schemas, versioned contracts and domain policies.
- `workflow-plane/`: Tekton pipelines, Argo workflow definitions, executors and Argo CD configuration.
- `data-plane/`: delivered Terraform modules, product blueprints, environments and consumer workloads.
- `observability-plane/`: telemetry definitions, dashboards, alerts and audit.
- `platform/`: kind clusters, bootstrap scripts, Helm, ingress and deployments of the IDP itself.
- `docs/`: architecture and ADRs; `labs/`: reproducible learning exercises; `tests/`: cross-component integration/E2E tests.

Do not confuse application **source** (`apps/`) with application **deployment** (`platform/`). Do not place IDP hosting resources in `data-plane/`. Controller installation belongs in `platform/`; workflow definitions belong in `workflow-plane/`. Do not create empty directories solely to match the target tree.

## Engineering conventions
- Keep modules small, independently testable and aligned to their directory owner.
- Prefer Python for API services and Terraform for initial IaC examples.
- Write tests for validation, authorization and orchestration behavior.
- Never commit credentials, tokens, state files, kubeconfigs, or real organization identifiers.
- Use mock execution by default; no real AWS changes without explicit approval.
- Propose changes on feature branches and through pull requests.
- Record architectural decisions as ADRs; update the repository guide when ownership boundaries change.

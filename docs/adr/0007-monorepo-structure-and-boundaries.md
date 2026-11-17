# ADR-0007: Monorepo structure and ownership boundaries

- **Status:** Accepted for POC
- **Date:** 2026-11-17

## Context
The POC has distinct architectural planes and needs clear code ownership. Mixing executable API code, product definitions, orchestration, consumer resources and platform installation manifests makes lifecycle management and future extraction difficult.

## Decision
Adopt a single Git repository with the following top-level directories:

| Directory | Responsibility |
| --- | --- |
| `apps/` | Executable application source: Control Plane API, orchestration service, future portal/CLI |
| `control-plane/` | Versioned product contracts, catalog definitions, domain policies and schemas |
| `workflow-plane/` | Tekton CI, Argo Workflows templates, executor adapters/configuration and Argo CD integration definitions |
| `data-plane/` | Terraform modules, blueprints, product implementations, target environment examples and consumer workload definitions |
| `observability-plane/` | Telemetry configuration, dashboards, alerts, audit schemas and observability integration |
| `platform/` | Infrastructure that **hosts the IDP**: kind clusters, bootstrap, Helm releases, ingress and platform component deployment |
| `docs/` | Architecture, ADRs, repository conventions and operational guides |
| `labs/` | Progressive, reproducible learning exercises |
| `tests/` | Cross-component integration and end-to-end tests |

Keep unit tests close to the application/module they test when appropriate; reserve root `tests/` for cross-component tests.

## Placement rules
1. Application **implementation** goes in `apps/`; its product-facing **contracts** and catalog definitions go in `control-plane/`.
2. Deployment manifests and Helm values for running an application **as part of the IDP** belong in `platform/`, even if the application source lives in `apps/`.
3. Tekton pipeline and Argo workflow **definitions** belong in `workflow-plane/`; Helm installation of their controllers belongs in `platform/`.
4. AWS resources and Kubernetes workloads **delivered to consumers** belong in `data-plane/`, not `platform/`.
5. Telemetry dashboards, collectors and alert configuration belong in `observability-plane/`; bootstrap installation wiring for the monitoring stack belongs in `platform/`.
6. Do not duplicate canonical manifests or contracts across directories. Reference reusable modules rather than copying them.
7. Future Backstage and Crossplane integrations must follow these ownership rules, without requiring a new top-level directory by default.

## Consequences
- The folder tree reflects **ownership**, not a one-to-one mapping to Kubernetes clusters or namespaces.
- Cross-plane integrations are explicit and documented; dependencies should point to stable APIs/contracts.
- Empty directories need not be committed until they contain real content. This ADR documents the target layout, not an already implemented tree.
- A monorepo simplifies coordinated POC changes; if independent release lifecycles become important, selected components can later be extracted.

## Alternatives considered
- **One folder per tool:** rejected because it obscures the distinction between platform installation, execution definitions and product delivery.
- **Separate repositories per plane:** deferred due to overhead for a learning POC.
- **Put all Control Plane code under `control-plane/`:** rejected to distinguish executable services from declarative contracts.

# ADR-0008: Monorepo mirrors planes and environment ownership

- **Status:** Accepted as POC target
- **Date:** 2026-11-18

## Context
The source tree should clearly expose the Hub, Workflow/Observability responsibilities and two Spokes without duplicating templates.

## Decision
Use control-plane/, workflow-plane/, observability-plane/, platform/hub/, data-plane/{uat,prod}/{management-eks,eks-pci,eks-non-pci,aws-runtime}/ plus docs/, labs/, tests/. Reusable implementation remains in owning plane and overlays in each spoke. Labs are progressive, observable and safe; placeholders document boundaries until implementation exists.

## Consequences and trade-offs
Clear ownership aids contribution, but boundaries are not automatic runtime isolation. Reassess path layout when real reusable infrastructure modules or controllers are introduced.

## Validation
Validate through phased labs before real AWS provisioning; update design when a constraint is disproven.

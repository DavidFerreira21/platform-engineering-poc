# Platform Engineering Lab

Personal, vendor-neutral learning POC for a **Kubernetes-native platform API** using a hierarchical **hub-and-spoke** architecture. This repository is a **design baseline**, not a claim that the stack has been provisioned.

## Architecture
- **Experience:** Backstage, CLI, GitOps (Git as the source of desired state), AI/MCP.
- **Hub — Platform Management Cluster:** Kubernetes API Server; three contract **families** (`AWSEnvironment`, `Domain`, specialized `Capability` resource kinds); Platform Controllers; logical Workflow and Observability planes.
- **Workflow:** Argo CD for GitOps, Argo Workflows for orchestration and manual gates, Terraform execution; Crossplane is a candidate additional reconciler, introduced progressively.
- **Spokes — UAT and PROD:** each represents **one separate AWS account**; each target account contains an Environment Management EKS, EKS PCI, EKS Non-PCI and AWS Runtime (network, RDS, S3, EC2 and managed services).
- **Observability:** telemetry, logs, metrics, traces and audit across hub and spokes. Collection and gateway components may run locally.

The architecture is logical and a **target topology**: its seven Kubernetes clusters need not all be deployed to learn or validate the patterns. The initial local implementation uses kind and mocks; actual EKS and AWS integration are explicit later steps.

## Start here
- [Architecture and data flows](docs/architecture.md)
- [Architecture decisions](docs/adr/README.md)
- [Repository boundaries](docs/repository-structure.md)
- [Progressive learning roadmap](docs/learning-roadmap.md)
- [Agent instructions](AGENTS.md)

## Principles
- No organization/customer names, corporate identifiers, credentials or production config.
- Git owns desired-state manifests; Kubernetes owns operational resource status; no ungoverned direct changes to GitOps-managed specs.
- Native Kubernetes API/CRDs instead of a mandatory custom REST API; use `spec` and `status.conditions`.
- One environment = one AWS account. UAT and PROD are separate spokes; PCI and Non-PCI are clusters **within** each spoke.
- Separate intended capability from implementation engine. No generic `CapabilityClaim` for the initial design.
- Human dependencies are explicit workflow gates with evidence, authorized resume and technical verification; never equate a ticket opened with work completed.
- No real AWS provisioning, destructive apply or secrets without explicit authorization.
- Each lab is progressive, observable, reproducible, and has a safe failure/recovery exercise.

**Current state:** architecture and learning plan only. Implementation folders contain scope notes, not running components.

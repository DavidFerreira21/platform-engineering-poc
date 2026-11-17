# Platform Engineering POC

Proof of concept for an Internal Developer Platform (IDP), focused on self-service infrastructure and clear architectural boundaries.

## Architecture
- **Control Plane:** API, contracts, validation and orchestration.
- **Workflow/GitOps Plane:** Argo Workflows executes Terraform; Argo CD reconciles platform components.
- **Data Plane:** AWS foundations and resources in isolated target environments.
- **Observability Plane:** planned for later.

Crossplane will be introduced later through an executor interface; the MVP starts with Terraform.

See [architecture](docs/architecture.md), [repository structure](docs/repository-structure.md), [learning roadmap](docs/learning-roadmap.md), [ADRs](docs/adr/README.md) and [agent instructions](AGENTS.md).

## Milestones
1. Bootstrap repository, documentation and conventions.
2. Build a Control Plane API with request validation and a mock executor.
3. Integrate Argo Workflows and Terraform plan.
4. Add approval gates and controlled apply.
5. Integrate Argo CD, then a Crossplane executor.
6. Extend observability, security policies and developer experience.

## Principles
- Git is the source of truth.
- No credentials, real infrastructure identifiers or Terraform state in Git.
- Mock execution first; real AWS changes require explicit authorization.
- Keep local development reproducible.

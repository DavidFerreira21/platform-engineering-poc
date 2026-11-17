# ADR-0005: Platform runtime, persistence and AWS integrations

- **Status:** Accepted for POC
- **Date:** 2026-11-16

## Decision
- **Python + FastAPI + Pydantic** for the Control Plane API and versioned product contracts.
- **PostgreSQL** for operational platform data: requests, actors, execution metadata, state transitions and audit references. Git remains the source of truth for declarative product definitions.
- **Terraform** for the initial infrastructure executor; AWS resources are provisioned only after explicit target selection and approval.
- **Amazon ECR** as the container image registry for images produced by Tekton.
- **Traefik** as ingress for local platform services.
- **Helm** for packaging and initial installation of platform components; decide which components Argo CD manages after bootstrap.

## Boundaries
PostgreSQL is the **Control Plane application's** database; it is **not** a database required by Crossplane. Separate platform management services from the AWS and workload resources offered as products. Do not hardcode AWS account numbers, credentials or production endpoints.

## Consequences
ECR and AWS Secrets Manager are managed AWS dependencies with costs and access requirements. A local/offline mock path must remain available before AWS credentials are configured. PostgreSQL backup/restore, migrations and retention will be designed when persistence is implemented.

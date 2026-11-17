# Architecture Decision Records

ADRs document **why** a decision was made, its scope, trade-offs and consequences. These records describe the target POC architecture; they do **not** imply that components are already deployed.

| ADR | Decision | Status |
| --- | --- | --- |
| [0001](0001-kubernetes-topology.md) | Two local kind clusters and four logical planes | Accepted for POC |
| [0002](0002-delivery-and-orchestration.md) | Tekton CI, Argo Workflows execution, Argo CD GitOps | Accepted for POC |
| [0003](0003-security-and-secrets.md) | Kyverno, Trivy, AWS Secrets Manager, ESO | Accepted for POC |
| [0004](0004-observability.md) | OpenTelemetry, Prometheus, Grafana, Loki | Accepted for POC |
| [0005](0005-platform-runtime-and-data.md) | FastAPI, PostgreSQL, Traefik, Helm, Terraform, ECR | Accepted for POC |
| [0006](0006-identity-and-future-capabilities.md) | Keycloak SSO and deferred capabilities | Accepted for POC (identity) |

## Conventions
- Status **Accepted for POC** records a design choice, not a production endorsement.
- Production use requires separate security, reliability, licensing and cost assessments.
- Architecture is reviewed via pull requests; superseded decisions receive a new ADR.
- No production secrets, cloud account IDs or Terraform state in the repository.

## Implementation sequence
1. Establish two kind clusters and a reproducible platform bootstrap.
2. Build minimal FastAPI contract and mock execution with structured logs.
3. Add Tekton CI and Trivy, publish images to ECR when AWS access is configured.
4. Add Argo Workflows Terraform plan; require explicit approval before apply.
5. Add Argo CD GitOps, Kyverno and ESO integrations.
6. Expand observability, policy and identity; defer Crossplane and Backstage.

# Progressive learning roadmap — Platform Engineering POC

## Purpose
Build an Internal Developer Platform **incrementally**, learning how each component works instead of installing the entire stack in one bootstrap. Each lab must produce a visible, testable outcome and a deliberate failure/recovery experiment.

This is a learning sequence, not a claim that the platform is deployed. Architecture and tool decisions remain in [ADRs](adr/README.md); directory ownership is defined in [repository structure](repository-structure.md).

## Conceptual learning objectives

Each lab must distinguish **Platform Engineering concepts** from the technologies used to demonstrate them. Tools are means to deliver platform capabilities, not the learning objective. The book *Platform Engineering in Practice* can inform discussions, but the lab order and outcomes are defined by this POC; do not imply that the book prescribes this sequence or cite chapters without checking the source.

| Lab | Platform Engineering concepts | Technology demonstration | Core question |
| --- | --- | --- | --- |
| 01 | Platform foundations, abstractions, reconciliation, operational ownership | Kubernetes, kind, Traefik | What capability does this foundation provide beyond a running cluster? |
| 02 | Platform as a Product, self-service, API-first, versioned contracts | FastAPI, Pydantic, mock executor | What does the consumer request, and what implementation details stay hidden? |
| 03 | Central identity, authentication vs authorization, least privilege | Keycloak, OIDC, PostgreSQL | Who can request, approve and observe each product? |
| 04 | Developer Experience, feedback loops, software supply chain, CI | Tekton, Trivy, ECR | How does the platform provide reliable build feedback and guardrails? |
| 05 | Orchestration, execution abstraction, separation of concerns | Argo Workflows, Terraform plan | Why should the API not run Terraform directly? |
| 06 | Governance, policy, approvals, auditability | Authorization, approval workflow, execution records | How do we enable autonomy without removing controls? |
| 07 | Desired state, GitOps, reconciliation, drift, Control/Data Plane boundary | Argo CD, two kind clusters | Who owns desired state and how does the system correct drift? |
| 08 | Policy as Code, secrets lifecycle, least privilege | Kyverno, ESO, Secrets Manager | Which controls are guardrails versus manual gates? |
| 09 | Observability as a platform capability, SLI/SLO basics, incident diagnosis | OTel, Prometheus, Grafana, Loki | Can a consumer and operator understand a request failure? |
| 10 | Golden Paths, platform product lifecycle, discoverability, adoption and DX | End-to-end self-service flow | Is this a reusable product capability or merely a collection of installed tools? |

For each concept, capture a concise definition, why it matters, the concrete architectural decision, an observable experiment, and a short trade-off. Avoid equating product success with a green Kubernetes deployment.

## Learning method
Each lab must include:
1. **Understand:** problem, architecture, boundaries, key Platform Engineering concepts and expected behavior.
2. **Build:** introduce only the components needed for this lab; inspect the manifests and resources.
3. **Observe:** run commands, view controller state, events, logs, API responses and/or UI.
4. **Break and diagnose:** intentionally cause a safe failure, explain symptoms and recover.
5. **Validate:** demonstrate acceptance criteria, explain the platform capability delivered, and answer review questions.
6. **Automate and reset:** only after the manual path is understood, add reproducible commands and cleanup/reset instructions.

Use small PRs; separate infrastructure/tool installation from meaningful behavioral exercises when a change would otherwise be too large. No real AWS apply without explicit approval. Never commit credentials, kubeconfigs, tokens or Terraform state.

## Lab progression

| Lab | Theme | Prerequisites | Visible outcome | Failure/recovery experiment | Completion gate |
| --- | --- | --- | --- | --- | --- |
| 01 | Kubernetes foundation and Traefik | Docker-compatible runtime, kind, kubectl, Helm | `platform-management` cluster; namespaces, test Deployment/Service and ingress reachable | Delete pod, inspect recreation and events; troubleshoot bad Service selector | Explain Deployment, Service, ingress and reconciliation; reproduce cluster |
| 02 | Control Plane API and product contracts | Lab 01 | FastAPI request validation and mock execution status/logs | Reject invalid schema; simulate failed execution | Valid request advances through defined states; invalid request never executes |
| 03 | PostgreSQL, Keycloak and OIDC | Lab 02 | Persist request state; authenticate users and test roles | Expired/invalid token, denied role, restart API and verify persistence | Demonstrate authentication vs authorization and persisted requests |
| 04 | Tekton CI, Trivy and ECR | Labs 01–03; AWS access only for ECR publishing | Pipeline test/build/scan; image published when authorized | Failing test or scan blocks publish | Trace TaskRuns and explain CI vs workflow orchestration |
| 05 | Argo Workflows and Terraform plan | Labs 02–04 | API-triggered workflow; Terraform plan artifact and status | Invalid Terraform or failed workflow step; inspect retries/logs | Explain executor contract; plan runs without apply |
| 06 | Approval and governance | Labs 03 and 05 | Separate request, approval and controlled apply lifecycle | Unauthorized approver or rejected request blocks apply | No apply without explicit authorized approval; audit trail exists |
| 07 | Argo CD GitOps and workload cluster | Labs 01, 02 and 06 | Create `platform-workload`, register with Argo CD, deploy sample consumer workload | Manual drift and reconciliation; broken manifest | Show desired vs actual state and recover safely |
| 08 | Kyverno and secrets | Labs 03 and 07; safe AWS identity mechanism | Admission policy enforcement and ESO/Secrets Manager integration | Reject noncompliant workload; observe safe secret refresh | Demonstrate policy denial and no secrets in Git |
| 09 | Observability end to end | Labs 02, 05 and 07 | OTel instrumentation, Prometheus, Grafana, Loki; correlate request/workflow logs | Trigger failure and locate it via correlation ID | Diagnose a request using metrics and logs; document trace backend as future choice |
| 10 | Golden Path and platform as a product | Labs 01–09 | Consumer requests generic product through API; contract → approval → execution → delivery → visibility | Invalid request, denied approval and execution failure | Demonstrate full lifecycle, ownership and recovery |

### Sequencing notes
- The final target is **two kind clusters**, but Lab 01 intentionally creates only `platform-management`; Lab 07 introduces `platform-workload` to make multi-cluster GitOps an observable learning milestone.
- Observability is **transversal from Lab 01**: basic Kubernetes events/logs, structured application logs and correlation IDs begin early. Lab 09 expands to the full observability stack, rather than introducing visibility for the first time.
- Lab 03 introduces centralized identity early so that later approval and audit exercises use real user identities. Individual UIs integrate OIDC as they are introduced; do not assume all tools have identical SSO/RBAC behavior.
- Lab 04 may run tests and scanning locally before AWS ECR credentials are available; publishing to ECR is an explicitly authorized extension, not a prerequisite for local learning.
- Lab 05 begins with `terraform plan` only. Lab 06 adds controlled apply; AWS resource provisioning remains optional and requires explicit approval and scoped credentials.
- Lab 08 must not assume local kind has IRSA. Design and validate a suitable AWS authentication method before integrating live Secrets Manager.
- No Backstage, Crossplane, enterprise multi-tenancy or advanced AI/SRE automation is required for Lab 10.

## Repository convention

```text
labs/
  01-kubernetes-foundation/
  02-control-plane/
  03-identity-and-persistence/
  04-tekton-ci/
  05-argo-workflows/
  06-approvals-governance/
  07-argocd-gitops/
  08-security-secrets/
  09-observability/
  10-golden-path/
```

These are **planned** lab directories; create each when its lab starts. Lab READMEs should reference canonical source under `apps/`, `control-plane/`, `workflow-plane/`, `data-plane/`, `observability-plane/` and `platform/`, not copy those sources.

## Standard lab README template
- Goal and architecture diagram (before/after)
- **Platform Engineering concepts:** definition, motivation, architectural manifestation and trade-off
- **Technical implementation:** tools, components and why they were selected
- **Practical evidence:** experiment proving where each concept appears in the running platform
- Concepts and prerequisite checks
- What will change in the platform
- Step-by-step manual actions and expected outputs
- Observation checklist: resources, events, logs, API/UI
- Safe failure injection, troubleshooting and recovery
- Acceptance checklist and review questions, including: **Is this a reusable platform capability or just an installed tool?**
- Reproducible automation and cleanup/reset
- Links to source directories, ADRs and related PRs

## Next PR after roadmap approval
Implement **Lab 01 only**: local prerequisites, one kind management cluster, namespace layout, Traefik and a disposable test service. Do not install the complete tool stack or create AWS resources.

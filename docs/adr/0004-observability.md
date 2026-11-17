# ADR-0004: Cross-cutting observability plane

- **Status:** Accepted for POC
- **Date:** 2026-11-16

## Context
Platform requests span API, CI, workflow engines, GitOps reconciliation and infrastructure targets. Execution status alone is insufficient for diagnosis.

## Decision
- **OpenTelemetry** for instrumentation and standardized telemetry export.
- **Prometheus** for metrics.
- **Grafana** for dashboards and investigation.
- **Loki** for centralized logs.
- Observability is a **cross-cutting logical plane**, initially hosted on `platform-management`, not a third Kubernetes cluster.
- Begin with structured JSON logs, stable request/execution IDs, correlation across components, and basic health/metrics endpoints.
- Add distributed tracing backend (e.g., Tempo) only when needed; OpenTelemetry does not itself store traces.

## Consequences
Instrument Control Plane requests and Terraform workflow executions first; expand to Tekton, Argo CD, policy decisions and consumer workloads over time. Define retention, redaction, access controls and resource budgets. Logs must never include secrets or Terraform sensitive outputs.

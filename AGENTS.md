# Agent instructions

## Objective
Build a small, reproducible Platform Engineering proof of concept, with Git as the source of truth.

## Architecture
- Consumer: CLI/API first; Backstage portal later.
- Control Plane: validates requests, manages contracts, and dispatches executions.
- Workflow/GitOps Plane: Argo Workflows runs Terraform; Argo CD manages declarative components.
- Data Plane: AWS foundations and resources in isolated accounts/environments.
- Crossplane: future executor behind an interface; not required for MVP.
- Observability Plane: planned, not part of initial MVP.

## Engineering conventions
- Keep modules small and independently testable.
- Prefer Python for API/control-plane services and OpenTofu/Terraform for IaC examples.
- Write tests for validation and orchestration behavior.
- Never commit credentials, tokens, state files, kubeconfigs, or real organization identifiers.
- Use mock execution by default; no real AWS changes without explicit approval.
- Propose changes on feature branches and through pull requests.
- Document assumptions and decisions in docs/architecture.md.

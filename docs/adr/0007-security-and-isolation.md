# ADR-0007: Per-environment boundaries, secure identities and safe deletion

- **Status:** Accepted as POC target
- **Date:** 2026-11-18

## Context
UAT and PROD are separate AWS accounts. An Argo CD with control over both PCI and Non-PCI may cross a PCI administrative boundary.

## Decision
Separate credentials and least-privilege federated roles for Hub/spokes, environments and service identities. Enforce Kubernetes RBAC/admission, restricted Git PR approvals and independent manual-gate authorization. Do not assume namespaces are hard boundaries or that one local Argo CD is PCI-compliant without security validation. No secrets, real IDs or state in Git. Deleting an AWSEnvironment CR must not automatically delete its AWS account.

## Consequences and trade-offs
Management EKS may need distinct control scopes for PCI. Real AWS operations, accounts, expenses and destructive actions remain opt-in. Identity provider and secret provider are implementation choices, not hardcoded for this baseline.

## Validation
Validate through phased labs before real AWS provisioning; update design when a constraint is disproven.

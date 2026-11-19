# ADR-0005: AWSEnvironment runs long-lived workflows with human dependencies

- **Status:** Accepted as POC target
- **Date:** 2026-11-18

## Context
Creating an account requires IP allocations, firewall/routes, certificates and third-party manual tasks that may last days.

## Decision
AWSEnvironment expresses outcomes/profile, not procedural tasks. Controller starts a versioned Argo Workflow with Terraform steps, bounded technical validations, suspend/resume manual gates and evidence checks. Track step status plus conditions AccountReady, NetworkReady, SecurityReady, ClustersReady, Ready. Ticket creation alone never satisfies a gate. Support pre-existing account adoption and require explicit decommission to destroy an AWS account.

## Consequences and trade-offs
Workflow retries/idempotency, audit of approvers, pause duration, evidence integrity, cross-system checks, status references, and recovery/retention policies are needed. No ServiceNow integration in v1.

## Validation
Validate through phased labs before real AWS provisioning; update design when a constraint is disproven.

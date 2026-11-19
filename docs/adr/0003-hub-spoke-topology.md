# ADR-0003: Hierarchical Hub-and-Spoke with two account spokes

- **Status:** Accepted as POC target
- **Date:** 2026-11-18

## Context
The personal POC needs an organization-neutral target reflecting separation between global coordination and environment-local operation.

## Decision
One Platform Management Cluster (Hub) contains logically co-located Control, Workflow and Observability planes. Data Plane has only UAT and PROD spokes, each exactly one AWS account. Each spoke's target layout contains Environment Management EKS, EKS PCI, EKS Non-PCI and AWS Runtime (RDS/S3/EC2/network/managed services). PCI and Non-PCI are not separate spokes. Model with kind/mocks before real EKS.

## Consequences and trade-offs
Target is seven Kubernetes clusters, not necessarily seven deployed clusters during learning. Hub-spoke means management, not assumed routing; segregation across accounts and environments remains explicit.

## Validation
Validate through phased labs before real AWS provisioning; update design when a constraint is disproven.

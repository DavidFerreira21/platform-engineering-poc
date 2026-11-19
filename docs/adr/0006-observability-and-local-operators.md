# ADR-0006: Cross-cutting telemetry and environment-local management

- **Status:** Accepted as POC target
- **Date:** 2026-11-18

## Context
Monitoring and operators can need access close to an isolated spoke's resources, but global diagnosis must remain possible.

## Decision
Observability Plane is logical and spans Hub and Spokes: OpenTelemetry, metrics, logs/traces, audit and alerts. Each Environment Management EKS can host a telemetry gateway and remotely capable operators/controllers. Each PCI/Non-PCI EKS retains cluster-local agents/operators when required; backends/storage and data-routing are separate design decisions.

## Consequences and trade-offs
Operator suitability must be validated per product/API and authorization. PCI telemetry routing, sensitivity, aggregation, offline operation and resource consumption require tests.

## Validation
Validate through phased labs before real AWS provisioning; update design when a constraint is disproven.

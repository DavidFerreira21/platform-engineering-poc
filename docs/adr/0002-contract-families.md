# ADR-0002: Three contract families and specialized capabilities

- **Status:** Accepted as POC target
- **Date:** 2026-11-18

## Context
The ecosystem must represent AWS account environments, business domains and provisionable products without a generic catch-all request object.

## Decision
Use three logical families: AWSEnvironment (exactly one AWS account per environment, foundations and cluster profile); Domain (business/application ecosystem, many domains can share an account); Capability (specialized kinds such as Database, ObjectStorage, ComputeInstance and Messaging). Each specialized CR represents its own consumption. No generic CapabilityClaim in v1; Capability does not imply kind Capability. Cross-resource links use explicit refs validated by controllers.

## Consequences and trade-offs
Simpler typed API and fewer indirections, but more CRD schemas; must design Domain identity across environments, refs, deletion protection and lifecycle semantics carefully.

## Validation
Validate through phased labs before real AWS provisioning; update design when a constraint is disproven.

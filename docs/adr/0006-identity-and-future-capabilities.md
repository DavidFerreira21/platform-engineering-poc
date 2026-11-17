# ADR-0006: Centralized identity with Keycloak and deferred capabilities

- **Status:** Accepted for POC (identity); future capabilities deferred
- **Date:** 2026-11-17

## Context
The platform needs centralized human identities rather than separately maintained users for Argo CD, Grafana, the Control Plane and other tools. Authentication (who the user is) must be distinct from authorization (what they may do).

## Decision
- Adopt **Keycloak** as the self-hosted Identity Provider (IdP) for the POC.
- Use **OpenID Connect (OIDC)** for Single Sign-On wherever a component supports it, including the Control Plane API, Argo CD, Grafana and Argo Workflows UI/server where integration is supported.
- Use a dedicated `platform` realm for the lab, with initial groups `platform-admins`, `platform-developers`, `platform-viewers`. Map claims to **application-specific** permissions; SSO does not grant universal admin rights.
- Store Keycloak data in PostgreSQL with a separate logical database and dedicated database credentials from the Control Plane API.
- Run Keycloak in the management cluster. Expose only required endpoints, use appropriate TLS for OIDC redirect flows, and protect the administrative interface.
- Tekton pipelines use Kubernetes identities and RBAC for execution; human access is mediated through the Control Plane or a separately authenticated interface rather than assuming Tekton itself provides a shared user login.
- Keep **workload/service identities** separate from human SSO: Kubernetes ServiceAccounts, least-privilege AWS IAM and AWS Secrets Manager/ESO remain the machine-to-machine controls.

## Consequences
- One identity per human can be used across integrated interfaces, but each application still requires explicit OIDC client configuration and authorization mapping.
- Keycloak availability becomes a login dependency. Backup, recovery, secret handling and session/token lifetime must be considered.
- Bootstrap admin credentials must never be committed to Git. Development-only authentication shortcuts must not become production defaults.
- Initial realm/groups are a lab starting point, not a complete enterprise RBAC model.

## Deferred capabilities
- **Backstage** until Control Plane contracts and product lifecycle stabilize.
- **Crossplane** until Terraform execution and the executor interface are proven.
- AI-assisted SRE, advanced tracing, artifact signing and enterprise multi-tenancy.

## Follow-up decisions
1. Define fine-grained roles, OIDC client mappings and authorization policy for each application.
2. Define catalog schema, product lifecycle and request/approval contracts.
3. Select GitOps bootstrap ownership and an AWS identity mechanism for local kind.

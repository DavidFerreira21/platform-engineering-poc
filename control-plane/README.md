# 🧩 Control Plane

O **Kubernetes API Server** é a API da plataforma para produtos técnicos Kubernetes-native. **Crossplane central** implementará XRDs, Compositions e Providers para **Capabilities** AWS.

| Conceito | Forma na POC |
| :--- | :--- |
| Entity | Registro de organização no Git |
| Environment | Registro YAML de UAT/PROD e conta/região AWS, não CRD |
| Domain | Registro lógico de ownership/consumo, não XRD |
| `ObjectStorage` | Primeira Capability via XRD + Composition + Provider |
| `Database` | Segunda Capability: **Aurora PostgreSQL**, dependente de conta UAT autorizada, VPC e controle de custo |
| `Messaging`, `ComputeInstance` | Capabilities posteriores |

Crossplane **não observa diretamente Git**. Argo CD central sincroniza somente CRs de Capability com a API. A definição dos **perfis versionados de Domain Bootstrap** e dos contratos de Capability pertence à plataforma. O Workflow Plane gera os manifests obrigatórios por Domain/Environment; Crossplane continua sendo o reconciliador de AWS, e não o executor do Bootstrap. A seleção de credenciais precisa vir de mapeamento validado; Domain não escolhe uma role arbitrária. O Argo CD local possui os objetos Kubernetes dos workloads e namespaces.

## TerraformRun (proposta)

A API do Hub também prevê um **`TerraformRun` CRD e controller próprios** para solicitar a execução de módulos Terraform de Foundation/Entity Core. O controller valida a requisição e aciona **Argo Workflows + Terraform Executor** (plan → gate → apply → status), com idempotência e isolamento; **não** é uma Capability XRD nem executa `apply` continuamente a cada reconcile.

**Admission / Policies:** Kyverno/OPA são tecnologias a avaliar. O contrato/controller TerraformRun será desenvolvido em lab futuro (08B); não está implementado neste repositório. [ADR-0006](../docs/adr/0006-terraformrun-and-application-ci.md).

O [Lab 03](../labs/03-crossplane-core.md) instala o Crossplane Core no Hub via Argo CD e valida um XRD namespaced que compõe um ConfigMap local.

[← Início](../README.md) · [Arquitetura](../docs/architecture.md)

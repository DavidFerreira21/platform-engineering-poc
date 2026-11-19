# 🧩 Control Plane

O **Kubernetes API Server** é a API da plataforma para produtos técnicos Kubernetes-native. **Crossplane central** implementará XRDs, Compositions e Providers para **Capabilities** AWS.

| Conceito | Forma na POC |
| :--- | :--- |
| Entity | Registro de organização no Git |
| Environment | Registro YAML de UAT/PROD e conta/região AWS, não CRD |
| Domain | Registro lógico de ownership/consumo, não XRD |
| `ObjectStorage` | Primeira Capability via XRD + Composition + Provider |
| `Database`, `Messaging`, `ComputeInstance` | Capabilities futuras |

Crossplane **não observa diretamente Git**. Argo CD central sincroniza somente CRs de Capability com a API. A definição dos **perfis versionados de Domain Bootstrap** e dos contratos de Capability pertence à plataforma. O Workflow Plane gera os manifests obrigatórios por Domain/Environment; Crossplane continua sendo o reconciliador de AWS, e não o executor do Bootstrap. A seleção de credenciais precisa vir de mapeamento validado; Domain não escolhe uma role arbitrária. O Argo CD local possui os objetos Kubernetes dos workloads e namespaces.

> [!NOTE]
> **Estado:** desenho/documentação; nada instalado automaticamente.

[← Início](../README.md) · [Arquitetura](../docs/architecture.md)

# ⚙️ Workflow Plane

O Workflow Plane inclui o **Domain Bootstrap**, que valida o perfil versionado e gera manifests obrigatórios para cada Domain/Environment em PR; ele não provisiona recursos diretamente e não concorre com Argo CD/Crossplane. A verificação da prontidão deve agregar evidências de ambos. 

O Workflow Plane mantém processos, GitOps central, aprovação e execução, **sem Crossplane**.

| Componente | Papel |
| :--- | :--- |
| **Argo CD central** | Sincroniza CRs Kubernetes de Capability na API do Hub |
| **Argo Workflows** | Orquestra passos, gates, espera e auditoria de Foundation/Core |
| **Terraform Executor** | `plan`, aprovação e `apply` de Foundation/EKS, com state seguro |

`Entity`, `Environment` e `Domain` são registros lógicos no Git. Não exigem Environment Controller próprio ou XRD de Domain nesta fase. A criação de contas será reconsiderada no futuro quando houver necessidade de lifecycle automatizado. O **Argo CD local** roda em cada Management EKS autorizado, junto do Data Plane.

> [!NOTE]
> **Estado:** documentação; nenhum componente provisionado.

[← Início](../README.md) · [Arquitetura](../docs/architecture.md)

# ⚙️ Workflow Plane

O Workflow Plane mantém processos, GitOps central, aprovação e execução. O Argo Workflows é instalado no Hub após o Argo CD central. A Etapa 03 publica a interface em `https://workflows.localhost:8443` pelo Traefik e usa um token de leitura temporário para observar as execuções. [Guia da Etapa 03](../labs/03-argo-workflows.md)

| Componente | Papel |
| :--- | :--- |
| **Argo CD central** | Sincroniza CRs Kubernetes de `TerraformStack` e de Capability na API do Hub |
| **Argo Workflows** | Orquestra a pipeline canônica, a avaliação OPA e o gate de aprovação. O Simulator usa esses steps na Etapa 06 |
| **Terraform Executor** | Imagem genérica de ferramentas que executa os scripts de Terraform do WorkflowTemplate |

No Hub `kind`, Argo CD central é autogerenciado. No Hub EKS, adotar EKS Capability for Argo CD para sincronização local e cross-account por ARN e Access Entries, sem rotas entre VPCs. [ADR-0008](../docs/adr/0008-managed-argocd-connectivity.md).

`Entity`, `Environment` e `Domain` são registros lógicos no Git. Não exigem Environment Controller próprio ou XRD de Domain nesta fase. A criação de contas será reconsiderada no futuro quando houver necessidade de lifecycle automatizado. O **Argo CD local** roda em cada Management EKS autorizado, junto do Data Plane.

## Aplicações e execuções declarativas

- **TerraformStack.** O controller Go fica no Control Plane e coordena o fluxo por seu adapter Argo. WorkflowTemplate, Terraform Executor, scripts e Policy Engine formam a implementação de referência confiável da POC; não são dependências compiladas do controller. [Contrato e fronteira](../docs/adr/0009-terraformstack-runtime.md). A Etapa 13 valida o `plan`, a aprovação humana e o `apply` do mesmo plano. A Etapa 06 exercita as decisões OPA e drift localmente. As Etapas 28–29 validam esses mecanismos na AWS. `TerraformRun` permanece um nome opcional para a execução individual.
- **Domain Applications:** GitHub Actions fará o **CI/CD em um repositório de aplicação de teste separado**, a integrar durante o lab correspondente. Build/test/scan/publicação → atualização GitOps revisada → Argo CD **local** no Management EKS faz deploy nos clusters autorizados. Não adicionar código ou workflow CI da aplicação neste monorepo.

[ADR-0006](../docs/adr/0006-terraformrun-and-application-ci.md)

O Workflow Plane usa o Hub local. A integração com o Terraform Executor faz parte da [Fase 2 do roteiro](../docs/learning-roadmap.md#fase-2--control-plane-local).

[← Início](../README.md) · [Arquitetura](../docs/architecture.md)

**Policy Engine:** [módulos Rego e perfis YAML](policy-engine/README.md) compõem as regras por stack, selecionadas pelo catálogo confiável.

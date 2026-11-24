# ⚙️ Workflow Plane

O Workflow Plane mantém processos, GitOps central, aprovação e execução. O Argo Workflows é instalado no Hub após o Argo CD central. A Etapa 03 publica a interface em `https://workflows.localhost:8443` pelo Traefik e usa um token de leitura temporário para observar as execuções. [Guia da Etapa 03](../labs/03-argo-workflows.md)

| Componente | Papel |
| :--- | :--- |
| **Argo CD central** | Sincroniza CRs Kubernetes de `TerraformStack` e de Capability na API do Hub |
| **Argo Workflows** | Executa fluxos locais na Etapa 03, integra o executor simulado na Etapa 06 e coordena o primeiro plan e apply aprovados na Etapa 09 |
| **Terraform Executor** | `plan`, aprovação e `apply` de Foundation/EKS, com state seguro |

`Entity`, `Environment` e `Domain` são registros lógicos no Git. Não exigem Environment Controller próprio ou XRD de Domain nesta fase. A criação de contas será reconsiderada no futuro quando houver necessidade de lifecycle automatizado. O **Argo CD local** roda em cada Management EKS autorizado, junto do Data Plane.

## Aplicações e execuções declarativas (planejadas)

- **TerraformStack (proposto).** O controller fica no Control Plane e coordena o fluxo executado pelo Argo Workflows. A Etapa 09 valida o `plan`, a aprovação humana e o `apply` do mesmo plano. As Etapas 18 e 19 acrescentam detecção de drift e reconciliação por políticas. `TerraformRun` permanece um nome opcional para a execução individual.
- **Domain Applications:** GitHub Actions fará o **CI/CD em um repositório de aplicação de teste separado**, a integrar durante o lab correspondente. Build/test/scan/publicação → atualização GitOps revisada → Argo CD **local** no Management EKS faz deploy nos clusters autorizados. Não adicionar código ou workflow CI da aplicação neste monorepo.

[ADR-0006](../docs/adr/0006-terraformrun-and-application-ci.md)

O Workflow Plane usa o Hub local. A integração com o Terraform Executor faz parte da [Fase 2 do roteiro](../docs/learning-roadmap.md#fase-2--terraformstack-e-primeira-execução-na-aws).

[← Início](../README.md) · [Arquitetura](../docs/architecture.md)

# ⚙️ Workflow Plane

A **proposta adiada** de Domain Bootstrap previa um gerador que valida o perfil versionado e gera a estrutura e os manifests iniciais de cada novo Domain, para os Environments selecionados, em PR; ele não provisiona recursos diretamente e não concorre com Argo CD/Crossplane. A verificação da prontidão deve agregar evidências de ambos. 

O Workflow Plane mantém processos, GitOps central, aprovação e execução, **sem Crossplane**. **Argo Workflows é instalado logo após Argo CD ([Etapa 03](../labs/03-argo-workflows.md))**, antes de Crossplane; as Etapas 10 e 11 aprofundam a integração com Terraform Executor e aprovações.

| Componente | Papel |
| :--- | :--- |
| **Argo CD central** | Sincroniza CRs Kubernetes de Capability na API do Hub |
| **Argo Workflows** | Instalado via GitOps no Hub na Etapa 03. Começa com uma execução local e evolui para passos, gates, espera e auditoria de Foundation e Core nas Etapas 10 e 11 |
| **Terraform Executor** | `plan`, aprovação e `apply` de Foundation/EKS, com state seguro |

`Entity`, `Environment` e `Domain` são registros lógicos no Git. Não exigem Environment Controller próprio ou XRD de Domain nesta fase. A criação de contas será reconsiderada no futuro quando houver necessidade de lifecycle automatizado. O **Argo CD local** roda em cada Management EKS autorizado, junto do Data Plane.

## Aplicações e execuções declarativas (planejadas)

- **TerraformStack (proposto):** o controller fica no Control Plane; **Argo Workflows** executa `plan`, avaliação de policies, gate e `apply` quando autorizado. O controller começa com execução simulada na Etapa 10. A Etapa 11 valida replan, detecção de drift e as decisões de aprovação, bloqueio e auto apply autorizado. `TerraformRun` permanece um nome opcional para a execução individual.
- **Domain Applications:** GitHub Actions fará o **CI/CD em um repositório de aplicação de teste separado**, a integrar durante o lab correspondente. Build/test/scan/publicação → atualização GitOps revisada → Argo CD **local** no Management EKS faz deploy nos clusters autorizados. Não adicionar código ou workflow CI da aplicação neste monorepo.
- **Domain Bootstrap:** conceito preservado no ADR-0005, mas **adiado e em reavaliação**. Não torná-lo pré-requisito das etapas.

[ADR-0006](../docs/adr/0006-terraformrun-and-application-ci.md)

O Workflow Plane usa o Hub local. Consulte a [Fase 3 do roteiro](../docs/learning-roadmap.md#fase-3--workflow-plane-e-execução-autorizada) para a integração com Terraform Executor.

[← Início](../README.md) · [Arquitetura](../docs/architecture.md)

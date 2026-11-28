# ⚙️ Workflow Plane

O Workflow Plane mantém processos, GitOps central, aprovação e execução. O Argo Workflows é instalado no Hub após o Argo CD central. A Etapa 03 publica a interface em `https://workflows.localhost:8443` pelo Traefik e usa um token de leitura temporário para observar as execuções. [Guia da Etapa 03](../labs/03-argo-workflows.md)

| Componente | Papel |
| :--- | :--- |
| **Argo CD central** | Sincroniza CRs Kubernetes de `TerraformStack` e de Capability na API do Hub |
| **Argo Workflows** | Executa os WorkflowTemplates cadastrados; a referência inclui avaliação OPA e gate de aprovação. O Simulator usa esses steps na Etapa 06 |
| **Terraform Executor** | Imagem genérica com Terraform, OpenTofu e as ferramentas usadas pelos scripts do WorkflowTemplate |

No Hub `kind`, Argo CD central é autogerenciado. No Hub EKS, adotar EKS Capability for Argo CD para sincronização local e cross-account por ARN e Access Entries, sem rotas entre VPCs. [ADR-0008](../docs/adr/0008-managed-argocd-connectivity.md).

`Entity`, `Environment` e `Domain` são registros lógicos no Git. Não exigem Environment Controller próprio ou XRD de Domain nesta fase. A criação de contas será reconsiderada no futuro quando houver necessidade de lifecycle automatizado. O **Argo CD local** roda em cada Management EKS autorizado, junto do Data Plane.

## Aplicações e execuções declarativas

- **TerraformStack.** O controller Go fica no Control Plane e coordena o fluxo pelo adapter Argo. O CR seleciona um alias cadastrado ou o default. Cada integração define namespace, template, ServiceAccount e handshake de autorização (habilitado por padrão). Outros templates compatíveis podem escolher suas ferramentas e gates. O controller não interpreta seus steps nem permite duas tentativas ativas por CR. A pipeline de referência resolve Git, contexto, backend, variáveis e política. WorkflowTemplate, runner, scripts e Policy Engine implementam o [Execution Contract](../docs/flows/terraformstack-execution-contract.md). A Etapa 06 comprova as decisões OPA, o claim e o drift localmente. A Etapa 13 prevê o primeiro apply da Network UAT, com aprovação vinculada ao plano. As Etapas 28–29 exercitam drift e políticas na AWS.
- **Domain Applications:** GitHub Actions fará o **CI/CD em um repositório de aplicação de teste separado**, a integrar durante o lab correspondente. Build/test/scan/publicação → atualização GitOps revisada → Argo CD **local** no Management EKS faz deploy nos clusters autorizados. Não adicionar código ou workflow CI da aplicação neste monorepo.

[ADR-0006](../docs/adr/0006-terraformrun-and-application-ci.md)

O Workflow Plane usa o Hub local. A integração com o Terraform Executor faz parte da [Fase 2 do roteiro](../docs/learning-roadmap.md#fase-2--control-plane-no-hub).

[← Início](../README.md) · [Arquitetura](../docs/architecture.md)

**Policy Engine:** [módulos Rego e perfis YAML](policy-engine/README.md) compõem as regras por stack, selecionadas pelo perfil de execução confiável.

## Binário de IaC no Simulator

O runner inclui Terraform 1.13.5 e OpenTofu 1.13.9. O perfil local seleciona `iac: terraform`. Os steps encaminham o binário selecionado em `IAC_BIN` para init, validate, geração e leitura do plano, revalidação, apply e verificação. O destroy usa o mesmo binário e as mesmas garantias.

Para exercitar OpenTofu, use um perfil com `iac: tofu` e um PVC limpo dedicado à stack de teste. Conclua o destroy e remova os Workflows antigos antes de liberar o PVC de teste. O vínculo persistido inclui a ferramenta. Mantenha separados os states, os planos salvos e os arquivos de lock de dependências dos dois CLIs. A POC usa somente o backend local neste exercício. [Perfil do Simulator](argo-workflows/terraformstack/simulator.json) · [Instalação oficial do OpenTofu em imagens](https://opentofu.org/docs/intro/install/docker/)

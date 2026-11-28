# Guias das etapas

O lab usa o **Hub EKS Auto Mode**. As etapas seguem a ordem de dependências do roteiro. A coluna de disponibilidade indica se o guia já traz comandos ou descreve a entrega prevista. [Estado da POC e evidências](../docs/learning-roadmap.md#estado-da-poc)

| Etapa | Guia | Disponibilidade |
| :--- | :--- | :--- |
| **00** | [Documentação e monorepo](../docs/repository-structure.md) | Referência da estrutura |
| **01** | [Bootstrap do Hub EKS Auto Mode](01-platform-management-cluster.md) | Procedimento e workload de teste validados |
| **02** | [Primeiro sync no Argo CD gerenciado](02-argocd-central.md) | Procedimento e primeiro sync comprovados no EKS |
| **03** | [Argo Workflows](03-argo-workflows.md) | Smoke test com Helm no EKS |
| **04** | [API do TerraformStack via GitOps](04-terraformstack-crd.md) | CRD estabelecido e sincronizado no EKS |
| **05** | [Controller e pipeline do TerraformStack](05-terraformstack-runtime.md) | Runtime com ECR e duas réplicas no EKS |
| **06** | [Terraform Stack Simulator de ponta a ponta](06-terraformstack-simulator.md) | Primeiro apply validado no EKS |
| **07** | [Crossplane Core e smoke test](07-crossplane-core.md) | Core e API de demonstração validados no EKS |
| **08** | [Crossplane de ponta a ponta](08-crossplane-gitops.md) | XR sincronizado e drift corrigido no EKS |

A instalação dos add-ons pelo catálogo e os cenários restantes do Simulator seguem com aceites próprios. As etapas seguintes estão no [roteiro](../docs/learning-roadmap.md).

> **Nota histórica.** Os primeiros testes do Control Plane e do Workflow Plane foram feitos em `kind`. Quem quiser repetir o ensaio em um cluster local pode adaptar os manifests por conta própria.

## Como ler e reproduzir uma etapa

Cada guia parte do estado alcançado pela etapa anterior. **Objetivo** e **O que será feito** delimitam a entrega. **Por que fizemos assim** explica as decisões. **O que observar** liga a configuração instalada ao resultado.

- **Observar.** Use consultas, eventos e logs para conferir as dependências e os sinais do sistema.
- **Sincronizar.** Confira a revisão, o destino e o diff de cada Application antes do sync manual. Reutilize os manifests canônicos dos blocos.
- **Executar.** Aplique a validação sobre os componentes instalados. Limpe os recursos do teste quando aplicável e preserve a base compartilhada.

[← Página inicial](../README.md) · [Arquitetura](../docs/architecture.md) · [Fundamentos](../docs/fundamentos/README.md)

# Guias das etapas

O [roteiro prático](../docs/learning-roadmap.md) organiza a montagem cumulativa da plataforma. Os guias descrevem os blocos disponíveis e as evidências que permitem avançar. As próximas entregas permanecem no roadmap.

| Etapa | Guia |
| :--- | :--- |
| **00** | [Documentação e monorepo](../docs/repository-structure.md) |
| **01** | [Bootstrap do Hub EKS Auto Mode](01-platform-management-cluster.md) |
| **02** | [Primeiro sync no Argo CD gerenciado](02-argocd-central.md) |
| **03** | [Argo Workflows](03-argo-workflows.md) |
| **04** | [API do TerraformStack via GitOps](04-terraformstack-crd.md) |
| **05** | [Controller e pipeline do TerraformStack](05-terraformstack-runtime.md) |
| **06** | [Terraform Stack Simulator de ponta a ponta no kind](06-terraformstack-simulator.md) |
| **07** | [Crossplane Core e smoke test](07-crossplane-core.md) |
| **08** | [Crossplane de ponta a ponta no kind](08-crossplane-gitops.md) |

> O Hub EKS é o ponto de partida da POC. Os guias das Etapas 03–08 ainda precisam adequar contextos, manifests ou evidências ao EKS antes de serem executados nesse destino.

## Como ler e reproduzir uma etapa

Cada guia parte do estado alcançado pela etapa anterior. **Objetivo** e **O que será feito** delimitam a entrega. **Por que fizemos assim** explica as decisões. **O que observar** liga a configuração instalada ao resultado.

- **Observar.** Use consultas, eventos e logs para conferir as dependências e os sinais do sistema.
- **Sincronizar.** Confira a revisão, o destino e o diff de cada Application antes do sync manual. Reutilize os manifests canônicos dos blocos.
- **Executar.** Aplique a validação sobre os componentes instalados. Limpe os recursos do teste quando aplicável e preserve a base compartilhada.

[← Página inicial](../README.md) · [Arquitetura](../docs/architecture.md) · [Fundamentos](../docs/fundamentos/README.md)

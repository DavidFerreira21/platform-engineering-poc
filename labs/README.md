# Guias das etapas

O [roteiro prático](../docs/learning-roadmap.md) reúne as oito fases e o aceite de cada etapa. Após a Etapa 04, a sequência começa pelo CRD `TerraformStack` e pela execução simulada. O guia de UAT passa a ser usado na primeira integração com AWS.

| Etapa | Guia |
| :--- | :--- |
| **00** | [Documentação e monorepo](../docs/repository-structure.md) |
| **01** | [Hub kind e Traefik](01-platform-management-cluster.md) |
| **02** | [Argo CD central e catálogo de add-ons](02-argocd-central.md) |
| **03** | [Argo Workflows](03-argo-workflows.md) |
| **04** | [Crossplane Core](04-crossplane-core.md) |
| **05** | [API do TerraformStack via GitOps](05-terraformstack-crd.md) |
| **06** | [Controller do TerraformStack e executor simulado](06-terraformstack-controller.md) |
| **07** | [Registro e validação de UAT](07-registro-ambientes-aws.md) |

## Como ler e reproduzir uma etapa

Os guias registram como a POC foi construída e por que cada decisão entrou naquele ponto. **Objetivo** e **O que será feito** delimitam o exercício. **Por que fizemos assim** explica a escolha. **O que observar** liga o resultado aos sinais do sistema. Os comandos reproduzem a construção e verificam o comportamento esperado. As orientações de diagnóstico ajudam a investigar problemas que apareçam durante o percurso.

- **Observar.** Use consultas, eventos e logs para conferir o estado atual do Hub. Verifique as dependências da etapa, mesmo quando o Hub já tiver componentes de etapas posteriores.
- **Sincronizar.** Quando a etapa pedir uma mudança, confira a `Application`, a revisão esperada e seu diff antes do **Sync**. Cada componente com ciclo próprio tem sua `Application`. Uma etapa pode usar componentes que já estão saudáveis.
- **Executar.** Submeta a execução do exercício no ponto indicado e acompanhe seus resultados. Mais adiante, a pausa de aprovação do Terraform acontece dentro do Workflow. Alterações em recursos AWS exigem autorização por operação.

Em um Hub já em uso, limpe os recursos identificados pelo exercício e preserve a base compartilhada. [Decisões de reconciliação](../docs/adr/0003-reconciliation-and-gitops.md)

[← Página inicial](../README.md) · [Arquitetura](../docs/architecture.md) · [Fundamentos](../docs/fundamentos/README.md)

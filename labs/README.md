# Guias das etapas

O [roteiro prático](../docs/learning-roadmap.md) reúne as sete fases e o aceite de cada etapa. Após a Etapa 04, a sequência começa pelo CRD `TerraformStack` e pela execução simulada. O guia de UAT passa a ser usado na primeira integração com AWS.

| Etapa | Guia |
| :--- | :--- |
| **00** | [Documentação e monorepo](../docs/repository-structure.md) |
| **01** | [Hub kind e Traefik](01-platform-management-cluster.md) |
| **02** | [Argo CD central](02-argocd-central.md) |
| **03** | [Argo Workflows](03-argo-workflows.md) |
| **04** | [Crossplane Core](04-crossplane-core.md) |
| **07** | [Registro e validação de UAT](07-registro-ambientes-aws.md) |

Cada guia começa com **Objetivo** e **O que será feito**, para indicar o resultado esperado e o que será criado, alterado ou validado. Em seguida vêm os pré-requisitos, os passos, as evidências, o diagnóstico, a recuperação e a limpeza. Recursos AWS reais exigem autorização por operação.

[← Página inicial](../README.md) · [Arquitetura](../docs/architecture.md) · [Fundamentos](../docs/fundamentos/README.md)

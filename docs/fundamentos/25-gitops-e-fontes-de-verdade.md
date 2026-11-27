# GitOps e fontes de verdade

GitOps é um modelo operacional no qual o estado desejado é declarativo, versionado e aplicado por reconciliação automatizada. Um commit não é uma ordem imperativa de execução: descreve a intenção que o sistema deve perseguir. O histórico Git permite revisar mudanças e recuperar versões, mas não substitui os demais estados da plataforma.

## Três estados diferentes

| Estado | Pergunta que responde |
| --- | --- |
| Git | Qual intenção foi revisada e versionada? |
| API Kubernetes | Qual solicitação foi entregue e qual resultado o controller conhece? |
| Infraestrutura e state do IaC | Quais recursos existem e qual vínculo a ferramenta mantém com eles? |

Argo CD pode marcar uma Application como Synced porque o CR corresponde ao Git, mesmo quando a infraestrutura ainda não existe, está bloqueada ou divergiu. A condição Ready do recurso e o estado real precisam ser examinados separadamente. Reverter um commit propõe outra intenção; não recupera automaticamente dados, efeitos irreversíveis ou recursos removidos.

## Aplicação na POC

O Argo CD central entrega TerraformStack e XRs ao Hub. O controller e o Crossplane reconciliam seus respectivos recursos; a squad mantém seus manifests em outro repositório, entregue por Argo CD local. A aprovação do PR não equivale à aprovação de um plano de infraestrutura. O catálogo de add-ons também não substitui os contratos de consumo.

No Simulator, sincronizar a mesma revisão durante uma espera de aprovação não deveria criar outra tentativa. Uma nova revisão pode modificar a intenção, mas o controller preserva as garantias de concorrência.

## Leitura complementar

- [Fluxo GitOps da POC](../flows/gitops.md)
- [OpenGitOps — princípios](https://opengitops.dev/)

---

[Aula anterior](24-ciclo-de-vida-produtos.md) · [Índice](README.md) · [Próxima aula](26-reconciliacao-e-orquestracao.md)

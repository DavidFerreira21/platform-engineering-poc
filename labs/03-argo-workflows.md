# Etapa 03 — Argo Workflows via GitOps

[← Etapa 02](02-argocd-central.md) · [Guias das etapas](README.md)

> A execução desta etapa no Hub EKS exige adaptar o catálogo à capability Argo CD e revisar o acesso ao Argo Server. Os comandos abaixo ainda usam o contexto e a rota do ambiente anterior.

## Objetivo

Executar um Workflow de teste no Hub EKS e acompanhar na interface do Argo Workflows o estado da execução, o Pod e seus logs.

## O que será feito

- Sincronizar a Application `argo-workflows`, gerada pelo `ApplicationSet` da Etapa 02. Conferir o controller, os CRDs, a ServiceAccount do executor e o Argo Server, que serve a interface.
- Abrir a interface com acesso de leitura autorizado e observar a execução de exemplo.
- Remover a execução de exemplo e seu Pod. O controller fica disponível para as próximas etapas.

O exercício usa um Pod descartável no Hub EKS.

## Por que fizemos assim

O controller e os CRDs do Argo Workflows ficam em uma Application compartilhada. O manifesto do exercício define uma tarefa pequena, e cada submissão cria uma execução com nome próprio. A interface do Argo Workflows usa o Argo Server para reunir o estado e os logs da execução. O acesso nesta etapa é de leitura e depende da configuração de identidade e conectividade do Hub EKS. A Etapa 06 usa identidade individual para aprovar o plano na UI.

> A instalação e o aceite desta etapa no Hub EKS estão pendentes. Os critérios abaixo descrevem a entrega prevista.

## Aceite no Hub EKS

Depois do primeiro sync da Etapa 02, validar os AppProjects e o destino por ARN da Application de Argo Workflows. Confirmar os CRDs, o rollout do controller e o acesso autorizado à interface.

## O que observar

Uma execução curta cria um Pod, conclui com sucesso e expõe fase, resultado e logs ao operador autorizado. Remover os recursos do exercício após a validação.

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

> O smoke test no Hub EKS está registrado abaixo. A instalação pelo catálogo GitOps segue com aceite próprio.

## Aceite no Hub EKS

Depois do primeiro sync da Etapa 02, validar os AppProjects e o destino por ARN da Application de Argo Workflows. Confirmar os CRDs, o rollout do controller e o acesso autorizado à interface.

## O que observar

Uma execução curta cria um Pod, conclui com sucesso e expõe fase, resultado e logs ao operador autorizado. Remover os recursos do exercício após a validação.

## Reproduzir o smoke test com Helm

Usar o contexto do Hub conferido na Etapa 02, Helm instalado e um namespace sem outra instalação do Argo Workflows. Se a Application do catálogo já gerencia esse componente, executar apenas o Workflow de teste após conferir o rollout.

```bash
helm upgrade --install argo-workflows argo-workflows \
  --repo https://argoproj.github.io/argo-helm --version 2.0.18 \
  --kube-context "$hub_arn" --namespace argo-workflows --create-namespace \
  --set crds.install=true --set crds.full=true --set singleNamespace=true \
  --set server.enabled=true --set server.secure=true \
  --set server.serviceType=ClusterIP --set 'server.authModes[0]=client' \
  --set workflow.serviceAccount.create=true --set workflow.rbac.create=true \
  --timeout 5m --wait

hub -n argo-workflows create -f - <<'EOF_WORKFLOW'
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  name: hub-workflow-smoke
spec:
  serviceAccountName: argo-workflow
  entrypoint: smoke
  templates:
    - name: smoke
      container:
        image: busybox:1.37.0
        command: [sh, -ec]
        args: ['echo "Workflow executado no Hub EKS"']
EOF_WORKFLOW
hub -n argo-workflows wait --for=jsonpath='{.status.phase}'=Succeeded \
  workflow/hub-workflow-smoke --timeout=180s
hub -n argo-workflows logs -l workflows.argoproj.io/workflow=hub-workflow-smoke -c main
hub -n argo-workflows delete workflow hub-workflow-smoke
```

## Evidência no Hub EKS

Em 28/11/2026, instalamos o chart `argo-workflows` 2.0.18 com Helm no namespace `argo-workflows` para testar o motor. O workflow controller e o Argo Server ficaram disponíveis. Um Workflow descartável terminou em `Succeeded`, e a API entregou os logs do container. O servidor permaneceu como ClusterIP. O aceite da instalação pelo catálogo GitOps é separado desse teste.

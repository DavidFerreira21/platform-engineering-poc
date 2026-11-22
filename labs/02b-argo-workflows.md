# Lab 02B — Argo Workflows via GitOps

[← Lab 02](02-argocd-central.md) · [Índice dos labs](README.md)

## Objetivo

Executar um Workflow local e usar o estado da execução, os Pods e os logs para diagnosticar uma falha e confirmar a recuperação.

## O que será feito

- Criar o namespace `argo-workflows`, o AppProject `platform-workflows` e a Application `argo-workflows`. Sincronizar pelo Argo CD a instalação do controller, dos CRDs e da ServiceAccount do executor.
- Submeter três execuções do mesmo manifesto e observar os estados `Succeeded`, `Failed` e `Succeeded`. A segunda execução usa um parâmetro que provoca uma falha no contêiner.
- Remover os três Workflows e seus Pods. O controller fica disponível para os próximos labs.

O exercício roda no Hub local e tem custo AWS zero.

## Pré-requisitos

- Lab 02 concluído, com o Argo CD central acessível.
- `kubectl` com o contexto `kind-platform-management`. O Hub precisa de acesso ao repositório Helm e aos registries das imagens.
- Execute os comandos na raiz do monorepo e no mesmo terminal.

```bash
hub() { kubectl --context kind-platform-management "$@"; }
```

## 1. Sincronizar o controller e os CRDs

```bash
bootstrap=workflow-plane/argo-workflows/bootstrap
hub apply -f "$bootstrap/namespace.yaml" -f "$bootstrap/project.yaml"
hub apply -f "$bootstrap/application.yaml"
```

Na UI do Argo CD, abra **Applications → argo-workflows**. Após a primeira comparação indicar `OutOfSync`, selecione **Sync → Synchronize**. O chart `2.0.18` usa Argo Workflows `v4.1.6`. Um Job de pré-sincronização instala os CRDs completos. O controller executa Workflows apenas em `argo-workflows`. A observação do exercício é feita com `kubectl`.

```bash
hub -n argocd get application argo-workflows
hub wait --for=condition=Established --timeout=90s \
  crd/workflows.argoproj.io crd/workflowtaskresults.argoproj.io
hub -n argo-workflows rollout status deployment/argo-workflows-workflow-controller --timeout=120s
```

A Application deve ficar `Synced` e `Healthy`. Se o sync falhar antes de criar o controller, consulte `hub -n argocd describe application argo-workflows` e, se o Job ainda existir, `hub -n argo-workflows logs job/argo-workflows-crd-install`.

## 2. Executar e observar

```bash
manifest=workflow-plane/argo-workflows/demo/workflow.yaml
run=$(hub create -f "$manifest" -o name)
hub -n argo-workflows wait "$run" --for='jsonpath={.status.phase}=Succeeded' --timeout=180s
hub -n argo-workflows get "$run"
hub -n argo-workflows get pods -l "workflows.argoproj.io/workflow=${run##*/}"
hub -n argo-workflows logs -l "workflows.argoproj.io/workflow=${run##*/}" -c main
```

O log esperado é `lab02b-workflow-ok`. O Pod usa a ServiceAccount `argo-workflow`, com permissões para o executor publicar resultados em `workflowtaskresults`. Cada submissão recebe um nome próprio.

## 3. Provocar uma falha e reexecutar

Submeta o mesmo manifesto com um parâmetro inválido para a lógica do contêiner. O patch altera apenas a cópia submetida.

```bash
failed=$(hub patch --local -f "$manifest" --type=json \
  -p '[{"op":"replace","path":"/spec/arguments/parameters/0/value","value":"falhar"}]' \
  -o yaml | hub create -f - -o name)
hub -n argo-workflows wait "$failed" --for='jsonpath={.status.phase}=Failed' --timeout=180s
hub -n argo-workflows get "$failed"
hub -n argo-workflows get pods -l "workflows.argoproj.io/workflow=${failed##*/}"
hub -n argo-workflows logs -l "workflows.argoproj.io/workflow=${failed##*/}" -c main
```

O log deve mostrar `falha controlada: resultado esperado é ok`. Se o Workflow parar em `Error` ou o Pod não iniciar, inspecione `hub -n argo-workflows describe "$failed"`, os eventos do Pod e os logs do controller.

Reenvie o manifesto com o parâmetro padrão. O Workflow anterior permanece disponível para comparar as duas execuções.

```bash
recovered=$(hub create -f "$manifest" -o name)
hub -n argo-workflows wait "$recovered" --for='jsonpath={.status.phase}=Succeeded' --timeout=180s
hub -n argo-workflows logs -l "workflows.argoproj.io/workflow=${recovered##*/}" -c main
```

## 4. Limpar o exercício

```bash
hub -n argo-workflows delete "$run" "$failed" "$recovered" \
  --ignore-not-found --cascade=foreground --timeout=120s
hub -n argo-workflows get workflows -l lab.platform.example/exercise=lab02b
hub -n argo-workflows get pods -l "workflows.argoproj.io/workflow in (${run##*/},${failed##*/},${recovered##*/})"
```

As duas consultas devem retornar listas vazias. Mantenha o controller instalado para uso nos próximos labs. [Documentação do Argo Workflows](https://argo-workflows.readthedocs.io/en/latest/installation/)

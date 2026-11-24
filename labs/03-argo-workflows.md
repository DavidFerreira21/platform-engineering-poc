# Etapa 03 — Argo Workflows via GitOps

[← Etapa 02](02-argocd-central.md) · [Guias das etapas](README.md)

## Objetivo

Executar um Workflow local e usar o estado da execução, os Pods e os logs para diagnosticar uma falha e confirmar a recuperação.

## O que será feito

- Sincronizar a Application `argo-workflows`, gerada pelo `ApplicationSet` da Etapa 02. Conferir o controller, os CRDs e a ServiceAccount do executor.
- Submeter três execuções do mesmo manifesto e observar os estados `Succeeded`, `Failed` e `Succeeded`. A segunda execução usa um parâmetro que provoca uma falha no contêiner.
- Remover os três Workflows e seus Pods. O controller fica disponível para as próximas etapas.

O exercício roda no Hub local e tem custo AWS zero.

## Por que fizemos assim

O controller e os CRDs do Argo Workflows ficam em uma Application compartilhada. Cada submissão do exemplo cria uma execução com nome próprio. Isso permite comparar a falha e a recuperação pelo estado, pelos Pods e pelos logs de cada execução.

## Pré-requisitos

- Etapa 02 concluída, com o Argo CD central acessível e `hub-addons` como owner da Application `argo-workflows`.
- `kubectl` com o contexto `kind-platform-management`. O Hub precisa de acesso ao repositório Helm e aos registries das imagens.
- Execute os comandos na raiz do monorepo e no mesmo terminal.

```bash
hub() { kubectl --context kind-platform-management "$@"; }
```

## 1. Conferir o controller e os CRDs

Na UI do Argo CD, abra **Applications → argo-workflows**. Se o componente já estiver saudável, confira a instalação com os comandos abaixo. Se a etapa exigir a instalação, confira a versão do chart no catálogo e inspecione o **Diff**. Sincronize a Application quando o diff corresponder à instalação esperada. Um Job de pré-sincronização instala os CRDs completos. O controller executa Workflows no namespace `argo-workflows`. [Entrada do catálogo](../platform/hub/addons/catalog/argo-workflows.yaml)

```bash
hub -n argocd get application argo-workflows
hub wait --for=condition=Established --timeout=90s \
  crd/workflows.argoproj.io crd/workflowtaskresults.argoproj.io
hub -n argo-workflows rollout status deployment/argo-workflows-workflow-controller --timeout=120s
```

Após sincronizar, confira `Synced` e `Healthy`. Em um Hub existente, a disponibilidade do controller e dos CRDs é o ponto de partida do exercício. Se a Application estiver ausente, confira o `ApplicationSet` na Etapa 02. Se o sync falhar antes de criar o controller, consulte `hub -n argocd describe application argo-workflows` e, se o Job ainda existir, `hub -n argo-workflows logs job/argo-workflows-crd-install`.

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

## O que observar

- Cada nome de Workflow identifica seus Pods e seus logs.
- Compare o parâmetro e o log da execução `Failed` com os da execução recuperada `Succeeded`. O controller compartilhado atende às duas.

## 4. Limpar o exercício

```bash
hub -n argo-workflows delete "$run" "$failed" "$recovered" \
  --ignore-not-found --cascade=foreground --timeout=120s
hub -n argo-workflows get "$run" "$failed" "$recovered" --ignore-not-found
hub -n argo-workflows get pods -l "workflows.argoproj.io/workflow in (${run##*/},${failed##*/},${recovered##*/})"
```

As duas consultas devem retornar listas vazias. Mantenha o controller instalado para uso nas próximas etapas. [Documentação do Argo Workflows](https://argo-workflows.readthedocs.io/en/latest/installation/)

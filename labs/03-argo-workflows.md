# Etapa 03 — Argo Workflows via GitOps

[← Etapa 02](02-argocd-central.md) · [Guias das etapas](README.md)

## Objetivo

Executar um Workflow local e acompanhar na interface do Argo Workflows o estado da execução, o Pod e seus logs.

## O que será feito

- Sincronizar a Application `argo-workflows`, gerada pelo `ApplicationSet` da Etapa 02. Conferir o controller, os CRDs, a ServiceAccount do executor e o Argo Server, que serve a interface.
- Abrir a interface local pelo Traefik com um token temporário de leitura e observar a execução de exemplo.
- Remover a execução de exemplo e seu Pod. O controller fica disponível para as próximas etapas.

O exercício roda no Hub local e tem custo AWS zero.

## Por que fizemos assim

O controller e os CRDs do Argo Workflows ficam em uma Application compartilhada. O manifesto do exercício define uma tarefa pequena, e cada submissão cria uma execução com nome próprio. A interface do Argo Workflows usa o Argo Server para reunir o estado e os logs da execução. O Traefik encaminha a conexão HTTPS de `workflows.localhost` ao Argo Server, que termina o TLS. O acesso nesta etapa é local e de leitura. A Etapa 06 usa identidade individual para aprovar o plano na UI.

## Pré-requisitos

- Etapa 02 concluída, com o Argo CD central acessível e `hub-addons` como owner da Application `argo-workflows`.
- `kubectl` com o contexto `kind-platform-management`. O Hub precisa de acesso ao repositório Helm e aos registries das imagens.
- Execute os comandos na raiz do monorepo e no mesmo terminal.

```bash
hub() { kubectl --context kind-platform-management "$@"; }
```

## 1. Conferir a instalação do Argo Workflows

Se o Hub já existia antes da configuração da rota, atualize o Traefik da Etapa 01 e reaplique o bootstrap da Etapa 02. Isso habilita o provider de CRDs do Traefik e permite a rota no AppProject.

```bash
bash platform/hub/scripts/install-traefik.sh
hub apply -f platform/hub/addons/bootstrap.yaml
```

Na UI do Argo CD, abra **Applications → argo-workflows**. Se o controller, o Argo Server e a rota já estiverem disponíveis, confira o estado com os comandos abaixo. Para instalar ou publicar a interface, confira a versão do chart no catálogo e inspecione o **Diff**. Sincronize a Application e aguarde a operação terminar antes de executar as consultas. Um Job de pré-sincronização instala os CRDs completos. O controller executa Workflows no namespace `argo-workflows`. [Entrada do catálogo](../platform/hub/addons/catalog/argo-workflows.yaml)

```bash
hub -n argocd get application argo-workflows
hub wait --for=condition=Established --timeout=90s \
  crd/workflows.argoproj.io crd/workflowtaskresults.argoproj.io
hub -n argo-workflows rollout status deployment/argo-workflows-workflow-controller --timeout=120s
hub -n argo-workflows rollout status deployment/argo-workflows-server --timeout=120s
hub -n argo-workflows get service/argo-workflows-server ingressroutetcp/argo-workflows-server
```

Após sincronizar, confira `Synced` e `Healthy`. Em um Hub existente, a disponibilidade do controller e dos CRDs é o ponto de partida do exercício. Se a Application estiver ausente, confira o `ApplicationSet` na Etapa 02. Se o sync falhar antes de criar o controller, consulte `hub -n argocd describe application argo-workflows` e, se o Job ainda existir, `hub -n argo-workflows logs job/argo-workflows-crd-install`.

## 2. Abrir a interface do Argo Workflows

Abra **https://workflows.localhost:8443**. O Argo Server apresenta o certificado autoassinado do laboratório. Gere um token temporário no terminal do exercício.

```bash
printf 'Bearer %s\n' "$(hub -n argo-workflows create token argo-workflows-viewer --duration=30m)"
```

Cole a linha inteira, incluindo `Bearer`, no campo de token da interface. Selecione o namespace `argo-workflows`. Se a sessão expirar, gere outro token. [Modo de autenticação](https://argo-workflows.readthedocs.io/en/release-4.1/argo-server-auth-mode/) · [Rota TCP do Traefik](https://doc.traefik.io/traefik/reference/routing-configuration/kubernetes/crd/tcp/ingressroutetcp/)

Se a URL responder com `404 page not found`, confira se a Application está sincronizada, se o `IngressRouteTCP` aparece no comando do passo 1 e se o Traefik foi atualizado. Para investigar a rota, consulte `hub -n traefik logs deployment/traefik --tail=100`.

## 3. Executar e observar

```bash
manifest=workflow-plane/argo-workflows/demo/workflow.yaml
run=$(hub create -f "$manifest" -o name)
hub -n argo-workflows wait "$run" --for='jsonpath={.status.phase}=Succeeded' --timeout=180s
hub -n argo-workflows get "$run"
hub -n argo-workflows get pods -l "workflows.argoproj.io/workflow=${run##*/}"
hub -n argo-workflows logs -l "workflows.argoproj.io/workflow=${run##*/}" -c main
```

O log esperado é `lab02b-workflow-ok`. O Pod usa a ServiceAccount `argo-workflow`, com permissões para o executor publicar resultados em `workflowtaskresults`. Cada submissão recebe um nome próprio.

Na interface do Argo Workflows, selecione a execução pelo nome retornado por `hub -n argo-workflows get "$run"`. No grafo, clique no círculo do único nó, rotulado com o nome da execução. No painel do nó, abra **Logs** e confira a mensagem no container `main`. Se a execução não chegar a `Succeeded`, consulte `hub -n argo-workflows describe "$run"`, os eventos do Pod e os logs do controller.

## O que observar

- A interface e o `kubectl` mostram o mesmo Workflow em `Succeeded`. O nome da execução identifica seu Pod e seus logs.
- O log `lab02b-workflow-ok` é a saída da tarefa declarada no manifesto.

## 4. Limpar o exercício

```bash
hub -n argo-workflows delete "$run" \
  --ignore-not-found --cascade=foreground --timeout=120s
hub -n argo-workflows get "$run" --ignore-not-found
hub -n argo-workflows get pods -l "workflows.argoproj.io/workflow=${run##*/}"
```

As duas consultas devem retornar listas vazias. O controller e a interface ficam disponíveis para as próximas etapas. [Documentação do Argo Workflows](https://argo-workflows.readthedocs.io/en/latest/installation/)

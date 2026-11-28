# Etapa 08 — Crossplane de ponta a ponta no kind

[← Etapa 07](07-crossplane-core.md) · [Guias das etapas](README.md)

## Objetivo

Comprovar o ciclo Git → Argo CD → Kubernetes API → Crossplane no Hub local, da criação de um `LabMessage` à remoção do ConfigMap composto.

## O que será feito

- Sincronizar a API de demonstração e sua primeira solicitação em Applications com sync manual.
- Confirmar o schema, o vínculo com a Composition, as condições do XR e o conteúdo do ConfigMap.
- Atualizar a mensagem via Git, repetir o sync e observar o Crossplane corrigir um desvio no ConfigMap.
- Verificar os limites dos AppProjects e concluir o ciclo com prune do XR.

## Por que fizemos assim

`crossplane-core` instala o motor compartilhado. `crossplane-demo-api` instala Function, XRD e Composition. `crossplane-demo` entrega apenas o `LabMessage`, depois de a API estar disponível. Os dois últimos reutilizam os manifests da Etapa 07. O Argo CD entrega a intenção e o Crossplane gerencia o ConfigMap composto. O ensaio usa recursos Kubernetes no `kind` e é independente do Simulator. [Catálogo do Crossplane](../platform/hub/addons/catalog/crossplane-core.yaml) · [ADR-0003](../docs/adr/0003-reconciliation-and-gitops.md)

## Pré-requisitos

Crossplane Core saudável e smoke test da Etapa 07 limpo. Execute na raiz do monorepo com acesso de administração ao Hub local, à UI do Argo CD e a uma branch publicável no repositório autorizado. Use uma revisão que contenha o catálogo da Etapa 08. O gerador `hub-addons` lê `main`. [Etapa 07](07-crossplane-core.md) · [Etapa 02](02-argocd-central.md)

## 1. Entregar a API e a primeira solicitação

Aplique o bootstrap atualizado conforme a Etapa 02 para registrar os AppProjects. Se o catálogo ainda estiver em revisão, no Hub de teste aponte temporariamente `spec.generators[0].git.revision` do `ApplicationSet/hub-addons` para o SHA publicado da Etapa 08. Mantenha os outros add-ons sem novo sync e reaplique o bootstrap ao concluir. [Bootstrap](../platform/hub/addons/bootstrap.yaml)

```bash
hub() { kubectl --context kind-platform-management "$@"; }
hub -n argocd wait --timeout=180s \
  --for='jsonpath={.metadata.ownerReferences[?(@.kind=="ApplicationSet")].name}=hub-addons' \
  application/crossplane-demo-api application/crossplane-demo
hub -n argocd get application crossplane-core crossplane-demo-api crossplane-demo
```

Na UI, confira **Diff**, destino no Hub e **Sync → Revision** antes de cada sync. Para uma branch de teste, informe o mesmo SHA publicado nos dois primeiros syncs.

1. Sincronize **crossplane-demo-api**. O diff inicial deve mostrar a Function `lab03-patch-and-transform`, o XRD `labmessages.lab.platform.example` e a Composition `lab03-message`.
2. Aguarde as condições abaixo. Se expirarem, confira `hub describe function.pkg.crossplane.io lab03-patch-and-transform`, `hub describe xrd labmessages.lab.platform.example` e o resultado da operação no Argo CD.

```bash
hub wait --for=condition=Healthy --timeout=180s function.pkg.crossplane.io/lab03-patch-and-transform
hub wait --for=create --timeout=90s crd/labmessages.lab.platform.example
hub wait --for=condition=Established --timeout=90s crd/labmessages.lab.platform.example
hub explain labmessage.spec.message --api-version=lab.platform.example/v1alpha1
```

3. Sincronize **crossplane-demo**. O diff inicial deve mostrar um único `LabMessage/lab03-demo` no namespace `platform-system`. Confirme a revisão sincronizada e aguarde o XR.

```bash
xr() { hub -n platform-system get labmessage lab03-demo "$@"; }
cm() { hub -n platform-system get configmap "$configmap" "$@"; }
mensagem() { cm -o jsonpath='{.data.mensagem}{"\n"}'; }
hub -n argocd get application crossplane-demo -o   'custom-columns=SYNC:.status.sync.status,REVISION:.status.sync.revision'
hub -n platform-system wait labmessage/lab03-demo --for=condition=Synced --timeout=120s
hub -n platform-system wait labmessage/lab03-demo --for=condition=Ready --timeout=120s
xr -o yaml
configmap="$(xr -o jsonpath='{.spec.crossplane.resourceRefs[?(@.kind=="ConfigMap")].name}')"
xr_uid="$(xr -o jsonpath='{.metadata.uid}')"
cm_uid="$(cm -o jsonpath='{.metadata.uid}')"
test "$xr_uid" = "$(cm -o jsonpath='{.metadata.ownerReferences[?(@.controller==true)].uid}')"
test "$(mensagem)" = reconciliado-pelo-crossplane
```

Se o XR não convergir, examine `hub -n platform-system describe labmessage lab03-demo` e `hub -n crossplane-system logs deployment/crossplane --since=10m`. Confira na árvore do Argo CD que o ConfigMap aparece como recurso filho do XR.

## 2. Validar schema e limites do GitOps

Execute um dry-run no servidor com uma mensagem vazia. A API deve recusar `spec.message` por comprimento mínimo. Confira também a lista de recursos autorizados nos dois AppProjects. O projeto da solicitação permite `LabMessage` em `platform-system`. O projeto da API permite Function, XRD e Composition. Para verificar a restrição pelo Argo CD, publique um commit de teste que acrescente um segundo documento YAML ao `xr.yaml` com um ConfigMap chamado `lab08-fora-do-contrato` em `platform-system`. Confira o diff e sincronize o SHA com **Prune** desmarcado. A operação deve recusar o ConfigMap por tipo de recurso e o objeto deve continuar ausente. Volte ao SHA inicial antes da atualização da mensagem. [AppProjects](../platform/hub/addons/bootstrap.yaml)

```bash
hub -n platform-system patch labmessage lab03-demo --type=merge --dry-run=server   -p '{"spec":{"message":""}}'
hub -n argocd get appproject platform-crossplane-demo platform-crossplane-demo-api -o yaml
test "$(mensagem)" = reconciliado-pelo-crossplane
```

## 3. Atualizar, repetir o sync e observar drift

Use uma branch de exercício a partir do SHA da primeira sincronização, sem o ConfigMap do teste de permissão. Altere `spec.message` em [xr.yaml](../control-plane/crossplane/demo/xr.yaml) para `atualizado-via-gitops`, versione a mudança e publique o commit.

```bash
git add control-plane/crossplane/demo/xr.yaml
git commit -m "Exercita atualização do LabMessage via GitOps"
git push -u origin HEAD
git rev-parse HEAD
```

Sincronize **crossplane-demo** com o SHA publicado. Confira no diff a troca de `spec.message` e aguarde a propagação ao ConfigMap. Sincronize o mesmo SHA mais uma vez e repita as consultas.

```bash
hub -n platform-system wait "configmap/$configmap"   --for='jsonpath={.data.mensagem}=atualizado-via-gitops' --timeout=120s
test "$xr_uid" = "$(xr -o jsonpath='{.metadata.uid}')"
test "$cm_uid" = "$(cm -o jsonpath='{.metadata.uid}')"
```

Simule uma edição direta no ConfigMap do exercício e aguarde o retorno à mensagem do Git sem novo sync.

```bash
hub -n platform-system patch "configmap/$configmap" --type=merge   -p '{"data":{"mensagem":"desvio-local"}}'
hub -n platform-system wait "configmap/$configmap"   --for='jsonpath={.data.mensagem}=atualizado-via-gitops' --timeout=180s
```

## 4. Remover a solicitação via GitOps

Na mesma branch de exercício, remova somente `control-plane/crossplane/demo/xr.yaml`, faça commit e publique o SHA. Sincronize **crossplane-demo** nessa revisão com **Prune** marcado. Confirme que a única remoção apresentada é o `LabMessage/lab03-demo`. Confira a exclusão do XR e do ConfigMap composto.

```bash
hub -n platform-system wait --for=delete labmessage/lab03-demo --timeout=120s
hub -n platform-system wait --for=delete "configmap/$configmap" --timeout=120s
```

Depois de ambos sumirem, exclua pela árvore de **crossplane-demo-api** a Composition, o XRD e a Function do exercício, nessa ordem. Confirme que o CRD gerado sumiu e que **crossplane-core** continua saudável. As Applications do catálogo permanecem com sync manual e podem ser reutilizadas na próxima execução. Se o gerador foi apontado para um SHA de teste, reaplique o bootstrap agora para voltar a `main`.

## O que observar

| Cenário | Evidência esperada |
| :--- | :--- |
| Instalação da API | Function `Healthy`, CRD `Established` e campo `spec.message` publicado |
| Criação | SHA da operação no Argo CD, XR `Synced` e `Ready`, ConfigMap com UID do XR no owner reference |
| Schema e escopo | API recusa mensagem vazia, AppProjects delimitam destino e tipos de recursos |
| Atualização e re-sync | Nova mensagem no ConfigMap e mesmos UIDs |
| Drift | Crossplane restaura a mensagem sem nova operação no Argo CD |
| Prune | XR e ConfigMap ausentes com Core saudável |

Registre o chart, os SHAs sincronizados, as condições do XR e os resultados observados. O aceite usa o Hub local. A reinstalação no Hub EKS e o Provider AWS têm validações próprias no roteiro. [Roadmap](../docs/learning-roadmap.md)

## Evidências de aceite

Em 27/11/2026, sincronizamos o Core 2.4.3 no Hub `kind` `platform-management`. A Function `v0.8.2` ficou saudável e o CRD foi estabelecido. Com os AppProjects da etapa e manifests locais fornecidos ao Argo CD, a criação e a atualização do XR passaram. O ConfigMap manteve o UID no re-sync e o Crossplane corrigiu o desvio sem outro sync. O dry-run recusou a mensagem vazia e o AppProject recusou o ConfigMap proposto como recurso da Application.

O Argo CD também sincronizou o XR a partir do commit `1b48874` de `main`. Para exercitar o prune com essa mesma revisão, excluímos temporariamente `xr.yaml` pelo filtro da Application. O XR e o ConfigMap desapareceram e o Core permaneceu `Synced` e `Healthy`. O roteiro usa commits publicados para a atualização e a remoção. A descoberta automática das duas entradas do catálogo e a sequência completa de commits ficam para a execução com a revisão publicada da Etapa 08.

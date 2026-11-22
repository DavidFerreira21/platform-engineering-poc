# Etapa 04 — Crossplane Core via GitOps

[← Etapa 03](03-argo-workflows.md) · [Guias das etapas](README.md)

## Objetivo

Validar a reconciliação de um recurso composto no Hub e observar como o schema de um XRD controla o que a API aceita.

## O que será feito

- Criar o AppProject `platform-crossplane` e a Application `crossplane-core` para instalar o Crossplane Core e seus CRDs via Argo CD.
- Instalar a função de composição do exercício, aplicar o XRD de `LabMessage` e sua Composition e criar um `LabMessage` em `platform-system`. Conferir as condições do recurso e o ConfigMap criado pelo Crossplane.
- Tentar gravar uma mensagem vazia, observar a rejeição pela API e confirmar que valores válidos são reconciliados no ConfigMap.
- Remover os recursos do exercício e manter o Crossplane Core instalado.

O exercício roda no Hub local e tem custo AWS zero.

## Pré-requisitos

- Hub da Etapa 01 com Argo CD central acessível conforme a Etapa 02 e Argo Workflows verificado na Etapa 03.
- `kubectl` com o contexto `kind-platform-management` e conectividade do Hub ao repositório Helm e aos registries das imagens e da função.
- Execute os comandos da raiz do monorepo. Nos exemplos, `hub` seleciona explicitamente o contexto local.

```bash
hub() { kubectl --context kind-platform-management "$@"; }
hub get namespace argocd crossplane-system platform-system
```

## 1. Sincronizar o Crossplane Core

```bash
hub apply -f control-plane/crossplane/bootstrap/project.yaml
hub apply -f control-plane/crossplane/bootstrap/application.yaml
```

Na UI do Argo CD, abra **Applications → crossplane-core**. Aguarde a primeira comparação mostrar `OutOfSync` e então selecione **Sync → Synchronize**. A Application usa o chart Crossplane `2.4.3` e o AppProject limita a origem, o destino e os tipos usados pelo chart.

```bash
hub -n argocd get application crossplane-core -o \
  'custom-columns=SYNC:.status.sync.status,HEALTH:.status.health.status,REVISION:.status.sync.revision'
hub -n crossplane-system rollout status deployment/crossplane --timeout=180s
hub -n crossplane-system rollout status deployment/crossplane-rbac-manager --timeout=180s
hub wait --for=condition=Established --timeout=120s \
  crd/compositeresourcedefinitions.apiextensions.crossplane.io \
  crd/compositions.apiextensions.crossplane.io crd/functions.pkg.crossplane.io
hub get providers.pkg.crossplane.io
```

Confirme `Synced`, `Healthy` e a versão `2.4.3`. Num Hub dedicado a estas etapas, a lista de Providers deve estar vazia. Se o chart não sincronizar, consulte `hub -n argocd describe application crossplane-core` e confira a origem Helm a partir dos logs de `hub -n argocd logs deployment/argocd-repo-server`.

## 2. Registrar e consumir uma API local

O exercício aplica os manifests em etapas para tornar visível quando a função fica saudável e quando o novo tipo passa a existir. O XRD declara o schema, a Composition define o ConfigMap e o XR fornece a mensagem.

```bash
demo=control-plane/crossplane/demo
hub apply -f "$demo/function.yaml"
hub wait --for=condition=Healthy --timeout=180s function.pkg.crossplane.io/lab03-patch-and-transform
hub apply -f "$demo/composition.yaml" -f "$demo/xrd.yaml"
hub wait --for=create --timeout=90s crd/labmessages.lab.platform.example
hub wait --for=condition=Established --timeout=90s crd/labmessages.lab.platform.example
hub apply -f "$demo/xr.yaml"
hub -n platform-system wait --for=condition=Synced --timeout=120s labmessage/lab03-demo
hub -n platform-system wait --for=condition=Ready --timeout=120s labmessage/lab03-demo
hub -n platform-system get labmessage lab03-demo
hub -n platform-system get configmap -l lab.platform.example/exercise=lab03 \
  -o 'custom-columns=NAME:.metadata.name,OWNER:.metadata.ownerReferences[0].name,MESSAGE:.data.mensagem'
```

A saída deve mostrar um ConfigMap cujo owner é `lab03-demo` e cuja mensagem é `reconciliado-pelo-crossplane`. O `Ready` depende de a função observar o ConfigMap com o label do exercício. Se os waits expirarem, consulte `hub describe function.pkg.crossplane.io lab03-patch-and-transform`, `hub describe xrd labmessages.lab.platform.example` e `hub -n platform-system describe labmessage lab03-demo`.

## 3. Provocar uma falha de validação e recuperar

Tente gravar uma mensagem vazia. A API deve rejeitar o patch pois o schema exige ao menos um caractere. Observe a mensagem do servidor e confira que o XR manteve o valor anterior.

```bash
hub -n platform-system patch labmessage lab03-demo --type merge \
  -p '{"spec":{"message":""}}'
hub -n platform-system get labmessage lab03-demo -o 'jsonpath={.spec.message}{"\n"}'
```

Agora use uma mensagem válida. Aguarde o ConfigMap refletir o novo valor sem escrever diretamente nele, e reaplique o manifesto para restaurar o valor de referência.

```bash
hub -n platform-system patch labmessage lab03-demo --type merge \
  -p '{"spec":{"message":"mensagem-recuperada"}}'
hub -n platform-system wait configmap -l lab.platform.example/exercise=lab03 \
  --for='jsonpath={.data.mensagem}=mensagem-recuperada' --timeout=120s
hub apply -f "$demo/xr.yaml"
hub -n platform-system wait configmap -l lab.platform.example/exercise=lab03 \
  --for='jsonpath={.data.mensagem}=reconciliado-pelo-crossplane' --timeout=120s
```

## 4. Limpar o exercício

Remova primeiro o XR enquanto a função e o Crossplane ainda estão ativos. Confirme que o ConfigMap composto sumiu antes de remover a API e a função.

```bash
hub delete -f "$demo/xr.yaml" --ignore-not-found --timeout=120s
hub -n platform-system wait --for=delete configmap \
  -l lab.platform.example/exercise=lab03 --timeout=120s
hub delete -f "$demo/composition.yaml" -f "$demo/xrd.yaml" -f "$demo/function.yaml" \
  --ignore-not-found --timeout=120s
```

Mantenha a Application `crossplane-core` sincronizada para as próximas etapas. [Documentação do Crossplane](https://docs.crossplane.io/latest/guides/function-patch-and-transform/)

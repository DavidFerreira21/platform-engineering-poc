# Etapa 07 — Crossplane Core via GitOps

[← Etapa 06](06-terraformstack-simulator.md) · [Guias das etapas](README.md)

## Objetivo

Validar a reconciliação de um recurso composto no Hub e observar como o schema de um XRD controla o que a API aceita.

## O que será feito

- Sincronizar a Application `crossplane-core`, gerada pelo `ApplicationSet` da Etapa 02, para instalar o Crossplane Core e seus CRDs.
- Instalar a função de composição do exercício, aplicar o XRD de `LabMessage` e sua Composition e criar um `LabMessage` em `platform-system`. Conferir as condições do recurso e o ConfigMap criado pelo Crossplane.
- Consultar o schema publicado pelo XRD e conferir a mensagem do `LabMessage` no ConfigMap.
- Remover os recursos do exercício e manter o Crossplane Core instalado.

O exercício roda no Hub local, sem provisionar recursos AWS. Crossplane Core é instalado via GitOps; os manifests da demonstração são aplicados diretamente à API para observar cada etapa. Isso é um smoke test, não o aceite GitOps completo, reservado à Etapa 08.

## Por que fizemos assim

O Crossplane Core é compartilhado pelas etapas. O `LabMessage` usa um XRD namespaced da API v2 e uma Composition de demonstração para tornar a reconciliação visível em um ConfigMap do Hub. Assim é possível rever a validação do contrato e a reconciliação local mesmo em um Hub que já tenha outras Capabilities.

## Pré-requisitos

- Argo CD central acessível, com `hub-addons` como owner da Application `crossplane-core`, e o Simulator verificado na Etapa 06.
- `kubectl` com o contexto `kind-platform-management` e conectividade do Hub ao repositório Helm e aos registries das imagens e da função.
- Execute os comandos da raiz do monorepo. Nos exemplos, `hub` seleciona explicitamente o contexto local.

```bash
hub() { kubectl --context kind-platform-management "$@"; }
hub get namespace argocd crossplane-system platform-system
```

## 1. Conferir o Crossplane Core

Na UI do Argo CD, abra **Applications → crossplane-core**. Se o componente já estiver saudável, confira a instalação com os comandos abaixo. Se a etapa exigir a instalação, confira a versão do chart no catálogo e inspecione o **Diff**. Sincronize a Application quando o diff corresponder à instalação esperada. Se ela estiver ausente, confira o `ApplicationSet` na Etapa 02. [Entrada do catálogo](../platform/hub/addons/catalog/crossplane-core.yaml)

```bash
hub -n argocd get application crossplane-core -o \
  'custom-columns=SYNC:.status.sync.status,HEALTH:.status.health.status,REVISION:.status.sync.revision'
hub -n crossplane-system rollout status deployment/crossplane --timeout=180s
hub -n crossplane-system rollout status deployment/crossplane-rbac-manager --timeout=180s
hub wait --for=condition=Established --timeout=120s \
  crd/compositeresourcedefinitions.apiextensions.crossplane.io \
  crd/compositions.apiextensions.crossplane.io crd/functions.pkg.crossplane.io
```

Após sincronizar, confirme `Synced`, `Healthy` e a revisão do chart. Em um Hub existente, o exercício verifica os Deployments e os CRDs de que o `LabMessage` depende. Se o chart não sincronizar, consulte `hub -n argocd describe application crossplane-core` e confira a origem Helm nos logs de `hub -n argocd logs deployment/argocd-repo-server`.

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
hub explain labmessage.spec.message --api-version=lab.platform.example/v1alpha1
hub -n platform-system get labmessage lab03-demo
hub -n platform-system get configmap -l lab.platform.example/exercise=lab03 \
  -o 'custom-columns=NAME:.metadata.name,OWNER:.metadata.ownerReferences[0].name,MESSAGE:.data.mensagem'
```

A saída deve mostrar um ConfigMap cujo owner é `lab03-demo` e cuja mensagem é `reconciliado-pelo-crossplane`. O `Ready` depende de a função observar o ConfigMap com o label do exercício. Se os waits expirarem, consulte `hub describe function.pkg.crossplane.io lab03-patch-and-transform`, `hub describe xrd labmessages.lab.platform.example` e `hub -n platform-system describe labmessage lab03-demo`.

## O que observar

- O campo `spec.message` aparece na API a partir do schema do XRD. O XR é o recurso observado pelo Crossplane; o XRD define o tipo e não representa uma solicitação de infraestrutura.
- O `LabMessage` chega a `Synced` e `Ready`. O ConfigMap identifica o XR como owner e contém a mensagem declarada no manifesto.

## 3. Limpar o exercício

Remova primeiro o XR enquanto a função e o Crossplane ainda estão ativos. Confirme que o ConfigMap composto sumiu antes de remover a API e a função.

```bash
hub delete -f "$demo/xr.yaml" --ignore-not-found --timeout=120s
hub -n platform-system wait --for=delete configmap \
  -l lab.platform.example/exercise=lab03 --timeout=120s
hub delete -f "$demo/composition.yaml" -f "$demo/xrd.yaml" -f "$demo/function.yaml" \
  --ignore-not-found --timeout=120s
```

Mantenha a Application `crossplane-core` sincronizada para as próximas etapas. [Documentação do Crossplane](https://docs.crossplane.io/latest/guides/function-patch-and-transform/)

A Etapa 08 evolui este fluxo para o aceite via GitOps antes do bootstrap do Hub EKS. [Roteiro](../docs/learning-roadmap.md#fase-2--control-plane-local)

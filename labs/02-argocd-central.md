# Etapa 02 — Argo CD central no Hub kind

[← Etapa 01](01-platform-management-cluster.md) · [Guias das etapas](README.md)

## Objetivo

Configurar o Argo CD para descobrir os add-ons do Hub no Git e demonstrar como ele restaura um recurso alterado no cluster.

## O que será feito

- Instalar o Argo CD central via Helm e acessar sua interface por port-forward local.
- Aplicar o bootstrap do `ApplicationSet` `hub-addons` e conferir as Applications de Argo Workflows e Crossplane Core geradas a partir do catálogo no Git.
- Criar o AppProject `lab02` e a Application `lab02-gitops-demo`. Sincronizar manualmente um commit publicado para criar o ConfigMap de demonstração em `platform-system`.
- Alterar o ConfigMap no cluster, observar `OutOfSync` e sincronizar novamente para restaurar o valor do Git.
- Remover os recursos de demonstração e manter o bootstrap dos add-ons para as etapas seguintes.

O exercício roda no Hub local e tem custo AWS zero.

## Por que fizemos assim

O Argo CD é o ponto de entrada para entregar os componentes do Hub a partir do Git. O catálogo gera uma Application por componente para que cada um tenha sua revisão e seu diff. Para observar drift, usamos uma Application separada com um ConfigMap e um SHA fixo. Assim a comparação usa o mesmo manifesto antes e depois da alteração no cluster.

## Pré-requisitos

- Etapa 01 concluída, com `kind`, `kubectl` e `helm` disponíveis.
- Catálogo dos add-ons publicado em `main`, a revisão que o `ApplicationSet` monitora.
- Uma revisão publicada neste monorepo que contenha `workflow-plane/argocd/demo/configmap.yaml`.
- Conectividade do Hub ao GitHub e aos registries das imagens.

Execute os comandos da raiz do monorepo.

Se o Argo CD e o `ApplicationSet` `hub-addons` já estiverem disponíveis, use as consultas do passo 2 para observar as Applications. Para repetir a demonstração de drift, comece no passo 3.

## 1. Instalar e acessar o Argo CD

```bash
bash platform/hub/scripts/verify.sh
bash platform/hub/scripts/install-argocd.sh
kubectl --context kind-platform-management -n argocd get deployments,statefulsets,pods,svc
```

O chart fixado instala os CRDs de `Application`, `AppProject` e `ApplicationSet`, além do controller de `ApplicationSet`. O Service `argocd-server` deve aparecer como `ClusterIP`. [Versão do chart](../workflow-plane/argocd/chart-version.txt)

Deixe o port-forward aberto em outro terminal.

```bash
kubectl --context kind-platform-management -n argocd port-forward \
  --address 127.0.0.1 svc/argocd-server 8081:443
```

Abra `https://localhost:8081`. O servidor usa um certificado local autoassinado. Para o primeiro acesso, use o usuário `admin` e consulte a senha inicial no terminal local.

```bash
kubectl --context kind-platform-management -n argocd \
  get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 --decode
printf '\n'
```

Após o login, troque a senha em **User Info → Update Password**. Mantenha as credenciais fora do Git e de registros compartilhados.

## 2. Configurar os add-ons do Hub

```bash
hub() { kubectl --context kind-platform-management "$@"; }
hub apply -f platform/hub/addons/bootstrap.yaml
hub -n argocd wait --for=condition=ResourcesUpToDate --timeout=180s \
  applicationset/hub-addons
hub -n argocd wait --timeout=180s \
  --for='jsonpath={.metadata.ownerReferences[?(@.kind=="ApplicationSet")].name}=hub-addons' \
  application/argo-workflows application/crossplane-core
hub -n argocd get applicationset hub-addons
hub -n argocd get application argo-workflows crossplane-core \
  -o 'custom-columns=APP:.metadata.name,OWNER:.metadata.ownerReferences[?(@.kind=="ApplicationSet")].name,SYNC:.status.sync.status'
```

O `ApplicationSet` lê `platform/hub/addons/catalog/*.yaml` em `main` e gera uma `Application` por arquivo. O bootstrap também declara os AppProjects que limitam a origem, o destino e os tipos de recurso de cada add-on. A coluna `OWNER` deve mostrar `hub-addons` nas duas linhas. Em um Hub novo, cada add-on aguarda o sync manual indicado nas Etapas 03 e 04. [Bootstrap do Hub](../platform/hub/addons/bootstrap.yaml)

Se o Hub já tiver essas Applications, use os mesmos comandos. O `ApplicationSet` assume as Applications com os nomes existentes e mantém as versões dos charts, as configurações e os nomes dos releases. Confira a coluna `OWNER` antes de seguir.

Se uma Application estiver ausente ou sem owner, consulte `hub -n argocd describe applicationset hub-addons` e os logs de `deployment/argocd-applicationset-controller` no namespace `argocd`. Confira se a revisão `main` contém o catálogo e se o Hub alcança o GitHub. Em seguida, repita a verificação.

## 3. Sincronizar uma revisão publicada

O monorepo deste lab é público. O Argo CD lê a URL HTTPS declarada na Application sem credenciais. Comece pelo comando abaixo.

```bash
read -r -p 'SHA completo do commit publicado: ' lab02_sha
bash workflow-plane/argocd/bootstrap/apply-demo.sh "$lab02_sha"
```

O AppProject `lab02` restringe a origem ao monorepo e permite apenas ConfigMaps em `platform-system` no Hub. Na UI, abra `lab02-gitops-demo`. Confira que a revisão é o SHA informado e que o diff se limita ao ConfigMap `lab02-gitops-demo`. Se estiver `OutOfSync` e o diff corresponder ao exercício, selecione **Sync → Synchronize**.

```bash
kubectl --context kind-platform-management -n argocd get application lab02-gitops-demo \
  -o 'custom-columns=SYNC:.status.sync.status,HEALTH:.status.health.status,REVISION:.status.sync.revision'
kubectl --context kind-platform-management -n platform-system get configmap lab02-gitops-demo \
  -o 'jsonpath={.data.mensagem}{"\n"}'
```

O valor esperado é `definido-no-git`. Se a aplicação mostrar `Unknown` ou `ComparisonError`, inspecione `kubectl --context kind-platform-management -n argocd describe application lab02-gitops-demo`. Confira a conectividade ao GitHub, o SHA publicado e o caminho. Consulte os logs de `deployment/argocd-repo-server` no namespace `argocd` se a comparação continuar falhando.

## 4. Provocar drift e recuperar

```bash
kubectl --context kind-platform-management -n platform-system patch configmap lab02-gitops-demo \
  --type merge -p '{"data":{"mensagem":"alterado-no-cluster"}}'
```

Use **Refresh** na aplicação e observe `OutOfSync`. Abra **Diff** e compare o valor no cluster com o valor do Git. Execute **Sync** novamente e repita os comandos de verificação. A revisão deve continuar a mesma e o valor deve voltar a `definido-no-git`.

## O que observar

- As Applications `argo-workflows` e `crossplane-core` têm `hub-addons` como owner e podem ser sincronizadas separadamente.
- Na Application de demonstração, o valor editado aparece no diff. O sync restaura `definido-no-git` usando a revisão original.

## 5. Limpar o exercício

```bash
kubectl --context kind-platform-management -n argocd delete application lab02-gitops-demo --ignore-not-found
kubectl --context kind-platform-management -n platform-system delete configmap lab02-gitops-demo --ignore-not-found
kubectl --context kind-platform-management -n argocd delete appproject lab02 --ignore-not-found
bash platform/hub/scripts/verify.sh
```

Encerre o port-forward com Ctrl+C. Mantenha o Argo CD e o `ApplicationSet` `hub-addons` para a Etapa 03. [Operações do Hub](../platform/hub/README.md)

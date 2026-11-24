# Etapa 02 — Argo CD central no Hub kind

[← Etapa 01](01-platform-management-cluster.md) · [Guias das etapas](README.md)

## Objetivo

Configurar o Argo CD para descobrir os add-ons do Hub no Git e acompanhar a entrega de um recurso a partir da branch `main`.

## O que será feito

- Instalar o Argo CD central via Helm e abrir sua interface HTTPS pelo Traefik.
- Aplicar o bootstrap do `ApplicationSet` `hub-addons` e conferir as Applications de Argo Workflows e Crossplane Core geradas a partir do catálogo no Git.
- Criar o AppProject `lab02` e a Application `lab02-gitops-demo`. Sincronizar manualmente a Application para entregar o ConfigMap de demonstração de `main` em `platform-system`.
- Remover os recursos de demonstração e manter o bootstrap dos add-ons para as etapas seguintes.

O exercício roda no Hub local e tem custo AWS zero.

## Por que fizemos assim

O Argo CD entrega os componentes do Hub a partir do Git. A rota da sua interface entra no mesmo release Helm e usa o Traefik da Etapa 01 para encaminhar HTTPS até o Argo CD. O catálogo gera uma Application por componente para que cada um tenha sua revisão e seu diff. A Application de demonstração acompanha `main` e permite observar como o Argo CD aplica o manifesto do ConfigMap no Hub.

## Pré-requisitos

- Etapa 01 concluída, com `kind`, `kubectl`, `helm` e o provider de CRDs do Traefik habilitado.
- Catálogo dos add-ons e manifesto do ConfigMap de demonstração publicados em `main`.
- Conectividade do Hub ao GitHub e aos registries das imagens.

Execute os comandos da raiz do monorepo.

Se a interface e o `ApplicationSet` `hub-addons` já estiverem disponíveis, use as consultas do passo 2 para observar as Applications. Para repetir a entrega do ConfigMap, comece no passo 3.

## 1. Instalar e acessar o Argo CD

Se o Hub ainda usa a configuração antiga do Traefik, execute `bash platform/hub/scripts/install-traefik.sh` com a versão atual do repositório antes de instalar ou atualizar o Argo CD.

```bash
bash platform/hub/scripts/verify.sh
bash platform/hub/scripts/install-argocd.sh
kubectl --context kind-platform-management -n argocd get deployments,statefulsets,pods,svc
kubectl --context kind-platform-management -n argocd get ingressroutetcp argocd-server
```

O chart fixado instala os CRDs de `Application`, `AppProject` e `ApplicationSet`, além do controller de `ApplicationSet`. O Service `argocd-server` deve aparecer como `ClusterIP`. [Versão do chart](../workflow-plane/argocd/chart-version.txt)

Abra **https://argocd.localhost:8443**. O Traefik encaminha a conexão ao Argo CD, que apresenta o certificado autoassinado do laboratório. Para o primeiro acesso, use o usuário `admin` e consulte a senha inicial no terminal local. [Referência do IngressRouteTCP](https://doc.traefik.io/traefik/reference/routing-configuration/kubernetes/crd/tcp/ingressroutetcp/)

```bash
kubectl --context kind-platform-management -n argocd \
  get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 --decode
printf '\n'
```

Após o login, troque a senha em **User Info → Update Password**. Mantenha as credenciais fora do Git e de registros compartilhados. Se a URL responder com `404 page not found`, confira o `IngressRouteTCP` e a atualização do Traefik no passo 1.

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

## 3. Sincronizar o ConfigMap a partir de main

O monorepo é público. A Application aponta para `main` e para o diretório do manifesto de demonstração.

```bash
kubectl --context kind-platform-management -n argocd apply \
  -f workflow-plane/argocd/bootstrap/project.yaml \
  -f workflow-plane/argocd/bootstrap/application.yaml
```

O AppProject `lab02` restringe a origem ao monorepo e permite apenas ConfigMaps em `platform-system` no Hub. Na UI, abra `lab02-gitops-demo`. Confira no diff o ConfigMap `lab02-gitops-demo`. Antes do primeiro sync, a Application aparece como `OutOfSync` e o ConfigMap como `Missing`. Selecione **Sync → Synchronize** para aplicar o manifesto.

```bash
kubectl --context kind-platform-management -n argocd get application lab02-gitops-demo \
  -o 'custom-columns=SYNC:.status.sync.status,HEALTH:.status.health.status,REVISION:.status.sync.revision'
kubectl --context kind-platform-management -n platform-system get configmap lab02-gitops-demo \
  -o 'jsonpath={.data.mensagem}{"\n"}'
```

O valor esperado é `definido-no-git`. Se a aplicação mostrar `Unknown` ou `ComparisonError`, inspecione `kubectl --context kind-platform-management -n argocd describe application lab02-gitops-demo`. Confira a conectividade ao GitHub e o caminho do manifesto em `main`. Consulte os logs de `deployment/argocd-repo-server` no namespace `argocd` se a comparação continuar falhando.

## O que observar

- As Applications `argo-workflows` e `crossplane-core` têm `hub-addons` como owner e podem ser sincronizadas separadamente.
- A Application de demonstração fica `Synced` e `Healthy`. A coluna `REVISION` mostra o commit de `main` que o Argo CD resolveu. O ConfigMap contém `definido-no-git`.

## 4. Limpar o exercício

```bash
kubectl --context kind-platform-management -n argocd delete application lab02-gitops-demo --ignore-not-found
kubectl --context kind-platform-management -n platform-system delete configmap lab02-gitops-demo --ignore-not-found
kubectl --context kind-platform-management -n argocd delete appproject lab02 --ignore-not-found
bash platform/hub/scripts/verify.sh
```

Mantenha o Argo CD, a rota e o `ApplicationSet` `hub-addons` para a Etapa 03. [Operações do Hub](../platform/hub/README.md)

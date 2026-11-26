# Etapa 05 — Controller e pipeline do TerraformStack

## Objetivo

Instalar o controller e a pipeline canônica que executa stacks Terraform a partir da intenção sincronizada pelo Argo CD.

## O que será feito

Construir a imagem do controller e o runner genérico. Publicar o WorkflowTemplate, seus scripts e a política OPA em ConfigMaps versionados. Instalar o controller com o catálogo do Hub local.

## Por que fizemos assim

O controller Go utiliza Controller Runtime no padrão Kubebuilder, com watches, cache, work queue e Leader Election. Dois Pods disputam uma Lease; apenas o líder reconcilia. Drift e retry usam temporizadores por stack. [Implementação](../control-plane/terraformstack/controller/README.md).

O catálogo seleciona fonte, destino, backend e identidade. O controller fixa os SHAs de código e configuração, grava a referência da execução no CR e acompanha o Workflow. Os steps Terraform usam a imagem comum com ferramentas e executam o script montado. OPA usa sua imagem oficial e monta o pacote versionado do [Policy Engine](../workflow-plane/policy-engine/README.md), com perfis selecionados pelo catálogo. O overlay dá ao controller o catálogo do Hub e pode promover o runner por digest via `images`. O step `verify` executa `verify.sh` quando a raiz aprovada inclui uma verificação específica do recurso. [Fluxo da pipeline](../docs/flows/terraformstack.md#3-pipeline-e-aprovação)

## Pré-requisitos

Hub `platform-management`, Argo CD e Argo Workflows instalados. CRD estabelecido pela Etapa 04. Execute os comandos na raiz do repositório. [Hub local](01-platform-management-cluster.md) · [Etapa 04](04-terraformstack-crd.md)

## 1. Construir ou reconstruir as imagens

Confira se o checkout corresponde à revisão que será sincronizada. Os nomes do contrato precisam corresponder ao WorkflowTemplate. Kustomize monta a configuração imutável no controller. Scripts, RBAC e admissão acompanham o mesmo contrato. [Configuração](../control-plane/terraformstack/controller/controller-config.yaml)

No Hub existente, aguarde os Workflows ativos e os gates de aprovação antes de substituir o runner na mesma tag ou alterar o contrato. Se precisar impedir novas execuções durante a atualização, suspenda as solicitações pelo GitOps após concluir as tentativas em curso.

Use este bloco na primeira instalação e para reconstruir o controller ou o runner. O build atualiza as imagens no Docker. O `kind load` carrega as imagens no nó do Hub.

```bash
hub() { kubectl --context kind-platform-management "$@"; }

docker build -t terraformstack-controller:local control-plane/terraformstack/controller &&
docker build -t terraformstack-runner:local workflow-plane/terraform-executor &&
kind load docker-image --name platform-management \
  terraformstack-controller:local terraformstack-runner:local
```

Se a imagem do controller mudou com a mesma tag e a sincronização do manifesto deixou o Deployment com os Pods anteriores, use o restart após o `kind load`. Na primeira instalação, siga para o sync dos componentes.

```bash
hub -n platform-system rollout restart deployment/terraformstack-controller &&
hub -n platform-system rollout status deployment/terraformstack-controller --timeout=120s
hub -n platform-system get pods -l app=terraformstack-controller \
  -o 'custom-columns=NAME:.metadata.name,IMAGE:.spec.containers[0].image,IMAGEID:.status.containerStatuses[0].imageID'
```

Confira as duas réplicas e os image IDs. Em `ImagePullBackOff` com tag `:local`, consulte `docker exec platform-management-control-plane crictl images` para verificar a presença das imagens no nó. O comando abaixo confirma se a tag do Docker aponta para o controller Go. O resultado esperado é `["/manager"]`.

```bash
docker image inspect --format '{{json .Config.Entrypoint}}' terraformstack-controller:local
```

Após a atualização, restaure as solicitações suspensas via GitOps e confira a imagem do runner em um novo Workflow do Simulator. A promoção para AWS usa imagens publicadas por digest. [Guia do Simulator](06-terraformstack-simulator.md#1-preparar-o-hub-e-sincronizar)

## 2. Sincronizar os componentes

Use o bootstrap do Hub para registrar as Applications. Se já estiverem cadastradas e a revisão atualizar apenas a imagem local, basta reconstruir, carregar e reiniciar o Deployment conforme o passo anterior.

```bash
hub apply -f platform/hub/addons/bootstrap.yaml
```

No Argo CD, confira a revisão e o destino Hub antes de cada sync manual.

1. **terraformstack-workflows** em `argo-workflows`. O diff deve mostrar o WorkflowTemplate, a ServiceAccount, o RBAC e os ConfigMaps imutáveis com scripts e Rego.
2. **terraformstack-controller** em `platform-system`. O diff deve mostrar Deployment, configuração do controller e catálogo imutáveis, permissões, namespace do Simulator, PVCs e políticas de admissão.

Se o catálogo aparecer sem hash no diff, confira o build do overlay com `kubectl kustomize control-plane/terraformstack/overlays/kind`.

Para testar uma revisão publicada em branch, informe o SHA em **Sync → Revision** na própria Application de cada componente. Aguarde a operação terminar e confira o SHA no resultado do sync. O Argo CD pode mostrar `OutOfSync` enquanto compara com o destino `main` configurado no catálogo. [Catálogo do Workflow](../platform/hub/addons/catalog/terraformstack-workflows.yaml) · [Catálogo do controller](../platform/hub/addons/catalog/terraformstack-controller.yaml)

```bash
hub -n platform-system rollout status deployment/terraformstack-controller --timeout=120s
hub -n platform-system get pods -l app=terraformstack-controller
hub -n platform-system get lease terraformstack-controller.platform.example.io
hub -n platform-system logs -l app=terraformstack-controller --tail=30
hub -n argo-workflows get workflowtemplate terraformstack
hub -n platform-system auth can-i get secrets --as=system:serviceaccount:platform-system:terraformstack-controller
hub -n argo-workflows get pvc simulator-network simulator-dependent
```

## O que observar

Observe duas réplicas, a Lease com holderIdentity e os logs estruturados do líder. O controller fica disponível e o WorkflowTemplate lista os steps. A consulta de Secrets retorna `no`. Os PVCs podem ficar `Pending` até o primeiro Pod quando a classe usa `WaitForFirstConsumer`. A etapa seguinte sincroniza as solicitações e valida os efeitos reais no Simulator.

[← Etapa 04](04-terraformstack-crd.md) · [Etapa 06 →](06-terraformstack-simulator.md)

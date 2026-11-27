# Etapa 05 — Controller e pipeline do TerraformStack

## Objetivo

Instalar o controller e a pipeline canônica que executa stacks Terraform a partir da intenção sincronizada pelo Argo CD.

## O que será feito

Construir a imagem do controller e o runner genérico. Publicar o WorkflowTemplate, seus scripts, o perfil local e a política OPA em ConfigMaps imutáveis. Instalar o controller com a configuração de integração Argo.

## Por que fizemos assim

O controller Go utiliza Controller Runtime no padrão Kubebuilder, com watches, cache, work queue e Leader Election. Dois Pods disputam uma Lease; apenas o líder reconcilia. Drift e retry usam temporizadores por stack. [Implementação](../control-plane/terraformstack/controller/README.md).

O controller registra a intenção, verifica dependências e acompanha os sinais do Workflow. A pipeline resolve Git, contexto, PVC, backend e configuração a partir do perfil aprovado. Os steps Terraform usam a imagem comum e executam o script montado. O OPA roda em sua imagem oficial e monta o pacote versionado. O step `verify` pode executar `verify.sh` na raiz da stack. [Fluxo da pipeline](../docs/flows/terraformstack.md#3-pipeline-e-aprovação) · [Perfil do Simulator](../workflow-plane/argo-workflows/terraformstack/simulator.json)

## Pré-requisitos

Hub `platform-management`, Argo CD e Argo Workflows instalados. CRD estabelecido pela Etapa 04. Execute os comandos na raiz do repositório. [Hub local](01-platform-management-cluster.md) · [Etapa 04](04-terraformstack-crd.md)

## 1. Construir ou reconstruir as imagens

Confira se o checkout corresponde à revisão que será sincronizada. Os nomes do contrato precisam corresponder ao WorkflowTemplate. O `controller-config.yaml` cadastra a integração de referência com `authorization: true`; os CRs do Simulator podem omitir `workflowRef` e usar o default. Não é necessário criar outras pipelines para concluir este lab. Kustomize monta a configuração imutável no controller. Scripts, RBAC e admissão acompanham o mesmo contrato. [Configuração](../control-plane/terraformstack/controller/controller-config.yaml)

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

1. **terraformstack-workflows** em `argo-workflows`. O diff deve mostrar o WorkflowTemplate, a ServiceAccount, o RBAC e os ConfigMaps imutáveis com scripts, perfil do Simulator e Rego.
2. **terraformstack-controller** em `platform-system`. O diff deve mostrar Deployment, configuração imutável do controller, permissões, namespace do Simulator, PVCs e políticas de admissão.

Confira os nomes por hash dos ConfigMaps nos builds de Kustomize. A Application de Workflows monta os scripts, o perfil e a política. O Deployment monta a configuração de integração.

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

## Atualizar um Hub da revisão anterior

Conclua os Workflows antigos e suspenda os CRs pelo GitOps. Antes de trocar de contrato, conclua o destroy aprovado dos recursos descartáveis do Simulator com a revisão antiga. Depois da verificação, remova as solicitações antigas e descarte apenas seus PVCs de teste. Um PVC com state anterior sem vínculo fica bloqueado no runtime novo. Preserve qualquer state que precise ser recuperado e mantenha a revisão que o criou disponível. A POC exige uma migração explícita para reaproveitar esse state.

Para a evolução de seleção de Workflows, sem mudança de binding ou state, basta drenar as tentativas e atualizar CRD, configuração e imagem do controller; não é necessário descartar o PVC por causa da seleção de alias. Mantenha a integração de referência com autorização habilitada.

Sincronize o CRD da Etapa 04, carregue as imagens reconstruídas, sincronize os Workflows e o controller e depois sincronize o Simulator com a mesma revisão. Atualize os manifestos das solicitações para `sourceRef.repository` e `context`. O perfil e os outputs devem corresponder à configuração instalada. [Etapa 04](04-terraformstack-crd.md) · [Execution Contract](../docs/flows/terraformstack-execution-contract.md)

## O que observar

Observe duas réplicas, a Lease com holderIdentity e os logs estruturados do líder. O controller fica disponível e o WorkflowTemplate lista os steps. A consulta de Secrets retorna `no`. Os PVCs podem ficar `Pending` até o primeiro Pod quando a classe usa `WaitForFirstConsumer`. A etapa seguinte sincroniza as solicitações e valida os efeitos reais no Simulator.

[← Etapa 04](04-terraformstack-crd.md) · [Etapa 06 →](06-terraformstack-simulator.md)

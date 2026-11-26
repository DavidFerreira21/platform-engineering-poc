# Etapa 05 — Controller e pipeline do TerraformStack

## Objetivo

Instalar o controller e a pipeline canônica que executa stacks Terraform a partir da intenção sincronizada pelo Argo CD.

## O que será feito

Construir a imagem do controller e o runner genérico. Publicar o WorkflowTemplate, seus scripts e a política OPA em ConfigMaps versionados. Instalar o controller com o catálogo do Hub local.

## Por que fizemos assim

O catálogo fixa fonte, destino, backend e identidade. O controller resolve a intenção, coordena o lifecycle do Workflow e publica o estado no CR. O Argo Workflows torna visíveis os steps de preparação, init, validate, plan, OPA, aprovação, apply e verify. Os steps Terraform usam a imagem comum com ferramentas. OPA usa sua imagem oficial e monta as regras do Git. O manifest base do controller recebe valores específicos do cluster pelo overlay. Um overlay pode promover o runner por digest via `images`. O step Verify executa `verify.sh` quando a raiz aprovada inclui uma verificação específica do recurso.

## Pré-requisitos

Hub `kind`, Argo CD e Argo Workflows instalados. CRD estabelecido pela [Etapa 04](04-terraformstack-crd.md). Execute os comandos na raiz do repositório.

## 1. Construir e carregar as imagens

```bash
cluster="${HUB_KIND_NAME:-platform-management}"
hub() { kubectl --context "kind-$cluster" "$@"; }

docker build -t terraformstack-controller:local control-plane/terraformstack/controller
docker build -t terraformstack-runner:local workflow-plane/terraform-executor
kind load docker-image --name "$cluster" terraformstack-controller:local terraformstack-runner:local
hub apply -f platform/hub/addons/bootstrap.yaml
```

Se o diff da Application do controller mostrar o catálogo sem hash de conteúdo, confira o build do overlay com `kubectl kustomize control-plane/terraformstack/overlays/kind`. Se houver erro de imagem, compare o nome e a tag da revisão sincronizada com as imagens carregadas no nó.

## 2. Sincronizar os componentes

No Argo CD, confira a revisão e o destino Hub antes de cada sync manual.

1. **terraformstack-workflows** em `argo-workflows`. O diff deve mostrar o WorkflowTemplate, a ServiceAccount, o RBAC e os ConfigMaps imutáveis com scripts e Rego.
2. **terraformstack-controller** em `platform-system`. O diff deve mostrar Deployment, catálogo, permissões, namespace do Simulator, PVCs e políticas de admissão.

Para uma revisão ainda em branch, use Applications de preview com os mesmos projetos, paths e destinos e um SHA acessível ao Hub. Sincronize cada bloco por uma única Application. [Catálogo do Workflow](../platform/hub/addons/catalog/terraformstack-workflows.yaml) · [Catálogo do controller](../platform/hub/addons/catalog/terraformstack-controller.yaml)

```bash
hub -n platform-system rollout status deployment/terraformstack-controller --timeout=120s
hub -n argo-workflows get workflowtemplate terraformstack
hub -n platform-system auth can-i get secrets --as=system:serviceaccount:platform-system:terraformstack-controller
hub -n argo-workflows get pvc simulator-network simulator-dependent
```

## O que observar

O controller fica disponível e o WorkflowTemplate lista os steps. A consulta de Secrets retorna `no`. Os PVCs podem ficar `Pending` até o primeiro Pod quando a classe usa `WaitForFirstConsumer`. A etapa seguinte sincroniza as solicitações e valida os efeitos reais no Simulator.

[← Etapa 04](04-terraformstack-crd.md) · [Etapa 06 →](06-terraformstack-simulator.md)

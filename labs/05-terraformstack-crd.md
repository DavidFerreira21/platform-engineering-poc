# Etapa 05 — API do TerraformStack via GitOps

[← Etapa 04](04-crossplane-core.md) · [Guias das etapas](README.md)

## Objetivo

Publicar o contrato `TerraformStack` no Hub e observar como a API Kubernetes valida uma solicitação de infraestrutura.

## O que será feito

- Atualizar o bootstrap do Hub com o AppProject `platform-terraformstack-crd` e conferir a Application `terraformstack-crd` gerada pelo catálogo.
- Sincronizar o CRD e inspecionar os campos de `spec` e o subrecurso `/status`.
- Validar uma solicitação fictícia com dry-run no servidor. O exercício roda no Hub local e tem custo AWS zero.

## Por que fizemos assim

Publicamos o contrato em uma Application própria para inspecionar o que a API aceita antes de conectar um executor. O `spec` descreve uma **stack** pela referência `sourceRef` e o destino lógico. A API também aceita o formato legado `module`, mas **não os dois no mesmo CR**. O `/status` é a interface que o controller usará para registrar o que aconteceu com a solicitação. O dry-run permite revisitar a etapa quando o controller já estiver ativo. [ADR-0006](../docs/adr/0006-terraformrun-and-application-ci.md)

## Pré-requisitos

- Hub `kind-platform-management` com o Argo CD central e o `ApplicationSet` `hub-addons` da Etapa 02. A interface está em **https://argocd.localhost:8443**.
- Checkout atualizado da `main` publicada com o CRD e a entrada do catálogo. Execute os comandos na raiz do monorepo.

## 1. Conferir a Application no Argo CD

```bash
hub() { kubectl --context kind-platform-management "$@"; }
hub apply -f platform/hub/addons/bootstrap.yaml
hub -n argocd get appproject platform-terraformstack-crd
hub -n argocd wait --timeout=180s \
  --for='jsonpath={.status.resources[?(@.name=="terraformstack-crd")].name}=terraformstack-crd' \
  applicationset/hub-addons
hub -n argocd get application terraformstack-crd \
  -o 'custom-columns=NAME:.metadata.name,PROJECT:.spec.project,TARGET:.spec.source.targetRevision,SYNC:.status.sync.status'
```

Na interface do Argo CD, abra **Applications → terraformstack-crd**. Confira que a origem é `main` e que o recurso esperado no **Diff** é o CRD `terraformstacks.platform.example.io`. Se a Application estiver `OutOfSync` com essa alteração, selecione **Sync → Synchronize**. Em um Hub já sincronizado, siga para a inspeção.

Se a Application estiver ausente, confirme que a entrada do catálogo está publicada em `main` e consulte `hub -n argocd describe applicationset hub-addons`. Se o sync indicar que o CRD é proibido, confira o AppProject e a aplicação do bootstrap atualizado. [Entrada do catálogo](../platform/hub/addons/catalog/terraformstack-crd.yaml)

## 2. Inspecionar o contrato e validar uma solicitação

```bash
hub wait --for=condition=Established --timeout=90s crd/terraformstacks.platform.example.io
hub api-resources --api-group=platform.example.io
```

A descoberta pode levar alguns segundos após a condição `Established`. Repita a consulta até `terraformstacks` aparecer. Em seguida, inspecione o contrato e valide a solicitação.

```bash
hub explain terraformstacks.spec --api-version=platform.example.io/v1alpha1 --recursive
hub explain terraformstacks.status.conditions --api-version=platform.example.io/v1alpha1
hub get --raw /apis/platform.example.io/v1alpha1
hub create --dry-run=server --validate=strict \
  -f control-plane/terraformstack/demo/terraformstack.yaml -o yaml
```

A descoberta da API lista `terraformstacks` e `terraformstacks/status`. A última saída mostra o exemplo aceito pela API. A fonte fictícia `foundation-stacks`, o path `stacks/lab-network` e o destino `entity-a/uat` são dados do exercício. A ausência de `sourceRef.version` e de `configuration.revision` representa `latest` no contrato; **o dry-run não resolve nenhuma versão no Registry/Git e não provisiona recursos**. A resolução dos destinos AWS entra mais adiante no roteiro. [Schema do CRD](../control-plane/terraformstack/crd/terraformstack.yaml) · [Solicitação de exemplo](../control-plane/terraformstack/demo/terraformstack.yaml)

Nas Etapas 08–09, as stacks Terraform passam a receber valores dos YAMLs em `config/`. O módulo auxiliar Configuration resolve a herança entre padrão global, entidade e ambiente. O `TerraformStack` seleciona a revisão desses arquivos em `spec.configuration.revision`. [Decisão sobre configuração por entidade e ambiente](../docs/adr/0006-terraformrun-and-application-ci.md#configuração-por-entidade-e-ambiente)

## O que observar

- O Argo CD sincroniza a definição do tipo e o Kubernetes informa a condição `Established`.
- O `spec` descreve a intenção. O subrecurso `/status` fica disponível para o controller que será ligado na Etapa 06.
- O servidor devolve a solicitação válida. O dry-run termina sem gravar uma stack para reconciliação.

Mantenha a Application e o CRD no Hub para a Etapa 06.

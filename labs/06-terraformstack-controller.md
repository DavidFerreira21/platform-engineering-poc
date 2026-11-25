# Etapa 06 — Controller do TerraformStack e executor simulado

[← Etapa 05](05-terraformstack-crd.md) · [Guias das etapas](README.md) · [Etapa 07 →](07-registro-ambientes-aws.md)

## Objetivo

Sincronizar um `TerraformStack` pelo Argo CD, acompanhar o Workflow criado pelo controller e verificar o resultado da simulação no status da solicitação.

## O que será feito

- Carregar a imagem do controller no Hub local e sincronizar a Application `terraformstack-controller`.
- Sincronizar a Application `terraformstack-demo` e acompanhar a execução simulada usando os logs do Argo Workflows.
- Conferir o status, a idempotência e o tratamento de falha da execução.

## Por que fizemos assim

A solicitação representa a intenção no Git. O controller acompanha a geração do CR e cria um Workflow com nome derivado do UID e da geração. O Workflow recebe o módulo, a versão e o destino lógico e registra o plan simulado. Essa divisão permite testar GitOps, reconciliação e execução no Hub antes de introduzir Terraform e identidade AWS. O status informa o resultado apenas da geração observada. [ADR-0006](../docs/adr/0006-terraformrun-and-application-ci.md) · [WorkflowTemplate](https://argo-workflows.readthedocs.io/en/release-4.1/workflow-templates/)

O controller usa dois watches da API Kubernetes. O primeiro acompanha a stack de exemplo e ignora alterações que afetam somente o status. O segundo acompanha os Workflows do lab no namespace `argo-workflows`. Os dois sinalizam o mesmo loop de reconciliação. A leitura inicial, o restabelecimento dos watches e o polling recuperam mudanças mesmo se o Pod reiniciar ou perder a conexão.

| Variável do Deployment | Padrão | Função |
| --- | --- | --- |
| `POLL_INTERVAL_SECONDS` | 30 | Espera máxima por um evento antes de reconciliar novamente. |
| `WATCH_TIMEOUT_SECONDS` | 60 | Duração solicitada para cada conexão watch antes de renovar a observação. |
| `WATCH_RETRY_SECONDS` | 2 | Pausa antes de restabelecer um watch que falhou. |

Os valores são inteiros positivos em segundos. O polling relê a stack e o Workflow da geração atual. O agendamento de novos plans para detectar drift entra na Etapa 24.

## Pré-requisitos

- Docker ativo e Hub `kind-platform-management` acessível. Executar os comandos na raiz do monorepo.
- Argo CD central e catálogo do Hub da [Etapa 02](02-argocd-central.md), Argo Workflows da [Etapa 03](03-argo-workflows.md) e CRD estabelecido da [Etapa 05](05-terraformstack-crd.md).

## 1. Preparar o Hub e sincronizar o controller

```bash
hub() { kubectl --context kind-platform-management "$@"; }
hub -n argo-workflows rollout status deployment/argo-workflows-workflow-controller --timeout=120s
hub -n argo-workflows get serviceaccount argo-workflow
hub wait --for=condition=Established --timeout=90s crd/terraformstacks.platform.example.io

docker build -t terraformstack-controller:lab06-v2 control-plane/terraformstack/controller
kind load docker-image terraformstack-controller:lab06-v2 --name platform-management

kubectl kustomize control-plane/terraformstack/lab06 | hub apply --dry-run=server -f -
kubectl kustomize control-plane/terraformstack/demo/lab06 | hub apply --dry-run=server -f -
hub apply -f platform/hub/addons/bootstrap.yaml
```

No Argo CD, abra **terraformstack-controller**. Confira a revisão `main`, o destino Hub local e o diff esperado com um Deployment, uma ServiceAccount, duas Roles, dois RoleBindings e um WorkflowTemplate. Sincronize manualmente e confirme a disponibilidade:

```bash
hub -n platform-system rollout status deployment/terraformstack-controller --timeout=120s
hub -n argo-workflows get workflowtemplate terraformstack-plan-simulado
```

Se a Application estiver ausente, confira a revisão publicada e o `ApplicationSet` `hub-addons`. Em `ErrImageNeverPull`, confira se a tag do Deployment é a mesma carregada no cluster. Para reproduzir antes do merge, crie Applications temporárias no Argo CD com os mesmos projetos, origens, caminhos e destinos dos dois arquivos do catálogo, fixe `targetRevision` no SHA publicado da branch e use nomes com sufixo `-preview`. Use as Applications temporárias em um Hub onde esses recursos ainda não estejam gerenciados pelas Applications da `main` e remova-as após o exercício. [Entrada do controller](../platform/hub/addons/catalog/terraformstack-controller.yaml) · [Entrada da solicitação](../platform/hub/addons/catalog/terraformstack-demo.yaml)

## 2. Sincronizar e acompanhar a solicitação

Abra **terraformstack-demo**. Confira a mesma revisão, o destino `platform-system` do Hub e o diff do `TerraformStack` `lab06-network`. Sincronize manualmente.

```bash
stack=terraformstack/lab06-network
hub -n platform-system wait --for=condition=SimulatedPlanReady --timeout=240s "$stack"
hub -n platform-system get "$stack" -o yaml

stack_uid=$(hub -n platform-system get "$stack" -o jsonpath='{.metadata.uid}')
stack_generation=$(hub -n platform-system get "$stack" -o jsonpath='{.metadata.generation}')
workflow="lab06-plan-${stack_uid}-${stack_generation}"
hub -n argo-workflows get workflow "$workflow" -o jsonpath='{.status.phase}{"\n"}'
hub -n argo-workflows logs -l "workflows.argoproj.io/workflow=$workflow" -c main
```

## O que observar

- No CR, `SimulatedPlanReady=True`, `reason=WorkflowSucceeded` e `observedGeneration` igual a `metadata.generation`. A mensagem identifica a execução e a fase.
- Em `argo-workflows`, a fase é `Succeeded`. O log esperado é `lab06-plan-simulado: módulo=lab-network versão=1.0.0 destino=entity-a/uat`. UAT é um identificador lógico nesta simulação.
- Se a execução não concluir, consulte `hub -n platform-system logs deployment/terraformstack-controller` e `hub -n argo-workflows describe workflow "$workflow"`. A condição começa com `Unknown` enquanto a execução está em andamento. Em falha terminal, o controller registra `False`. Uma nova tentativa requer nova geração ou recriação da solicitação.

## 3. Verificar idempotência, falha e permissões

Reinicie o Deployment, aguarde o rollout e repita o sync de `terraformstack-demo` na mesma revisão. O seletor deve retornar um único Workflow, com o mesmo nome e UID de antes. A identidade do controller deve receber `no` nas consultas de Secrets.

```bash
selector="lab.platform.example/stack-uid=$stack_uid,lab.platform.example/stack-generation=$stack_generation"
hub -n argo-workflows get workflows -l "$selector" -o custom-columns=NAME:.metadata.name,UID:.metadata.uid
hub -n platform-system rollout restart deployment/terraformstack-controller
hub -n platform-system rollout status deployment/terraformstack-controller --timeout=120s
```

Após repetir o sync na interface:

```bash
hub -n argo-workflows get workflows -l "$selector" -o custom-columns=NAME:.metadata.name,UID:.metadata.uid
hub -n platform-system auth can-i get secrets --as=system:serviceaccount:platform-system:terraformstack-controller
hub -n argo-workflows auth can-i get secrets --as=system:serviceaccount:platform-system:terraformstack-controller
```

Para exercitar a falha no Hub, substitua temporariamente o comando do template do lab por `false` e recrie a solicitação pelo sync manual. O patch é local ao exercício e deve aparecer no diff da Application do controller.

```bash
hub -n argo-workflows patch workflowtemplate terraformstack-plan-simulado --type=json \
  -p='[{"op":"replace","path":"/spec/templates/0/container/command","value":["false"]},{"op":"remove","path":"/spec/templates/0/container/args"}]'
hub -n platform-system delete "$stack"
```

Sincronize `terraformstack-demo` novamente. Aguarde a condição de falha e confira se o motivo é `WorkflowFailed` ou `WorkflowError` para a nova solicitação.

```bash
hub -n platform-system wait --for=condition=SimulatedPlanReady=False --timeout=240s "$stack"
hub -n platform-system get "$stack" -o yaml
hub -n argo-workflows get workflows -l lab.platform.example/exercise=lab06
```

Sincronize `terraformstack-controller` para restaurar o template versionado antes da limpeza.

## 4. Limpar os recursos do exercício

Remova a solicitação antes dos Workflows. Mantenha o CRD, o Argo CD e o Argo Workflows instalados para reutilização.

```bash
hub -n platform-system delete terraformstack lab06-network --ignore-not-found
hub -n argo-workflows delete workflows -l lab.platform.example/exercise=lab06
hub delete -k control-plane/terraformstack/lab06 --ignore-not-found
hub -n argo-workflows get workflows -l lab.platform.example/exercise=lab06
```

As Applications do catálogo ficam disponíveis para um novo sync e podem aparecer como `OutOfSync` após a limpeza. Caso tenha usado Applications temporárias, remova apenas as duas com sufixo `-preview`.

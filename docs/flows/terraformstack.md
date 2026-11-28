# Fluxo do TerraformStack

[Arquitetura](../architecture.md) · [GitOps](gitops.md) · [ADR-0006](../adr/0006-terraformrun-and-application-ci.md) · [Roadmap](../learning-roadmap.md)

O `TerraformStack` declara a infraestrutura desejada de Foundation e Entity Core. Na Etapa 06, o aceite verifica o fluxo de ponta a ponta no Hub `kind`. O Terraform Stack Simulator usa o provider Kubernetes para gerenciar um ConfigMap com a configuração efetiva de Network. A integração com a AWS segue nas etapas do roadmap. [Guia do Simulator](../../labs/06-terraformstack-simulator.md)

O diagrama mostra os caminhos de execução, política e drift validados no Simulator. O painel da conta AWS representa o destino das próximas fases. A seção de status descreve os campos e as condições da API instalada.

![Fluxo do TerraformStack e reconciliação de drift](terraformstack-flow.png)

O diagrama ilustra a pipeline governada de referência com autorização habilitada; não prescreve os steps de integrações alternativas.

## 1. Do Git ao Workflow

1. O Git versiona o manifesto e os YAMLs de configuração. O Argo CD sincroniza o CR na API do Hub e injeta o SHA resolvido pela Application em `platform.example.io/revision`. O estado `Synced` indica que o manifesto foi entregue à API. [Application do Simulator](../../platform/hub/addons/catalog/terraformstack-simulator.yaml)
2. O controller Go usa Controller Runtime (padrão Kubebuilder), watches de TerraformStacks e Workflows, cache de informers e fila com rate limiting. Eventos de dependências e alterações de status ou revisão GitOps reativam a reconciliação; drift e retry usam RequeueAfter. Ele valida a estrutura da intenção, verifica as dependências e grava em `status.execution` a intenção declarada e o nome do Workflow antes de criá-lo. Em seguida, cria ou acompanha o Workflow na integração selecionada por `spec.workflowRef.name` ou pelo default da instalação; na POC, o namespace é `argo-workflows`. O CR permanece como a identidade da stack e cada Workflow representa uma tentativa. [Controller](../../control-plane/terraformstack/controller/internal/controller/reconciler.go)
3. Na pipeline de referência, o Argo Workflows executa a pipeline governada. Terraform calcula o plano, OPA avalia as mudanças e o apply depende da decisão e da autorização para aquela execução. O adapter Argo traduz os resultados e a fase do Workflow para a interface interna; o reconciliador decide as transições e atualiza o CR. [WorkflowTemplate](../../workflow-plane/argo-workflows/terraformstack/workflowtemplate.yaml)

## 2. Contrato e resolução de revisões

O CRD publica o tipo `TerraformStack` com API `platform.example.io/v1alpha1`, escopo Namespaced e subrecurso `/status`. Este exemplo mostra a solicitação principal do Simulator com a fonte, o contexto e o intervalo de drift. [Schema](../../control-plane/terraformstack/crd/terraformstack.yaml) · [Manifestos completos](../../control-plane/terraformstack/examples/simulator/stacks.yaml)

```yaml
apiVersion: platform.example.io/v1alpha1
kind: TerraformStack
metadata:
  name: simulator-network
  namespace: platform-system
spec:
  sourceRef:
    repository: https://github.com/DavidFerreira21/platform-engineering-poc.git
    path: foundation/stacks/simulator
  context:
    entity: entity-a
    environment: simulator
  reconciliation:
    driftCheckInterval: 45s
    retryPolicy:
      strategy: Fixed
      interval: 10s
      maxRetries: 1
```

- `workflowRef.name` é opcional e seleciona apenas um alias cadastrado pelo Platform Team. A instalação da POC cadastra uma única integração de referência por padrão; outras podem coexistir, mas cada CR conserva no máximo uma tentativa ativa. Outras integrações podem implementar suas próprias ferramentas e etapas; a autorização operacional é habilitada por padrão e pode ser desabilitada na configuração da integração, não diretamente no CR. [Execution Contract](terraformstack-execution-contract.md)
- `sourceRef.repository` identifica a origem e `sourceRef.path` aponta para a raiz da stack. O campo opcional `context` aceita pares de strings que seguem até o executor. O perfil do Simulator autoriza uma combinação de repositório, caminho, namespace, nome da stack e contexto. Ele resolve o caminho de Configuration, PVC, backend, variáveis e perfil OPA. A ServiceAccount é definida na instalação do runtime. [Perfil do Simulator](../../workflow-plane/argo-workflows/terraformstack/simulator.json)
- A pipeline resolve `sourceRef.version` para um commit imutável. Quando a origem coincide com o repositório GitOps configurado, `latest` usa o SHA da Application. Sem essa revisão ou para outra origem previamente autorizada, `latest` resolve `HEAD`. Versões explícitas aceitam referência Git válida, incluindo SHA, branch e tag. A pipeline busca a mesma revisão para código e configuração e confere o checkout. [Pipeline](../../workflow-plane/argo-workflows/terraformstack/pipeline.sh)
- O controller considera o UID, a geração e a revisão GitOps na intenção. Um commit que muda apenas YAML pode iniciar outro plan após o sync, porque a anotação muda mesmo com o mesmo `metadata.generation`.
- A raiz Terraform consome o módulo genérico Configuration. O Simulator usa `network.yaml` com o perfil de dados `entity-a/uat` para o destino lógico `entity-a/simulator`. Os arquivos são compostos na ordem global → entidade → ambiente. A camada mais específica substitui escalares e listas, e objetos são mesclados recursivamente. [Raiz do Simulator](../../foundation/stacks/simulator/main.tf) · [Configuration](../../foundation/README.md#configuration)

## 3. Pipeline e aprovação

Os steps Terraform usam a mesma imagem de runner com ferramentas, e o `pipeline.sh` define os comandos de cada step. Scripts e regras Rego são montados a partir de ConfigMaps imutáveis com nomes derivados do conteúdo. O OPA roda em sua própria imagem. O backend local usa um PVC por stack, com state em `/state/terraform.tfstate` e arquivos de cada tentativa em `/state/<workflow>`. O locking do backend protege cada operação e é liberado enquanto o Workflow aguarda aprovação. O perfil seleciona o CLI da tentativa, com Terraform como padrão. Testes com OpenTofu usam outro state. [Binário de IaC no Simulator](../../workflow-plane/README.md#binário-de-iac-no-simulator) [Runner](../../workflow-plane/terraform-executor/Dockerfile) · [Kustomize](../../workflow-plane/argo-workflows/terraformstack/kustomization.yaml)

| Step | Comportamento |
| --- | --- |
| `resolve` | Verifica o perfil, fixa o SHA e seleciona o PVC do destino. |
| `prepare` | Confere o vínculo no PVC, publica a referência e o vínculo confirmados, faz checkout do commit imutável e monta o arquivo de variáveis. |
| `init` e `validate` | Inicializam backend e providers e validam a configuração. |
| `plan` | Gera o binário e o JSON do plano. Calcula SHA-256 do binário, do JSON e do contexto da decisão. |
| `opa` e `decision` | Avaliam as ações propostas sobre o JSON e o contexto. A pipeline valida os hashes devolvidos e publica o estado operacional, a decisão, o hash do plano e o digest do pacote de políticas. |
| `approval` | Abre um nó suspenso na UI quando a política retorna `RequireApproval`. |
| `authorize` | Emite o hash como pedido de apply quando a decisão permite. No caminho manual, exige `Approve` e o hash correto. |
| `apply` e `verify` | Revalidam e aplicam o binário salvo, executam o `verify.sh` da stack quando presente e exigem um plan pós-apply sem mudanças. |

| Decisão | Exemplo no Simulator | Efeito |
| --- | --- | --- |
| `AutoApply` | Criar o ConfigMap com `10.81.0.0/16` ou restaurar esse baseline após drift. | Segue para autorização e apply sem pausa. |
| `RequireApproval` | Alterar o CIDR desejado para `10.82.0.0/16` ou executar um destroy permitido. | Pausa para aprovação do plano. |
| `Deny` | Tentar configurar `0.0.0.0/0` ou mudar recursos fora do contrato do Simulator. | Encerra antes do apply e preserva o recurso. |
| `NoChanges` | Plan regular sem diferenças. | Encerra a verificação sem apply. |

Uma falha de avaliação do OPA ou uma decisão incompatível com o plano bloqueia o apply. No caminho manual, o aprovador autorizado abre o nó **approval**, seleciona **Resume**, escolhe `Approve` e informa o hash mostrado em `status.result.details.planHash`. O gate tem `Reject` como padrão. RBAC e admissão limitam a edição do Workflow ao gate ativo, e a auditoria do Hub registra quem retomou a execução. [Rego](../../workflow-plane/policy-engine/README.md) · [Aprovação pela UI](../../labs/06-terraformstack-simulator.md#2-exercitar-as-decisões-opa-pelo-gitops) · [Admissão local](../../control-plane/terraformstack/overlays/kind/admission.yaml)

### Autorização final do apply

1. O controller recebe o pedido de apply produzido por `authorize`. Se o pedido pertence à execução e a intenção ainda está vigente e elegível, registra o token opaco em `status.execution.authorization`.
2. O step `apply` copia o plano salvo, confere seu SHA-256 e gera um plano de revalidação. Ele compara `resource_changes`, `resource_drift` e `output_changes` com os do binário salvo. Uma divergência encerra a tentativa antes do apply.
3. A função `claim()` confere UID, geração, revisão GitOps, Workflow, intenção, estado de exclusão, suspensão e hash autorizado. Faz um patch condicional com `resourceVersion` para registrar `status.claim` com Workflow, hash da intenção e autorização. Se a intenção mudar ou a autorização ficar fora de vigência, a execução encerra sem aplicar.
4. Com o claim obtido, Terraform aplica a cópia verificada do binário original. O plano de revalidação serve para comparação. O step `verify` verifica o recurso e publica o resultado para o controller.

O `claimName` do WorkflowTemplate identifica o PVC que os Pods montam. O `claim()` do script marca o início do apply no CR. [Pipeline](../../workflow-plane/argo-workflows/terraformstack/pipeline.sh) · [Controller](../../control-plane/terraformstack/controller/internal/controller/reconciler.go)

## 4. Gatilhos, drift e concorrência

| `status.execution.trigger` | Quando o controller agenda |
| --- | --- |
| `GitOps` | Primeira intenção válida ou mudança na intenção declarada, inclusive geração, revisão GitOps ou exclusão que pede teardown. |
| `DriftCheck` | A última execução ficou `Ready` e já decorreu o `driftCheckInterval` desde `status.execution.finishedAt`. |
| `Retry` | A última execução ficou `Failed` e `retryPolicy` permite outra tentativa após o intervalo. |

Drift e retry preservam a referência imutável já resolvida. A pipeline do Simulator sinaliza mudanças externas pelo campo `resource_drift` do plan e aplica as regras OPA para decidir a remediação. O controller registra `lastCheckTime` no fim da tentativa (inclusive falha), conserva a última detecção em `lastDriftTime` e mantém `driftPending` após drift reportado até confirmação de convergência. A ausência de `driftPending` significa estado desconhecido, não ausência de drift. O acionamento manual dedicado por annotation, o jitter e a exposição da próxima checagem são evoluções planejadas. [Ensaio de drift](../../labs/06-terraformstack-simulator.md#3-observar-drift-e-reconciliação) · [Roadmap](../learning-roadmap.md#fase-6--observabilidade-operação-e-governança)

A exclusividade de execução é uma garantia estrutural: no máximo um Workflow ativo por CR, independentemente de alias, intenção ou modo de autorização. CRs independentes podem executar em paralelo; dependências são respeitadas. O lock de state continua necessário, inclusive quando CRs distintos apontam para o mesmo backend. Um sync repetido com a mesma intenção ou um reinício do controller conserva a execução ativa. Com autorização habilitada, se chega uma nova revisão enquanto o Workflow espera aprovação, o controller encerra o Workflow anterior antes de agendar o novo. Se o apply já fez o claim, ele aguarda a conclusão da execução anterior. `suspend: true` impede novas execuções e cancela a ativa quando a autorização está habilitada e ela ainda não marcou o início do apply. Sem autorização, o controller aguarda o término de qualquer Workflow ativo, inclusive diante de nova intenção; não há claim que delimite o início da alteração. Antes de submeter, o controller persiste `launchAttempted`; se a submissão ficar ambígua ou o Workflow desaparecer, bloqueia novas tentativas com `ExecutionMissing` até investigação. Alterar a configuração de uma integração ativa também bloqueia a reconciliação até restaurar o snapshot original. [Reconciliação](../../control-plane/terraformstack/controller/internal/controller/reconciler.go)

O retry operacional cria outro Workflow com outro plano e, quando exigido, outra aprovação. A política aceita `None`, `Fixed` e `ExponentialBackoff`. O runtime efetivamente faz zero retries quando `maxRetries` é omitido. Na pipeline de referência, `PolicyDenied`, `PolicyError` e `ApprovalRejected` terminam em `ReconcileBlocked` e encerram a tentativa sem retry automático. [Contrato de retry](../../control-plane/terraformstack/crd/terraformstack.yaml) · [Controller](../../control-plane/terraformstack/controller/internal/controller/reconciler.go)

## 5. Dependências e destruição

`dependsOn` referencia stacks no mesmo namespace. O controller detecta ciclos e verifica as condições das dependências sem interpretar seus contextos. Para aplicar, cada dependência precisa estar `Ready` para a geração e a intenção atuais. Para destruir, uma dependente que ainda referencia a stack precisa estar `Destroyed` para sua intenção atual. [Exemplo com duas stacks](../../control-plane/terraformstack/examples/simulator/stacks.yaml)

Na tabela, plan, política e aprovação descrevem a pipeline de referência; o controller exige o resultado tipado de destroy e o claim somente nas integrações que habilitam o handshake.

| Solicitação | Comportamento |
| --- | --- |
| `spec.lifecycle.destroy: true` | Gera `terraform plan -destroy`, passa pela política e pela aprovação humana e mantém o CR em `Destroyed / Verified` após verificar o teardown. |
| Excluir o CR com o padrão `destroyResourcesOnDeletion: false` | Retém o recurso e o PVC. O controller libera o finalizer após tratar a execução ativa. |
| Excluir o CR com `destroyResourcesOnDeletion: true` | Registra `status.deletionIntent` e mantém o finalizer durante o teardown. Libera a exclusão após o destroy aprovado e verificado. Negativa ou rejeição conserva o CR em exclusão e preserva o recurso. |

A aprovação do PR autoriza a alteração da intenção no Git. Na referência governada, a aprovação do Workflow autoriza o plano específico, inclusive no destroy. Essa política não é imposta pelo adapter a outros templates. [Exercício de teardown](../../labs/06-terraformstack-simulator.md#4-exercitar-destroy-e-limpar)

## 6. Status e diagnóstico

O controller publica a condição corrente em `status.conditions`, preservando condições anteriores com status False. A condição ativa usa True, inclusive em Failed ou ReconcileBlocked. Consulte por type e status, não pela posição na lista. Leia `type` e `reason` juntos. Consulte a fase no objeto `Workflow` indicado em `status.execution.workflow`. [Schema de status](../../control-plane/terraformstack/crd/terraformstack.yaml) · [Condições do controller](../../control-plane/terraformstack/controller/internal/controller/reconciler.go)

| `type` | `reason` | Como interpretar |
| --- | --- | --- |
| `Reconciling` | `Scheduling`, `Submitting`, `Running`, `AuthorizationRequested`, `Applying` | Tentativa registrada, em execução, ou com claim do apply registrado. Para o step exato, consulte o Workflow. |
| `ExecutionWaiting` | `Waiting` | A pipeline aguarda uma condição externa. No Simulator, consulte a decisão e o hash em `status.result.details`. |
| `DriftDetected` | `ExternalChange` | O plan identificou divergência externa. Consulte `lastDriftTime` após a transição de condição. |
| `Ready` | `NoChanges`, `Verified` | O plan saiu sem mudanças ou o apply passou na verificação. |
| `ReconcilePending` | `Suspended`, `DependencyBlocked`, `DependencyCycle`, `Cancelling`, `PreviousExecutionFinished` | Execução suspensa, dependência pendente ou transição entre intenções. |
| `ReconcileBlocked` | `InvalidIntent`, `UnknownWorkflow`, `IntegrationChanged`, `InvalidExecution`, `InvalidExecutionContract`, `ExecutionBlocked`, `Cancelled` | Validação, política, aprovação ou cancelamento interrompeu a tentativa. |
| `Failed` | `ExecutionFailed`, `ExecutionMissing` | O Workflow falhou ou desapareceu. Consulte os Pods e a última execução antes de intervir no state. |
| `Destroyed` | `Verified` | O teardown foi aplicado e verificado. |

Correlacione `status.observedGeneration` com a geração do CR e `status.resolvedSource.reference` com o commit efetivo. `status.resolvedSource.binding` identifica o destino autorizado. `status.execution` guarda a intenção, a fase, a referência do Workflow e o gatilho. O `status.result` informa o resultado e o diagnóstico do executor, como `PolicyDenied`, `PolicyError`, `ApprovalRejected` e `StateBindingMismatch`. O resultado acompanha a tentativa registrada em `status.execution`. Uma suspensão ou dependência pode atualizar a condição antes de existir uma nova tentativa. O binário, o JSON integral e o state ficam no volume. [Consultas do exercício](../../labs/06-terraformstack-simulator.md#o-que-observar)

## 7. Operação do controller

O manager usa quatro workers por padrão, Leader Election por Lease e duas réplicas, com apenas um líder ativo. A API Kubernetes conserva a intenção e o claim com resourceVersion; o controller lê diretamente a stack, o Workflow e as dependências antes de autorizar mudanças; o cache/index é usado nos watches. O controller publica logs JSON, Kubernetes Events e métricas nativas e customizadas em `:8080/metrics`, com probes em `:8081`. `internal/observability` deriva gauges do status no cache e emite contadores de transições persistidas. A integração dos sinais do Hub com CloudWatch está prevista nas Etapas 26–27, conforme o [fluxo de observabilidade](observability.md). Consulte o [catálogo de métricas e diagnóstico](../../control-plane/terraformstack/controller/README.md#observabilidade-operacional). Mudanças na configuração do controller são aplicadas por rollout. [Implementação Go](../../control-plane/terraformstack/controller/README.md)

## 8. Evolução na AWS

- **Etapas 10–12.** Reinstalar os blocos no Hub EKS, registrar UAT e validar identidade temporária do Pod, role autorizada na conta de destino e backend S3 com state segregado e locking. O destino é resolvido pelo cadastro da plataforma.
- **Etapas 13–15.** Aplicar o plano aprovado da Network UAT e validar a dependência dos EKS. O primeiro apply na AWS exige aprovação humana.
- **Etapa 17.** Publicar apenas os outputs autorizados após a verificação, pelo [GitOps Bridge](gitops-bridge.md).
- **Etapas 28–29.** Validar detecção de drift na AWS primeiro e depois remediação, auto apply sob política explícita e destroy governado.

O Workflow define limites por step no template e pode executar o `verify.sh` da raiz aprovada. O gate local pode ficar suspenso sem prazo fixo e usa revalidação antes do apply. [Roteiro de evolução](../learning-roadmap.md) · [Decisões de arquitetura](../adr/0006-terraformrun-and-application-ci.md)

## Referências de arquitetura

O contrato da POC se inspira no [Tofu Controller](https://github.com/flux-iac/tofu-controller) para reconciliação declarativa, no [HCP Terraform Operator](https://developer.hashicorp.com/validated-designs/terraform/user-guide/hcp-terraform-operator-for-kubernetes) para solicitação via API Kubernetes e no [Atlantis](https://www.runatlantis.io/guide) para gates de aprovação. A implementação do TerraformStack usa Argo CD para entregar a intenção e Argo Workflows para executar a pipeline.

## Contrato Controller / Workflow

O [Execution Contract](terraformstack-execution-contract.md) define a intenção, os estados operacionais, a autorização, o claim e os resultados. O [controller-config.yaml](../../control-plane/terraformstack/controller/controller-config.yaml) configura os nomes de transporte, a ServiceAccount e o template. O snapshot por tentativa protege a interpretação de execuções ativas. Drene os Workflows antes de mudar o contrato e revise RBAC, políticas de admissão e GitOps quando mudar a integração. [ADR-0009](../adr/0009-terraformstack-runtime.md)

# TerraformStack Controller

O controller Go reconcilia solicitações `TerraformStack` por meio de um adapter Argo. O [schema do CRD](../crd/terraformstack.yaml) define `sourceRef.repository/path/version`, a referência opcional `workflowRef.name`, o contexto e os controles de dependência, suspensão, reconciliação e lifecycle.

## Runtime

- Observa TerraformStacks e Workflows, indexa as dependências do namespace e usa fila com rate limiting. As dependências precisam estar prontas para a intenção atual. Ciclos bloqueiam a execução.
- Usa quatro workers por padrão e Leader Election por Lease com duas réplicas.
- Persiste a intenção e o nome determinístico antes de criar o Workflow. Agenda drift e retry por stack com `RequeueAfter`.
- Usa leitura direta da API para a stack, o Workflow e as dependências antes da autorização. Atualiza o status com `resourceVersion` para concorrer com o claim do executor.
- Expõe métricas em `:8080` e probes de saúde em `:8081` no Hub.

## Integração

O [controller-config.yaml](controller-config.yaml) é mantido pelo Platform Team. A lista `workflows` registra uma ou mais integrações autorizadas, cada uma com alias único, namespace, WorkflowTemplate, ServiceAccount e nomes do transporte. Com uma integração, o default é implícito; com mais de uma, `defaultWorkflow` é obrigatório. A instalação anterior com `workflow` singular também é aceita, mas não junto de `workflows`.

O CR pode selecionar uma integração por `spec.workflowRef.name`. Sem esse campo, usa o default. O alias precisa existir na configuração; não é acesso arbitrário a um WorkflowTemplate. A configuração da POC continua registrando apenas a pipeline de referência, com autorização habilitada. Para adicionar uma integração, registre um WorkflowTemplate compatível, configure sua identidade e acrescente uma entrada na lista. Não há tipos fixos de pipeline nem seleção automática por ferramenta ou módulo.

```yaml
# Exemplo de configuração de uma instalação com uma integração customizada.
controller:
  stackNamespace: platform-system
  workers: 4
  leaderElection: true
defaultWorkflow: custom
workflows:
  - name: custom
    namespace: argo-workflows
    templateRef:
      name: custom-template
    serviceAccountName: custom-executor
    contract:
      authorization: true
    parameters:
      input:
        intent: intent
      output:
        state: execution-state
        authorization: authorization-request
        result: execution-result
        resolvedSource: resolved-source
```

```yaml
# Trecho de spec do TerraformStack
workflowRef:
  name: custom
```

`contract.authorization` é opcional e seu default é `true`. Quando `false`, a pipeline compatível não solicita autorização nem claim; `parameters.output.authorization` pode ser omitido. Alterar esse campo não modifica os steps do WorkflowTemplate: a pipeline de referência atual exige o handshake e permanece com `true`. OPA, aprovação humana e scanners são independentes e pertencem à implementação do Workflow. Os outputs de resolução e resultado continuam obrigatórios para sucesso; o progresso é opcional.

A configuração é publicada como ConfigMap imutável pelo Kustomize. Cada tentativa conserva o alias em `status.execution.workflowRef` e o snapshot da integração. Drene as execuções antes de mudar o default, remover uma integração, alterar o contrato ou atualizar os templates. Uma tentativa ativa não muda de integração. Sem autorização, uma execução já submetida não é cancelada automaticamente por mudança de spec, suspensão ou dependência; o controller aguarda o término. Se o Workflow desaparecer ou a submissão ficar ambígua, o controller bloqueia novas execuções até investigação e recuperação.

A allowlist não autoriza individualmente o uso de cada alias: restrinja por RBAC, admissão e revisão GitOps quais solicitantes/recursos podem selecionar integrações com menor governança. Conceda RBAC de Workflow apenas nos namespaces cadastrados e proteja templates, ServiceAccounts e outputs contra edição não autorizada. As políticas do Simulator protegem a integração de referência; novas integrações precisam de políticas próprias. A pipeline deve validar a identidade do destino e o binding antes de qualquer alteração, inclusive ao trocar de Workflow, e garantir locking e integridade do plano conforme sua implementação.

`internal/controller` mantém lifecycle, dependências, agendamento e autorização. `internal/execution` define os tipos de intenção, progresso, autorização, resultado e resolução. `internal/execution/argo` traduz parâmetros e fases do Argo e verifica identidade, template e argumentos. A pipeline resolve fonte, destino, state, variáveis e política. Consulte o [Execution Contract](../../../docs/flows/terraformstack-execution-contract.md) para implementar um WorkflowTemplate compatível.

## Testes

```bash
go mod tidy -diff
go test -race ./...
go vet ./...
```

A CI também executa envtest com API Kubernetes e etcd. A Etapa 06 usa o Hub para validar Argo, OPA e Terraform em conjunto. [Instalação e atualização](../../../labs/05-terraformstack-runtime.md) · [Fluxo e diagnóstico](../../../docs/flows/terraformstack.md)

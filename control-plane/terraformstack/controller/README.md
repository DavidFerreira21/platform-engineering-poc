# TerraformStack Controller

O controller Go reconcilia solicitações `TerraformStack` por meio de um adapter Argo. O [schema do CRD](../crd/terraformstack.yaml) define `sourceRef.repository/path/version`, a referência opcional `workflowRef.name`, o contexto e os controles de dependência, suspensão, reconciliação e lifecycle.

## Runtime

- Observa TerraformStacks e Workflows, indexa as dependências do namespace e usa fila com rate limiting. As dependências precisam estar prontas para a intenção atual. Ciclos bloqueiam a execução.
- Usa quatro workers por padrão e Leader Election por Lease com duas réplicas.
- Persiste a intenção e o nome determinístico antes de criar o Workflow. Mantém no máximo uma tentativa ativa por CR, sem parâmetro para desabilitar essa exclusividade. CRs independentes podem avançar em paralelo. Agenda drift e retry por stack com `RequeueAfter`.
- Usa leitura direta da API para a stack, o Workflow e as dependências antes da autorização. Atualiza o status com `resourceVersion` para concorrer com o claim do executor.
- Expõe métricas em `:8080` e probes de saúde em `:8081` no Hub.

## Integração

O [controller-config.yaml](controller-config.yaml) é mantido pelo Platform Team. A lista `workflows` registra uma ou mais integrações autorizadas, cada uma com alias único, namespace, WorkflowTemplate, ServiceAccount e nomes do transporte. Com uma integração, o default pode ser implícito; com mais de uma, `defaultWorkflow` é obrigatório. A instalação anterior com `workflow` singular também é aceita, mas não junto de `workflows`.

O CR pode selecionar uma integração por `spec.workflowRef.name`. Sem esse campo, usa o default. O alias precisa existir na configuração; não é acesso arbitrário a um WorkflowTemplate. A configuração da POC continua registrando apenas a pipeline de referência, com autorização habilitada. Para adicionar uma integração, registre um WorkflowTemplate compatível, configure sua identidade e acrescente uma entrada na lista. Não há tipos fixos de pipeline nem seleção automática por ferramenta ou módulo.

```yaml
# Exemplo ilustrativo de instalação com duas integrações, não incluídas no lab.
controller:
  stackNamespace: platform-system
  workers: 4
  leaderElection: true
defaultWorkflow: terraformstack
workflows:
  - name: terraformstack
    namespace: argo-workflows
    templateRef:
      name: terraformstack
    serviceAccountName: terraformstack-executor
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
  - name: custom
    namespace: argo-workflows
    templateRef:
      name: custom-template
    serviceAccountName: custom-executor
    contract:
      authorization: false
    parameters:
      input:
        intent: intent
      output:
        state: execution-state
        result: execution-result
        resolvedSource: resolved-source
```

```yaml
# Trecho de spec do TerraformStack
workflowRef:
  name: custom
```

`contract.authorization` é opcional e seu default é `true`. Quando `false`, a pipeline compatível não solicita autorização nem claim; `parameters.output.authorization` pode ser omitido. Alterar esse campo não modifica os steps do WorkflowTemplate: a pipeline de referência atual exige o handshake e permanece com `true`. OPA, aprovação humana e scanners são independentes e pertencem à implementação do Workflow. Os outputs de resolução e resultado continuam obrigatórios para sucesso; o progresso é opcional.

A configuração é publicada como ConfigMap imutável pelo Kustomize. Cada tentativa conserva o alias em `status.execution.workflowRef` e o snapshot da integração. Drene as execuções antes de mudar o default, remover uma integração, alterar o contrato ou atualizar os templates. Uma tentativa ativa não muda de integração. Sem autorização, uma execução já submetida não é cancelada automaticamente por mudança de spec, suspensão ou dependência; o controller aguarda o término. Sem handshake, o controller registra `launchAttempted` antes de submeter; se o Workflow desaparecer ou a submissão ficar ambígua, o controller bloqueia novas execuções até investigação e recuperação. Com handshake, também bloqueia quando perde um Workflow já criado ou com claim. O lock do backend continua necessário; exclusividade por CR não protege dois CRs distintos que apontem para o mesmo state.

A allowlist não autoriza individualmente o uso de cada alias: restrinja por RBAC, admissão e revisão GitOps quais solicitantes/recursos podem selecionar integrações com menor governança. Conceda RBAC de Workflow apenas nos namespaces cadastrados e proteja templates, ServiceAccounts e outputs contra edição não autorizada. As políticas do Simulator protegem a integração de referência; novas integrações precisam de políticas próprias. A pipeline deve validar a identidade do destino e o binding antes de qualquer alteração, inclusive ao trocar de Workflow, e garantir locking e integridade do plano conforme sua implementação.

`internal/controller` mantém lifecycle, dependências, agendamento e autorização. `internal/execution` define os tipos de intenção, progresso, autorização, resultado e resolução. `internal/execution/argo` traduz parâmetros e fases do Argo e verifica identidade, template e argumentos. `internal/observability` observa transições persistidas, enriquece o status operacional e expõe telemetria sem tomar decisões de reconciliação. A pipeline resolve fonte, destino, state, variáveis e política. Consulte o [Execution Contract](../../../docs/flows/terraformstack-execution-contract.md) para implementar um WorkflowTemplate compatível.

## Testes

```bash
go mod tidy -diff
go test -race ./...
go vet ./...
```

A CI também executa envtest com API Kubernetes e etcd. A Etapa 06 usa o Hub para validar Argo, OPA e Terraform em conjunto. [Instalação e atualização](../../../labs/05-terraformstack-runtime.md) · [Fluxo e diagnóstico](../../../docs/flows/terraformstack.md)

## Troubleshooting e reporte de bugs

O controller utiliza Zap com JSON e nível `info` por padrão. Para diagnóstico temporário, altere `--zap-log-level=debug` no Deployment e restaure `info` após a investigação. Os logs incluem identificadores da stack e do Workflow quando disponíveis. Eventos Kubernetes associam a condição e a fase ao CR, sem copiar payloads. Filtre por `workflow` ou `intentHash` para correlacionar ciclos de reconciliação. No nível debug, o adapter informa a categoria e o nome do output inválido, sem registrar seu conteúdo.

```bash
kubectl logs -n platform-system -l app=terraformstack-controller --all-containers=true --prefix=true --since=30m
kubectl get terraformstack <stack> -n platform-system -o jsonpath='{.status.execution.workflow}'
kubectl get terraformstack <stack> -n platform-system -o json | jq '.status'
```

Inspecione também os steps e logs do Workflow correspondente no Argo. Antes de compartilhar diagnósticos, sanitize manifests, status e logs (credenciais, tokens, variáveis, identificadores privados e planos). Informe a revisão Git, versões de Kubernetes/Argo, passos de reprodução e resultado esperado/observado.


## Observabilidade operacional

O controller expõe logs JSON em stdout, Events na API Kubernetes e métricas no endpoint existente `:8080/metrics`. Não depende de Prometheus Server, Grafana, Loki ou coletores. O operador decide como consumir os sinais. O package `internal/observability` não interpreta ferramentas ou etapas da pipeline.

| Métrica (prefixo terraformstack_) | Semântica |
| --- | --- |
| condition, execution_phase | Condição e fase persistidas; séries atuais com valor 1, labels `condition`/`reason` ou `phase` |
| execution_active, suspended, observed_generation_current | Tentativa registrada sem `finishedAt` (não comprova Pod em execução), suspensão e igualdade de geração observada com a atual |
| drift_pending | Drift reportado ainda não confirmado convergente; ausência significa desconhecido |
| last_check_timestamp_seconds, last_success_timestamp_seconds, last_drift_timestamp_seconds | Última tentativa concluída, inclusive falha; último resultado `Ready`; última detecção de drift. Instantes conhecidos em Unix seconds |
| execution_scheduled_timestamp_seconds, retry_attempts | Agendamento da tentativa registrada e número de retry |
| active_executions, blocked_executions | Quantidades de tentativas registradas sem término e de CRs com condição `ReconcileBlocked` para a geração corrente |
| resource_health_available | 1 se a leitura do cache teve sucesso; 0 se indisponível |
| executions_total, contract_errors_total, drift_detected_total, retries_total, authorization_requests_total | Contadores locais de transições persistidas observadas pelo processo |
| execution_duration_seconds | Histograma do agendamento até término, incluindo espera |

Os gauges de recursos são derivados do cache Kubernetes no scrape; as séries individuais usam `namespace` e `stack`. Um gauge ausente não equivale a zero. `resource_health_available` sinaliza sucesso de leitura, não garante que o cache esteja perfeitamente atualizado. A geração observada não comprova prontidão nem sincronização da anotação GitOps. Use condição, fase, revisão e logs em conjunto. Em instalações com múltiplas réplicas, deduplicar os gauges por namespace/stack e dimensão (não somar réplicas); usar apenas amostras de caches disponíveis. Os contadores são locais ao processo e reiniciam com ele; não são um histórico de auditoria exatamente uma vez. As métricas não incluem tokens, payloads, hashes ou nomes individuais de Workflow como labels. Logs e Events não copiam mensagens livres do executor. O status persiste `status.execution.scheduledAt`, `status.lastSuccessTime` e `status.driftPending` sem alterar o contrato do executor. `driftPending` é ativado por drift reportado e limpo por resultado `Ready`; após `Destroyed` o campo é removido. Não confundir falta de sinal com infraestrutura sem drift. Erros internos que não persistem status devem ser acompanhados pelos logs e métricas nativas de reconciliação. Registros antigos podem não possuir esses campos até novas transições.

```bash
kubectl -n platform-system describe terraformstack simulator-network
kubectl -n platform-system port-forward deployment/terraformstack-controller 8080:8080
# Em outro terminal:
curl -fsS http://localhost:8080/metrics | grep '^terraformstack_'
```

Exemplos de consultas para um consumidor Prometheus (não são regras de alerta instaladas):

```promql
terraformstack_condition{condition="Failed"} == 1
terraformstack_drift_pending == 1
terraformstack_execution_active == 1
```

Combine as consultas com a disponibilidade do cache e deduplicação entre réplicas; para duração de execução, use o instante de agendamento somente quando a tentativa estiver ativa. `authorization_requests_total` conta autorizações observadas no status, não todos os pedidos do executor. O endpoint existente não deve ser publicado externamente sem controles de acesso e rede. A instrumentação não instala ServiceMonitor, dashboards ou regras de alerta. Ao atualizar, drenar execuções e sincronizar CRD e controller na mesma revisão; preservar o state.

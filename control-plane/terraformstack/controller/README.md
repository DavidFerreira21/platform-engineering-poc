# TerraformStack Controller

O controller Go reconcilia solicitações `TerraformStack` por meio de um adapter Argo. O [schema do CRD](../crd/terraformstack.yaml) define `sourceRef.repository/path/version`, o contexto opcional e os controles de dependência, suspensão, reconciliação e lifecycle.

## Runtime

- Observa TerraformStacks e Workflows, indexa as dependências do namespace e usa fila com rate limiting. As dependências precisam estar prontas para a intenção atual. Ciclos bloqueiam a execução.
- Usa quatro workers por padrão e Leader Election por Lease com duas réplicas.
- Persiste a intenção e o nome determinístico antes de criar o Workflow. Agenda drift e retry por stack com `RequeueAfter`.
- Usa leitura direta da API para a stack, o Workflow e as dependências antes da autorização. Atualiza o status com `resourceVersion` para concorrer com o claim do executor.
- Expõe métricas em `:8080` e probes de saúde em `:8081` no Hub.

## Integração

O [controller-config.yaml](controller-config.yaml) é mantido pelo Platform Team. Define namespace observado, workers, Leader Election, namespace de Workflows, template, ServiceAccount e nomes do parâmetro de entrada e dos quatro outputs tipados. Kustomize publica a configuração como ConfigMap imutável e atualiza o Deployment quando ela muda. Drene as execuções ativas antes de aplicar mudanças incompatíveis. A tentativa conserva um snapshot da integração.

`internal/controller` mantém lifecycle, dependências, agendamento e autorização. `internal/execution` define os tipos de intenção, progresso, autorização, resultado e resolução. `internal/execution/argo` traduz parâmetros e fases do Argo e verifica identidade, template e argumentos. A pipeline resolve fonte, destino, state, variáveis e política. Consulte o [Execution Contract](../../../docs/flows/terraformstack-execution-contract.md) para implementar um WorkflowTemplate compatível.

## Testes

```bash
go mod tidy -diff
go test -race ./...
go vet ./...
```

A CI também executa envtest com API Kubernetes e etcd. A Etapa 06 usa o Hub para validar Argo, OPA e Terraform em conjunto. [Instalação e atualização](../../../labs/05-terraformstack-runtime.md) · [Fluxo e diagnóstico](../../../docs/flows/terraformstack.md)

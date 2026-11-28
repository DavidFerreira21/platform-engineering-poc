# ADR-0009 — Runtime e contrato de execução do TerraformStack

[← ADRs](README.md) · [Fluxo](../flows/terraformstack.md) · [Execution Contract](../flows/terraformstack-execution-contract.md)

**Status** — decisão vigente, com runtime Go e pipeline de referência versionados. O aceite no Hub EKS está previsto nas Etapas 05–06.

## Contexto

O TerraformStack mantém a intenção declarativa de stacks completas. O runtime acompanha a execução e conserva autorizações e status na Kubernetes API. O Workflow governa execução, política e aprovação com identidade e state segregados.

## Decisão

- Usar Go e controller-runtime com watches, fila, cache de dependências, workers, Leader Election, métricas e probes. Drift e retry usam `RequeueAfter` por stack.
- Declarar fonte por `sourceRef.repository/path/version`, contexto como mapa opcional e `workflowRef.name` como seleção opcional de uma integração cadastrada. O núcleo valida lifecycle, dependências e elegibilidade e encaminha a intenção ao adapter.
- Tipar os sinais de progresso, drift, autorização, resultado e referência resolvida em `internal/execution`. O adapter Argo traduz parâmetros e fases e valida a identidade do Workflow. A pipeline decide como resolver Git, destino, credenciais, state, variáveis e política.
- Com `contract.authorization: true` (default), registrar `status.claim` por JSON Patch condicionado ao `resourceVersion`. O controller disputa a mesma versão ao revogar a autorização e preserva a execução quando o apply já obteve o claim.
- Permitir uma ou mais integrações no `controller-config.yaml`, com default e aliases autorizados pelo Platform Team. O CR seleciona apenas um alias cadastrado, sem impor ferramentas ou tipos de pipeline. O Platform Team pode desabilitar o handshake de autorização por integração; nesse caso, preservamos o Workflow ativo até término, sem inferir o início do apply, e bloqueamos a repetição em caso de submissão ambígua ou perda do Workflow.
- Garantir no máximo um Workflow ativo por CR, sem opção de desabilitar a exclusividade; CRs independentes podem executar em paralelo. A proteção cobre todo o lifecycle, inclusive esperas sem lock de state. O backend continua responsável pelo lock.
- Expor logs estruturados JSON, Kubernetes Events e métricas nativas/customizadas (incluindo saúde por CR) em package `internal/observability`. Derivar gauges do status persistido no cache e emitir eventos/contadores após atualização de status bem-sucedida. Não inferir ausência de drift nem interpretar steps da pipeline; coleta, armazenamento, dashboards e alertas pertencem ao operador. [Operação do controller](../../control-plane/terraformstack/controller/README.md)
- Manter a configuração de integração e o WorkflowTemplate sob controle do Platform Team. O snapshot por tentativa protege a semântica durante upgrades. Mudar namespace, template ou identidade exige revisar RBAC, admissão e GitOps. [Configuração do controller](../../control-plane/terraformstack/controller/controller-config.yaml)

## Consequências e aceite

As pipelines compatíveis podem escolher sua organização interna com o mesmo contrato. A referência atual continua sendo uma única integração com autorização habilitada. A escolha de políticas, scanners, aprovação e ferramenta pertence ao Workflow, não ao controller. A seleção de aliases por solicitante deve ser protegida por admissão e GitOps quando houver requisitos distintos de segurança. A pipeline de referência usa runner genérico, OPA em step separado, aprovação no Argo Workflows e um vínculo de destino persistido no PVC. O `apply` executa o plano binário originalmente autorizado. O contrato anterior é substituído na POC. Atualize o runtime com as execuções drenadas e preserve o state que precisar de recuperação. [Instalação e atualização](../../labs/05-terraformstack-runtime.md)

A CI verifica código, contrato, empacotamento e testes de observabilidade. O runtime consolidado passou pela CI do PR #62; isso não comprova um novo aceite de ponta a ponta no Hub após a refatoração. A Etapa 06 usa o Hub para verificar o efeito da reconciliação no ConfigMap e no status. [Guia do Simulator](../../labs/06-terraformstack-simulator.md)

# Execution Contract do TerraformStack

[Fluxo do TerraformStack](terraformstack.md) · [ADR-0009](../adr/0009-terraformstack-runtime.md)

O contrato liga o controller, o adapter Argo e o WorkflowTemplate pela Kubernetes API. O adapter envia uma solicitação JSON e normaliza os parâmetros globais de saída. O Platform Team registra uma ou mais integrações autorizadas, com nomes de transporte, ServiceAccount, template e capacidade de autorização. Uma pipeline alternativa organiza seus steps livremente, sem tipos predefinidos, e cumpre as garantias do contrato configurado.

## Solicitação

O adapter publica um parâmetro global de entrada chamado `intent` na pipeline de referência. Seu conteúdo tem esta estrutura.

```json
{
  "workflow": "tf-123-95fa9fa945dd-2",
  "intentHash": "95fa9fa945dd3c0c4f674a88710607e76d171ba334688eaec85bf63d125ce151",
  "intent": {
    "stack": "simulator-network",
    "namespace": "platform-system",
    "uid": "123",
    "generation": 2,
    "revisionAnnotation": "a8d6be1284360d87e2546a9c2bacf84757623817",
    "deleting": false,
    "sourceRef": {
      "repository": "https://github.com/DavidFerreira21/platform-engineering-poc.git",
      "path": "foundation/stacks/simulator",
      "version": "latest"
    },
    "context": {"entity": "entity-a", "environment": "simulator"},
    "operation": "apply"
  },
  "pinnedReference": "a8d6be1284360d87e2546a9c2bacf84757623817",
  "binding": "simulator-binding-01"
}
```

`workflow` e `intentHash` identificam a tentativa e sua intenção. O hash cobre a representação canônica da intenção, incluindo fonte, contexto, UID, generation, anotação de revisão e operação. Quando presente, `intent.workflowRef.name` integra o hash da intenção e identifica o alias solicitado. Quando ausente, o controller seleciona a integração padrão e fixa o alias resolvido no status da tentativa. `context` é um mapa de strings encaminhado integralmente. O controller retira o finalizer para `retain`. As demais operações encaminhadas são `apply` e `destroy`.

`pinnedReference` e `binding` são opcionais. O executor devolve a referência imutável e o vínculo efetivo do destino. O controller fixa a referência usada na tentativa anterior nos próximos retries e checagens periódicas. Para destroy, fixa a referência conhecida. O vínculo acompanha as próximas solicitações da stack. A pipeline verifica seu significado e recusa deriva de backend ou destino.

## Saídas globais

As saídas são strings JSON com `workflow` e `intentHash` iguais aos da solicitação. Na pipeline de referência, os nomes têm este mapeamento.

| Nome | Campos específicos | Semântica |
| --- | --- | --- |
| `resolved-source` | `reference`, `binding` | Identificadores imutáveis e opacos. Publicar antes de progresso, pedido de autorização ou sucesso |
| `execution-state` | `phase` (`Running` ou `Waiting`), `driftDetected` opcional, diagnósticos | Sinaliza atividade ou espera por condição externa e informa drift detectado |
| `authorization-request` | `token` | Valor opaco que identifica a alteração autorizada pela pipeline; somente quando `contract.authorization: true` |
| `execution-result` | `outcome`, diagnósticos | Resultado funcional terminal da tentativa |

Exemplo de publicação depois que o gate da pipeline libera uma alteração.

```json
{
  "workflow": "tf-123-95fa9fa945dd-2",
  "intentHash": "95fa9fa945dd3c0c4f674a88710607e76d171ba334688eaec85bf63d125ce151",
  "token": "plan-version-42"
}
```

O resultado aceita `NoChanges` sem alteração para `apply`, `Verified` após alteração verificada (com claim correspondente quando a autorização está habilitada), `Blocked` para recusa funcional sem autorização e `Cancelled` para execução que terminou sem iniciar a alteração. `Verified` aplicado a `destroy` encerra a stack em `Destroyed`. `NoChanges` e `Verified` exigem fase nativa `Succeeded`. Falha nativa sem resultado funcional vira falha operacional e pode habilitar retry. Sucesso nativo sem resultado válido bloqueia a execução. O executor pode publicar `Blocked` antes da resolução para sinalizar uma recusa na entrada.

Os diagnósticos opcionais são `reason`, `message` e `details` como mapa curto de strings. Códigos técnicos, hashes adequados para consulta e mensagens curtas são aceitos. Mantenha segredos, planos, state e configuração sensível fora dos parâmetros e do status. A lógica de autorização usa os campos tipados. O adapter recusa envelopes incompatíveis, duplicados, maiores que 4 KiB, de identidade divergente ou com resultados contraditórios.

## Estados, autorização e claim

Com `contract.authorization` ausente ou `true`, o controller registra `Pending` no agendamento e observa `Running`, `Waiting` e `AuthorizationRequested` pelo adapter. Confere identidade, intenção vigente, dependências, suspensão e cancelamento antes de gravar `status.execution.authorization`. O executor espera o token correspondente, verifica a intenção na API e registra somente o claim.

```yaml
status:
  claim:
    workflow: tf-123-95fa9fa945dd-2
    intentHash: 95fa9fa945dd3c0c4f674a88710607e76d171ba334688eaec85bf63d125ce151
    authorization: plan-version-42
```

O executor usa JSON Patch com teste do `metadata.resourceVersion`. O controller também usa `resourceVersion` na atualização de status. Se o cancelamento vencer, a autorização é revogada antes de terminar o Workflow. Se o claim vencer, o controller registra `Applying` e acompanha o término daquela alteração antes de prosseguir com outra intenção. A política de admissão do Hub limita o executor a registrar o claim autorizado e vigente.

Ao terminar, o adapter e o controller distinguem `Succeeded`, `Failed`, `Blocked` e `Cancelled`. O status do CR registra a condição da reconciliação, a resolução observada e o diagnóstico. A autorização de negócio, a verificação do plano e a política são implementadas pela pipeline confiável. A exclusividade por CR vale para toda a tentativa, com ou sem handshake, e independe do lock de state. Não é configurável. CRs independentes podem avançar em paralelo.

### Execução sem handshake de autorização

O Platform Team pode configurar `contract.authorization: false` para uma integração cuja implementação não solicita autorização nem registra claim. Não existe um tipo de pipeline associado a essa opção. Aprovação humana e política continuam independentes. Não publique pedido de autorização nesse contrato; o mapeamento do output pode ser omitido. O controller aceita `Verified` com resolução válida, fase nativa `Succeeded` e sem autorização/claim. `NoChanges`, `Blocked` e `Cancelled` preservam suas semânticas. O executor continua responsável por verificar o destino e o binding **antes** de modificar infraestrutura, pela integridade de suas operações e pelo locking de state.

Sem o claim, o controller não consegue identificar atomicamente o início do apply. Por isso não infere a fase `Applying` nem interrompe um Workflow ativo por mudança de intenção, suspensão, dependência ou erro de contrato; observa seu término antes de reconciliar outra intenção. Uma violação do contrato observada durante a execução invalida o resultado. Antes de submeter a execução, registra `status.execution.launchAttempted` com CAS. Se o Workflow não puder ser encontrado após esse registro, mantém a tentativa bloqueada com `ExecutionMissing`, inclusive se a intenção mudar; não reenvia automaticamente. A investigação deve verificar Pods, state/lock e resultado antes de qualquer recuperação administrativa. A perda de disponibilidade é intencional para evitar duas alterações concorrentes.

## Configuração

Os nomes estão em [controller-config.yaml](../../control-plane/terraformstack/controller/controller-config.yaml). A lista `workflows` aceita aliases únicos; o CR usa `spec.workflowRef.name` ou o default. Uma instalação com um único Workflow continua suficiente. O controller guarda o alias resolvido e o snapshot dessa configuração durante a tentativa. O progresso é opcional e seu mapeamento pode ser omitido; resolução e resultado são necessários para sucesso. O hash da intenção não incorpora automaticamente o snapshot da integração. Drene os Workflows ao mudar default, template, namespace, ServiceAccount, mapeamento ou capacidade de autorização. Se a integração ativa for removida ou seu snapshot divergir, restaure a configuração anterior para acompanhar a tentativa. Garanta que apenas identidades autorizadas criem o Workflow e editem seus outputs e que o executor obtenha somente os acessos requeridos. [Instalação e atualização do runtime](../../labs/05-terraformstack-runtime.md)

## Observabilidade e limites do contrato

O controller deriva sua telemetria das transições de reconciliação e do status persistido. `status.execution.scheduledAt`, `status.lastSuccessTime` e `status.driftPending` são mantidos pelo controller e não são novos outputs do executor. `driftDetected` continua opcional: sua ausência não significa que o executor comprovou ausência de drift. A pipeline mantém seus diagnósticos próprios; o controller não interpreta ferramentas, steps ou mensagens livres. Não copie parâmetros, tokens, planos ou state para logs, Events ou labels de métricas. [Operação e catálogo de métricas](../../control-plane/terraformstack/controller/README.md#observabilidade-operacional).

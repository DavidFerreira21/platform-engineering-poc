# Execution Contract do TerraformStack

[Fluxo do TerraformStack](terraformstack.md) · [ADR-0009](../adr/0009-terraformstack-runtime.md)

O contrato liga o controller, o adapter Argo e o WorkflowTemplate pela Kubernetes API. O adapter envia uma solicitação JSON e normaliza quatro parâmetros globais de saída. O Platform Team configura os nomes de transporte, a ServiceAccount e o template. Uma pipeline alternativa pode organizar seus steps livremente e precisa cumprir a mesma semântica.

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

`workflow` e `intentHash` identificam a tentativa e sua intenção. O hash cobre a representação canônica da intenção, incluindo fonte, contexto, UID, generation, anotação de revisão e operação. `context` é um mapa de strings encaminhado integralmente. O controller retira o finalizer para `retain`. As demais operações encaminhadas são `apply` e `destroy`.

`pinnedReference` e `binding` são opcionais. O executor devolve a referência imutável e o vínculo efetivo do destino. O controller fixa a referência usada na tentativa anterior nos próximos retries e checagens periódicas. Para destroy, fixa a referência conhecida. O vínculo acompanha as próximas solicitações da stack. A pipeline verifica seu significado e recusa deriva de backend ou destino.

## Saídas globais

As saídas são strings JSON com `workflow` e `intentHash` iguais aos da solicitação. Na pipeline de referência, os nomes têm este mapeamento.

| Nome | Campos específicos | Semântica |
| --- | --- | --- |
| `resolved-source` | `reference`, `binding` | Identificadores imutáveis e opacos. Publicar antes de progresso, pedido de autorização ou sucesso |
| `execution-state` | `phase` (`Running` ou `Waiting`), `driftDetected` opcional, diagnósticos | Sinaliza atividade ou espera por condição externa e informa drift detectado |
| `authorization-request` | `token` | Valor opaco que identifica a alteração autorizada pela pipeline |
| `execution-result` | `outcome`, diagnósticos | Resultado funcional terminal da tentativa |

Exemplo de publicação depois que o gate da pipeline libera uma alteração.

```json
{
  "workflow": "tf-123-95fa9fa945dd-2",
  "intentHash": "95fa9fa945dd3c0c4f674a88710607e76d171ba334688eaec85bf63d125ce151",
  "token": "plan-version-42"
}
```

O resultado aceita `NoChanges` sem alteração para `apply`, `Verified` após alteração verificada e claim correspondente, `Blocked` para recusa funcional sem autorização e `Cancelled` para execução que terminou sem iniciar a alteração. `Verified` aplicado a `destroy` encerra a stack em `Destroyed`. `NoChanges` e `Verified` exigem fase nativa `Succeeded`. Falha nativa sem resultado funcional vira falha operacional e pode habilitar retry. Sucesso nativo sem resultado válido bloqueia a execução. O executor pode publicar `Blocked` antes da resolução para sinalizar uma recusa na entrada.

Os diagnósticos opcionais são `reason`, `message` e `details` como mapa curto de strings. Códigos técnicos, hashes adequados para consulta e mensagens curtas são aceitos. Mantenha segredos, planos, state e configuração sensível fora dos parâmetros e do status. A lógica de autorização usa os campos tipados. O adapter recusa envelopes incompatíveis, duplicados, maiores que 4 KiB, de identidade divergente ou com resultados contraditórios.

## Estados, autorização e claim

O controller registra `Pending` no agendamento e observa `Running`, `Waiting` e `AuthorizationRequested` pelo adapter. Confere identidade, intenção vigente, dependências, suspensão e cancelamento antes de gravar `status.execution.authorization`. O executor espera o token correspondente, verifica a intenção na API e registra somente o claim.

```yaml
status:
  claim:
    workflow: tf-123-95fa9fa945dd-2
    intentHash: 95fa9fa945dd3c0c4f674a88710607e76d171ba334688eaec85bf63d125ce151
    authorization: plan-version-42
```

O executor usa JSON Patch com teste do `metadata.resourceVersion`. O controller também usa `resourceVersion` na atualização de status. Se o cancelamento vencer, a autorização é revogada antes de terminar o Workflow. Se o claim vencer, o controller registra `Applying` e acompanha o término daquela alteração antes de prosseguir com outra intenção. A política de admissão do Hub limita o executor a registrar o claim autorizado e vigente.

Ao terminar, o adapter e o controller distinguem `Succeeded`, `Failed`, `Blocked` e `Cancelled`. O status do CR registra a condição da reconciliação, a resolução observada e o diagnóstico. A autorização de negócio, a verificação do plano e a política são implementadas pela pipeline confiável.

## Configuração

Os nomes estão em [controller-config.yaml](../../control-plane/terraformstack/controller/controller-config.yaml). O controller guarda um snapshot dessa configuração durante a tentativa. Drene os Workflows ao mudar template, namespace, ServiceAccount ou mapeamento. Garanta que apenas identidades autorizadas criem o Workflow e editem seus outputs e que o executor obtenha somente os acessos requeridos. [Instalação e atualização do runtime](../../labs/05-terraformstack-runtime.md)

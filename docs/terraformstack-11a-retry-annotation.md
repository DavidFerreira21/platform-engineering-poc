# TerraformStack 11A — solicitação explícita de nova execução via annotation

[← ADR-0006](adr/0006-terraformrun-and-application-ci.md) · [Roteiro](learning-roadmap.md) · [Aprovação na Argo UI](terraformstack-11a-argo-client-rbac.md)

## Decisão

Uma tentativa encerrada (`Failed` ou `Blocked`, inclusive cancelamento antes do apply) **não é reiniciada automaticamente**. Uma pessoa do **Platform Team**, com permissão para alterar o TerraformStack, solicita uma nova tentativa por uma annotation no próprio recurso Kubernetes:

```yaml
metadata:
  annotations:
    platform.example.io/retry-request-id: "valor-unico-por-solicitacao"
```

Esta annotation é um **sinal operacional**, não parte de `spec` nem uma mudança na infraestrutura desejada. O Argo CD central continua sincronizando `spec` do Git; não é necessário um novo Sync para executar novamente a mesma intenção. Não confundir esta operação com a ação **Retry/Resubmit** nativa da Argo Workflows UI: o retry deve ser feito pelo **TerraformStack Controller**, criando outro Workflow e outro Terraform Plan.

## Comando do laboratório

```bash
hub() { kubectl --context kind-platform-management "$@"; }
# Usar uma Stack de teste já existente, com última tentativa terminada.
stack=network-uat
ns=platform-system  # namespace previsto para CRs do Hub; confirmar no CRD implementado
request_id=$(python3 -c 'import uuid; print(uuid.uuid4())')

hub -n "$ns" annotate terraformstack "$stack" \
  "platform.example.io/retry-request-id=$request_id" --overwrite
hub -n "$ns" get terraformstack "$stack" -o yaml
```

O namespace `platform-system` é **proposto**, não implantado. Validar grupo/plural do CRD e namespace efetivo antes de executar. O operador não aplica Terraform por esse comando: somente cria uma intenção de reexecução.

## Comportamento do controller

1. Receber evento de alteração do objeto (annotation no `metadata`); `metadata.generation` **não precisa mudar**.
2. Ler a annotation e comparar o ID com `status.lastHandledRetryRequestId` **e** com o histórico persistido de IDs consumidos (`status.handledRetryRequests` ou storage durável equivalente). IDs já vistos são ignorados mesmo após restart/relist do controller ou um retry posterior. Apenas guardar o último ID seria insuficiente para impedir replay de IDs antigos.
3. Exigir que não exista Workflow ativo ou apply em execução para aquela Stack/state; aceitar apenas uma tentativa terminal e elegível a retry. Se o mesmo `request-id` chegar enquanto há execução ativa, registrar rejeição/condição apropriada **sem colocá-lo numa fila para aplicação futura**.
4. Resolver novamente `spec`, revisão/destino e permissões; criar um novo ID de execução e Workflow. Um pedido aceito produz **novo plan** e passa novamente por `suspend` na Argo UI. O plano antigo não é reaproveitado.
5. Persistir o consumo do request id de maneira idempotente e reconciliar usando identidade/nome determinístico do Workflow ou mecanismo de reserva durável, evitando duplicações entre concorrência, crash e atualização de `status`. Somente o controller escreve `status`.
6. Guardar `status.lastHandledRetryRequestId`, o histórico de IDs tratados, `status.currentExecution.id` e referência ao Workflow; preservar `lastSuccessfulExecution`. Na POC o histórico poderá ser pequeno e persistido, mas não remover IDs antigos sem outro mecanismo de antirreplay. Falha no novo plan/execução não apaga histórico anterior.
7. Não limpar a annotation automaticamente: o próximo retry exige **outro UUID**. Um mesmo ID nunca cria duas tentativas, inclusive após GitOps resync, restart do controller ou processamento repetido de eventos.

**Sem prazo de aprovação:** o Workflow permanece em `suspend: {}` até retomada ou cancelamento. A ausência de TTL não elimina a validação de `spec`, generation, hash, destino e state antes do apply.

**Alteração de spec concorrente:** se o Git atualizar `spec` durante a aprovação, bloquear a aplicação do plano anterior. A nova generation segue o fluxo normal de execução; a annotation não autoriza saltar o gate de aprovação.

## Argo CD e a annotation

A annotation é acrescentada operacionalmente **no objeto live**, não em um commit Git. O Argo CD normalmente não exige novo Sync para alterações apenas de metadata. Como o comportamento exato depende das opções de apply, do campo gerenciado e de eventuais políticas de self-heal, **testar na Etapa 10/11A se um Sync ou self-heal remove a annotation**. Se houver conflito, configurar o Argo CD para preservar somente este path específico de metadata (por exemplo, `ignoreDifferences` junto com `RespectIgnoreDifferences`, após testar), evitando ignorar toda a seção de annotations. O controller já deve ter persistido o pedido aceito em `status`; não dependa da permanência da annotation para processar uma tentativa já registrada.

## Permissões

- **Platform Team / Retry Operator (identidade confiável):** `get` e `patch` de `terraformstacks` no namespace apropriado; apenas essa identidade solicita novo retry. No laboratório, a mesma equipe pode administrar a Stack, mas o papel de aprovação do Argo UI **não recebe** essa permissão.
- **Approver (Argo UI):** `get/list/watch/update/patch` exclusivamente em `workflows.argoproj.io` no namespace `argo-workflows`, conforme o [guia de RBAC](terraformstack-11a-argo-client-rbac.md). Isso não concede `patch` em TerraformStacks.
- **Controller:** leitura de TerraformStacks, atualização da subresource `terraformstacks/status`, e criação/observação de Workflows no namespace autorizado. O controller precisa acompanhar alterações em metadata, além de `spec`.
- **Viewer e squads:** observação quando autorizada, **sem patch de TerraformStack**.

**Limite de segurança:** o Kubernetes RBAC não restringe `patch` a apenas uma annotation. Quem puder aplicar patch no TerraformStack poderá potencialmente alterar outros campos de metadata/spec, a menos que haja validação/admission específica. Por isso o direito de retry fica restrito a operadores confiáveis; não apresentar o Role abaixo como isolamento de campo. Políticas de admission mais precisas poderão ser implementadas em uma evolução posterior.

### Exemplo de Role (para identidade confiável, não aplicar automaticamente)

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: terraformstack-retry-operator
  namespace: platform-system
rules:
  - apiGroups: ["platform.example.io"]
    resources: ["terraformstacks"]
    verbs: ["get", "patch"]
```

O grupo `platform.example.io`, plural `terraformstacks` e namespace devem ser confirmados na Etapa 07. O binding não está incluído neste exemplo, pois a identidade real do operador ainda não foi definida.

## Critérios de aceite

| Cenário | Evidência esperada |
| :--- | :--- |
| Cancelamento antes do apply | Workflow termina, Stack fica `Blocked` e **nenhuma nova tentativa** é iniciada automaticamente |
| Patch autorizado após cancelamento/falha | UUID novo cria exatamente um Workflow com novo Plan e novo gate na Argo UI |
| Mesmo UUID novamente | Não cria Workflow adicional, inclusive após restart do controller |
| Argo CD Sync sem mudança de spec | Não reinicia Workflow nem executa Terraform |
| Pedido durante Workflow ativo | Não cria Workflow concorrente, registra rejeição e não deixa pedido pendente para execução posterior |
| Operador sem `patch` | Kubernetes responde `Forbidden`; Approver da UI não consegue solicitar retry |
| Falha entre criação e atualização de status | Reconcile recupera o Workflow já existente sem duplicar execução |
| Novo workflow aprovado | Aplica **apenas** seu próprio plano após validação; plano anterior nunca é reutilizado |
| Mudança na spec enquanto aguarda | Bloqueia plano antigo, e a nova generation segue o ciclo normal |
| Self-heal/Sync do Argo CD | Annotation não gera repetição; preservar o sinal ou consumi-lo duravelmente pelo controller |

**Fora do escopo:** implementar o controller/CRD, API de retry própria, IAM AWS, política de auto-remediação ou rollback de um apply parcialmente executado. Este guia registra contrato, comandos previstos e testes para a implementação nas Etapas 07, 10 e 11A.

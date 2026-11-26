# TerraformStack — contrato de execução

> Proposta em implementação nesta branch. Não representa o runtime atual até o aceite dos testes.

## Fronteira

O CR declara sourceRef (repository, path, version), context opcional (mapa string/string), dependsOn, suspend, reconciliation e lifecycle. Não há target, configuration, timeout ou healthChecks. Não há compatibilidade com o contrato anterior nesta POC.

O controller mantém identidade, dependências, lifecycle, agendamento, retry, autorização, concorrência e conditions. Não resolve Git, conta, região, identidade cloud, backend, PVC, variáveis ou políticas. A pipeline e o IaC resolvem esses elementos. Argo Workflows permanece a integração suportada. A compatibilidade com OpenTofu será validada na pipeline, com state separado.

## Comunicação

Não há novo componente ou CRD. O adapter cria o Workflow com a intenção JSON e identidade da execução, observa eventos por labels, lê phase e outputs globais e normaliza para tipos internos. O reconciler é a autoridade sobre conditions. O runner só escreve o claim restrito.

Proposta de mapeamento no controller-config.yaml:
- input intent
- outputs execution-state, authorization-request, execution-result, resolved-source

Os nomes de transporte são configuráveis; a semântica não. A ServiceAccount e o template são definidos pelo Platform Team. Não exigir nomes de steps de OPA ou aprovação.

## Semântica

- Pending/Running: execução não terminal.
- Waiting: condição externa; reason e message para diagnóstico.
- AuthorizationRequested: identificador opaco não vazio vinculado a execução e intenção vigentes; controller concede somente se elegível e não cancelando.
- Applying: derivado do claim persistido, não de uma declaração arbitrária de output.
- Succeeded: resultado funcional válido e fase terminal compatível, distinguindo NoChanges, Verified e teardown verificado.
- Failed: falha operacional; retry conforme política.
- Blocked: bloqueio funcional, sem retry automático.
- Cancelled: cancelamento, sem sucesso implícito.

Drift é um sinal tipado, não uma string de diagnóstico. Detalhes opcionais são limitados e não contêm segredos. Resultados ausentes, inválidos ou contraditórios não autorizam apply nem Ready. Eventos podem ser repetidos ou fora de ordem; ler o estado corrente da API e validar identidade. Não interpretar reason ou message como autorização.

## Identidade, autorização e versões

Persistir intenção e nome determinístico antes de criar Workflow; manter snapshot do contrato. Vincular autorização ao UID, geração, revisão GitOps, identidade da execução e identificador opaco. Claim com resourceVersion disputa atomicamente com cancelamento. Não cancelar apply com claim obtido; aguardar término. Manter RBAC e admissão de proteção do protocolo.

A pipeline resolve e valida a revisão Git e publica referência imutável; controller conserva o valor opaco por intenção para drift/retry. Para latest, a revisão GitOps identifica a mudança da intenção, não necessariamente um commit do repositório de origem externo. Retry antes de resolução bem-sucedida não pode fingir que existe referência fixada. A pipeline mantém hashes, assinatura comparável, revalidação, plano original e locking. Contexto resolvido e identidade de state devem permanecer estáveis entre plan/apply e entre reconciliações; mudança de destino/backend exige tratamento explícito, sem migração implícita.

## Aceite

Validar CRD, Go, adapter, envtest, race, OPA, Kustomize, CI e Simulator completo (plan/apply/verify, no changes, aprovação, bloqueio, drift, retry, dependências, suspensão, restart, concorrência, destroy, resultados inválidos). Atualizar ADR-0009, arquitetura, fluxo e Labs 04–06 junto da implementação. Não provisionar AWS nesta refatoração.

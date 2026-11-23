# TerraformStack 11A — substituição automática de execução por nova geração

[← ADR-0006](adr/0006-terraformrun-and-application-ci.md) · [Roadmap](learning-roadmap.md) · [Retry manual](terraformstack-11a-retry-annotation.md)

## Decisão

Se o `spec` de um `TerraformStack` mudar **enquanto o Workflow anterior aguarda aprovação**, o controller deve **invalidar a execução anterior, solicitar seu encerramento e iniciar automaticamente outra execução para a generation nova**. O novo Workflow produz outro `terraform plan` e aguarda outra aprovação na Argo Workflows UI. **Não exige annotation de retry nem um Sync adicional do Argo CD.**

Distinguir os gatilhos:

| Evento | Comportamento |
| :--- | :--- |
| `metadata.generation` aumenta por mudança real no `spec` | Substituição automática da execução pendente; novo plan e novo gate |
| Workflow é cancelado pelo operador sem mudança de `spec` | Stack `Blocked`; novo Workflow **somente** por annotation `retry-request-id` |
| Annotation de retry muda com a mesma generation | Nova tentativa explicitamente autorizada, se execução anterior terminal |
| Argo CD Sync com `spec` idêntico | Não cria tentativa adicional |
| Metadata/labels/annotation alteradas sem mudar `spec` | Não dispara execução por nova generation; apenas retry-id novo, quando elegível, pode fazê-lo |

**Importante:** `generation` é a revisão do `spec` Kubernetes, não o número de commits ou de Sync do Git.

## Fluxo seguro

1. O reconciler percebe `metadata.generation > status.currentExecution.generation`, com tentativa anterior em `AwaitingApproval`.
2. Marca o plano antigo como **superseded** em seu registro interno e impede sua aplicação; a etapa pré-apply do Workflow valida de novo UID da Stack, generation, revisão e destino **mesmo após um Resume tardio**.
3. Solicita `terminate`/encerramento da execução antiga e espera evidência de que ela chegou a estado terminal. **Não criar Workflow novo ainda**, nem liberar apply antigo. Se o encerramento falhar, marcar bloqueio operacional e tentar recuperação controlada, sem reativar o plano antigo.
4. Somente após confirmar execução antiga terminal e ausência de apply/lock concorrente, reserva de modo idempotente a nova tentativa com identidade determinística baseada em Stack UID + generation + execução. O controller cria um novo Workflow e novo plan.
5. A nova tentativa passa por `Planning → AwaitingApproval`. A aprovação anterior nunca é reaproveitada; a conclusão antiga (ou eventual `resume` atrasado) não pode atualizar `Ready=True` da generation atual.
6. Ao recuperar-se de restart/relist, o controller reconstrói as referências dos Workflows e executa o mesmo algoritmo sem duplicar tentativas.

**Escopo desta decisão:** substituição automática somente enquanto a execução anterior aguarda aprovação, **antes do início do apply**. Se o `spec` mudar durante `Applying`, **não cancelar cegamente o Terraform** nem assumir rollback: sinalizar necessidade de serialização/inspeção e bloquear nova execução até o estado da infraestrutura ficar conhecido. Esse caso terá tratamento específico de implementação.

Se chegarem várias alterações de `spec` durante a substituição, preferir a **última generation válida**; não criar um Workflow para cada revisão intermediária que nunca chegou ao plan.

## Transições de status

Mantemos `status.conditions` como fonte de verdade e `status.phase` como resumo derivado. Na troca de generation, `Ready` deve ficar `False` para a intenção nova. O último sucesso histórico pode continuar em `lastSuccessfulExecution`, mas não torna a nova intenção pronta.

| Evento | Phase da Stack | Conditions relevantes |
| :--- | :--- | :--- |
| Generation N aguardando aprovação | `AwaitingApproval` | `Ready=False`, `Reconciling=True`, `ApprovalRequired=True` (`ApprovalPending`) |
| Spec muda para N+1, encerramento iniciado | `Pending` | `Ready=False`, `Reconciling=True` (`SupersedingExecution`), `ApprovalRequired=False`; `Stalled=False` |
| Workflow N confirmado terminal; Workflow N+1 criado | `Planning` | `Ready=False`, `Reconciling=True` (`PlanInProgress`), `ApprovalRequired=False` |
| Plan N+1 concluído com alterações | `AwaitingApproval` | `Ready=False`, `Reconciling=True`, `ApprovalRequired=True` (`ApprovalPending`) |
| N+1 aprovado e em apply | `Applying` | `Ready=False`, `Reconciling=True`, `ApprovalRequired=False` |
| Apply N+1 verificado | `Ready` | `Ready=True`, `Reconciling=False`, `ApprovalRequired=False` |
| Falha em encerrar N com segurança | `Blocked` | `Ready=False`, `Stalled=True` (`SupersessionBlocked`); `ApprovalRequired=False` |

Durante a troca, `status.observedGeneration` só avança quando a geração nova for efetivamente reconhecida/processada pelo reconciler; cada condition usa `observedGeneration` correspondente à avaliação. `currentExecution` referencia a tentativa efetivamente ativa; a anterior permanece identificável por referência/histórico não sensível com motivo `Superseded`. Evitar criar um campo paralelo de `phase` para o Workflow; seu estado nativo já existe no Argo.

`lastTransitionTime` só muda ao mudar o valor `status` da condition, seguindo o contrato já estabelecido; mudanças de `reason`/mensagem sem mudança de `status` não devem reescrever esse timestamp.

## Critérios de aceite

| Cenário de teste | Evidência esperada |
| :--- | :--- |
| N está `AwaitingApproval`; alterar spec | Workflow N é invalidado e encerrado; plano N não pode ser aplicado |
| Encerramento de N confirmado | **Um** Workflow novo para a última generation, com novo plan e gate suspend |
| Resume atrasado do Workflow N | Apply N negado por generation/plano inválido, sem escrever `Ready=True` |
| Falha de encerramento | `Blocked/SupersessionBlocked`; sem segundo Workflow concorrente e sem apply do antigo |
| Nova generation N+1 em aprovação | `Ready=False`, `ApprovalRequired=True`, `observedGeneration=N+1` |
| Múltiplas mudanças N→N+1→N+2 | Prioriza N+2, sem Workflows extras para revisões já superadas |
| Sync Argo CD sem mudança de spec | Não inicia Workflow novo |
| Cancelamento manual com spec inalterado | Não reinicia automaticamente; annotation de retry continua obrigatória |
| Reinício do controller durante encerramento | Recupera estado do Workflow e não duplica execução |
| Spec muda durante `Applying` | Não inicia outro apply concorrente; registra necessidade de inspeção e não presume rollback |

Nenhum teste deve exigir AWS real nesta fase de simulação. Os testes de controle do `TerraformStack Controller` serão implementados nas Etapas 10/11A; este documento especifica o resultado esperado, sem afirmar que já foram executados.

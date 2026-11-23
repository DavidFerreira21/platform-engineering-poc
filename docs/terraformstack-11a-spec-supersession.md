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

**Mudança de spec durante `Applying` (decisão vigente):** o Workflow já aprovado **continua executando exatamente o plano salvo e aprovado**, mesmo que a Kubernetes API receba uma generation mais nova. A mudança de intenção **não** cancela o Apply nem é classificada como drift externo. O controller marca a intenção mais recente como pendente, acompanha a execução atual até um estado terminal e não inicia um segundo Terraform simultaneamente sobre o mesmo state.

- **Apply anterior concluiu com sucesso e verificação consistente:** preservar sua evidência como `lastSuccessfulExecution` para a generation antiga; não atribuir `Ready=True` à generation mais nova. Iniciar **automaticamente** novo Workflow/Plan para a **última generation** e exigir aprovação humana na Argo UI. Sem annotation de retry e sem resync manual.
- **Apply anterior falhou ou terminou com resultado incerto:** **não iniciar automaticamente** um novo Apply. Registrar `Failed`/`Stalled=True` com motivo `PreviousApplyFailed` ou `ApplyOutcomeUnknown`, exigir diagnóstico do state/infraestrutura e recuperação explícita. Evitar sobrepor execução sobre estado parcialmente aplicado ou desconhecido.
- **Mudança de spec antes do Apply (Planning/AwaitingApproval):** continua valendo a regra anterior de encerrar o Workflow obsoleto e replanejar. O cancelamento seguro só ocorre antes da fase de Apply.
- **Múltiplas generations recebidas:** processar apenas a **última** generation ao terminar o Apply anterior; não gerar plans intermediários desnecessários.

**Termo importante:** `spec` diferente da infraestrutura atualmente aplicada é **mudança declarativa pendente** (`DesiredStateChanged`), não `DriftDetected`. Na 11B, drift será detectado para alterações externas à intenção Git/CR.

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

## Transições durante Apply em andamento

| Evento | Phase | Conditions |
| :--- | :--- | :--- |
| Generation N em apply e spec avança para N+1 | `Applying` | `Ready=False`, `Reconciling=True` (`ApplyingPreviousGeneration`); `ApprovalRequired=False` |
| Apply N bem-sucedido, N+1 ainda pendente | `Pending` → `Planning` | `Ready=False`, `Reconciling=True` (`DesiredStateChanged` → `PlanInProgress`) |
| Plan N+1 possui mudanças | `AwaitingApproval` | `ApprovalRequired=True`, `Ready=False` |
| Apply N falhou ou resultado desconhecido | `Failed` ou `Blocked` | `Stalled=True` com `PreviousApplyFailed` ou `ApplyOutcomeUnknown`; sem novo workflow automático |

`Ready=True` refere-se somente à generation desejada e verificada; não marcar a Stack pronta para N+1 com o sucesso de N. O controller guarda a generation/processo efetivamente em Apply e acompanha a generation desejada mais recente sem substituir o Workflow que já está aplicando.

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
| Spec muda durante `Applying`, execução anterior bem-sucedida | Conclui o Apply sem interromper, preserva histórico e inicia novo Plan para última generation com nova aprovação |
| Spec muda durante `Applying`, execução anterior falha/incerta | Não inicia novo Apply automaticamente; `Failed/Stalled=True` e diagnóstico manual antes de recuperar |

Nenhum teste deve exigir AWS real nesta fase de simulação. Os testes de controle do `TerraformStack Controller` serão implementados nas Etapas 10/11A; este documento especifica o resultado esperado, sem afirmar que já foram executados.

# TerraformStack — revisão consolidada dos critérios de aceite da Etapa 11A

[← Roadmap](learning-roadmap.md) · [ADR-0006](adr/0006-terraformrun-and-application-ci.md) · [Lab de aprovação](terraformstack-11a-argo-client-rbac.md)

**Estado desta revisão:** critérios definidos; **não** equivale a implementação ou execução de testes. Os manifests existentes para 11A são um smoke test com `echo` e RBAC; o controller, Terraform Executor real, armazenamento do plan e CRD ainda precisam ser desenvolvidos nas etapas correspondentes.

## Fronteiras do escopo

- **07:** instalar CRD com schema validável, subresource `/status` e impressões de phase/Ready.
- **10:** controller inicial com Workflow simulado, idempotência, transições, supersession e retry sem AWS.
- **11A:** executor Terraform em destino explicitamente autorizado, state seguro e segregado, plan salvo, gate humano na Argo UI, verificação pré-apply e evidências.
- **11B:** deteção periódica de drift externo sem remediação.
- **11C:** aplicação automática condicionada a políticas; **fora da 11A**.

## Gates de aceite por cenário

| ID | Cenário | Evidência exigida | Onde validar |
| :--- | :--- | :--- | :--- |
| AC-01 | CRD inválido / target não autorizado | API/admission ou controller rejeita com condição legível; nenhuma chamada AWS | 07/10 |
| AC-02 | Primeira intenção GitOps | Argo CD sincroniza uma Stack e controller cria **um** Workflow da generation correspondente | 10 |
| AC-03 | Plan com mudanças | Artefato binário privado, digest e referência de execução; resumo legível para revisão, sem dados sensíveis | 11A |
| AC-04 | Plan sem mudanças | `Ready=True` para generation atual sem pedir aprovação nem executar apply | 11A |
| AC-05 | Gate humano | `suspend: {}` sem TTL: Viewer não retoma; Approver retoma via Client Auth/RBAC; sem resume não há apply | Smoke + 11A |
| AC-06 | Apply fiel ao plan | Execução aplica **o plano salvo** (não gera novo plan silencioso); verifica resultado/estado e atualiza status | 11A |
| AC-07 | Alteração maliciosa/obsoleta | Hash, target, revisão, generation ou state inválidos **antes** do apply bloqueiam a execução | 11A |
| AC-08 | Spec muda durante aprovação | Cancela Workflow suspenso e só cria novo após término confirmado; novo plan e novo gate | 10/11A |
| AC-09 | Spec muda durante Apply | Apply corrente termina; após sucesso confirmado gera novo plan para última generation e nova aprovação; sem concorrência | 10/11A |
| AC-10 | Resultado do Apply falha/incerto | Não lança outro apply automaticamente; `Failed/Blocked`, `Stalled=True`, diagnóstico obrigatório | 11A |
| AC-11 | Cancelamento manual pré-apply | Stack permanece declarada, sem novo Workflow; `Blocked/ExecutionCancelled` | 10 |
| AC-12 | Retry explícito | Annotation `platform.example.io/retry-request-id` nova e autorizada cria um novo Workflow; nunca reutiliza plan/aprovação | 10/11A |
| AC-13 | Deduplicação após restart/replay | Mesmo ID antigo ou evento repetido não inicia execução; reinício entre criação de Workflow/status não cria duplicatas | 10 |
| AC-14 | Concorrência de state | Duas tentativas na mesma Stack/state não aplicam ao mesmo tempo; lock é validado | 10/11A |
| AC-15 | Condições/phase | `Ready`, `Reconciling`, `ApprovalRequired`, `Stalled` e `observedGeneration` são coerentes; phase derivada | 07/10/11A |
| AC-16 | Argo CD resync | Sincronização idêntica ou reconciliação de metadata não recria execução já consumida | 10 |
| AC-17 | Segurança | Aprover não tem IAM, Secrets, state ou binário plan; tokens não vão ao Git; execução usa identidade independente | Smoke + 11A |
| AC-18 | Exclusão/Destroy | Remover CR não dispara `terraform destroy`; operação destrutiva não recebe autorização implícita | 10/11A |

## Máquina de estados (resumo)

| Gatilho | Phase | Condições de referência |
| :--- | :--- | :--- |
| Intenção ainda não processada | `Pending` | `Ready=Unknown/False` |
| Plan em curso | `Planning` | `Ready=False`, `Reconciling=True` |
| Plan com mudanças esperando Resume | `AwaitingApproval` | `ApprovalRequired=True`, `Ready=False` |
| Apply iniciado | `Applying` | `Reconciling=True`, `Ready=False` |
| Plan sem mudanças / apply verificado, generation atual | `Ready` | `Ready=True`, `Reconciling=False` |
| Cancelamento/reprovação/target proibido | `Blocked` | `Stalled=True`, `Ready=False` |
| Falha de plan/apply | `Failed` | `Stalled=True`, `Ready=False` |

Se o `spec` mudar durante o Apply, não atribuir `Ready=True` à nova generation quando a antiga concluir. Uma alteração declarativa **não é drift externo**. Apenas na 11B `DriftDetected` poderá indicar divergências externas verificadas.

## Dependências e limites conscientes

1. **Aprovação é simples:** `update/patch` do Workflow no RBAC **não equivale à autorização exclusiva para resume**, nem garante assinatura de plano por usuário; é uma escolha didática da POC. Revisar antes de usar com workloads não confiáveis, PROD ou PCI.
2. **Sem prazo de aprovação:** `suspend: {}` sem `duration`; tokens de login podem expirar sem cancelar o Workflow. O plano pode tornar-se obsoleto por alteração de state/infraestrutura, verificada antes do apply.
3. **Durabilidade e antirreplay:** guardar somente `lastHandledRetryRequestId` não impede reuso de IDs anteriores; é necessário persistir histórico ou identidade imutável por execução. Garantir recuperação após crash entre criação do Workflow e atualização de status.
4. **Armazenamento e visualização do plan:** definir artefato privado, retenção, permissões, hash e como o operador vê um resumo seguro na UI. O Workflow atual só imprime um `echo`, então não prova AC-03/06/07.
5. **Argo Server:** atualmente `server.enabled: false`; habilitar via GitOps em 11A, exclusivamente `client`, testar Viewer/Approver no `kind`. Não criar exposição pública.
6. **Identidade de executor:** não confundir credencial do usuário da UI com a ServiceAccount/role de execução; credenciais AWS temporárias e por destino.
7. **Falha durante apply:** sem rollback presumido; bloquear nova execução até inspeção da situação real e do state.
8. **Observabilidade:** Workflows, condições e logs devem ser correlacionados por Stack UID, generation e execution ID, sem guardar tokens/state em status.

## Recomendação de implementação

Seguir **07 → 10 → 11A**, usando os IDs AC-01 a AC-18 como casos de teste automatizados e evidência dos labs. Priorizar primeiro um controller mínimo com mock executor que demonstre status, retry e supersession; depois integrar plan/state e o Apply com aprovação. Não implementar drift nem Policy Engine de 11C prematuramente.

**Pendências de implementação, não novas decisões de produto:** grupo/namespace definitivos do CRD, caminho para persistência das chaves de execução, modelo de artifact storage do plan, RBAC efetivo na versão instalada do Argo e visualização segura do resumo na UI.

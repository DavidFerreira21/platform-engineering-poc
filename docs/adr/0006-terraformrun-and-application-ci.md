# ADR-0006 — TerraformStack declarativo, drift e CI/CD de Domain Applications

**Estado:** decisão de arquitetura-alvo; implementação pendente · **Data original:** 21/11/2026 · **Revisão:** 22/11/2026

## Contexto e objetivo

A POC usa a **Kubernetes API como contrato técnico** e quatro camadas de ownership: **Foundation → Entity Core Infrastructure → Domain Infrastructure → Domain Applications**.

- **Foundation / Entity Core:** Terraform/OpenTofu para rede, IAM, dependências, e os três EKS por Environment (Management, PCI e Non-PCI).
- **Domain Infrastructure:** Crossplane v2 central com XRDs/Compositions/Providers específicos para recursos AWS como ObjectStorage e Database.
- **Domain Applications:** Application Team mantém **código, GitHub Actions e manifests Kubernetes no mesmo repositório de aplicação de teste separado**, a integrar na Etapa 17 do roteiro; Argo CD **local no Management EKS** lê esses manifests GitOps e entrega em EKS PCI/Non-PCI autorizados. O DevOps Federado mantém somente os manifests de Domain Infrastructure.

**Intenção desta decisão:** oferecer, para módulos Terraform, **experiência declarativa semelhante à do Crossplane**: manifestos de intenção versionados no Git, provisionamento inicial condicionado a políticas, detecção recorrente de drift, e remediação automática ou com aprovação conforme riscos. Não se trata apenas de disparar uma execução única de Terraform.

## Decisão 1 — TerraformStack representa a infraestrutura desejada

O recurso de consumo passa a ter o nome de trabalho **`TerraformStack`**, implementado futuramente como **CRD Kubernetes próprio**, observado por um **TerraformStack Controller** no Control Plane do Hub. **Não é XRD Crossplane**.

- O `spec` referencia **módulo aprovado e versão/revisão fixada**, entradas autorizadas e o destino lógico `Entity + Environment`; o controller resolve, em cadastro confiável, credenciais AWS, backend/state e permissões.
- O recurso mantém o ciclo de vida da **infraestrutura desejada**; não representa apenas um `terraform apply`.
- O `status` expõe condições, última verificação, revisão observada, resumo do último plan, classificação, referência de workflow, estado de aprovação e divergência detectada. Não guardar segredos nem todo o plan em `status`.
- `TerraformRun` fica **reservado como nome opcional para uma execução imutável** (ou pode ser substituído por um Argo Workflow criado pelo controller). **Não criar um segundo CRD só para representar runs antes de existir justificativa.**
- **Entity/Environment/Domain continuam registros lógicos Git**, não CRDs criados por esta decisão.

### Exemplo ilustrativo de contrato (NÃO implementado)

```yaml
apiVersion: platform.example.io/v1alpha1
kind: TerraformStack
metadata:
  name: network-uat
spec:
  module:
    name: aws-network
    version: "1.0.0"
  target:
    entity: entity-a
    environment: uat
  inputs:
    networkProfile: standard
  reconciliation:
    interval: 30m
    policyRef: foundation-default
```

Este YAML é **intenção declarativa, não manifesto funcional**. Schema, nomes, catálogo de módulos, versionamento e parâmetros finais serão validados nas etapas do roteiro.

## Decisão 2 — GitOps aplica a intenção; Argo Workflows executa Terraform

```text
Backstage / CLI / pessoa → PR + validações + aprovação Git
  → Argo CD CENTRAL → Kubernetes API / TerraformStack (objeto Kubernetes)
  → TerraformStack Controller (validação + agendamento)
  → Argo Workflows → Terraform Executor → terraform plan
  → 11A: aprovação humana do plano imutável → terraform apply do mesmo plano
  → 11B: verificação periódica e relato de drift, sem apply
  → 11C: decisão versionada auto apply / aprovação / bloqueio
  → AWS Foundation / Entity Core → verificação → status do TerraformStack
```

**Separação de responsabilidades:**
- **Argo CD central:** sincroniza `TerraformStack` do Git para Kubernetes, **não executa Terraform nem cria workflows por conta própria**.
- **TerraformStack Controller:** observa CRs, valida módulo/destino/permissões, dispara ou agenda verificações, controla concorrência/retries, invoca o motor de políticas e atualiza status.
- **Argo Workflows:** orquestra plan, avaliação das políticas, gates/aprovação, apply e verificações.
- **Terraform:** calcula mudanças e executa a infraestrutura na AWS.

**Argo CD `Synced` significa Git == CR Kubernetes, não AWS sem drift.** Drift de infraestrutura é informado em `TerraformStack.status.conditions` (por exemplo `DriftDetected`, `Ready`, `ApprovalRequired`, `ReconcileBlocked`) e pode ser exposto por health check customizado, CLI ou observabilidade. A POC **não depende** de interface/health customizados do Argo CD.

## Decisão 3 — três etapas incrementais da POC

### Etapa 11A — Deploy declarativo ponta a ponta com aprovação

- Aplicar `TerraformStack` pelo GitOps central (ou diretamente na API para teste local), e ver o controller iniciar Workflow/plan.
- **Toda execução de apply na 11A exige aprovação humana explícita do plano**, mesmo para criação em UAT. Auto apply só será avaliado na 11C, com política versionada.
- Validar schema, autorização, idempotência por generation/revisão, erros/retry, status e evidências. Começar com **executor simulado na Etapa 10**; na 11A, validar o fluxo autorizado com state segregado, lock, plano imutável e aprovação vinculada a hash, revisão Git e destino.

### Decisão 11A — aprovação por RBAC na Argo Workflows UI

**Escolha para o laboratório:** autorização simples por **Argo Server + RBAC**, com `suspend: {}` entre plan e apply. Usuários autorizados retomam o Workflow pela UI; a retomada autorizada é o **gate de aprovação operacional da 11A**. Não haverá serviço próprio de aprovação, assinatura individual do plano, nem validação adicional da identidade do aprovador depois do `resume`. O `apply` permanece bloqueado até a retomada manual. A aprovação de PR não substitui a retomada na UI.

**Autenticação decidida:** Argo Server em modo **`client` exclusivamente**, usando Kubernetes bearer token e RBAC namespace-scoped. Viewer pode consultar; Approver tem `get/list/watch/update/patch` em Workflows no namespace `argo-workflows`. O chart instalado na Etapa 03 deixa `server.enabled: false`; a UI será habilitada por mudança revisada na Application GitOps somente na 11A. Exemplo de RBAC, Workflow simulado, comandos e testes em [guia de Client Auth e RBAC](../terraformstack-11a-argo-client-rbac.md). Não armazenar tokens no Git.

**Fronteira de confiança:** esta decisão é proporcional à POC em ambiente autorizado. O RBAC controla quem pode modificar/retomar o Workflow, mas não garante aprovação criptograficamente vinculada a uma pessoa e a um hash específico. Não afirmar compliance/auditoria individual de aprovação; reavaliar para ambientes produtivos ou PCI. Permissões de `update/patch` em Workflows podem conceder mais capacidades do que apenas `resume`, portanto o acesso precisa ficar limitado a um namespace de laboratório e a identidades confiáveis. **Quem puder modificar o template, plano ou artefatos de execução não deve ser tratado como mero aprovador.**

**Implementação prevista (não implantada por este ADR):**

1. O `TerraformStack Controller` cria o Workflow somente a partir de Stack/destino/módulo autorizados e associa Stack UID, `metadata.generation`, revisão Git e identificador único de execução. O executor usa IAM mínimo e state segregado com lock; o aprovador **não recebe permissões AWS nem acesso ao state**.
2. O executor executa `terraform plan -out` e salva o arquivo binário em storage restrito, com identificador/digest e retenção segura, sem expiração por tempo de aprovação na 11A. A UI exibe um resumo de mudanças **sanitizado** e a referência do plano; ela não expõe state, segredo ou conteúdo sensível do plano.
3. Um passo `suspend: {}` (sem `duration`) aguarda a retomada humana na **Argo Workflows UI**. O Argo Server deve exigir autenticação e RBAC por namespace, sem acesso administrativo irrestrito nem modo de autorização compartilhado. Um papel observador só lê; um papel aprovador confiável pode executar a operação necessária para retomar Workflows naquele namespace. O verbo efetivo (`update`/`patch` e eventual subresource) será confirmado em teste na versão instalada.
4. Após o `resume`, o executor **não revalida a identidade individual de quem clicou**; confia no gate de RBAC da UI. Contudo, antes do apply verifica correspondência entre execução e Stack, `generation`, revisão, módulo/versão, destino, hash do plano e compatibilidade com state/backend, e lock das operações Terraform. Qualquer divergência bloqueia o apply e exige novo plan e nova aprovação via UI.
5. O executor aplica o **arquivo binário planejado** (não um plan regenerado), valida o resultado e informa o controller, que atualiza `status`. Retomadas duplicadas, retries, cancelamentos e falhas não podem disparar apply indevido; `destroy` não é implícito.

**Evidência mínima na 11A:** referência ao Workflow, à Stack/generation, ao digest do plano, ao instante de retomada observado e ao resultado da execução. O identificador do aprovador, se disponível em logs de acesso, é diagnóstico opcional e **não é garantia de auditoria individual**. Um `resume` por usuário autorizado é suficiente para o gate operacional desta POC; RBAC não deve ser confundido com aprovação forte de produção.

**Critérios de aceite da decisão:**

- Um observador consegue consultar Workflow e resumo de plano mas não retomar; a tentativa retorna negação de autorização.
- Um aprovador com RBAC permitido retoma o Workflow na UI, e o processo segue à verificação pré-apply e ao apply do arquivo planejado; ninguém precisa editar o CR manualmente.
- Sem retomada manual, não há apply; não configurar retomada automática por prazo. Cancelar/encerrar a execução não deve liberar apply.
- Alteração de hash, destino, módulo, `generation`, revisão ou state/lock depois do plan bloqueia o apply, mesmo se o Workflow for retomado.
- O papel de aprovação não inclui acesso a IAM, secrets, state ou plano binário; testar tentativas de escalonamento/modificação e registrar limitações do RBAC efetivo.
- `status.conditions`, `phase`, logs do Workflow e referência do plano permanecem consistentes, sem conter material sensível.

**Decisão sobre prazo e cancelamento (11A):** não haverá TTL de aprovação nem `duration` no `suspend`; o Workflow aguarda indefinidamente até ação humana ou encerramento operacional. Isso **não dispensa** a validação pré-apply de generation, revisão, hash, destino, backend/state e lock, nem impede que um plano fique obsoleto devido a mudanças externas. Não manter lock exclusivo de state durante toda a espera humana; adquirir o lock nas operações Terraform e revalidar o estado ao retomar.

**Cancelamento e nova tentativa:** encerrar um Workflow aguardando aprovação não remove o `TerraformStack`, não cria outro Workflow automaticamente e não exige Sync do Argo CD. O controller observa a execução encerrada, registra `phase: Blocked`, `Stalled=True` com `reason: ExecutionCancelled`, `Ready=False`, e preserva `lastSuccessfulExecution`. Para reexecutar a **mesma intenção**, um operador autorizado precisa sinalizar um **novo pedido de execução** ao controller (API/annotation/CLI ainda a definir); ele cria um **novo Workflow com novo plan e novo gate**. Mudança de `spec` via Git inicia outra generation como caminho normal de mudança de infraestrutura. Um Sync isolado da mesma definição não deve disparar execução adicional. Cancelamento enquanto `apply` já está em andamento exige tratamento próprio: não presumir rollback ou ausência de alterações parciais.

**Evolução futura:** aprovação forte vinculada a identidade e digest por serviço/contrato dedicado fica como opção fora da 11A; não é pré-requisito de 11B/11C.

### Etapa 11B — Drift Detection

- Com intervalo configurado, controller agenda `terraform plan` (ou acionamento manual controlado) para verificar divergência entre configuração/state e AWS.
- Diferenciar **mudança desejada via Git** de **drift externo** e registrar a informação no status/observabilidade; checar os resultados estruturados do plan, incluindo ações e diferenças observadas no refresh.
- Nesta etapa **não executar remediação de drift**: detectar, registrar e notificar, sem apply automático. Concluir com evidências de que um recurso modificado externamente é detectado.

### Etapa 11C — Policy-Based Reconciliation

- Um **Policy Engine versionado/testável** classifica o plan segundo identidade, Environment, origem (PR vs drift), recurso, atributos, custo/impacto e operações (`create/update/delete/replace`). A tecnologia específica (por exemplo OPA) ainda está **a decidir**.
- Três resultados possíveis: **Auto Apply** se expressamente permitido; **Require Approval** se sensível; **Deny/Block** se não autorizado.
- Exemplos *ilustrativos*, não permissões default: remediação de tags permitidas; adição de recurso de baixo risco em UAT; aprovação para modificação de rota/rede crítica; bloqueio de exposição insegura ou destino não autorizado.
- O **controller coordena o ciclo de vida**, e o **Workflow executa a decisão**. Não espalhar a política como regras ad hoc difíceis de auditar dentro dos Workflows.

## Decisão 4 — contrato de status (Phase + Conditions)

**Escolha da POC:** `status.conditions` é a fonte de verdade; `status.phase` é um resumo **derivado** delas e do ciclo de execução. Os dois são escritos pelo mesmo reconciler na subresource `/status`, nunca por usuários, Argo CD ou Workflow diretamente. `status.observedGeneration` indica a revisão de `spec` já processada. Quando `metadata.generation` for maior, o status anterior é **stale** e não pode aprovar/apply a nova intenção.

### Campos do status v1alpha1

| Campo | Obrigação | Semântica |
| :--- | :--- | :--- |
| `observedGeneration` | Sim após primeiro reconcile | Última generation avaliada pelo controller. Não significa deploy bem-sucedido. |
| `phase` | Sim após primeiro reconcile | Resumo operacional: `Pending`, `Planning`, `AwaitingApproval`, `Applying`, `Ready`, `Failed` ou `Blocked`. Não é um controle de segurança. |
| `conditions[]` | Sim | Condições padrão Kubernetes com `type/status/reason/message/observedGeneration/lastTransitionTime`. Cada tipo aparece no máximo uma vez. |
| `currentExecution` | Durante execução | `id`, `workflowRef`, `generation`, `startedAt`, `planRef` (referência opaca), `planDigest`, `approval` e, após conclusão, `finishedAt`. |
| `lastSuccessfulExecution` | Opcional | Resumo da última execução aplicada com sucesso (`id`, `generation`, `finishedAt`), mantido mesmo quando uma nova execução falha. |
| `lastPlan` | Após plan | `id`, `generation`, `hasChanges`, `planDigest` e referência restrita ao artefato. Sem conteúdo do plano nem credenciais. |

A aprovação em `currentExecution.approval` usa `Pending/Approved/Rejected` na 11A (`Expired` fica reservado para possível política futura), com `decidedAt` (quando observável) e referência não sensível ao Workflow; nunca contém token ou credenciais. Na 11A, `Approved` significa **gate suspend retomado por sessão autorizada pelo RBAC do Argo**, não comprovação individual de quem aprovou nem assinatura do plano. O executor ainda valida hash, generation, destino, revisão e lock antes do apply.

### Conditions v1alpha1

| Tipo | True | False | Unknown |
| :--- | :--- | :--- | :--- |
| `Ready` | Última intenção observada foi aplicada/verificada (ou plan válido sem mudanças, sem execução pendente) | Não pronta, falha ou aprovação pendente | Estado ainda não apurado |
| `Reconciling` | Processando plan, esperando aprovação ou aplicando | Sem execução ativa | Estado não verificável |
| `ApprovalRequired` | Plan com mudanças aguardando aprovação humana | Não há aprovação pendente | Aprovação ainda não apurada |
| `Stalled` | Erro ou bloqueio que impede avanço sem intervenção/novo evento | Sem bloqueio terminal | Indeterminado |

**Reservado para 11B:** `DriftDetected` (`True/False/Unknown`) acompanhado de `lastDriftCheckAt`. Em 11A **não** declarar `DriftDetected=False`, porque nenhum monitoramento periódico ainda aconteceu. A ausência da condição significa funcionalidade não habilitada; `Unknown` significa monitoramento habilitado mas resultado inconclusivo.

`reason` usa identificadores de máquina em CamelCase (ex.: `PlanInProgress`, `ApprovalPending`, `ApplyInProgress`, `ApplySucceeded`, `PlanFailed`, `ApprovalRejected`, `PlanExpired`, `TargetForbidden`). `message` é para humanos, não lógica. `lastTransitionTime` só muda se o valor `status` da condição mudar; `observedGeneration` acompanha a intenção avaliada.

### Mapeamento determinístico da phase

1. Destino inválido, execução cancelada, aprovação rejeitada ou conflito irreconciliável → `Blocked` com `Stalled=True`.
2. Falha de plan/apply/workflow que exige intervenção → `Failed` com `Ready=False`, `Stalled=True`.
3. Apply em progresso → `Applying`; plan em progresso → `Planning`.
4. Plano aguardando autorização → `AwaitingApproval` com `ApprovalRequired=True` e `Reconciling=True`.
5. Resultado verificado para a generation atual, sem execução ativa → `Ready` com `Ready=True` e `Reconciling=False`.
6. Sem avaliação concluída → `Pending`, `Ready=Unknown` ou `False`, conforme evidência existente.

`Ready=True` **não garante ausência de drift externo**; até 11B, apenas atesta o resultado da última verificação da intenção. O status antigo não pode ser interpretado como atual se a generation divergir. O controller não inicia um novo apply apenas porque foi reconciliado.

### Exemplo ilustrativo (não implementado)

```yaml
status:
  observedGeneration: 3
  phase: AwaitingApproval
  currentExecution:
    id: network-uat-g3-01
    workflowRef: terraform-network-uat-g3-01
    generation: 3
    startedAt: "2026-11-22T20:00:00Z"
    planDigest: "sha256:<digest-ilustrativo>"
    planRef: "artifact://restrito/network-uat-g3-01"
    approval:
      state: Pending
  lastPlan:
    id: network-uat-g3-01
    generation: 3
    hasChanges: true
    planDigest: "sha256:<digest-ilustrativo>"
  conditions:
    - type: Ready
      status: "False"
      reason: ApprovalPending
      observedGeneration: 3
      lastTransitionTime: "2026-11-22T20:00:00Z"
      message: "Plano aguardando aprovação"
    - type: Reconciling
      status: "True"
      reason: ApprovalPending
      observedGeneration: 3
      lastTransitionTime: "2026-11-22T20:00:00Z"
    - type: ApprovalRequired
      status: "True"
      reason: ApprovalPending
      observedGeneration: 3
      lastTransitionTime: "2026-11-22T20:00:00Z"
```

### Critérios de aceite de status da Etapa 11A

1. `kubectl get/describe TerraformStack` revela fase, generation e condições coerentes, e o CRD define colunas de impressão úteis (`PHASE`, `READY`, `AGE`).
2. Testes percorrem `Pending → Planning → AwaitingApproval → Applying → Ready` e verificam `observedGeneration`, razões e timestamps. Um plan sem mudanças vai a `Ready` sem apply.
3. Testes de autorização negada, aprovação rejeitada, falha de plan/apply e cancelamento expõem `Blocked`/`Failed` corretamente, preservando o último sucesso sem dar `Ready=True` à intenção não concluída.
4. Mudar `spec` durante aprovação invalida a aprovação anterior; geração divergente impede apply e força novo plan/aprovação.
5. Duas reconciliações concorrentes não iniciam dois Workflows nem dois applies sobre o mesmo state; retry é idempotente e o lock é observado.
6. `status` não expõe plano binário, state, segredos, credenciais ou detalhes sensíveis. Artefatos completos ficam em storage privado e trilha de auditoria autorizada.
7. Alterações em conditions não causam laço de reconciliação nem mudanças artificiais em `lastTransitionTime`; `phase` é calculada pelas condições/eventos e não pode contradizê-las.
8. Observador não consegue retomar o Workflow; usuário autorizado pelo RBAC consegue. A verificação da identidade individual depois de `resume` **não faz parte da 11A**. Testar exposição acidental de permissões amplas e impedir apply quando houver mudança no plano ou destino.

**Implementação incremental:** Etapa 07 estabelece schema e condições básicas; Etapa 10 exercita transições simuladas; Etapa 11A comprova fluxo, aprovação e erro em destino autorizado. Etapas 11B/11C acrescentam detecção de drift e remediação sem redefinir o contrato básico.

## Garantias obrigatórias

1. **Sem apply em todo reconcile:** reconcile Kubernetes é idempotente; novas execuções só ocorrem por solicitação nova/alterada, verificação agendada ou ação explícita. Evitar runs duplicadas e execução simultânea sobre o mesmo state.
2. **Plan e gate de aprovação:** gerar plano imutável em armazenamento restrito, relacionando-o a hash, revisão, destino e Workflow. Na 11A o `resume` autorizado por RBAC libera o gate operacional, sem assinatura por pessoa/plano; antes do apply, validar o vínculo técnico e aplicar **exatamente o plano planejado**. Se a intenção ou o estado relevante mudar, replanejar e exigir nova retomada; não há TTL de aprovação na 11A.
3. **Isolamento:** módulos/versões permitidos, targets por Entity/Environment resolvidos em mapping confiável, IAM mínimo; usuários não selecionam account/role/backend arbitrários. Validar UAT/PROD e ambientes críticos.
4. **Drift com contexto:** não autorizar só porque a ação é `create` ou `update`; analisar a semântica de atributos e recursos. Não tratar qualquer `refresh-only` como correção automática da infraestrutura.
5. **Estado e remoção:** state fora do Git e segregado; política de deletion/retain explícita; remover CR do Git **não implica automaticamente `destroy`**. Operações destrutivas exigem avaliação e gate apropriado.
6. **Evidências:** manter referências do plan, Workflow, retomada observada e verificações, com correlação observável. Na 11C acrescentar policy decision; na 11A não atribuir identidade de aprovador sem fonte confiável. Sem credenciais, state ou plan sensível em CR/status.

## Relação com os demais componentes

O código, a pipeline CI/CD **e os manifests Kubernetes da aplicação de exemplo** **não ficarão neste monorepo**; o repositório da aplicação de demonstração será definido e integrado em etapa posterior. GitHub Actions realizará build, tests, scans e publicação de imagem; a entrega será por alteração declarativa GitOps e **Argo CD local**, e não por deploy imperativo padrão da pipeline ao EKS. O workflow `.github/workflows/validate.yml` atual da POC continua exclusivo para validação de documentação e guias. As três personas e o modelo de Golden Paths são definidos no [ADR-0007](0007-personas-and-developer-experience.md).

**Domain Applications** permanecem independentes: GitHub Actions de uma aplicação de teste rodará em **outro repositório**; após build/test/scan/publicação, uma atualização de manifesto GitOps será revisada/aprovada e aplicada pelo **Argo CD local**. Nenhuma pipeline de aplicação é criada neste monorepo agora.

**Observabilidade** continua com OpenTelemetry, Grafana, Loki e Tempo; coletor/gateway local em cada Management EKS, Operators a definir.

## Faseamento e pendências

- **Etapa 07:** CRD `TerraformStack` e validação do contrato na API Kubernetes.
- **Etapa 09:** identidades de workload e destinos autorizados.
- **Etapa 10:** controller, Argo Workflows e execução simulada.
- **Etapa 11A:** Terraform Executor, state segregado, lock e deploy ponta a ponta com aprovação humana do plano imutável.
- **Etapa 11B:** plan recorrente e detecção de drift com status/alerta, sem remediação.
- **Etapa 11C:** reconciliação com Policy Engine versionado, decisões de auto apply, aprovação ou bloqueio.
- **A validar em implementação:** requeue/scheduling, verbos efetivos de RBAC para `resume` na versão do Argo instalada, limites de alteração de Workflow, engine de políticas da 11C, catálogo e semântica de cancelamento/lock. O contrato inicial de status está definido acima.

**Nenhum controller/CRD, remediação, workflow de infraestrutura real ou recurso AWS é implantado por este ADR.** A documentação não autoriza `apply` ou `destroy`.

[Arquitetura](../architecture.md) · [ADRs](README.md) · [Roadmap](../learning-roadmap.md)

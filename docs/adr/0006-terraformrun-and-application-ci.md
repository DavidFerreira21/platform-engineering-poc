# ADR-0006 — TerraformStack declarativo, drift e CI/CD de Domain Applications

**Estado:** decisão de arquitetura-alvo; implementação pendente · **Data original:** 21/11/2026 · **Revisão:** 21/11/2026

## Contexto e objetivo

A POC usa a **Kubernetes API como contrato técnico** e quatro camadas de ownership: **Foundation → Entity Core Infrastructure → Domain Infrastructure → Domain Applications**.

- **Foundation / Entity Core:** Terraform/OpenTofu para rede, IAM, dependências, e os três EKS por Environment (Management, PCI e Non-PCI).
- **Domain Infrastructure:** Crossplane v2 central com XRDs/Compositions/Providers específicos para recursos AWS como ObjectStorage e Database.
- **Domain Applications:** Application Team mantém **código, GitHub Actions e manifests Kubernetes no mesmo repositório de aplicação de teste separado**, a integrar na Etapa 12 do roteiro; Argo CD **local no Management EKS** lê esses manifests GitOps e entrega em EKS PCI/Non-PCI autorizados. O DevOps Federado mantém somente os manifests de Domain Infrastructure.

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
  → Etapa 09: aprovação humana do plano imutável → terraform apply do mesmo plano
  → Etapa 18: verificação periódica e relato de drift, sem apply
  → Etapa 19: decisão versionada auto apply / aprovação / bloqueio
  → AWS Foundation / Entity Core → verificação → status do TerraformStack
```

**Separação de responsabilidades:**
- **Argo CD central:** sincroniza `TerraformStack` do Git para Kubernetes, **não executa Terraform nem cria workflows por conta própria**.
- **TerraformStack Controller:** observa CRs, valida módulo/destino/permissões, dispara ou agenda verificações, controla concorrência/retries, invoca o motor de políticas e atualiza status.
- **Argo Workflows:** orquestra plan, avaliação das políticas, gates/aprovação, apply e verificações.
- **Terraform:** calcula mudanças e executa a infraestrutura na AWS.

**Argo CD `Synced` significa Git == CR Kubernetes, não AWS sem drift.** Drift de infraestrutura é informado em `TerraformStack.status.conditions` (por exemplo `DriftDetected`, `Ready`, `ApprovalRequired`, `ReconcileBlocked`) e pode ser exposto por health check customizado, CLI ou observabilidade. A POC **não depende** de interface/health customizados do Argo CD.

## Decisão 3 — três etapas incrementais da POC

### Etapa 09 — Deploy declarativo ponta a ponta com aprovação

- Aplicar `TerraformStack` pelo GitOps central (ou diretamente na API para teste local), e ver o controller iniciar Workflow/plan.
- **Toda execução de apply exige aprovação humana explícita do plano** até a evolução prevista na Etapa 19. Auto apply só será avaliado com política versionada.
- Validar schema, autorização, idempotência por generation/revisão, erros/retry, status e evidências. A Etapa 06 exercita o executor simulado. A Etapa 09 aplica a rede UAT com state segregado, lock, plano imutável e aprovação vinculada a hash, revisão Git e destino. O mesmo fluxo de aprovação atende às stacks de EKS.

### Aprovação humana na Argo Workflows UI

**Canal de aprovação escolhido:** Argo Workflows UI, com um passo `suspend: {}` entre `terraform plan` e `terraform apply`. O operador autorizado visualiza a referência do plano e retoma o Workflow pela UI. Nesta fase **não existe auto apply**; a aprovação de PR/Git não substitui a aprovação do plano de execução.

**Implementação prevista (sem código ou infraestrutura criados por este ADR):**

1. O `TerraformStack Controller` dispara um Workflow com identificação imutável de Stack, `metadata.generation`, revisão Git, módulo/versão e destino resolvido de uma lista permitida; o Workflow adquire lock de state.
2. O executor gera `terraform plan -out` e persiste o plano binário **em armazenamento privado e com prazo de validade**, associado a hash SHA-256, state/serial e identidade do destino. Expõe apenas metadados não sensíveis para consulta.
3. O passo de aprovação da UI usa um template `suspend: {}`, **sem `duration` que retome automaticamente**. Acesso ao Argo Server deve usar autenticação individual e autorização restrita a aprovadores, com trilha auditável de quem retomou o Workflow. Um simples `resume` não é autorização suficiente sem essa identidade verificável.
4. Após o `resume`, um passo independente valida: identidade/autorização do aprovador, vínculo da decisão ao Workflow e ao hash do plano, revisão e `generation` ainda vigentes, validade do artefato, lock/state e destino. Qualquer falha, expiração ou plano alterado **bloqueia o apply**, exige novo plan e nova aprovação.
5. O executor aplica **o arquivo de plano aprovado**, não executa um novo `terraform plan` implícito, verifica o resultado e atualiza `status`. A confirmação de sucesso depende de verificação da execução; `Argo CD Synced` não comprova o resultado na AWS.
6. Reprovação, cancelamento, timeout e erro terminam sem apply; retry é controlado e nunca reaproveita aprovação para plano diferente. Operações `destroy` ou deleção de Stack não são autorizadas implicitamente.

**Critérios de aceite específicos da decisão:** (a) aprovação de pessoa autorizada libera exatamente o plano exibido; (b) usuário sem permissão de retomar Workflow não aprova; (c) `resume` sem evidência verificável de aprovação não executa apply; (d) hash/revisão/destino/generation alterados, expiração ou perda de lock impedem apply; (e) auditoria registra Workflow, plano, aprovador, instante e resultado; (f) todos os caminhos de falha permanecem sem apply indevido.

**Pendente para próxima decisão:** como representar os estados de execução e aprovação em `TerraformStack.status` (phase e conditions, ou apenas conditions) e quais referências não sensíveis expor.

### Etapa 18 — Drift Detection

- Com intervalo configurado, controller agenda `terraform plan` (ou acionamento manual controlado) para verificar divergência entre configuração/state e AWS.
- Diferenciar **mudança desejada via Git** de **drift externo** e registrar a informação no status/observabilidade; checar os resultados estruturados do plan, incluindo ações e diferenças observadas no refresh.
- Detectar, registrar e notificar drift sem remediação. A evidência é a detecção de uma mudança externa numa stack UAT já aplicada.

### Etapa 19 — Policy-Based Reconciliation

- Um **Policy Engine versionado/testável** classifica o plan segundo identidade, Environment, origem (PR vs drift), recurso, atributos, custo/impacto e operações (`create/update/delete/replace`). A tecnologia específica (por exemplo OPA) ainda está **a decidir**.
- Três resultados possíveis: **Auto Apply** se expressamente permitido; **Require Approval** se sensível; **Deny/Block** se não autorizado.
- Exemplos *ilustrativos*, não permissões default: remediação de tags permitidas; adição de recurso de baixo risco em UAT; aprovação para modificação de rota/rede crítica; bloqueio de exposição insegura ou destino não autorizado.
- O **controller coordena o ciclo de vida**, e o **Workflow executa a decisão**. Não espalhar a política como regras ad hoc difíceis de auditar dentro dos Workflows.

## Garantias obrigatórias

1. **Sem apply em todo reconcile:** reconcile Kubernetes é idempotente; novas execuções só ocorrem por solicitação nova/alterada, verificação agendada ou ação explícita. Evitar runs duplicadas e execução simultânea sobre o mesmo state.
2. **Plan e aprovação vinculados:** gerar plano versionado/imutável em armazenamento restrito, associar aprovação a hash, revisão, destino e execução; aplicar **exatamente o plano aprovado**. Se ele expirar, mudar estado/inputs ou perder validade, replanejar e reavaliar políticas antes de aplicar.
3. **Isolamento:** módulos/versões permitidos, targets por Entity/Environment resolvidos em mapping confiável, IAM mínimo; usuários não selecionam account/role/backend arbitrários. Validar UAT/PROD e ambientes críticos.
4. **Drift com contexto:** não autorizar só porque a ação é `create` ou `update`; analisar a semântica de atributos e recursos. Não tratar qualquer `refresh-only` como correção automática da infraestrutura.
5. **Estado e remoção:** state fora do Git e segregado; política de deletion/retain explícita; remover CR do Git **não implica automaticamente `destroy`**. Operações destrutivas exigem avaliação e gate apropriado.
6. **Evidências:** manter referência de plan, policy decision, aprovação, workflow, checks e correlação observável. Sem credenciais, conteúdo sensível de state ou plan em CR/status.

## Relação com os demais componentes

O código, a pipeline CI/CD **e os manifests Kubernetes da aplicação de exemplo** **não ficarão neste monorepo**; o repositório da aplicação de demonstração será definido e integrado em etapa posterior. GitHub Actions realizará build, tests, scans e publicação de imagem; a entrega será por alteração declarativa GitOps e **Argo CD local**, e não por deploy imperativo padrão da pipeline ao EKS. O workflow `.github/workflows/validate.yml` atual da POC continua exclusivo para validação de documentação e guias. As três personas e o modelo de Golden Paths são definidos no [ADR-0007](0007-personas-and-developer-experience.md).

**Domain Applications** permanecem independentes: GitHub Actions de uma aplicação de teste rodará em **outro repositório**; após build/test/scan/publicação, uma atualização de manifesto GitOps será revisada/aprovada e aplicada pelo **Argo CD local**. Nenhuma pipeline de aplicação é criada neste monorepo agora.

**Observabilidade** continua com OpenTelemetry, Grafana, Loki e Tempo; coletor/gateway local em cada Management EKS, Operators a definir.

## Faseamento e pendências

- **Etapa 05.** CRD `TerraformStack` entregue pelo catálogo de add-ons do Hub. A API Kubernetes valida exemplos com dados fictícios.
- **Etapa 06.** Argo CD sincroniza o controller, o `WorkflowTemplate` e a solicitação `TerraformStack`. O controller cria cada Workflow de execução simulada.
- **Etapas 07 e 08.** Registro de UAT, identidade temporária do executor e inicialização autorizada do backend de estado.
- **Etapa 09.** Rede UAT aplicada com state segregado, lock e aprovação humana do plano imutável.
- **Etapas 18 e 19.** Detecção de drift com status e alerta. Depois, reconciliação com Policy Engine versionado.
- **A validar:** formato final da API, status/conditions, requeue/scheduling, política de alterações de spec durante workflows, engine de políticas, versionamento de módulo, catálogo, lock de state e aprovação segura.

**Nenhum controller/CRD, remediação, workflow de infraestrutura real ou recurso AWS é implantado por este ADR.** A documentação não autoriza `apply` ou `destroy`.

[Arquitetura](../architecture.md) · [ADRs](README.md) · [Roadmap](../learning-roadmap.md)

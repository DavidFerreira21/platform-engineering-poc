# ADR-0006 — TerraformStack declarativo, drift e CI/CD de Domain Applications

**Estado:** decisão vigente com contrato inicial versionado · **Data original:** 21/11/2026 · **Revisão:** 21/11/2026

## Contexto e objetivo

A POC usa a **Kubernetes API como contrato técnico** e quatro camadas de ownership: **Foundation → Entity Core Infrastructure → Domain Infrastructure → Domain Applications**.

- **Foundation / Entity Core:** Terraform/OpenTofu para rede, IAM, dependências, e os três EKS por Environment (Management, PCI e Non-PCI).
- **Domain Infrastructure:** Crossplane v2 central com XRDs/Compositions/Providers específicos para recursos AWS como ObjectStorage e Database.
- **Domain Applications:** Application Team mantém **código, GitHub Actions e manifests Kubernetes no mesmo repositório de aplicação de teste separado**, a integrar na Fase 5 (Etapas 18–21) do roteiro; Argo CD **local no Management EKS** lê esses manifests GitOps e entrega em EKS PCI/Non-PCI autorizados. O DevOps Federado mantém somente os manifests de Domain Infrastructure.

**Intenção desta decisão:** oferecer, para módulos Terraform, **experiência declarativa semelhante à do Crossplane**: manifestos de intenção versionados no Git, provisionamento inicial condicionado a políticas, detecção recorrente de drift, e remediação automática ou com aprovação conforme riscos. Não se trata apenas de disparar uma execução única de Terraform.

## Decisão 1 — TerraformStack representa a infraestrutura desejada

O recurso **`TerraformStack`** tem um CRD próprio versionado na Etapa 05. Na Etapa 06, o **TerraformStack Controller** passa a observar esse recurso no Control Plane do Hub.

- O `spec` referencia **stack aprovada por `sourceRef` e sua versão**, o destino lógico `Entity + Environment` e a versão dos YAMLs de configuração. O controller resolve credenciais AWS, backend, state e permissões a partir do cadastro autorizado.
- O recurso mantém o ciclo de vida da **infraestrutura desejada**; não representa apenas um `terraform apply`.
- O `status` expõe condições, última verificação, revisão observada, resumo do último plan, classificação, referência de workflow, estado de aprovação e divergência detectada. Não guardar segredos nem todo o plan em `status`.
- `TerraformRun` fica **reservado como nome opcional para uma execução imutável** (ou pode ser substituído por um Argo Workflow criado pelo controller). **Não criar um segundo CRD só para representar runs antes de existir justificativa.**
- **Entity/Environment/Domain continuam registros lógicos Git**, não CRDs criados por esta decisão.

### Contrato inicial

A Etapa 05 publica `platform.example.io/v1alpha1` com escopo de namespace. O schema exige `spec.target.entity/environment` e `spec.sourceRef.name/path`. `sourceRef.version` e `configuration.version` são opcionais, com default `latest`. `spec.module` não é aceito. O subrecurso `/status` oferece `conditions` para o controller. O exemplo usa referências fictícias e pode ser validado com dry-run na API do Hub. Resolução de fontes Git/Registry, destinos e políticas de reconciliação evoluem nas etapas que passam a usá-las. [Schema](../../control-plane/terraformstack/crd/terraformstack.yaml) · [Exemplo](../../control-plane/terraformstack/demo/terraformstack.yaml) · [Guia da Etapa 05](../../labs/05-terraformstack-crd.md)

### Configuração por entidade e ambiente

O monorepo centraliza os valores dos módulos Terraform em `foundation/config/`. Cada módulo usa um arquivo próprio, como `network.yaml` ou `eks.yaml`. A composição vai do nível global para a entidade e depois para o ambiente. Esta é a estrutura prevista, usando `br` como entidade ilustrativa.

```text
foundation/config/
├── network.yaml
├── eks.yaml
└── br/
    ├── network.yaml
    └── uat/
        ├── network.yaml
        └── eks.yaml
```

O nível mais específico prevalece. Os mapas combinam chaves nos níveis definidos pelo schema. Se `br/uat/network.yaml` sobrescrever apenas `vpc.cidr`, os outros campos de `vpc` continuam herdados. Uma lista declarada no nível mais específico substitui a lista anterior inteira. Uma chave ou um arquivo ausente mantém os valores já herdados.

Exemplo de sobrescrita para `foundation/config/br/uat/network.yaml`.

```yaml
vpc:
  cidr: 10.42.0.0/16
```

Um módulo auxiliar Terraform chamado `configuration` lê os YAMLs com `yamldecode(file(...))` e compõe os mapas com `merge` explícito nos níveis definidos pelo schema. Ele será criado na Etapa 08, com o primeiro consumo real de `network.yaml` na Etapa 09. O módulo raiz passa a configuração resolvida ao módulo de infraestrutura por entradas tipadas, por exemplo `vpc_cidr = module.configuration.network.vpc.cidr`. O módulo de rede declara seus tipos, validações e valores padrão aplicáveis em `variables.tf`. [Leitura de YAML no Terraform](https://developer.hashicorp.com/terraform/language/functions/yamldecode) · [Mesclagem de mapas](https://developer.hashicorp.com/terraform/language/functions/merge) · [Variáveis de entrada](https://developer.hashicorp.com/terraform/language/values/variables)

**Stacks iniciais da POC:** `network` e `eks` são **stacks completas de infraestrutura**, cada uma composta pelos módulos Terraform internos e recursos necessários ao seu domínio, não apenas módulos isolados. `network` prepara o baseline da rede spoke (Etapa 09); `eks` prepara clusters, IAM e dependências de addons (Etapa 10, reutilizada nas 11 e 26). O módulo auxiliar **`configuration`** (Etapa 08) não provisiona recursos e é consumido por **todas as stacks**, lendo `network.yaml`, `eks.yaml` ou um arquivo correspondente de uma stack futura. O mesmo `TerraformStack` poderá gerenciar stacks adicionais, como `bootstrap`, sem novo CRD. Destino, região, role e backend continuam no cadastro confiável de Environment, não no módulo Configuration.

**Baseline e feature flags:** cada stack implementa um baseline abrangente de componentes previstos, com defaults definidos e validados. O Configuration resolve overrides hierárquicos para ativar ou desativar opções; certas dependências obrigatórias podem permanecer **forçadas pelo contrato/política**, sem override permitido. A stack EKS permite combinações independentes de `management`, `pci` e `nonPci`; o **perfil inicial autorizado de UAT** habilita Management e Non-PCI progressivamente, enquanto PCI exige autorização específica. Defaults não são licença para provisionar automaticamente recursos caros/PCI: cada execução respeita plan, custo, policy e aprovação. Flags de pré-requisitos de addons criam somente IAM/recursos AWS necessários; operadores Kubernetes pertencem ao Platform GitOps. Validar Karpenter contra EKS Auto Mode para não duplicar controladores.

**Outputs:** as stacks `network` e `eks` expõem saídas tipadas; a publicação de um subconjunto permitido para o Parameter Store da conta do Environment ocorre pelo **GitOps Bridge (Etapa 13)**, depois de apply e verificação. Não publicar state, plano ou segredos como output distribuído. A implementação deve verificar combinações de flags e tornar explícitas as dependências de rede/IAM, sem criar um quarto módulo obrigatório agora.

Na Etapa 09, o monorepo usa `main` como origem de `latest`. Os manifestos do lab omitem `configuration.version` e `sourceRef.version`. O merge de configuração seleciona as stacks consumidoras conforme arquivo e escopo de herança. A automação proposta registra no GitOps o SHA em uma anotação por stack. O controller deverá observar essa anotação além de `generation`, pois uma mudança apenas em metadados preserva a geração. Cada Workflow disparado pelo merge fixa o SHA sinalizado para carregar código e YAMLs. Na execução inicial sem sinal de merge, resolve `main` uma vez antes do plan. O apply usa o binário aprovado. Releases por componente e promoção explícita de versões ficam para a evolução do lab. O `environment.yaml` registra o destino AWS, e a plataforma resolve as credenciais, as roles e o backend autorizados para esse destino.

Na Etapa 09, os testes do módulo auxiliar cobrem os três níveis de herança, a preservação dos campos vizinhos, a substituição de listas e a rejeição de valores incompatíveis com o schema. A validação do fluxo confirma que um commit que altera apenas o YAML muda os valores do próximo plan com a mesma imagem e o mesmo conteúdo de código Terraform.

## Decisão 2 — GitOps aplica a intenção; Argo Workflows executa Terraform

```text
Backstage / CLI / pessoa → PR + validações + aprovação Git
  → Argo CD CENTRAL → Kubernetes API / TerraformStack (objeto Kubernetes)
  → TerraformStack Controller (validação + agendamento)
  → Argo Workflows → Terraform Executor → terraform plan
  → Etapa 09: aprovação humana do plano imutável → terraform apply do mesmo plano
  → Etapa 24: verificação periódica e relato de drift, sem apply
  → Etapa 25: decisão versionada auto apply / aprovação / bloqueio
  → AWS Foundation / Entity Core → verificação → status do TerraformStack
```

**Separação de responsabilidades:**
- **Argo CD central:** sincroniza `TerraformStack` do Git para Kubernetes, **não executa Terraform nem cria workflows por conta própria**.
- **TerraformStack Controller:** observa CRs, valida módulo/destino/permissões, dispara ou agenda verificações, controla concorrência/retries, acompanha o resultado da política avaliada no Workflow e atualiza status.
- **Argo Workflows:** orquestra plan, avaliação das políticas, gates/aprovação, apply e verificações.
- **Terraform:** calcula mudanças e executa a infraestrutura na AWS.

**Argo CD `Synced` significa Git == CR Kubernetes, não AWS sem drift.** Drift de infraestrutura é informado em `TerraformStack.status.conditions` (por exemplo `DriftDetected`, `Ready`, `ApprovalRequired`, `ReconcileBlocked`) e pode ser exposto por health check customizado, CLI ou observabilidade. A POC **não depende** de interface/health customizados do Argo CD.

### Executor Argo, identidade do destino e isolamento

**Decisão:** manter o `TerraformStack Controller` próprio com **Argo Workflows** como executor/orquestrador; não substituir pelo Tofu Controller nem criar um segundo Runner Controller. O controller observa o CR, resolve o destino autorizado e cria um Workflow governado. O Argo cria **Pods efêmeros** para executar a imagem Terraform Executor nos passos de `plan`, policy, revalidação, `apply` e `verify`.

Identidades **distintas e complementares**:

- **Kubernetes:** o controller tem RBAC apenas para suas operações de controle; os Pods do Workflow usam uma `serviceAccountName` de executor reservada e protegida contra uso por Workflows arbitrários.
- **AWS de origem:** quando o Hub migrar para EKS na conta de plataforma, uma identidade de Pod (EKS Pod Identity ou IRSA, conforme avaliação) poderá habilitar credenciais AWS temporárias. **O Hub `kind` atual não fornece EKS Pod Identity**; a Etapa 08 deve validar outra forma segura de bootstrap de credenciais temporárias, sem chaves estáticas no Git/CR/imagens.
- **AWS de destino:** `spec.target.entity/environment` é uma chave lógica para consultar o registro confiável. O executor obtém uma sessão STS na role **permitida** da conta/região daquele Environment; manifesto não aceita role arbitrária. Conferir `GetCallerIdentity`, backend e trust policy. Não é o Kubernetes ServiceAccount, isoladamente, que escolhe a conta AWS.

O mesmo plano imutável é transportado em artefato privado entre Pods; `plan` e `apply` podem ocorrer em Pods/sessões AWS distintos. **A aprovação não mantém Pod, credenciais nem lock de state abertos indefinidamente.** No retorno da pausa, uma nova sessão obtém o artefato, revalida identidade de aprovador, hash, geração, revisão, destino, state e validade antes de aplicar exatamente o plan aprovado. Bloquear se a checagem falhar. Bloqueios IAM, RBAC e admission impedem que um Workflow criado fora do controller contorne os gates.

### Dependências e checagem periódica configurável

**Contrato implementado no schema**, inspirado em práticas de operadores Terraform/OpenTofu: `spec.dependsOn` representa stacks dependentes e `spec.reconciliation.driftCheckInterval` expressa a frequência desejada de verificação. Os campos são aceitos pela API desde a Etapa 05; a execução ainda **não está implementada** no reconciliador de laboratório.

```yaml
spec:
  dependsOn:
    - name: network-uat
  reconciliation:
    driftCheckInterval: 6h
```

- `dependsOn`: antes de executar uma stack, conferir dependência `Ready` com `observedGeneration` atual. Resolver por nome estável, no mesmo namespace e Entity/Environment por padrão; rejeitar referências proibidas, dependências inexistentes e ciclos. Exemplo: EKS Management depende da stack de rede UAT. Dependência pronta **não autoriza** apply sem política e aprovação.
- `driftCheckInterval`: o controller agenda um `plan` de verificação idempotente, com `lastCheckAt` / `nextCheckAt`, jitter, deduplicação e proteção contra execuções concorrentes no mesmo state. **Intervalo é para detectar, não para corrigir automaticamente**; Etapa 24 detecta e relata, Etapa 25 aplica as políticas vigentes.
- Reconcile por evento Git, retry explícito por annotation e checagem agendada são gatilhos diferentes. Não tratar toda notificação Kubernetes como ordem de executar Terraform.

**Fases da implementação:** Etapa 06 exercita a criação idempotente de Workflow e a atualização do status com **execução simulada**; Etapa 08 implementa a identidade AWS; Etapa 09 prova a execução real com aprovação; Etapas 10–11 usam `dependsOn` em rede → EKS; Etapa 24 ativa intervalo de drift.

[Detalhamento do fluxo e da identidade](../flows/terraformstack.md#43-executor-por-pods-do-argo-workflows-e-identidade-aws).

### Contrato do CRD: sourceRef e lifecycle de destruição

**Contrato atualizado:** manter `TerraformStack` como API genérica de stacks completas e aceitar `spec.sourceRef.name/path/version` como referência à **fonte versionada e aprovada**. `sourceRef.version` opcional (`latest` por padrão) resolve Registry ou Git a uma versão/commit imutável no executor real; `configuration.version` também pode ser omitido para resolver `latest`, com pin antes do plan. O source é resolvido no catálogo sob controle do Platform Team e não aceita URL arbitrária/credenciais. A revisão de código da stack e a revisão de `configuration` são independentes e ficam fixadas em cada execução. O CRD v1alpha1 exige exclusivamente `sourceRef`; o controller da Etapa 06 usa esse contrato em **simulação**, mas a resolução remota e a execução real seguem pendentes. Não dependemos do `GitRepository` de Flux.

**Dois controles de destruição já presentes no CRD, com execução local na Etapa 06B e validação na AWS na Etapa 25; ambos sujeitos ao Policy Engine e à aprovação humana:**

- `spec.lifecycle.destroy: true`: solicita um `plan -destroy` e um `apply` **do mesmo plano aprovado**, mantendo o CR para acompanhar o resultado. Não reaproveitar aprovação de apply anterior.
- `spec.lifecycle.destroyResourcesOnDeletion: true`: habilita a **intenção** de teardown no momento da exclusão do CR. O default é `false`; excluir YAML/CR não constitui aprovação. O controller usa finalizer, snapshot imutável de source/config/destino/state e Workflow com aprovação de identidade verificável. Quando não autorizado, bloqueia sem destruir. Uma solução separada e auditada para `orphan/retain` deverá ser estudada para falhas e CRs presos em deletion.

**Garantias:** aplicar proteção das dependências reversas (não destruir rede enquanto EKS depender dela), state/lock, IAM, planejamento imutável, política de risco para delete/replace e registros de auditoria. `finalizer` protege o ciclo de exclusão, **não** substitui aprovação. Destruição explícita e destruição on-delete ficam previstas para a **Etapa 25**, nunca habilitadas silenciosamente nas etapas AWS iniciais.

Outros campos já validados pelo schema: `spec.suspend`, `spec.reconciliation.retryPolicy`, `spec.timeout` por etapa e `spec.healthChecks` opcionais. **Schema não implica implementação operacional**: além de `suspend` no lab simulado, retries, timeouts e health checks serão implementados e testados ao longo da execução real. `spec.reconciliation.driftCheckInterval` (Etapa 24) e `spec.dependsOn` (Etapas 10–11) já constam no schema, mas não estão implementados no reconciliador do lab. `serviceAccountName`, AWS role, backend e desbloqueio forçado **não** serão selecionáveis livremente pelo usuário via CR.

[Exemplo conceitual e fluxo de destroy](../flows/terraformstack.md#32-evolução-do-contrato-sourceref-destroy-e-controles-operacionais).

## Decisão 3 — três etapas incrementais da POC

A Etapa 06B valida primeiro o ciclo completo no `kind` com Terraform real, OPA, aprovação, destroy, drift periódico e status, usando apenas recursos locais. As Etapas 09, 24 e 25 retomam esses mecanismos no contexto AWS.

### Etapa 09 — Deploy declarativo ponta a ponta com aprovação

- Aplicar `TerraformStack` pelo GitOps central (ou diretamente na API para teste local), e ver o controller iniciar Workflow/plan.
- **Toda execução de apply na AWS exige aprovação humana explícita do plano** até a evolução prevista na Etapa 25. Auto apply só será avaliado com política versionada.
- Validar schema, autorização, idempotência por generation/revisão, erros/retry, status e evidências. A Etapa 06 exercita o executor simulado. A Etapa 09 aplica a rede UAT com state segregado, lock, plano imutável e aprovação vinculada a hash, revisão Git e destino. O mesmo fluxo de aprovação atende às stacks de EKS.

O passo de política do Workflow executa OPA com regras Rego versionadas sobre `terraform show -json` do plano salvo. No aceite, um plano permitido chega à pausa de aprovação humana e um plano com abertura de SSH para `0.0.0.0/0` termina bloqueado antes do apply. Exercitar o caso bloqueado somente até a decisão. Erro de avaliação ou decisão ausente também bloqueia. Registrar motivo, revisão da política e hash do plano, mantendo o JSON em armazenamento restrito. Após aprovação, o executor aplica o mesmo plano binário.

### Aprovação humana na Argo Workflows UI

**Canal de aprovação escolhido:** Argo Workflows UI, com um passo `suspend: {}` entre `terraform plan` e `terraform apply`. O operador autorizado visualiza a referência do plano e retoma o Workflow pela UI. Nesta fase **não existe auto apply**; a aprovação de PR/Git não substitui a aprovação do plano de execução.

**Implementação prevista (sem código ou infraestrutura criados por este ADR):**

1. O `TerraformStack Controller` dispara um Workflow com identificação imutável da stack, `metadata.generation`, revisão Git da solicitação, versão da configuração, versão resolvida da stack e destino autorizado. O backend protege as operações Terraform com lock quando necessário, **sem manter o lock retido durante uma aprovação pendente**.
2. O executor gera `terraform plan -out` com a configuração da revisão selecionada. Persiste o plano binário **em armazenamento privado e com prazo de validade**, associado ao hash SHA-256, ao state/serial, à revisão de configuração e à identidade do destino. Expõe apenas metadados não sensíveis para consulta.
3. O passo de aprovação da UI usa um template `suspend: {}`, **sem `duration` que retome automaticamente**. Acesso ao Argo Server deve usar autenticação individual e autorização restrita a aprovadores, com trilha auditável de quem retomou o Workflow. Um simples `resume` não é autorização suficiente sem essa identidade verificável.
4. Após o `resume`, um passo independente valida a identidade e a autorização do aprovador, o vínculo da decisão ao Workflow e ao hash do plano, as revisões da solicitação e da configuração, a `generation`, a validade do artefato, o state/lock efetivo no momento do apply e o destino. Qualquer falha, expiração ou mudança nesses vínculos **bloqueia o apply** e exige novo plan e nova aprovação.
5. O executor aplica **o arquivo de plano aprovado**, não executa um novo `terraform plan` implícito, verifica o resultado e atualiza `status`. A confirmação de sucesso depende de verificação da execução; `Argo CD Synced` não comprova o resultado na AWS.
6. Reprovação, cancelamento, timeout e erro terminam sem apply; retry é controlado e nunca reaproveita aprovação para plano diferente. Operações `destroy` ou deleção de Stack não são autorizadas implicitamente.

**Critérios de aceite específicos da decisão:** (a) aprovação de pessoa autorizada libera exatamente o plano exibido; (b) usuário sem permissão de retomar Workflow não aprova; (c) `resume` sem evidência verificável de aprovação não executa apply; (d) hash/revisão/destino/generation alterados, expiração ou perda de lock impedem apply; (e) auditoria registra Workflow, plano, aprovador, instante e resultado; (f) todos os caminhos de falha permanecem sem apply indevido.

**Pendente para próxima decisão:** como representar os estados de execução e aprovação em `TerraformStack.status` (phase e conditions, ou apenas conditions) e quais referências não sensíveis expor.

### Etapa 24 — Drift Detection

- Com intervalo configurado, controller agenda `terraform plan` (ou acionamento manual controlado) para verificar divergência entre configuração/state e AWS.
- Diferenciar **mudança desejada via Git** de **drift externo** e registrar a informação no status/observabilidade; checar os resultados estruturados do plan, incluindo ações e diferenças observadas no refresh.
- Detectar, registrar e notificar drift sem remediação. A evidência é a detecção de uma mudança externa numa stack UAT já aplicada.

### Etapa 25 — Policy-Based Reconciliation

- Um **Policy Engine versionado/testável** classifica o plan segundo identidade, Environment, origem (PR vs drift), recurso, atributos, custo/impacto e operações (`create/update/delete/replace`). A POC usa OPA e amplia as regras Rego exercitadas na Etapa 09. Validar as três decisões com planos distintos, incluindo um caso de baixo risco explicitamente autorizado para auto apply.
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

- **Etapa 05.** A Application `terraformstack-crd` entrega o CRD pelo catálogo do Hub. A API Kubernetes valida exemplos com dados fictícios.
- **Etapa 06.** A Application `terraformstack-controller` entrega o controller e o `WorkflowTemplate`. A Application `terraformstack-demo` sincroniza a solicitação de exemplo após a instalação do controller. O controller cria um Workflow simulado por geração e publica o resultado no status. [Guia da Etapa 06](../../labs/06-terraformstack-controller.md) [Divisão das Applications](0003-reconciliation-and-gitops.md#applications-e-sincronização-no-hub)
- **Etapas 07 e 08.** Registro de UAT, identidade temporária do executor (incluindo bootstrap seguro no Hub kind), resolução de STS cross-account e inicialização autorizada do backend de estado.
- **Etapa 09.** Rede UAT aplicada com a revisão dos YAMLs selecionada na stack, state segregado, lock e aprovação humana do plano imutável.
- **Etapas 10–11.** Validar dependência explícita das stacks de EKS em relação à rede UAT.
- **Etapas 24 e 25.** Intervalo de drift configurável, status e alerta; depois, reconciliação com Policy Engine versionado.
- **A validar:** implementação de `dependsOn` e `driftCheckInterval` no controller, pin de versões Git/Registry/configuração, status/conditions, política de intervalos padrão, bootstrap de identidade no Hub kind, engine de políticas, catálogo, lock de state e aprovação segura.

O guia da Etapa 05 descreve a instalação do CRD no Hub por GitOps. As operações AWS das etapas posteriores exigem autorização própria.

[Arquitetura](../architecture.md) · [ADRs](README.md) · [Roadmap](../learning-roadmap.md)

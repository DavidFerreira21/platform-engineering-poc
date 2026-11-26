# ADR-0006 — TerraformStack declarativo, drift e CI/CD de Domain Applications

**Estado:** decisão vigente com contrato versionado · **Data original:** 21/11/2026 · **Revisão:** 25/11/2026

## Contexto e objetivo

A POC usa a **Kubernetes API como contrato técnico** e quatro camadas de ownership: **Foundation → Entity Core Infrastructure → Domain Infrastructure → Domain Applications**.

- **Foundation / Entity Core:** Terraform/OpenTofu para rede, IAM, dependências, e os três EKS por Environment (Management, PCI e Non-PCI).
- **Domain Infrastructure:** Crossplane v2 central com XRDs/Compositions/Providers específicos para recursos AWS como ObjectStorage e Database.
- **Domain Applications:** Application Team mantém **código, GitHub Actions e manifests Kubernetes no mesmo repositório de aplicação de teste separado**, a integrar na Fase 5 (Etapas 22–25) do roteiro; Argo CD **local no Management EKS** lê esses manifests GitOps e entrega em EKS PCI/Non-PCI autorizados. O DevOps Federado mantém somente os manifests de Domain Infrastructure.

**Intenção desta decisão:** oferecer, para módulos Terraform, **experiência declarativa semelhante à do Crossplane**: manifestos de intenção versionados no Git, provisionamento inicial condicionado a políticas, detecção recorrente de drift, e remediação automática ou com aprovação conforme riscos. Não se trata apenas de disparar uma execução única de Terraform.

## Decisão 1 — TerraformStack representa a infraestrutura desejada

O recurso **`TerraformStack`** tem um CRD próprio versionado na Etapa 04. Na Etapa 05, o **TerraformStack Controller** passa a observar esse recurso no Control Plane do Hub.

- O `spec` referencia a stack aprovada por `sourceRef`, o destino lógico `Entity + Environment` e a versão dos YAMLs de configuração. O catálogo do Simulator seleciona o backend local, o PVC e a ServiceAccount. Na integração AWS, a plataforma resolve a role e o backend pelo cadastro autorizado.
- O recurso mantém o ciclo de vida da **infraestrutura desejada**; não representa apenas um `terraform apply`.
- O `status` expõe condições, última verificação, revisão observada, resumo do último plan, classificação, referência de workflow, estado de aprovação e divergência detectada. Não guardar segredos nem todo o plan em `status`.
- `TerraformRun` fica **reservado como nome opcional para uma execução imutável** (ou pode ser substituído por um Argo Workflow criado pelo controller). **Não criar um segundo CRD só para representar runs antes de existir justificativa.**
- **Entity/Environment/Domain continuam registros lógicos Git**, não CRDs criados por esta decisão.

### Contrato inicial

A Etapa 04 publica `platform.example.io/v1alpha1` com escopo de namespace. O schema exige `spec.target.entity/environment` e `spec.sourceRef.name/path`. `sourceRef.version` e `configuration.version` são opcionais, com default `latest`. O subrecurso `/status` expõe condições, execução, política e revisões resolvidas. A Etapa 04 valida o exemplo com dry-run na API do Hub, e a Etapa 06 usa a mesma solicitação para executar o Simulator. [Schema](../../control-plane/terraformstack/crd/terraformstack.yaml) · [Exemplo](../../control-plane/terraformstack/examples/simulator/stacks.yaml) · [Guia da Etapa 04](../../labs/04-terraformstack-crd.md)

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

O módulo `configuration` lê o arquivo `<stack>.yaml` nos níveis global, Entity e Environment e expõe o resultado em `module.configuration.values`. A mesclagem é recursiva nos objetos. Quando uma chave recebe novo valor, a camada mais específica substitui o valor anterior, inclusive para listas, booleanos, números e `null`. Os nomes das chaves e os tipos decodificados do YAML são preservados. A stack consumidora seleciona os campos e declara seus tipos e validações. O provider `isometry/deepmerge` fornece a função de mesclagem. [Leitura de YAML no Terraform](https://developer.hashicorp.com/terraform/language/functions/yamldecode) · [Função de mesclagem](https://registry.terraform.io/providers/isometry/deepmerge/1.3.1/docs/functions/mergo)

**Stacks iniciais da POC:** `network` e `eks` são **stacks completas de infraestrutura**, cada uma composta pelos módulos Terraform internos e recursos necessários ao seu domínio, não apenas módulos isolados. `network` prepara o baseline da rede spoke (Etapa 13); `eks` prepara clusters, IAM e dependências de addons (Etapas 10, 14, 15 e 30). O módulo auxiliar **`configuration`** não provisiona recursos e é consumido por **todas as stacks**, lendo `network.yaml`, `eks.yaml` ou um arquivo correspondente de uma stack futura. O mesmo `TerraformStack` poderá gerenciar stacks adicionais, como `bootstrap`, sem novo CRD. Destino, região, role e backend continuam no cadastro confiável de Environment, não no módulo Configuration.

**Baseline e feature flags:** cada stack implementa um baseline abrangente de componentes previstos, com defaults definidos e validados. O Configuration resolve overrides hierárquicos para ativar ou desativar opções; certas dependências obrigatórias podem permanecer **forçadas pelo contrato/política**, sem override permitido. A stack EKS permite combinações independentes de `management`, `pci` e `nonPci`; o **perfil inicial autorizado de UAT** habilita Management e Non-PCI progressivamente, enquanto PCI exige autorização específica. Defaults não são licença para provisionar automaticamente recursos caros/PCI: cada execução respeita plan, custo, policy e aprovação. Flags de pré-requisitos de addons criam somente IAM/recursos AWS necessários; operadores Kubernetes pertencem ao Platform GitOps. Validar Karpenter contra EKS Auto Mode para não duplicar controladores.

**Outputs:** as stacks `network` e `eks` expõem saídas tipadas; a publicação de um subconjunto permitido para o Parameter Store da conta do Environment ocorre pelo **GitOps Bridge (Etapa 17)**, depois de apply e verificação. Não publicar state, plano ou segredos como output distribuído. A implementação deve verificar combinações de flags e tornar explícitas as dependências de rede/IAM, sem criar um quarto módulo obrigatório agora.

No Simulator, a Application acompanha `main` e injeta o SHA efetivamente sincronizado em `platform.example.io/revision`. Quando `sourceRef.version` ou `configuration.version` é omitido ou recebe `latest`, o controller exige essa anotação com um SHA completo. Cada versão explícita também exige um SHA completo. O controller considera a anotação junto à geração para detectar commits que alteram apenas os YAMLs. O Workflow fixa as duas revisões e usa o binário do plano autorizado no apply. A Etapa 13 reutiliza o vínculo entre revisão, plano e aprovação com destino AWS registrado em `environment.yaml`. [Resolução de revisões](../flows/terraformstack.md#2-contrato-e-resolução-de-revisões)

Os testes do Configuration cobrem os três níveis de herança, a preservação de campos aninhados e a substituição de listas e valores escalares para uma stack de exemplo. A validação do fluxo confirma que um commit que altera apenas o YAML muda os valores do próximo plan com a mesma imagem e o mesmo conteúdo de código Terraform.

## Decisão 2 — GitOps aplica a intenção; Argo Workflows executa Terraform

```text
Backstage / CLI / pessoa → PR + validações + aprovação Git
  → Argo CD CENTRAL → Kubernetes API / TerraformStack (objeto Kubernetes)
  → TerraformStack Controller (validação + agendamento)
  → Argo Workflows → Terraform Executor → terraform plan
  → Etapa 13: aprovação humana do plano imutável → terraform apply do mesmo plano
  → Etapa 28: verificação periódica e relato de drift, sem apply
  → Etapa 29: decisão versionada auto apply / aprovação / bloqueio
  → AWS Foundation / Entity Core → verificação → status do TerraformStack
```

**Separação de responsabilidades:**
- **Argo CD central:** sincroniza `TerraformStack` do Git para Kubernetes, **não executa Terraform nem cria workflows por conta própria**.
- **TerraformStack Controller:** observa CRs, valida módulo/destino/permissões, dispara ou agenda verificações, controla concorrência/retries, acompanha o resultado da política avaliada no Workflow e atualiza status.
- **Argo Workflows:** orquestra plan, avaliação das políticas, gates/aprovação, apply e verificações.
- **Terraform:** calcula e aplica mudanças no ConfigMap do Simulator. As raízes AWS entram nas etapas de Foundation e Entity Core.

**Argo CD `Synced` significa Git == CR Kubernetes, não AWS sem drift.** Drift de infraestrutura é informado em `TerraformStack.status.conditions` (por exemplo `DriftDetected`, `Ready`, `ApprovalRequired`, `ReconcileBlocked`) e pode ser exposto por health check customizado, CLI ou observabilidade. A POC **não depende** de interface/health customizados do Argo CD.

### Executor Argo, identidade do destino e isolamento

**Decisão:** manter o `TerraformStack Controller` próprio com **Argo Workflows** como executor/orquestrador; não substituir pelo Tofu Controller nem criar um segundo Runner Controller. O controller observa o CR, resolve o destino autorizado e cria um Workflow governado. O Argo cria Pods efêmeros a partir do WorkflowTemplate canônico. Os passos Terraform usam o runner genérico e os scripts versionados. OPA usa sua imagem oficial e monta a versão imutável do Rego. O catálogo fixa a raiz, as revisões Git, o backend e a identidade de execução.

A Etapa 06 aplica o Simulator com o provider Kubernetes. Na Etapa 10, os mesmos blocos serão instalados no Hub EKS. A Etapa 12 valida a sessão de origem obtida por identidade de Pod e a assunção da role autorizada em UAT.

Identidades **distintas e complementares**:

- **Kubernetes:** o controller tem RBAC apenas para suas operações de controle; os Pods do Workflow usam uma `serviceAccountName` de executor reservada e protegida contra uso por Workflows arbitrários.
- **AWS de origem:** a identidade de Pod no Hub EKS fornece a sessão inicial temporária. A integração AWS é validada na Etapa 12.
- **AWS de destino:** `spec.target.entity/environment` é uma chave lógica para consultar o registro confiável. O executor obtém uma sessão STS na role **permitida** da conta/região daquele Environment; manifesto não aceita role arbitrária. Conferir `GetCallerIdentity`, backend e trust policy. Não é o Kubernetes ServiceAccount, isoladamente, que escolhe a conta AWS.

No Simulator, os Pods compartilham o binário do plano pelo PVC da stack. O backend libera o lock durante a pausa de aprovação. RBAC e admissão restringem a criação dos Workflows e a edição do gate. Na AWS, os Pods de plan e apply usam sessões temporárias e armazenamento privado. O fluxo deve validar a identidade e o destino da sessão de apply antes de consumir o plano aprovado. [Autorização final do apply](../flows/terraformstack.md#autorização-final-do-apply)

### Dependências e checagem periódica configurável

**Contrato implementado no schema**, inspirado em práticas de operadores Terraform/OpenTofu: `spec.dependsOn` representa stacks dependentes e `spec.reconciliation.driftCheckInterval` expressa a frequência desejada de verificação. Os campos são aceitos pela API desde a Etapa 04 e exercitados no Simulator na Etapa 06.

```yaml
spec:
  dependsOn:
    - name: network-uat
  reconciliation:
    driftCheckInterval: 6h
```

- `dependsOn`: antes de executar uma stack, conferir dependência `Ready` com `observedGeneration` atual. Resolver por nome estável, no mesmo namespace e Entity/Environment por padrão; rejeitar referências proibidas, dependências inexistentes e ciclos. Exemplo: EKS Management depende da stack de rede UAT. Dependência pronta **não autoriza** apply sem política e aprovação.
- `driftCheckInterval` agenda um plan após o intervalo desde a última execução `Ready`. O status expõe `lastCheckTime` e `lastDriftTime`. A política do Simulator pode autorizar a correção. A Etapa 28 prevê jitter, exposição da próxima checagem e detecção sem remediação na AWS. A Etapa 29 valida a correção governada na AWS.
- O controller agenda por `GitOps`, `DriftCheck` e `Retry`. O retry atual depende de falha operacional e da política de retry declarada. O acionamento manual dedicado por annotation permanece planejado.

**Fases da implementação:** Etapas 05–06 instalam e validam a criação idempotente de Workflows no Simulator; Etapa 12 valida a identidade AWS; Etapa 13 aplica Network com aprovação; Etapas 14–15 usam `dependsOn` em rede → EKS; Etapa 28 ativa intervalo de drift.

[Gatilhos, drift e concorrência](../flows/terraformstack.md#4-gatilhos-drift-e-concorrência) · [Evolução na AWS](../flows/terraformstack.md#7-evolução-na-aws)

### Contrato do CRD: sourceRef e lifecycle de destruição

**Contrato atualizado:** manter `TerraformStack` como API genérica de stacks completas e aceitar `spec.sourceRef.name/path/version` como referência à **fonte versionada e aprovada**. `sourceRef.version` opcional (`latest` por padrão) resolve Registry ou Git a uma versão/commit imutável no executor real; `configuration.version` também pode ser omitido para resolver `latest`, com pin antes do plan. O source é resolvido no catálogo sob controle do Platform Team e não aceita URL arbitrária/credenciais. A revisão de código da stack e a revisão de `configuration` são independentes e ficam fixadas em cada execução. O Simulator usa SHAs Git para código e configuração, com `latest` resolvido pela anotação GitOps. Não dependemos do `GitRepository` de Flux.

**Dois controles de destruição já presentes no CRD, com execução local na Etapa 06 e validação na AWS na Etapa 29; ambos sujeitos ao Policy Engine e à aprovação humana:**

- `spec.lifecycle.destroy: true`: solicita um `plan -destroy` e um `apply` **do mesmo plano aprovado**, mantendo o CR para acompanhar o resultado. Não reaproveitar aprovação de apply anterior.
- `spec.lifecycle.destroyResourcesOnDeletion: true`: habilita a **intenção** de teardown no momento da exclusão do CR. O default é `false`; excluir YAML/CR não constitui aprovação. O controller usa finalizer, snapshot imutável de source/config/destino/state e Workflow com aprovação de identidade verificável. Quando não autorizado, bloqueia sem destruir. Uma solução separada e auditada para `orphan/retain` deverá ser estudada para falhas e CRs presos em deletion.

**Garantias:** aplicar proteção das dependências reversas (não destruir rede enquanto EKS depender dela), state/lock, IAM, planejamento imutável, política de risco para delete/replace e registros de auditoria. `finalizer` protege o ciclo de exclusão, **não** substitui aprovação. Destruição explícita e destruição on-delete ficam previstas para a **Etapa 29**, nunca habilitadas silenciosamente nas etapas AWS iniciais.

Outros campos já validados pelo schema: `spec.suspend`, `spec.reconciliation.retryPolicy`, `spec.timeout` por etapa e `spec.healthChecks` opcionais. O Simulator executa suspensão, retry, checagem de drift e dependências. Timeouts configuráveis e health checks do schema ficam para sua integração operacional. `serviceAccountName`, AWS role, backend e desbloqueio forçado **não** serão selecionáveis livremente pelo usuário via CR.

[Dependências e destruição](../flows/terraformstack.md#5-dependências-e-destruição)

## Decisão 3 — três etapas incrementais da POC

A Etapa 06 valida primeiro o ciclo completo no `kind` com Terraform real, OPA, aprovação, destroy, drift periódico e status, usando apenas recursos locais. As Etapas 13, 28 e 29 retomam esses mecanismos no contexto AWS.

### Etapa 13 — Deploy declarativo ponta a ponta com aprovação

- Aplicar `TerraformStack` pelo GitOps central (ou diretamente na API para teste local), e ver o controller iniciar Workflow/plan.
- **Toda execução de apply na AWS exige aprovação humana explícita do plano** até a evolução prevista na Etapa 29. Auto apply só será avaliado com política versionada.
- Validar schema, autorização, idempotência por generation/revisão, erros/retry, status e evidências. A Etapa 06 exercita a pipeline com o Simulator. A Etapa 13 aplica a rede UAT com state segregado, lock, plano imutável e aprovação vinculada a hash, revisão Git e destino. O mesmo fluxo de aprovação atende às stacks de EKS.

O passo de política do Workflow executa OPA com regras Rego versionadas sobre `terraform show -json` do plano salvo. No aceite, um plano permitido chega à pausa de aprovação humana e um plano com abertura de SSH para `0.0.0.0/0` termina bloqueado antes do apply. Exercitar o caso bloqueado somente até a decisão. Erro de avaliação ou decisão ausente também bloqueia. Registrar motivo, revisão da política e hash do plano, mantendo o JSON em armazenamento restrito. Após aprovação, o executor aplica o mesmo plano binário.

### Aprovação humana na Argo Workflows UI

O canal de aprovação é a UI do Argo Workflows. No Simulator, o step `approval` suspende a execução quando o OPA retorna `RequireApproval`. O aprovador autorizado usa **Resume**, escolhe `Approve` e informa o hash do plano. O step `authorize` valida a resposta. O apply confere o binário e compara o efeito previsto com um novo plan. A função `claim()` confirma a intenção vigente e registra o início do apply no CR. A Etapa 13 exige esse gate humano para todo plano AWS permitido. [Pipeline e aprovação](../flows/terraformstack.md#3-pipeline-e-aprovação)

**Requisitos para a integração AWS:**

1. O `TerraformStack Controller` dispara um Workflow com identificação imutável da stack, `metadata.generation`, revisão Git da solicitação, versão da configuração, versão resolvida da stack e destino autorizado. O backend protege as operações Terraform com lock quando necessário, **sem manter o lock retido durante uma aprovação pendente**.
2. O executor gera `terraform plan -out` com a configuração da revisão selecionada. Persiste o plano binário **em armazenamento privado e com prazo de validade**, associado ao hash SHA-256, ao state/serial, à revisão de configuração e à identidade do destino. Expõe apenas metadados não sensíveis para consulta.
3. O passo de aprovação da UI usa um template `suspend: {}`, **sem `duration` que retome automaticamente**. Acesso ao Argo Server deve usar autenticação individual e autorização restrita a aprovadores, com trilha auditável de quem retomou o Workflow. Um simples `resume` não é autorização suficiente sem essa identidade verificável.
4. Após o `resume`, um passo independente valida a identidade e a autorização do aprovador, o vínculo da decisão ao Workflow e ao hash do plano, as revisões da solicitação e da configuração, a `generation`, a validade do artefato, o state/lock efetivo no momento do apply e o destino. Qualquer falha, expiração ou mudança nesses vínculos **bloqueia o apply** e exige novo plan e nova aprovação.
5. O executor aplica o arquivo de plano aprovado. Um plan adicional pode verificar a validade do plano aprovado antes do apply. O controller registra o resultado da verificação em `status`.
6. Reprovação, cancelamento, timeout e erro terminam sem apply; retry é controlado e nunca reaproveita aprovação para plano diferente. Operações `destroy` ou deleção de Stack não são autorizadas implicitamente.

**Critérios de aceite específicos da decisão:** (a) aprovação de pessoa autorizada libera exatamente o plano exibido; (b) usuário sem permissão de retomar Workflow não aprova; (c) `resume` sem evidência verificável de aprovação não executa apply; (d) hash/revisão/destino/generation alterados, expiração ou perda de lock impedem apply; (e) auditoria registra Workflow, plano, aprovador, instante e resultado; (f) todos os caminhos de falha permanecem sem apply indevido.

O status atual publica a condição ativa em `status.conditions` e os dados da tentativa em `status.execution`, `status.policy` e `status.resolvedSource`. [Condições e diagnóstico](../flows/terraformstack.md#6-status-e-diagnóstico)

### Etapa 28 — Drift Detection

- Com intervalo configurado, controller agenda `terraform plan` (ou acionamento manual controlado) para verificar divergência entre configuração/state e AWS.
- Diferenciar **mudança desejada via Git** de **drift externo** e registrar a informação no status/observabilidade; checar os resultados estruturados do plan, incluindo ações e diferenças observadas no refresh.
- Detectar, registrar e notificar drift sem remediação. A evidência é a detecção de uma mudança externa numa stack UAT já aplicada.

### Etapa 29 — Policy-Based Reconciliation

- Um **Policy Engine versionado/testável** classifica o plan segundo identidade, Environment, origem (PR vs drift), recurso, atributos, custo/impacto e operações (`create/update/delete/replace`). A POC usa OPA e amplia as regras Rego exercitadas na Etapa 13. Validar as três decisões com planos distintos, incluindo um caso de baixo risco explicitamente autorizado para auto apply.
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

- **Etapa 04.** A Application `terraformstack-crd` entrega o CRD pelo catálogo do Hub. A API Kubernetes valida exemplos com dados fictícios.
- **Etapas 05–06.** As Applications do controller e da pipeline publicam os blocos. A Application do Simulator sincroniza as solicitações de teste. [Guia da Etapa 05](../../labs/05-terraformstack-runtime.md) · [Simulator](../../labs/06-terraformstack-simulator.md)
- **Etapas 07–10.** Aceite Crossplane local, Foundation da conta de plataforma, bootstrap externo do Hub EKS e reinstalação dos mesmos blocos.
- **Etapas 11–12.** Registro UAT, identidade temporária do executor a partir do Hub EKS e inicialização autorizada do backend.
- **Etapa 13.** Rede UAT aplicada com a revisão dos YAMLs selecionada na stack, state segregado, lock e aprovação humana do plano imutável.
- **Etapas 14–15.** Validar dependência explícita das stacks de EKS em relação à rede UAT.
- **Etapas 28 e 29.** Intervalo de drift configurável, status e alerta; depois, reconciliação com Policy Engine versionado.
- **A validar na AWS.** Consumo de dependências por EKS, identidade temporária, segregação e locking no backend, políticas OPA para recursos AWS e auditoria do apply aprovado. Releases por componente, fontes de Registry, jitter e acionamento manual dedicado permanecem como evoluções.

O guia da Etapa 04 descreve a instalação do CRD no Hub por GitOps. As operações AWS das etapas posteriores exigem autorização própria.

[Arquitetura](../architecture.md) · [ADRs](README.md) · [Roadmap](../learning-roadmap.md)

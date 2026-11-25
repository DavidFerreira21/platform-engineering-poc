# Roteiro prático da POC — Roadmap V2

[← Início](../README.md) · [Arquitetura](architecture.md) · [Guias das etapas](../labs/README.md) · [GitOps Flow](flows/gitops.md) · [TerraformStack Flow](flows/terraformstack.md) · [GitOps Bridge Flow](flows/gitops-bridge.md)

A POC é construída em **oito fases e 32 etapas (00–31)**. Cada etapa adiciona uma capacidade técnica, reutiliza as anteriores e termina com uma verificação objetiva. O objetivo dos guias é **construir, integrar e observar o funcionamento da plataforma**, sem tentar ensinar exaustivamente cada ferramenta.

Os laboratórios existentes das **Etapas 00–08** estão documentados. A Etapa 08 prepara e testa **identidade AWS temporária, bucket S3 para state, executor em Argo Workflows e módulo Configuration**, sem `terraform apply` nem stack Network real. Requer permissões AWS e execução supervisionada; não representa recursos já implantados. As demais entregas descrevem o escopo planejado; ainda **não foram implementadas** só por aparecerem aqui. Os guias mantêm a estrutura atual: **Objetivo → O que será feito → Por que fizemos assim → Pré-requisitos → Execução → O que observar → Limpeza (quando aplicável)**. [Convenção](../labs/README.md#como-ler-e-reproduzir-uma-etapa).

## Visão das oito fases

| Fase | Etapas | Marco técnico |
| --- | --- | --- |
| **1 — Control Plane local** | 00–04 | Hub `kind` com Argo CD, Argo Workflows e Crossplane Core operacionais |
| **2 — TerraformStack** | 05–09 | CRD, controller, plan/aprovação/apply e primeiro recurso AWS real |
| **3 — EKS, GitOps e Bridge** | 10–14 | Environment UAT com EKS, onboarding, addons, Argo CD local e outputs consumidos |
| **4 — Domain Infrastructure** | 15–17 | Capabilities S3 e Aurora PostgreSQL via Crossplane |
| **5 — Primeira aplicação e CI/CD** | 18–21 | App FastAPI, Aurora, S3, GitHub Actions, ECR e deploy GitOps |
| **6 — Observabilidade e operação** | 22–27 | Telemetria, diagnóstico, drift, políticas e isolamento |
| **7 — Experience Layer** | 28–29 | Personas, permissões, catálogo e Starter Kit |
| **8 — Golden Path ponta a ponta** | 30–31 | Solicitação federada e entrega integrada da aplicação |

## Fase 1 — Preparação do Control Plane local

Preparar o Hub `kind` e instalar cada motor com um teste funcional pequeno. Os componentes permanecem instalados para as etapas seguintes; recursos apenas de demonstração podem ser removidos.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **00** | [Documentação e monorepo](repository-structure.md) | Estrutura, camadas, personas e ownership definidos |
| **01** | [Hub kind e Traefik](../labs/01-platform-management-cluster.md) | Cluster, namespaces, DNS, acesso e Ingress local verificados |
| **02** | [Argo CD central e catálogo de add-ons](../labs/02-argocd-central.md) | ApplicationSet gera Applications a partir do Git; ConfigMap de teste sincronizado |
| **03** | [Argo Workflows](../labs/03-argo-workflows.md) | Um Workflow executado, com Pod, resultado e logs observáveis |
| **04** | [Crossplane Core](../labs/04-crossplane-core.md) | XRD/Composition de teste reconciliam um ConfigMap local |

## Fase 2 — TerraformStack e primeira execução na AWS

Construir o contrato declarativo, o executor e as **primeiras stacks de infraestrutura** até o primeiro apply AWS autorizado, sem antecipar remediação de drift. **Network** e **EKS** são stacks completas: cada uma compõe quantos módulos Terraform e recursos forem necessários para entregar uma capacidade da spoke. **Configuration** é o módulo auxiliar compartilhado que resolve os valores de cada stack; não é uma terceira stack provisionável.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **05** | [CRD Kubernetes `TerraformStack` via GitOps](../labs/05-terraformstack-crd.md) | API valida `spec` fictício e expõe `status.conditions`, sem executar Terraform. [ADR-0006](adr/0006-terraformrun-and-application-ci.md) |
| **06** | [Controller do TerraformStack e executor simulado](../labs/06-terraformstack-controller.md) | Sincronizar CR com `sourceRef`, acompanhar Workflow simulado e status por geração; `suspend` bloqueia novas simulações. O controller ainda não resolve Registry/Git, dependências, drift ou destroy |
| **07** | [Registro e validação de UAT](../labs/07-registro-ambientes-aws.md) | `environment.yaml` associa `entity-a/uat` à conta/região; perfil humano SSO e STS validados sem Secrets no Kubernetes |
| **08** | [Terraform Executor, state S3 e Configuration](../labs/08-identidade-executor-aws.md) | Executar bootstrap SSO → STS → Pod; criar bucket S3 com locking nativo autorizado; carregar imagem Terraform Executor, validar YAMLs Network/EKS com Configuration e executar `init/validate/plan` sem `apply`. O executor ainda não é acionado pelo controller real |
| **09** | Stack `network` e primeira execução real em UAT | Construir a **stack Network** completa, composta pelos módulos necessários de VPC, subnets, rotas e requisitos da spoke, com feature flags; consumir `network.yaml` via Configuration e identificar fonte aprovada por `sourceRef.name/path/version`, resolvendo `latest` para artefato imutável antes do plan. TerraformStack/Workflow executa plan, aprovação, apply e verify do plano aprovado e publica outputs da rede |

**Execução e identidade:** Argo Workflows agenda Pods com uma ServiceAccount Kubernetes dedicada; o Terraform Executor usa credenciais AWS temporárias e assume apenas a role autorizada pelo registro de Entity/Environment. Em `kind`, o mecanismo inicial de credenciais é definido na Etapa 08; EKS Pod Identity é opção quando o Hub operar em EKS. Não criar um Runner Controller próprio. O schema já aceita `dependsOn` e `driftCheckInterval`; sua reconciliação efetiva permanece nas etapas posteriores. [Contrato](../control-plane/terraformstack/crd/terraformstack.yaml). [Decisão detalhada](adr/0006-terraformrun-and-application-ci.md#executor-argo-identidade-do-destino-e-isolamento).

**Promoção de Configuration na Etapa 09.** Após o merge de alterações em `config/`, a CI publica uma release SemVer imutável da configuração e abre um PR de promoção. Pelo tipo de stack e pela hierarquia global → entidade → ambiente, o PR atualiza `spec.configuration.version` uma vez em cada manifesto afetado. Uma alteração em `config/br/uat/network.yaml` seleciona a rede de BR/UAT. Uma alteração em `config/br/network.yaml` seleciona as redes de BR. Uma alteração em `config/network.yaml` seleciona as redes de todas as entidades consumidoras, inclusive BR e CL no exemplo. As stacks EKS continuam na versão anterior. Uma sobrescrita local pode resultar em plan sem mudanças. Após o merge e o sync dos CRs pelo Argo CD, cada nova geração inicia seu próprio plan. Para retroceder uma stack, promover nela uma release anterior e revisar o novo plano.

**Aceite de OPA na Etapa 09.** O passo de política do Workflow executa OPA com regras Rego versionadas sobre `terraform show -json` do plano salvo. No aceite, um plano permitido chega à pausa de aprovação humana e um plano com abertura de SSH para `0.0.0.0/0` termina bloqueado antes do apply. Exercitar o caso bloqueado somente até a decisão. Erro de avaliação ou decisão ausente também bloqueia. Registrar motivo, revisão da política e hash do plano, mantendo o JSON em armazenamento restrito. Após aprovação, o executor aplica o mesmo plano binário.

**Contrato desde a Etapa 05:** o CRD aceita `sourceRef` (name/path/version opcional), `configuration.version`, `dependsOn`, `suspend`, `reconciliation`, `lifecycle`, `timeout` e `healthChecks`. `sourceRef.version` ausente e `configuration.version` ausente representam `latest` **apenas na intenção**; executor real deve resolver e fixar versões imutáveis antes do plan. Não há compatibilidade com `spec.module`: apenas `sourceRef` é aceito. **Aceitação pelo CRD não significa funcionalidade implementada:** a Etapa 06 é um simulador e destroy/drift/retry/policies dependem das etapas de execução descritas abaixo.

**Modelo de stacks:** `TerraformStack` representa uma capacidade declarativa, como `network` ou `eks`, e aponta para sua implementação aprovada, que pode compor **vários módulos Terraform internos**. Futuras stacks (por exemplo, `bootstrap` ou uma stack específica) usam o mesmo controller, sem exigir novos tipos de CRD. `configuration` é um módulo auxiliar comum, invocado por cada stack, e não uma stack AWS. O destino/role permanece no cadastro confiável de Environment e no executor.

**Marco da fase:** `TerraformStack` é capaz de receber intenção via GitOps, executar Terraform com aprovação humana e produzir infraestrutura real. Terraform não aplica continuamente a cada reconcile. **Auto Apply, verificação periódica e correção de drift entram na Fase 6.**

## Fase 3 — Provisionamento de EKS, GitOps e GitOps Bridge

Usar o TerraformStack já operacional e construir a **stack EKS completa e reutilizável**, que compõe os módulos internos necessários para os clusters e os pré-requisitos AWS de addons, consumindo `eks.yaml` via Configuration. Feature flags controlam quais componentes do baseline são habilitados ou desabilitados, sem exigir stacks distintas por tipo de cluster. A instalação dos operadores Kubernetes permanece em GitOps. O registro do cluster, a autorização e a configuração por GitOps são ações independentes. Antes de registrar EKS privado, comprovar a conectividade e a identidade que o Argo CD usará; não presumir comunicação entre VPCs/contas.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **10** | Stack `eks`, Management EKS e bootstrap de gestão | Implementar a **stack EKS** e seus módulos internos para clusters, IAM e pré-requisitos dos addons; consumir `eks.yaml` via Configuration. Depender de Network `Ready`, aplicar inicialmente o perfil autorizado para Management, conferir outputs/acesso e iniciar onboarding e Argo CD local |
| **11** | Stack EKS: Non-PCI e onboarding por flags | Reutilizar a **mesma stack EKS**, ativando Non-PCI conforme perfil de UAT, sem exigir PCI; validar `dependsOn`, IAM, outputs, registro e labels `entity/environment/cluster-role/compliance`. Os clusters podem ser habilitados independentemente ou juntos, sempre sujeitos ao plan e à autorização |
| **12** | Platform GitOps multicluster e perfis de add-ons | ApplicationSets do Argo CD central selecionam clusters registrados e instalam addons corporativos por perfil. Validar os pré-requisitos AWS opcionais produzidos pela **stack EKS** (roles, permissões, identidade) e seu consumo pelo GitOps; confirmar ownership único |
| **13** | GitOps Bridge — outputs, Parameter Store e ESO | Integrar publicação **allowlisted** dos outputs de Network/EKS no Parameter Store da **conta UAT** após apply/verify; ESO no **Management EKS** lê o SSM e materializa um Secret consumido localmente. Publicar outputs não registra clusters |
| **14** | Smoke test de Domain Application no Non-PCI | Argo CD local sincroniza um workload HTTP mínimo a partir de Git, verifica serviço/health e remove apenas o teste; os componentes de plataforma permanecem |

**Feature flags das stacks:** cada stack descreve um **baseline completo**, inclusive componentes opcionais, e define defaults explícitos no schema. Configuration aceita overrides hierárquicos para **habilitar ou desabilitar** recursos; configurações obrigatórias podem ser forçadas por política e não podem ser desligadas pelo consumidor. A stack EKS suporta Management, PCI e Non-PCI independentes; o perfil de implantação inicial habilita apenas os clusters autorizados, mesmo que o catálogo da stack contemple todos. PCI permanece condicionado à Etapa 26 e aprovação de custo. Flags de pré-requisitos de Karpenter, quando aplicáveis, criam identidades/dependências AWS, **não instalam o operador**; respeitar EKS Auto Mode. Toda combinação deve passar por validação, plano, policy e aprovação. Outputs de Network/EKS alimentarão o GitOps Bridge da Etapa 13.

**Marco da fase:** UAT possui Management e Non-PCI operacionais, ambos sob os mecanismos GitOps autorizados. O EKS PCI será avaliado com custo e isolamento na **Etapa 26**. O [GitOps Bridge](flows/gitops-bridge.md) não requer que o ApplicationSet central leia os Secrets do Management EKS.

## Fase 4 — Domain Infrastructure com Crossplane

Entregar as primeiras Capabilities AWS a partir de contratos Kubernetes declarativos. Os Providers reconciliam recursos no Environment autorizado, sem Terraform Workflow para cada Capability.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **15** | Provider AWS e identidade Crossplane no Hub | Provider assume role temporária em UAT; conferir acesso correto, renovação e rejeição de destino não autorizado |
| **16** | `ObjectStorage` via Crossplane | XRD/Composition solicitam um bucket S3 na conta UAT; verificar condições, permissões e política de remoção |
| **17** | `Database` via Crossplane | XRD/Composition criam Aurora PostgreSQL em sub-redes autorizadas; conferir prontidão, acesso, custo e teardown planejado |

**Marco da fase:** `ObjectStorage` e `Database` são capacidades reais e disponíveis para a aplicação da Fase 5. O **Domain Infrastructure** é gerenciado pelo DevOps Federado no Golden Path; a primeira implantação técnica das Capabilities é conduzida pelo Platform Team.

## Fase 5 — Primeira aplicação e esteira CI/CD

Construir uma aplicação **mínima e funcional** usando a infraestrutura pronta. O cenário é **Rastreamento de Encomendas**: uma tela permite registrar o recebimento e consultar os últimos recebimentos. O código, os testes, a configuração de deployment e a pipeline CI ficam em **repositório separado, sob ownership da squad**, conforme o [ADR-0007](adr/0007-personas-and-developer-experience.md).

**Stack escolhida:** **Python + FastAPI** para backend HTTP/API e **Jinja2 + HTML/CSS** para uma interface renderizada pelo servidor. Não usar Django, React ou um segundo backend nesta POC. O banco será **Aurora PostgreSQL** da Capability `Database`; a API gravará um **JSON de evidência de recebimento** no S3 da Capability `ObjectStorage`. O S3 não substitui logs operacionais.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **18** | Aplicação de referência: FastAPI + Jinja2 | Frontend simples registra código da encomenda e lista recebimentos; API e testes locais funcionando |
| **19** | Integração da aplicação com Capabilities | API persiste código e timestamp no Aurora, grava evidência JSON no S3 e valida leituras, IAM mínimo e conectividade; sem credenciais fixas no código |
| **20** | CI da aplicação: GitHub Actions e Amazon ECR | Pipeline executa testes, build do contêiner, scan de segurança e push ao ECR com OIDC/credenciais temporárias; imagem identificada por digest |
| **21** | CD GitOps: promoção e deploy no EKS Non-PCI | CI propõe atualização de digest no manifesto/overlay versionado; após revisão, Argo CD local sincroniza a nova imagem; confirmar rollout e funcionamento da tela |

**Responsabilidades:** Platform Team oferece ECR, workflow/template reutilizável, RBAC e guardrails; a squad é owner do código, da versão e do release da aplicação; DevOps Federado é owner das solicitações de Capabilities. A pipeline **não faz `kubectl apply` imperativo**. Manifests e código permanecem inicialmente no mesmo repositório da aplicação, para reduzir complexidade. O deploy pode usar Helm ou Kustomize com defaults mantidos pela plataforma; não criar agora um CRD `Application` próprio.

**Decisão técnica ainda aberta para a Etapa 20:** escolher se o repositório ECR ficará na conta de plataforma ou na conta do Environment. O critério é permitir publicação autorizada por OIDC e pull da mesma imagem pelo EKS Non-PCI; se for cross-account, validar IAM e repository policy antes do deploy. Não decidir isso implicitamente pelo desenho do fluxo.

**Marco da fase:** uma alteração de código gera uma imagem versionada no ECR e chega ao Non-PCI via GitOps, usando Aurora e S3 reais. O objetivo aqui é implementar e validar a entrega; **a experiência self-service das três personas será demonstrada na Fase 8**, sem criar outra pipeline.

## Fase 6 — Observabilidade, operação e governança

Instrumentar os componentes implantados e concluir o lifecycle operacional do TerraformStack. Políticas preventivas essenciais já devem existir desde a introdução de cada componente; esta fase consolida diagnósticos e cenários mais avançados.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **22** | Telemetria de plataforma e aplicação | OpenTelemetry Collector/gateway no Data Plane e integração com Grafana, Loki e Tempo; logs/traces correlacionados da API, dos Workflows e dos controllers |
| **23** | Diagnóstico, recuperação e custos | Investigar uma falha controlada, recuperar a operação e conferir rastreabilidade e custos AWS |
| **24** | Drift Detection do TerraformStack | Implementar `driftCheckInterval` configurável, agendamento idempotente com jitter e verificação periódica; detectar alteração externa em stack UAT, registrar última/próxima checagem e condição/alerta **sem aplicar correção** |
| **25** | Reconciliação, destroy e políticas | Ampliar as regras OPA da Etapa 09 e comprovar **Auto Apply / Require Approval / Deny** com três planos de teste e evidências da etapa alcançada; testar drift, `plan -destroy` explícito, destroy on deletion com finalizer, proteção de dependências, aprovação humana e bloqueio de ações sensíveis |
| **26** | EKS PCI e isolamento PCI/Non-PCI em UAT | Reutilizar a **stack EKS** e habilitar PCI somente com autorização/custo aprovados; registrar com perfil restrito, aplicar addons autorizados e testar negações entre PCI e Non-PCI. Sem PCI real, declarar somente validação lógica, não isolamento operacional |
| **27** | Registro de PROD e isolamento por Environment | Registrar conta PROD distinta, validar identity/mapping/negações e políticas UAT/PROD. Provisionamento real em PROD depende de autorização separada; não presumir EKS PROD já criado |

**Lifecycle de destruição (Etapa 25):** os campos já estão no CRD, mas **não executam teardown** até implementação do controller completo. Implementar dois caminhos controlados de TerraformStack: solicitação explícita de `destroy` na spec e intenção `destroyResourcesOnDeletion` com default `false`, finalizer e aprovação humana verificável. Ambos usam plan de destruição imutável, proteção de dependências reversas e política de retenção/orphan; excluir CR ou YAML **não autoriza** automaticamente destroy. O CRD já inclui `sourceRef.version` opcional, `configuration.version` opcional, `suspend`, retry, timeout e health checks. O executor real ainda precisará implementar e testar esses comportamentos, sem manter compatibilidade com o contrato anterior. [Detalhamento](flows/terraformstack.md#32-evolução-do-contrato-sourceref-destroy-e-controles-operacionais).

**Marco da fase:** plataforma observável e governada, com drift controlado e provas de isolamento compatíveis com os ambientes **efetivamente provisionados**. Os testes avançados de PCI/PROD não autorizam automaticamente novas despesas AWS.

## Fase 7 — Experience Layer e Developer Experience

Formalizar os canais de consumo **sem colocar Backstage ou uma REST API adicional como pré-requisito**. A Kubernetes API continua sendo o contrato técnico das Capabilities e do TerraformStack; Git, templates e ferramentas de acesso são interfaces possíveis.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **28** | Personas, ownership e permissões | Platform Team, DevOps Federado e Application Team operam com caminhos Git, RBAC, IAM, AppProjects e políticas limitados ao seu escopo |
| **29** | Catálogo de contratos e Starter Kit | Configuração mínima de aplicação, template de CI, padrão de deployment e instruções de consumo das Capabilities acessíveis por Git/CLI; Backstage permanece opcional |

**Marco da fase:** a experiência padronizada está pronta para ser exercitada por cada persona, sem exigir que a squad mantenha addons corporativos ou módulos Terraform.

## Fase 8 — Golden Path ponta a ponta

Reutilizar a mesma aplicação de Rastreamento de Encomendas e a pipeline construída na Fase 5. Aqui o foco é **validar a jornada federada inteira**, não reimplementar ferramentas ou fazer outro projeto de aplicação.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **30** | Primeiro Domain e solicitações de Capabilities via GitOps | Registrar `Domain`, ownership e Environments permitidos; DevOps Federado solicita `ObjectStorage` e `Database` por PR e acompanha condições no Crossplane central |
| **31** | Entrega integrada da aplicação pelo Golden Path | Squad altera FastAPI, CI testa e publica imagem no ECR, PR atualiza GitOps, Argo CD local promove e aplica no Non-PCI; frontend registra e consulta recebimentos no Aurora, com evidência S3 e telemetria |

**Marco da fase:** fluxo completo demonstrado com responsabilidades separadas, contratos versionados, governança, infraestrutura de domínio, CI/CD, GitOps e observabilidade.

## Regras transversais da POC

- **Construção incremental:** cada etapa deve **criar/integrar**, executar uma verificação objetiva e preservar as capacidades necessárias às etapas posteriores. O guia não é um curso exaustivo de cada ferramenta.
- **Contratos e ownership:** `Entity`, `Environment` e `Domain` são registros lógicos em Git; Capabilities são XRDs/Compositions do Crossplane; `TerraformStack` é um CRD com controller próprio para Foundation/Entity Core. [ADRs](adr/README.md).
- **Um reconciliador por recurso:** Argo CD central mantém addons corporativos e entrega CRs de infraestrutura; Crossplane reconcilia Capabilities; TerraformStack governa Terraform; Argo CD local entrega Domain Applications.
- **CI separado de CD:** GitHub Actions testa, constrói, verifica e publica a imagem; Git registra o digest desejado; Argo CD local efetua o deploy. O workflow da aplicação pertence ao repositório da squad, não a este monorepo.
- **Segurança na etapa de origem:** IAM/STS, RBAC, validação de destino, state/plan protegido, custos e política de deleção são requisitos das operações correspondentes — não ficam pendentes até a Fase 6.
- **Arquitetura lógica versus implementação:** `kind` serve ao Hub inicial; EKS e integrações cross-account exigem testes reais de identidade/conectividade. UAT, PROD, PCI e Non-PCI são escopos distintos; não afirmar validação real de recursos não criados.
- **Custos e limpeza:** Aurora, EKS e outros serviços AWS têm custo; confirmar autorização e teardown antes de provisionar. Remover um CR `TerraformStack` não implica `destroy` automático.

## Evoluções opcionais

Avaliar posteriormente EKS Auto Mode (sem instalar Karpenter autogerenciado sobre seu autoscaling), EKS Argo CD Capability, portal Backstage, um contrato abstrato de aplicação e automação de criação de contas. Nenhum deles é necessário para concluir a POC V2.

[Arquitetura](architecture.md) · [GitOps](flows/gitops.md) · [TerraformStack](flows/terraformstack.md) · [GitOps Bridge](flows/gitops-bridge.md) · [Guia das etapas](../labs/README.md)

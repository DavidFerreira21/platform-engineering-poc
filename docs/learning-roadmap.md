# Roteiro prático da POC — Roadmap V2

[← Início](../README.md) · [Arquitetura](architecture.md) · [Guias das etapas](../labs/README.md) · [GitOps Flow](flows/gitops.md) · [TerraformStack Flow](flows/terraformstack.md) · [GitOps Bridge Flow](flows/gitops-bridge.md)

A POC é construída em **oito fases e 32 etapas (00–31)**. Cada etapa adiciona uma capacidade técnica, reutiliza as anteriores e termina com uma verificação objetiva. O objetivo dos guias é **construir, integrar e observar o funcionamento da plataforma**, sem tentar ensinar exaustivamente cada ferramenta.

Os laboratórios existentes das **Etapas 00–05 e 07** permanecem válidos. As demais entregas descrevem o escopo planejado; ainda **não foram implementadas** só por aparecerem aqui. Os guias mantêm a estrutura atual: **Objetivo → O que será feito → Por que fizemos assim → Pré-requisitos → Execução → O que observar → Limpeza (quando aplicável)**. [Convenção](../labs/README.md#como-ler-e-reproduzir-uma-etapa).

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

Construir o contrato declarativo e o motor de execução até **o primeiro apply AWS autorizado**, sem antecipar detecção ou remediação automática de drift.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **05** | [CRD Kubernetes `TerraformStack` via GitOps](../labs/05-terraformstack-crd.md) | API valida `spec` fictício e expõe `status.conditions`, sem executar Terraform. [ADR-0006](adr/0006-terraformrun-and-application-ci.md) |
| **06** | TerraformStack Controller + WorkflowTemplate e executor simulado | CR estável e versionado cria um Workflow; execução simulada e status são observados. Validar idempotência, cancelamento e autorização por testes do controller |
| **07** | [Registro e validação de UAT](../labs/07-registro-ambientes-aws.md) | `environment.yaml` associa `entity-a/uat` à conta/região; identidade humana verificada com Identity Center/STS |
| **08** | Identidade temporária do executor e backend de state | Role de execução, mapping confiável, state segregado, lock e rejeição de destino indevido verificados |
| **09** | Primeira stack real: rede UAT por TerraformStack | Argo Workflows produz plan, avalia o gate humano na UI, valida o plano imutável, aplica a VPC/rede autorizada e publica status/evidências |

**Marco da fase:** `TerraformStack` é capaz de receber intenção via GitOps, executar Terraform com aprovação humana e produzir infraestrutura real. Terraform não aplica continuamente a cada reconcile. **Auto Apply, verificação periódica e correção de drift entram na Fase 6.**

## Fase 3 — Provisionamento de EKS, GitOps e GitOps Bridge

Usar o TerraformStack já operacional para transformar recursos AWS em um **Environment UAT gerenciado**. O registro do cluster, a autorização e a configuração por GitOps são ações independentes. Antes de registrar EKS privado, comprovar a conectividade e a identidade que o Argo CD usará; não presumir comunicação entre VPCs/contas.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **10** | Management EKS UAT e bootstrap de gestão | TerraformStack provisiona Management EKS; acesso e registro no Argo CD central são validados; a plataforma instala o Argo CD local e a configuração-base autorizada |
| **11** | EKS Non-PCI e onboarding | TerraformStack cria o cluster Non-PCI; automação configura identidade, permissões, registro no Argo CD apropriado e labels `entity/environment/cluster-role/compliance` |
| **12** | Platform GitOps multicluster e perfis de add-ons | ApplicationSets do Argo CD central selecionam clusters registrados e instalam addons corporativos por perfil. Confirmar ownership único: central para addons, local para aplicações |
| **13** | GitOps Bridge — outputs, Parameter Store e ESO | Outputs permitidos da rede/EKS são publicados na **conta UAT** com escrita controlada; ESO no **Management EKS** lê o SSM e materializa um Secret consumido localmente. Publicação não registra clusters |
| **14** | Smoke test de Domain Application no Non-PCI | Argo CD local sincroniza um workload HTTP mínimo a partir de Git, verifica serviço/health e remove apenas o teste; os componentes de plataforma permanecem |

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
| **24** | Drift Detection do TerraformStack | Plan periódico detecta alteração externa em stack UAT existente, diferencia drift de nova intenção e registra condição/alerta **sem aplicar correção** |
| **25** | Reconciliação por políticas | Policy Engine versionado decide **Auto Apply / Require Approval / Deny** para mudança ou drift; testar correção permitida, gate humano e bloqueio de ação sensível |
| **26** | EKS PCI e isolamento PCI/Non-PCI em UAT | Provisionar EKS PCI **somente com autorização/custo aprovados**; quando existir, registrar com perfil restrito, aplicar addons autorizados e testar negações entre PCI e Non-PCI. Se não houver EKS PCI, declarar validação apenas lógica, não isolamento real |
| **27** | Registro de PROD e isolamento por Environment | Registrar conta PROD distinta, validar identity/mapping/negações e políticas UAT/PROD. Provisionamento real em PROD depende de autorização separada; não presumir EKS PROD já criado |

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

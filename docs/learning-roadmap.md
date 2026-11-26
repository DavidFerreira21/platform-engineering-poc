# Roteiro prático da POC — Roadmap V2

[← Início](../README.md) · [Arquitetura](architecture.md) · [Guias das etapas](../labs/README.md) · [GitOps](flows/gitops.md) · [TerraformStack](flows/terraformstack.md)

A POC é montada em **oito fases e 36 etapas (00–35)**. Cada etapa instala um bloco ou valida um comportamento sobre os blocos já instalados. Os diretórios de implementação seguem os componentes. A numeração organiza os guias e o roteiro. As etapas futuras descrevem entregas e evidências esperadas. [Convenção dos guias](../labs/README.md#como-ler-e-reproduzir-uma-etapa)

## Visão das oito fases

| Fase | Etapas | Marco técnico |
| --- | --- | --- |
| **1 — Base do Hub local** | 00–03 | Hub `kind`, Argo CD e Argo Workflows |
| **2 — Control Plane local** | 04–08 | TerraformStack e Crossplane validados com recursos Kubernetes locais |
| **3 — Hub EKS e spokes** | 09–18 | Hub EKS Auto Mode, rede e clusters UAT, GitOps e Bridge |
| **4 — Domain Infrastructure** | 19–21 | Capabilities S3 e Aurora PostgreSQL via Crossplane |
| **5 — Primeira aplicação e CI/CD** | 22–25 | FastAPI, Aurora, S3, GitHub Actions, ECR e deploy GitOps |
| **6 — Observabilidade e operação** | 26–31 | Telemetria, diagnóstico, drift, políticas e isolamento |
| **7 — Experience Layer** | 32–33 | Personas, permissões, catálogo e Starter Kit |
| **8 — Golden Path ponta a ponta** | 34–35 | Solicitação federada e entrega integrada da aplicação |

## Fase 1 — Base do Hub local

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **00** | [Documentação e monorepo](repository-structure.md) | Estrutura, camadas e ownership definidos |
| **01** | [Hub kind e Traefik](../labs/01-platform-management-cluster.md) | Cluster, namespaces, DNS, Ingress e auditoria local preparados |
| **02** | [Argo CD central e catálogo de add-ons](../labs/02-argocd-central.md) | ApplicationSet gera Applications e um ConfigMap de teste é sincronizado |
| **03** | [Argo Workflows](../labs/03-argo-workflows.md) | Workflow executado com Pod, resultado e logs observáveis |

## Fase 2 — Control Plane local

Concluir o ciclo TerraformStack antes do Crossplane. Cada reconciliador gerencia seu próprio recurso de teste no Kubernetes. Os aceites usam o Hub `kind` com dados fictícios.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **04** | [API do TerraformStack](../labs/04-terraformstack-crd.md) | CRD `v1alpha1` estabelecido e schema validado na API |
| **05** | [Controller e pipeline do TerraformStack](../labs/05-terraformstack-runtime.md) | Controller Go e adapter Argo, configuração e perfil local imutáveis, runner e WorkflowTemplate instalados via GitOps; duas réplicas saudáveis, Lease e RBAC verificados |
| **06** | [Terraform Stack Simulator](../labs/06-terraformstack-simulator.md) | GitOps, Terraform real, decisão OPA, aprovação, drift, retry, concorrência e destroy comprovados no ConfigMap e no status |
| **07** | [Crossplane Core e smoke test](../labs/07-crossplane-core.md) | Core disponível, Function saudável, XRD estabelecido e Composition local observável |
| **08** | Crossplane de ponta a ponta no `kind` | Argo CD sincroniza XR e o Crossplane cria, atualiza e remove um ConfigMap distinto do Simulator. Conferir schema, `Synced`, `Ready`, ownership, re-sync e permissões |

**Marco da fase.** TerraformStack e Crossplane reconciliam pelo GitOps com os mesmos blocos que serão reinstalados no Hub EKS. A identidade AWS do Provider Crossplane e as Capabilities reais entram nas Etapas 19–21.

## Fase 3 — Hub EKS, spokes e GitOps Bridge

Preparar a conta de plataforma e provisionar o Hub EKS Auto Mode por bootstrap externo. Reinstalar os blocos centrais e repetir as validações antes de preparar as spokes. Cada destino mantém identidade e state próprios. Aplicações e workloads de negócio entram nas fases seguintes.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **09** | Foundation da conta de plataforma e módulos AWS | Conferir acesso humano, conta, região, backend e rede de plataforma. Completar os módulos Configuration, Network e EKS usados pelas raízes reutilizáveis e validar os planos |
| **10** | Bootstrap externo e reinstalação do Hub EKS Auto Mode | Provisionar a rede e o Hub na conta de plataforma por execução autorizada com states separados. Publicar imagens por digest, validar acesso ao Git e aos registries, preparar armazenamento, criar EKS Capability for Argo CD com IAM Identity Center, registrar Hub por ARN, reinstalar os blocos do Git e repetir os dois aceites Kubernetes no EKS |
| **11** | Registro de UAT | Validar o registro de destino, conta, região, acesso humano por IAM Identity Center e STS |
| **12** | Identidade do executor e state | Provar a sessão temporária do Pod no Hub EKS, assunção da role UAT, bloqueio de destino indevido e operação do backend S3 com locking |
| **13** | Network e primeiro apply da spoke UAT | Resolver revisões imutáveis e aplicar o mesmo plan aprovado. Verificar e disponibilizar ID da VPC e sub-redes para EKS |
| **14** | EKS Management em UAT | Consumir outputs verificados de Network e exigir dependência `Ready`. Aplicar o perfil Management, testar acesso privado e iniciar o onboarding |
| **15** | EKS Non-PCI em UAT | Habilitar Non-PCI na mesma raiz EKS. Verificar dependências, IAM e outputs. PCI segue para a Etapa 30 |
| **16** | Platform GitOps multicluster e add-ons | Criar Access Entries, registrar ARNs e labels, comprovar sincronização de APIs privadas entre VPCs isoladas pela EKS Capability for Argo CD. ApplicationSets entregam add-ons por perfil; Argo CD local recebe destinos autorizados na própria VPC. [ADR-0008](adr/0008-managed-argocd-connectivity.md) |
| **17** | GitOps Bridge | Publicar outputs allowlisted no Parameter Store após verify e comprovar consumo via ESO no Management EKS |
| **18** | Smoke test HTTP do Non-PCI | Argo CD local sincroniza workload mínimo, verifica serviço e remove os recursos do teste |

### Etapa 11 — Registro de UAT

Usar o [modelo de Environment](../foundation/environment.example.yaml) e o registro de UAT da POC. Conferir conta e região com uma sessão temporária do operador. O registro fornece o destino confiável para as stacks da spoke.

### Etapa 12 — Identidade do executor e state

Validar a identidade de origem dos Pods do Hub EKS e a role autorizada de UAT usando STS. Preparar backend S3 e locks por destino e stack. Conferir autorização, região e custo antes das operações AWS. A etapa termina com identidade e backend verificados.

**Plan e apply da Etapa 13.** Usar as [raízes Terraform](../foundation/stacks/README.md) e a mesma pipeline do Simulator. A pipeline resolve fonte, destino, backend e identidade pelo cadastro autorizado. A revisão GitOps e os SHAs efetivos ligam o plano à configuração. O OPA encaminha um plano permitido à aprovação humana e bloqueia violações ou erro de avaliação. Os cenários de auto apply e remediação de drift na AWS entram na Etapa 29.

**Marco da fase.** O Hub EKS conduz o TerraformStack de UAT. Management e Non-PCI passam pelo GitOps, e o smoke test confirma a entrega antes da aplicação de negócio. A definição de perfil e os testes de isolamento governam o avanço de PCI e PROD.

## Fase 4 — Domain Infrastructure com Crossplane

Entregar as primeiras Capabilities AWS a partir de contratos Kubernetes declarativos. Os Providers reconciliam recursos no Environment autorizado, sem Terraform Workflow para cada Capability.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **19** | Provider AWS e identidade Crossplane no Hub | Provider assume role temporária em UAT; conferir acesso correto, renovação e rejeição de destino não autorizado |
| **20** | `ObjectStorage` via Crossplane | XRD/Composition solicitam um bucket S3 na conta UAT; verificar condições, permissões e política de remoção |
| **21** | `Database` via Crossplane | XRD/Composition criam Aurora PostgreSQL em sub-redes autorizadas; conferir prontidão, acesso, custo e teardown planejado |

**Marco da fase:** `ObjectStorage` e `Database` são capacidades reais e disponíveis para a aplicação da Fase 5. O **Domain Infrastructure** é gerenciado pelo DevOps Federado no Golden Path; a primeira implantação técnica das Capabilities é conduzida pelo Platform Team.

## Fase 5 — Primeira aplicação e esteira CI/CD

Construir uma aplicação **mínima e funcional** usando a infraestrutura pronta. O cenário é **Rastreamento de Encomendas**: uma tela permite registrar o recebimento e consultar os últimos recebimentos. O código, os testes, a configuração de deployment e a pipeline CI ficam em **repositório separado, sob ownership da squad**, conforme o [ADR-0007](adr/0007-personas-and-developer-experience.md).

**Stack escolhida:** **Python + FastAPI** para backend HTTP/API e **Jinja2 + HTML/CSS** para uma interface renderizada pelo servidor. Não usar Django, React ou um segundo backend nesta POC. O banco será **Aurora PostgreSQL** da Capability `Database`; a API gravará um **JSON de evidência de recebimento** no S3 da Capability `ObjectStorage`. O S3 não substitui logs operacionais.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **22** | Aplicação de referência: FastAPI + Jinja2 | Frontend simples registra código da encomenda e lista recebimentos; API e testes locais funcionando |
| **23** | Integração da aplicação com Capabilities | API persiste código e timestamp no Aurora, grava evidência JSON no S3 e valida leituras, IAM mínimo e conectividade; sem credenciais fixas no código |
| **24** | CI da aplicação: GitHub Actions e Amazon ECR | Pipeline executa testes, build do contêiner, scan de segurança e push ao ECR com OIDC/credenciais temporárias; imagem identificada por digest |
| **25** | CD GitOps: promoção e deploy no EKS Non-PCI | CI propõe atualização de digest no manifesto/overlay versionado; após revisão, Argo CD local sincroniza a nova imagem; confirmar rollout e funcionamento da tela |

**Responsabilidades:** Platform Team oferece ECR, workflow/template reutilizável, RBAC e guardrails; a squad é owner do código, da versão e do release da aplicação; DevOps Federado é owner das solicitações de Capabilities. A pipeline **não faz `kubectl apply` imperativo**. Manifests e código permanecem inicialmente no mesmo repositório da aplicação, para reduzir complexidade. O deploy pode usar Helm ou Kustomize com defaults mantidos pela plataforma; não criar agora um CRD `Application` próprio.

**Decisão técnica ainda aberta para a Etapa 24:** escolher se o repositório ECR ficará na conta de plataforma ou na conta do Environment. O critério é permitir publicação autorizada por OIDC e pull da mesma imagem pelo EKS Non-PCI; se for cross-account, validar IAM e repository policy antes do deploy. Não decidir isso implicitamente pelo desenho do fluxo.

**Marco da fase:** uma alteração de código gera uma imagem versionada no ECR e chega ao Non-PCI via GitOps, usando Aurora e S3 reais. O objetivo aqui é implementar e validar a entrega; **a experiência self-service das três personas será demonstrada na Fase 8**, sem criar outra pipeline.

## Fase 6 — Observabilidade, operação e governança

Instrumentar os componentes implantados e concluir o lifecycle operacional do TerraformStack. Políticas preventivas essenciais já devem existir desde a introdução de cada componente; esta fase consolida diagnósticos e cenários mais avançados.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **26** | Telemetria de plataforma e aplicação | OpenTelemetry Collector/gateway no Data Plane e integração com Grafana, Loki e Tempo; logs/traces correlacionados da API, dos Workflows e dos controllers |
| **27** | Diagnóstico, recuperação e custos | Investigar uma falha controlada, recuperar a operação e conferir rastreabilidade e custos AWS |
| **28** | Drift Detection do TerraformStack | Validar na AWS `driftCheckInterval` configurável, agendamento idempotente com jitter e verificação periódica; detectar alteração externa em stack UAT, registrar última/próxima checagem e condição/alerta **sem aplicar correção** |
| **29** | Reconciliação, destroy e políticas | Ampliar as regras OPA da Etapa 13 e comprovar **Auto Apply / Require Approval / Deny** com três planos de teste e evidências da etapa alcançada; testar drift, `plan -destroy` explícito, destroy on deletion com finalizer, proteção de dependências, aprovação humana e bloqueio de ações sensíveis |
| **30** | EKS PCI e isolamento PCI/Non-PCI em UAT | Reutilizar a **stack EKS** e habilitar PCI somente com autorização/custo aprovados; registrar com perfil restrito, aplicar addons autorizados e testar negações entre PCI e Non-PCI. Sem PCI real, declarar somente validação lógica, não isolamento operacional |
| **31** | Registro de PROD e isolamento por Environment | Registrar conta PROD distinta, validar identity/mapping/negações e políticas UAT/PROD. Provisionamento real em PROD depende de autorização separada; não presumir EKS PROD já criado |

**Lifecycle de destruição na Etapa 29.** Validar na AWS o destroy explícito e o destroy na exclusão com finalizer. O Simulator já exercita os dois caminhos com plano salvo, proteção de dependências e aprovação humana. A exclusão padrão retém o recurso. A validação AWS inclui a política de retenção/orphan e usa o controller e a pipeline aprovados no Hub. [Dependências e destruição](flows/terraformstack.md#5-dependências-e-destruição)

### Etapa 31 — Isolamento por Environment

Conferir a identidade e o mapeamento de PROD em uma conta distinta. Usar o registro declarativo e testar negações entre UAT e PROD antes de qualquer oferta de provisionamento em PROD.

**Marco da fase:** plataforma observável e governada, com drift controlado e provas de isolamento compatíveis com os ambientes **efetivamente provisionados**. Os testes avançados de PCI/PROD não autorizam automaticamente novas despesas AWS.

## Fase 7 — Experience Layer e Developer Experience

Formalizar os canais de consumo **sem colocar Backstage ou uma REST API adicional como pré-requisito**. A Kubernetes API continua sendo o contrato técnico das Capabilities e do TerraformStack; Git, templates e ferramentas de acesso são interfaces possíveis.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **32** | Personas, ownership e permissões | Platform Team, DevOps Federado e Application Team operam com caminhos Git, RBAC, IAM, AppProjects e políticas limitados ao seu escopo |
| **33** | Catálogo de contratos e Starter Kit | Configuração mínima de aplicação, template de CI, padrão de deployment e instruções de consumo das Capabilities acessíveis por Git/CLI; Backstage permanece opcional |

**Marco da fase:** a experiência padronizada está pronta para ser exercitada por cada persona, sem exigir que a squad mantenha addons corporativos ou módulos Terraform.

## Fase 8 — Golden Path ponta a ponta

Reutilizar a mesma aplicação de Rastreamento de Encomendas e a pipeline construída na Fase 5. Aqui o foco é **validar a jornada federada inteira**, não reimplementar ferramentas ou fazer outro projeto de aplicação.

| Etapa | Entrega | Evidência |
| :--- | :--- | :--- |
| **34** | Primeiro Domain e solicitações de Capabilities via GitOps | Registrar `Domain`, ownership e Environments permitidos; DevOps Federado solicita `ObjectStorage` e `Database` por PR e acompanha condições no Crossplane central |
| **35** | Entrega integrada da aplicação pelo Golden Path | Squad altera FastAPI, CI testa e publica imagem no ECR, PR atualiza GitOps, Argo CD local promove e aplica no Non-PCI; frontend registra e consulta recebimentos no Aurora, com evidência S3 e telemetria |

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

EKS Auto Mode e EKS Capability for Argo CD integram a arquitetura AWS nas Etapas 09–16; validar custo e disponibilidade. Avaliar posteriormente portal Backstage, um contrato abstrato de aplicação e automação de criação de contas.

[Arquitetura](architecture.md) · [GitOps](flows/gitops.md) · [TerraformStack](flows/terraformstack.md) · [GitOps Bridge](flows/gitops-bridge.md) · [Guia das etapas](../labs/README.md)

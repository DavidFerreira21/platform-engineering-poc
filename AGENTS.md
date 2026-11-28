# AGENTS.md — Instruções de implementação

## Fonte da verdade e idioma

Consultar [arquitetura](docs/architecture.md), [decisões](docs/adr/README.md), [estrutura](docs/repository-structure.md) e [roadmap](docs/learning-roadmap.md) antes de implementar. Documentar em **português brasileiro**, preservando nomes técnicos. Não reintroduzir modelos superados do histórico Git.

## Estilo da documentação

Escrever os documentos **na voz do próprio projeto**, em português brasileiro e tom técnico direto. Preferir construções como "nesta POC", "a arquitetura adota", "o repositório mantém" e "vamos validar" quando a primeira pessoa fizer sentido. **Não narrar a criação do projeto como um observador externo** nem atribuir decisões a interlocutores, conversas ou ferramentas de redação. Evitar comentários de bastidores e manter justificativas arquiteturais, referências técnicas e avisos de risco.

## Personas e níveis de abstração (ADR-0007)

- **Platform Team:** mantém os motores, contratos, Golden Paths, Foundation e Entity Core compartilhado; provê governança.
- **Federated DevOps Team:** ponte técnica entre a plataforma e as squads; consome contratos especializados para **Domain Infrastructure** (Database, ObjectStorage, Messaging), faz upgrades/alterações do Domain e pode solicitar evolução de Capability. Não administra EKS compartilhado nem implementa Controllers da plataforma por padrão.
- **Application Team / Developers:** mantém **código, testes, GitHub Actions, imagens e manifests Kubernetes no repositório da aplicação separado**; usa Application Capabilities e Golden Paths/Starter Kits para não precisar aprender Crossplane/Terraform/Argo internamente.
- **Infrastructure Capability** é produto técnico (XRD Crossplane). **Application Capability** é padrão reutilizável de app/runtime/CI/deploy/observabilidade, **não obrigatoriamente CRD/XRD**. Golden Path combina padrões, templates e automações.
- **Aplicação de demonstração:** FastAPI + Jinja2 + HTML/CSS, com tela simples de recebimento de encomendas, Aurora PostgreSQL e evidência JSON no S3 via Capabilities Crossplane. GitHub Actions → ECR → GitOps → Argo CD local. Construção e CI/CD nas Etapas 22–25; Golden Path nas 34–35. Evitar software de negócio complexo, portal ou gerador completo. Não provisionar Capabilities automaticamente sem o fluxo autorizado do DevOps Federado.
- **Monorepo de lab:** manter manifests de `Domain Infrastructure` por Domain/Environment sob DevOps Federado; **não adicionar** pastas `applications/` do Domain como destino obrigatório, pois esses manifests pertencem à aplicação separada. Revisão por path, IAM, RBAC, Admission e AppProject devem comprovar os limites de acesso.

## Modelo vigente

- **Experience Layer:** Backstage, CLI, GitOps e AI/MCP são canais paralelos. Alterações persistentes preferencialmente via PR.
- **Control Plane:** Kubernetes API Server central, Crossplane Core + XRDs/Compositions/Providers para **Capabilities**; começar `ObjectStorage`, depois `Database` com **Aurora PostgreSQL** em UAT. Crossplane único central, sem cascata.
- **Entity / Environment / Domain:** conceitos lógicos e **registros declarativos no Git**. Entity organiza; Environment mapeia UAT/PROD à conta AWS/região; Domain é unidade de negócio, consumo, políticas e ownership de aplicações, namespaces e Capabilities. **Não criar CRD/XRD** para esses três na primeira etapa.
- **Workflow Plane:** Argo CD central aplica CRs de Capability na Kubernetes API; Argo Workflows realiza Foundation e EKS por integrações cadastradas; o Terraform Executor e o handshake são usados na implementação de referência. Argo CD não executa Terraform. **Não implementar Environment Controller próprio como dependência inicial**.
- **GitOps AWS:** adotar EKS Capability for Argo CD como Argo CD central no Hub EKS; manter Argo CD autogerenciado em kind e Argo CD local nos Management EKS. VPCs Platform/UAT/PROD isoladas com NAT, sem Peering/TGW. Registrar Hub e spokes por ARN, Access Entries e RBAC; não presumir conectividade de Pods. [ADR-0008](docs/adr/0008-managed-argocd-connectivity.md).
- **Observability Plane:** Managed Grafana e CloudWatch do Hub na conta da Plataforma. Cada EKS envia métricas direto ao AMP do ambiente. Logs e traces seguem pelo Gateway para Loki/Tempo no Management EKS, com S3 do ambiente. Validar ingestão e consulta antes de habilitar PCI. [ADR-0010](docs/adr/0010-observability-topology.md).
- **Data Plane:** `data-plane/<entity>/core-infrastructure/{uat,prod}/` organiza recursos comuns; `data-plane/<entity>/domains/<domain>/` organiza uma identidade única e recursos/aplicações por ambiente. Ambos pertencem ao Data Plane; **Environment não equivale a Core**.
- **Applications:** Argo CD local do Management EKS aplica namespaces, policies e workloads em PCI/Non-PCI **somente se o isolamento permitir**.
- **Topologia-alvo:** UAT e PROD são contas AWS separadas; cada uma inclui Management EKS, EKS PCI, EKS Non-PCI e AWS Runtime. EKS Auto Mode preferido, sob validação de custo/operadores/PCI. O roteiro começa no Hub EKS Auto Mode. O lab `kind` preserva as evidências locais de TerraformStack e Crossplane. Sem MSK na primeira versão.

- **Policy Engine:** módulos Rego reutilizáveis e perfis YAML em `workflow-plane/policy-engine/`; perfil selecionado pela pipeline confiável. Regras AWS futuras não devem ser ativadas sem implementação e testes. [Policy Engine](workflow-plane/policy-engine/README.md).

## Ownership e contratos

**Foundation** é central (conta, landing zone, rede, IAM e Terraform Bootstrap). **Entity Core** é compartilhado, federado pode solicitar mas lifecycle de EKS é governado. **Domain Infrastructure** é federada e usa Crossplane para Capabilities AWS. **Domain Applications** é responsabilidade das squads: CI/CD em repositório de aplicação de teste separado (GitHub Actions) e deploy GitOps por Argo CD local.

Os YAMLs `domain.yaml` e `environment.yaml` são registros lógicos no Git. Validar o destino e a identidade autorizados antes de provisionar uma Capability e conferir o Domain para solicitações federadas. A plataforma resolve a role ou o ProviderConfig a partir do destino autorizado. Cada recurso externo tem um único reconciliador. Terraform cuida de Foundation e EKS. Crossplane cuida das Capabilities delegadas. Não criar `CapabilityClaim` genérico.

## Complemento vigente de arquitetura (21/11/2026)

- **Ownership padrão:** Foundation → Entity Core Infrastructure → Domain Infrastructure → **Domain Applications**.
- **TerraformStack.** O CRD da Etapa 04 declara `sourceRef.repository/path/version` (default `latest`), `context` opcional, `workflowRef.name` opcional, `dependsOn`, `suspend`, `reconciliation` e `lifecycle`. O Argo CD central sincroniza a intenção. O controller da Etapa 05 agenda uma tentativa ativa por CR, via adapter Argo, sem interpretar contexto, Git, backend ou resultado de OPA. Integrações são cadastradas com aliases e default; `contract.authorization` tem default `true` e pode ser `false` por integração compatível. Sem handshake, o controller não infere apply nem cancela automaticamente a execução ativa; submissão ambígua ou Workflow perdido bloqueiam replay. A pipeline de referência da Etapa 06 resolve Git e destino, fixa o vínculo do state, aplica o plano autorizado após claim atômico e publica o resultado tipado. Preserve idempotência, plan imutável, isolamento e destroy governado. O primeiro apply aprovado na AWS entra na Etapa 13; drift e políticas, nas Etapas 28 e 29. Cada tentativa é um Argo Workflow; não criar outro CRD de execução. Logs JSON, Events e métricas de saúde por CR são responsabilidade passiva do controller em `internal/observability`, sem dependência de backends externos ou alteração do contrato. Não registrar segredos, tokens, planos ou state em telemetria. [Execution Contract](docs/flows/terraformstack-execution-contract.md)
- **CI/CD de aplicação:** GitHub Actions **e os manifests Kubernetes da aplicação** deverão ficar junto ao código **no repositório de aplicação separado, previsto para etapa posterior**. Este monorepo mantém apenas manifests da **Domain Infrastructure**, configurações de plataforma e validações das etapas. Argo CD **local** entrega workloads via GitOps.
- **Management EKS:** Argo CD local, Operators a escolher, OTel Collector para os sinais do próprio cluster e OTel Gateway para logs e traces do Environment. O Collector coleta métricas do Argo CD local e as envia ao AMP. [Fluxo de observabilidade](docs/flows/observability.md).
- Este projeto é uma POC pessoal e **não deve usar nomes/recursos corporativos**. Documentação de arquitetura não significa execução implantada.

## Etapas de personas e Golden Path

- **Etapa 32:** validar os três owners; testar autorização de PRs por path/repo, IAM/RBAC, AppProjects e restrições de Environment.
- **Etapa 18:** smoke test HTTP mínimo no EKS Non-PCI, após o Argo CD local. A aplicação FastAPI de negócio e a pipeline CI/CD entram nas Etapas 22–25, em repositório separado da squad.
- **Etapas 32–33:** Experience Layer após Data Plane, observabilidade, segurança/governança e workflows; contratos, catálogo e Starter Kit via Git/CLI. Backstage opcional.
- **Etapas 34–35:** Golden Path final, reutilizando a aplicação e CI/CD já construídas para demonstrar solicitação de Capabilities, GitHub Actions, Argo CD local e telemetria.

## Sequência inicial e autenticação

1. Preparar a base do Hub EKS nas Etapas 01–03. Confirmar o primeiro workload e o primeiro sync com a EKS Capability for Argo CD antes de instalar os blocos seguintes.
2. Adequar os guias locais ao Hub EKS. Instalar o CRD e a pipeline canônica do TerraformStack nas Etapas 04–05. Concluir o aceite com o Simulator na Etapa 06.
3. Completar o fluxo Crossplane com recursos Kubernetes de teste nas Etapas 07–08. Preparar a Foundation para as spokes e consolidar as evidências do Hub EKS nas Etapas 09–10.
4. Validar o registro UAT e o acesso humano por IAM Identity Center na Etapa 11. Validar identidade temporária do executor e backend segregado na Etapa 12. Aplicar Network com aprovação explícita na Etapa 13. Fazer Management e Non-PCI nas Etapas 14–15, onboarding, add-ons e Bridge nas 16–17 e smoke test na 18.
5. Validar a identidade dedicada do Provider AWS no Hub EKS na Etapa 19. Validar `ObjectStorage` e Aurora nas Etapas 20–21. Validar PROD na Etapa 31 e introduzir o Domain na Etapa 34.

## Monorepo, fases e etapas

O arquivo [`docs/arquitetura Plataform.png`](docs/arquitetura%20Plataform.png) é o **diagrama de referência do projeto**, exibido no README e na arquitetura detalhada. **Manter essa imagem como referência; não criar desenhos alternativos sem uma decisão de atualização.** Se um ADR alterar personas, ownership, TerraformStack, topologia ou GitOps, verificar a consistência do diagrama e dos documentos, sem declarar componentes propostos como implantados. Não reintroduzir `TerraformRun` como contrato declarativo principal.

O monorepo organiza a implementação por bloco e usa a numeração nos guias e no roteiro. Reutilizar os manifests e a pipeline canônicos na montagem cumulativa. Criar pastas Domain/Foundation somente com conteúdo real das etapas. Manter navegação e links válidos e preservar as aulas em [fundamentos](docs/fundamentos/README.md). O [roteiro](docs/learning-roadmap.md) mantém a numeração contínua das etapas.

- Escrever cada guia como registro reproduzível de uma decisão da POC. Começar com **Objetivo** e **O que será feito**, explicar a decisão em **Por que fizemos assim** e usar **O que observar** para ligar os sinais do sistema ao resultado. O roteiro principal segue a construção e a verificação do comportamento esperado. Incluir diagnóstico junto aos pontos que podem precisar de investigação e limpar os recursos do exercício quando aplicável. [Convenção dos guias](labs/README.md)
- Organizar as `Applications` do Hub por componente com ciclo de sincronização próprio. Uma etapa pode reutilizar uma `Application` compartilhada. Manter sync manual nos add-ons das etapas locais e indicar o alvo, a revisão e o diff esperado antes de sincronizar. Para uma etapa de observação, usar consultas, eventos e logs. [ADR-0003](docs/adr/0003-reconciliation-and-gitops.md)
- Ao revisitar uma etapa, verificar as dependências e os sinais usados pelo exercício. Limpar apenas seus recursos. Quando uma mudança compartilhada alterar uma dessas dependências ou evidências, atualizar o guia afetado no mesmo PR. Preferir Markdown e Mermaid nativos. CODEOWNERS sem rulesets não é barreira de acesso.

## Segurança

Cada alteração AWS exige `sts get-caller-identity`, conta/região conferidas, credenciais temporárias, IAM mínimo, custo e autorização por operação. Terraform state fora do Git e segregado. Nunca commitar segredos, dados corporativos, certificados ou kubeconfigs. Os IDs reais das contas UAT e PROD desta POC podem ser versionados exclusivamente no campo `aws.accountId` dos dois registros `environment.yaml` sob `foundation/config/entity-a/`. Outros identificadores AWS reais ficam fora do Git. Destruição de conta nunca é efeito de remover YAML; aprovar docs não autoriza `apply`/`destroy`. Validar isolamento entre Domain, UAT/PROD e PCI. Trabalhar via branch e PR, **sem merge automático**.

- **Contrato de integração TerraformStack:** parametrizar WorkflowTemplate, ServiceAccount e nomes dos inputs/outputs no `controller-config.yaml`. Definir política e aprovação no WorkflowTemplate e nos scripts da pipeline. Permitir múltiplas integrações cadastradas pelo Platform Team, com seleção opcional de alias via `spec.workflowRef.name`. Não permitir referência arbitrária a templates ou ServiceAccounts. Manter autorização operacional e claim habilitados por padrão; desabilitação explícita por integração exige pipeline compatível e preservação da execução ativa até o término, com bloqueio em caso de submissão ambígua. Manter as garantias de lifecycle, identidade e binding. A escolha de aliases com diferentes permissões exige controles de admissão/GitOps. Manter a reconciliação separada do adapter Argo. Argo Workflows é a única integração suportada. WorkflowTemplate, runner e OPA são a implementação de referência confiável da POC. [ADR-0009](docs/adr/0009-terraformstack-runtime.md). Alterações de contrato exigem drenagem e revisão coordenada de RBAC/admissão.

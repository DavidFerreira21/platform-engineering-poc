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
- **Aplicação de demonstração:** preferir **Python + Flask mínimo** (serviço HTTP simples + health check), apenas para exercitar plataforma: GitHub Actions, imagem, manifests K8s, Argo CD local, OpenTelemetry e policies. **Java foi somente exemplo**. Golden Paths/Starter Kits são conceitos de DX, **não construir gerador completo ou aplicação de negócio** neste laboratório. Backstage, CLI e Git permanecem canais possíveis. **Não provisionar Database/S3 automaticamente sem DevOps Federado**.
- **Monorepo de lab:** manter manifests de `Domain Infrastructure` por Domain/Environment sob DevOps Federado; **não adicionar** pastas `applications/` do Domain como destino obrigatório, pois esses manifests pertencem à aplicação separada. Revisão por path, IAM, RBAC, Admission e AppProject devem comprovar os limites de acesso.

## Modelo vigente

- **Experience Layer:** Backstage, CLI, GitOps e AI/MCP são canais paralelos. Alterações persistentes preferencialmente via PR.
- **Control Plane:** Kubernetes API Server central, Crossplane Core + XRDs/Compositions/Providers para **Capabilities**; começar `ObjectStorage`, depois `Database` com **Aurora PostgreSQL** em UAT. Crossplane único central, sem cascata.
- **Entity / Environment / Domain:** conceitos lógicos e **registros declarativos no Git**. Entity organiza; Environment mapeia UAT/PROD à conta AWS/região; Domain é unidade de negócio, consumo, políticas e ownership de aplicações, namespaces e Capabilities. **Não criar CRD/XRD** para esses três na primeira etapa.
- **Workflow Plane:** Argo CD central aplica CRs de Capability na Kubernetes API; Argo Workflows e Terraform Executor realizam Foundation e EKS via execução autorizada. Argo CD não executa Terraform. **Não implementar Environment Controller próprio como dependência inicial**.
- **Observability Plane:** métricas, logs, traces, eventos e auditoria.
- **Data Plane:** `data-plane/<entity>/core-infrastructure/{uat,prod}/` organiza recursos comuns; `data-plane/<entity>/domains/<domain>/` organiza uma identidade única e recursos/aplicações por ambiente. Ambos pertencem ao Data Plane; **Environment não equivale a Core**.
- **Applications:** Argo CD local do Management EKS aplica namespaces, policies e workloads em PCI/Non-PCI **somente se o isolamento permitir**.
- **Topologia-alvo:** UAT e PROD são contas AWS separadas; cada uma inclui Management EKS, EKS PCI, EKS Non-PCI e AWS Runtime. EKS Auto Mode preferido, sob validação de custo/operadores/PCI. Hub é `kind` inicialmente. Sem MSK na primeira versão.

## Ownership e contratos

**Foundation** é central (conta, landing zone, rede, IAM e Terraform Bootstrap). **Entity Core** é compartilhado, federado pode solicitar mas lifecycle de EKS é governado. **Domain Infrastructure** é federada e usa Crossplane para Capabilities AWS. **Domain Applications** é responsabilidade das squads: CI/CD em repositório de aplicação de teste separado (GitHub Actions) e deploy GitOps por Argo CD local.

Os YAMLs `domain.yaml` e `environment.yaml` **não são CRs**; validar e resolver destino autorizado antes de provisionar Capability. Nunca permitir que consumidor selecione role/ProviderConfig de outra conta livremente. Um recurso externo tem um único owner: Terraform para Foundation/EKS, Crossplane para Capabilities delegadas. Não criar `CapabilityClaim` genérico.

## Complemento vigente de arquitetura (21/11/2026)

- **Ownership padrão:** Foundation → Entity Core Infrastructure → Domain Infrastructure → **Domain Applications**.
- **TerraformStack CRD + Controller:** **propostos**, representam a infraestrutura desejada em Foundation/Entity Core. Argo CD central sincroniza o CR Kubernetes; controller coordena plans iniciais/periódicos via Argo Workflows/Terraform Executor, sinaliza drift em `status.conditions` e consulta Policy Engine (auto apply / aprovação / bloqueio). Não é XRD Crossplane, não executa `apply` em todo reconcile e deve manter idempotência, plan aprovado imutável e state/identidades segregados. `TerraformRun` é somente nome opcional para execução individual, sem segundo CRD obrigatório. Etapa **07** para o CRD e Etapas **10 e 11** para controller, execução e políticas.
- **CI/CD de aplicação:** GitHub Actions **e os manifests Kubernetes da aplicação** deverão ficar junto ao código **no repositório de aplicação separado, previsto para etapa posterior**. Este monorepo mantém apenas manifests da **Domain Infrastructure**, configurações de plataforma e validações das etapas. Argo CD **local** entrega workloads via GitOps.
- **Management EKS:** Argo CD local, espaço para **Operators ainda não escolhidos**, e **OpenTelemetry Collector/gateway** de telemetria para backend central **Grafana, Loki e Tempo**. Prometheus/métricas e segregação PCI serão validados.
- Este projeto é uma POC pessoal e **não deve usar nomes/recursos corporativos**. Documentação de arquitetura não significa execução implantada.

## Etapas de personas e Golden Path

- **Etapa 21:** simular os três owners; testar autorização de PRs por path/repo, IAM/RBAC, AppProjects e restrições de Environment.
- **Etapa 17:** **smoke test técnico**, após EKS e Argo CD local, com app Python + Flask mínima (preferencial), manifestos e health check; sem exigência de portal, CI/CD sofisticado ou Golden Path. Integrar o repositório separado na etapa correspondente do roteiro, mediante decisão explícita de implementação.
- **Etapa 22:** **Developer Experience** depois de Data Plane, observabilidade, segurança/governança e workflows estarem validados; experimentar contratos, catálogo e canais Backstage/CLI/Git.
- **Etapa 24:** **último passo**, reutilizar a mesma app para demonstrar Golden Path e integração da plataforma de ponta a ponta, GitHub Actions, Argo CD local, Capabilities e telemetria. Não transformar em desenvolvimento de aplicação nem construir gerador completo.

## Sequência inicial e autenticação

1. Preparar o Hub `kind` e instalar as ferramentas nas Etapas 01 a 04. A Composition de teste da Etapa 04 roda no Hub local.
2. Registrar **UAT e PROD** na Etapa 05 em arquivos `environment.yaml` separados na Foundation, usando o modelo em `foundation/environment.example.yaml`. O operador usa IAM Identity Center.
3. Definir os contratos e validar as identidades de workload na sequência das Etapas 06 a 09. Validar a autenticação do Crossplane em `kind` sem presumir IRSA/Pod Identity ou usar credenciais permanentes.
4. Preparar o fluxo Terraform nas Etapas 10 e 11, aplicar a rede na 12 e validar `ObjectStorage` e **Aurora PostgreSQL** nas Etapas 13 e 14 com orçamento, IAM, rede e remoção aprovados.

## Monorepo, fases e etapas

O arquivo [`docs/arquitetura Plataform.png`](docs/arquitetura%20Plataform.png) é o **diagrama de referência do projeto**, exibido no README e na arquitetura detalhada. **Manter essa imagem como referência; não criar desenhos alternativos sem uma decisão de atualização.** Se um ADR alterar personas, ownership, TerraformStack, topologia ou GitOps, verificar a consistência do diagrama e dos documentos, sem declarar componentes propostos como implantados. Não reintroduzir `TerraformRun` como contrato declarativo principal.

Um único monorepo simula limites. Criar pastas Domain/Foundation somente com conteúdo real das etapas. Manter navegação e links válidos; preservar as 24 aulas em [fundamentos](docs/fundamentos/README.md). O [roteiro](docs/learning-roadmap.md) agrupa o trabalho em fases, com uma numeração contínua para as etapas. Cada guia começa com **Objetivo** e **O que será feito**. Declare o resultado esperado, o que será criado ou alterado e o que será validado. Em seguida, inclua pré-requisitos, passos e comandos, evidências, falha controlada, diagnóstico, recuperação e limpeza. Preferir Markdown/Mermaid nativo legível. CODEOWNERS sem rulesets não é barreira de acesso.

## Segurança

Cada alteração AWS exige `sts get-caller-identity`, conta/região conferidas, credenciais temporárias, IAM mínimo, custo e autorização por operação. Terraform state fora do Git e segregado. Nunca commitar segredos, dados corporativos, certificados ou kubeconfigs. Os IDs reais das contas UAT e PROD desta POC podem ser versionados exclusivamente no campo `aws.accountId` dos dois registros `environment.yaml` sob `foundation/config/entity-a/`. Outros identificadores AWS reais ficam fora do Git. Destruição de conta nunca é efeito de remover YAML; aprovar docs não autoriza `apply`/`destroy`. Validar isolamento entre Domain, UAT/PROD e PCI. Trabalhar via branch e PR, **sem merge automático**.

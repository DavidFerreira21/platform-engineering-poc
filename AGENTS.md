# AGENTS.md — Instruções de implementação

## Fonte da verdade e idioma

Consultar [arquitetura](docs/architecture.md), [decisões](docs/adr/README.md), [estrutura](docs/repository-structure.md) e [roadmap](docs/learning-roadmap.md) antes de implementar. Documentar em **português brasileiro**, preservando nomes técnicos. Não reintroduzir modelos superados do histórico Git.

## Modelo vigente

- **Experience Layer:** Backstage, CLI, GitOps e AI/MCP são canais paralelos. Alterações persistentes preferencialmente via PR.
- **Control Plane:** Kubernetes API Server central, Crossplane Core + XRDs/Compositions/Providers para **Capabilities**; começar `ObjectStorage`, depois `Database` com **Aurora PostgreSQL** em UAT. Crossplane único central, sem cascata.
- **Domain Bootstrap (hipótese anterior):** proposta de fluxo padronizado de criação de Domain por perfil versionado. **Está em reavaliação e sua implementação foi adiada**; não criar gerador/controller ou torná-lo pré-requisito até nova decisão. Perfil completo de referência exige Database, ObjectStorage, namespace e quotas/policies. Geração idempotente de manifests no Git + PR, Crossplane central aplica recursos AWS e Argo CD local os objetos Kubernetes. Validar `Ready` a partir das dependências; sem banco ainda implantado, não declarar perfil completo pronto. Perfil novo não altera os Domains existentes sem migração deliberada.
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
- **TerraformStack CRD + Controller:** **propostos**, representam infraestrutura desejada de Foundation/Entity Core na Kubernetes API, sincronizada pelo Argo CD central. O controller executa `plan` via Argo Workflows/Terraform Executor para **criação e drift periódico**, avalia **Policy Engine** com decisões auto apply / aprovação / bloqueio e atualiza status. `Synced` no Argo CD não significa ausência de drift AWS. Não é XRD Crossplane e não executa `apply` a cada reconcile. `TerraformRun` é nome reservado opcional de execução individual, sem CRD obrigatório. A aprovação deve corresponder ao plano imutável; state e destino são controlados pela plataforma.
- **CI/CD de aplicação:** GitHub Actions deverá operar em **repositório de aplicação de teste separado, ainda não fornecido**; este monorepo mantém apenas documentação/contratos e validações dos labs. Argo CD **local** entrega workloads via GitOps.
- **Management EKS:** Argo CD local, espaço para **Operators ainda não escolhidos**, e **OpenTelemetry Collector/gateway** de telemetria para backend central **Grafana, Loki e Tempo**. Prometheus/métricas e segregação PCI serão validados.
- **Domain Bootstrap:** somente documentado e **adiado/em reavaliação**; não implementá-lo agora.
- Este projeto é uma POC pessoal e **não deve usar nomes/recursos corporativos**. Documentação de arquitetura não significa execução implantada.

## Evolução didática do TerraformStack

1. **08B-A — Declarative Execution:** manifesto Git → Argo CD central → CR TerraformStack → controller → plan/aprovação/apply, inicialmente com executor simulado.
2. **08B-B — Drift Detection:** plans periódicos, status e diagnóstico, **sem auto apply** de drift.
3. **08B-C — Policy-Based Reconciliation:** políticas versionadas para auto apply, aprovação ou bloqueio; diferenciar drift externo de mudança pedida no Git, verificar atributos/criticidade e usar plano aprovado/revalidado.

As três etapas são **arquitetura-alvo**, não funcionalidades existentes. Nenhuma mudança documental autoriza apply/destroy AWS.

## Sequência inicial e autenticação

1. Lab 01 cria **somente** Platform Management Cluster local com `kind`; não depender de AWS.
2. Instalar Crossplane no Hub e estudar XRD/Composition sem provisionar recursos AWS.
3. Registrar **apenas UAT** com template em `foundation/config/entity-a/uat/environment.example.yaml`. **O método de autenticação AWS ainda não foi decidido**: não presumir IRSA/Pod Identity em `kind`, não commitar secrets e não usar credenciais permanentes como padrão.
4. Implementar `ObjectStorage` real e depois `Database` como **cluster Aurora PostgreSQL**, com orçamento, IAM, rede e teardown aprovados.

## Monorepo e labs

Um único monorepo simula limites. Criar pastas Domain/Foundation somente com conteúdo real dos labs. Manter navegação e links válidos; preservar as 24 aulas em [fundamentos](docs/fundamentos/README.md). O aprendizado exige laboratório observável: pré-requisitos, comandos, falha controlada, diagnóstico, recuperação e cleanup. Preferir Markdown/Mermaid nativo legível. CODEOWNERS sem rulesets não é barreira de acesso.

## Segurança

Cada alteração AWS exige `sts get-caller-identity`, conta/região conferidas, credenciais temporárias, IAM mínimo, custo e autorização por operação. Terraform state fora do Git e segregado. Nunca commitar segredos, dados corporativos, certificados, kubeconfigs ou identificadores AWS reais. Destruição de conta nunca é efeito de remover YAML; aprovar docs não autoriza `apply`/`destroy`. Validar isolamento entre Domain, UAT/PROD e PCI. Trabalhar via branch e PR, **sem merge automático**.

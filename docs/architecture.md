# 🏗️ Arquitetura da Platform Engineering POC

[← Início](../README.md) · [Decisões atuais](adr/README.md) · [Monorepo](repository-structure.md) · [Roadmap](learning-roadmap.md)

> [!NOTE]
> **Arquitetura-alvo, não implementada.** Um monorepo de laboratório simula ownership e autonomia federada. Nenhuma operação AWS está autorizada por este documento. Nomes de entidades e Domains são fictícios.

## 1. Modelo arquitetural

A POC demonstra uma plataforma **Kubernetes-native** de governança central e consumo federado. Existem duas dimensões distintas:

- **Planos:** Experience Layer, Control Plane, Workflow Plane, Observability Plane e Data Plane. Os três planos técnicos centrais coexistem no **Platform Management Cluster**, inicialmente `kind`.
- **Quatro camadas de ownership:** **Foundation → Entity Core Infrastructure → Domain Infrastructure → Domain Applications**. Elas classificam responsabilidade de mudança e lifecycle, **não** representam planos ou clusters adicionais.

A **Entity** reúne a operação organizacional. Um **Environment** (UAT/PROD) é o contexto de uma conta AWS. O **Domain** é a unidade lógica de negócio, consumo, ownership e governança: possui Capabilities e aplicações em um ou mais Environments. **Capability** é o produto técnico provisionável, como `ObjectStorage`.

**Environment não é sinônimo de Core Infrastructure**: o ambiente é a conta toda, incluindo Core e os recursos dos Domains.

## 2. Visão geral

![Diagrama de arquitetura de referência com Hub, contas AWS UAT/PROD, quatro camadas e três personas](arquitetura%20Plataform.png)

[**Abrir diagrama em tamanho completo**](arquitetura%20Plataform.png) · O diagrama é **arquitetura-alvo**; não confirma recursos implantados. `TerraformStack` e os mecanismos de política/drift estão propostos, e GitHub Actions de aplicações opera fora do Hub.

### Topologia lógica (Mermaid)

```mermaid
flowchart TB
  E["Experience Layer<br/>Backstage · CLI · GitOps · AI/MCP"]
  subgraph HUB["Platform Management Cluster · kind inicial"]
    direction TB
    CP["Control Plane<br/>Kubernetes API · Crossplane central · XRDs<br/>Admission/Policies · TerraformStack CRD/controller (propostos)"]
    WF["Workflow Plane<br/>Argo CD central · Argo Workflows · Terraform Executor"]
    OB["Observability Plane<br/>OpenTelemetry · Grafana · Loki · Tempo · auditoria"]
    CP --> WF
  end
  subgraph DP["Data Plane · Entity A"]
    direction TB
    subgraph U["UAT · conta AWS"]
      UM["Management EKS<br/>Argo CD local · Operators (a definir) · OTel gateway"]
      UP["EKS PCI"]
      UN["EKS Non-PCI"]
      UR["AWS Runtime<br/>Core + Domains"]
      UM --> UP
      UM --> UN
    end
    subgraph P["PROD · conta AWS"]
      PM["Management EKS<br/>Argo CD local · Operators (a definir) · OTel gateway"]
      PP["EKS PCI"]
      PN["EKS Non-PCI"]
      PR["AWS Runtime<br/>Core + Domains"]
      PM --> PP
      PM --> PN
    end
  end
  E --> CP
  CP -. "Capabilities" .-> UR
  CP -. "Capabilities" .-> PR
  WF -. "Foundation e EKS" .-> UM
  WF -. "Foundation e EKS" .-> PM
```

Linhas tracejadas representam integração lógica, **não redes AWS nem permissões concedidas**. A arquitetura permite várias entidades. Cada entidade-alvo contém UAT e PROD em **contas AWS distintas**, cada uma com **Management EKS, EKS PCI, EKS Non-PCI** e serviços AWS. Na representação de **uma entidade** são sete clusters no total (Hub + 2×3), não um limite global. EKS Auto Mode é a preferência para EKS reais, **sob validação de custos, operadores e PCI**; criar um cluster de cada vez. O Argo CD local pode exigir segregação de instâncias/credenciais para PCI.

| Plano | Responsabilidade |
| :--- | :--- |
| **Experience Layer** | Consumidores independentes: Backstage, CLI, PR GitOps, AI/MCP; Backstage não é gateway obrigatório |
| **Control Plane** | Kubernetes API Server central, XRDs/Compositions/Providers de Capabilities, Crossplane e políticas |
| **Workflow Plane** | Argo CD central, Argo Workflows, Terraform Executor, aprovações; GitHub Actions de aplicações em repositório separado |
| **Observability Plane** | OpenTelemetry, Grafana, Loki, Tempo, métricas, eventos e auditoria |
| **Data Plane** | Contas AWS das Entities, Core compartilhado, Domains, Management EKS com Argo CD local/Operators/gateway de logs e EKS PCI/Non-PCI |

## 3. Quatro conceitos e seus contratos

| Conceito | Definição | Implementação **inicial** |
| :--- | :--- | :--- |
| **Entity** | Identidade organizacional que contém Environments e Domains | Organização declarativa no Git |
| **Environment** | Ambiente UAT/PROD da Entity associado a conta AWS, região e perfis | Registro `environment.yaml` versionado/validado, **não CRD** |
| **Domain** | Unidade lógica de negócio que possui aplicações, namespaces, Capabilities, owners, padrões e políticas | Registro **único** `domain.yaml` por Domain, **não XRD** |
| **Capability** | Produto técnico consumível, com tipo, schema, lifecycle e status próprios | **Crossplane XRD + Composition + Provider** por produto |

A mesma identidade de Domain pode ter recursos diferentes em UAT e PROD. Um bucket, banco ou aplicação pertence ao par **Domain + Environment**. O namespace é a materialização do Domain em um cluster, não sua definição. O Core pertence ao Environment, nunca ao lifecycle de um Domain específico. **MSK está fora do escopo inicial**.

Exemplo **ilustrativo** de Environment, sob gestão da Foundation central:

```yaml
# foundation/config/entity-a/uat/environment.yaml
entity: entity-a
environment: uat
aws:
  accountId: "111122223333" # fictício
  region: us-east-1
foundation:
  profile: standard
core:
  profile: eks-standard
```

Exemplo **ilustrativo** de Domain (identidade única):

```yaml
# data-plane/entity-a/domains/payments/domain.yaml
domain: payments
entity: entity-a
owners: [team-payments]
environments: [uat, prod]
policies:
  profile: standard
```

Esses YAMLs **não são manifests Kubernetes** e não devem ser sincronizados diretamente pelo Argo CD central. Inicialmente são lidos por validações e ferramentas de workflow. A vinculação confiável **Entity + Environment + Domain → conta/IAM/namespace** precisa ser construída e testada: consumidores não podem escolher credenciais AWS arbitrárias.

**`ObjectStorage` e `Database` são as duas primeiras Capabilities da POC**, nessa ordem. A primeira implementação de `Database` usará **Amazon Aurora PostgreSQL** (cluster Aurora), sem assumir que uma definição lógica de `Database` tenha de se limitar a Aurora em versões futuras; `Messaging` e `ComputeInstance` ficam no roadmap. **Não criar `CapabilityClaim` genérico.** **Domain Bootstrap é proposta anterior adiada:** previa a criação de Capabilities e configurações iniciais do Domain, mas não será implementado agora. Capabilities e Deployments continuam solicitações independentes. Ver [Domain Bootstrap](#6-domain-bootstrap-proposta-anterior-em-reavaliação).

### Decisões complementares da arquitetura-alvo

- **Contrato Terraform (proposto):** `TerraformStack` é o CRD Kubernetes próprio do Hub que representa **infraestrutura desejada** de Foundation/Entity Core. O Argo CD central sincroniza o manifesto; o controller coordena Argo Workflows/Terraform Executor para `plan` inicial e periódico, detecção de drift, classificação por Policy Engine (**auto apply / aprovação / bloqueio**) e atualização de status. **Não executar `apply` a cada reconcile**. `TerraformRun` permanece nome opcional de uma execução isolada, não um segundo CRD obrigatório. Ver [ADR-0006](adr/0006-terraformrun-and-application-ci.md).
- **Domain Applications:** as squads manterão **código, GitHub Actions e manifests Kubernetes da aplicação em repositório próprio**, separado deste monorepo e previsto para uma etapa posterior. O pipeline faz build/testes/scans e publicação, propõe atualização declarativa GitOps; **Argo CD local** em cada Management EKS faz o deploy nos EKS PCI/Non-PCI autorizados. O workflow CI existente neste monorepo só valida a POC.
- **Management EKS:** **Operators ainda a estudar**; usar **OpenTelemetry Collector/gateway** para encaminhar telemetria à camada central, não assumir Loki instalado dentro de todos os Management EKS.
- **Observabilidade central:** **OpenTelemetry, Grafana, Loki e Tempo**; estratégia de métricas/Prometheus, retenção e fronteiras PCI a validar no laboratório.
- **Domain Bootstrap:** conceito mantido no ADR-0005 para rastreabilidade, mas **em reavaliação e fora das próximas implementações**. Não é pré-requisito para Foundation, Capabilities ou Domain Applications.

### Personas, limites de abstração e modelo de consumo

Existem **três personas**, não quatro equipes para as quatro camadas de ownership:

| Persona | Papel | Conhecimento e responsabilidades |
| :--- | :--- | :--- |
| **Platform Team** | Construir/operar a plataforma como produto | Control/Workflow/Observability Planes, contratos, módulos/Compositions, policies, Golden Paths, Foundation e Entity Core compartilhados |
| **DevOps Federado (Federated DevOps Team)** | Ponte técnica entre plataforma e squads | Conhece catálogo/contratos e mantém **Domain Infrastructure**, usando Capabilities para Database, ObjectStorage, Messaging, alterações e upgrades; solicita evolução da plataforma quando necessário |
| **Squads de Produto (Application Team / Developers)** | Construir o produto de negócio | Código, testes, CI/CD, imagens e **manifests Kubernetes da aplicação** no repositório da squad, consumindo Starter Kits e padrões da plataforma sem conhecer motores internos |

**O desenvolvimento não precisa aprender a implementação interna de Crossplane, Terraform ou Argo para criar um serviço.** O DevOps Federado tem conhecimento mais aprofundado de consumo e lifecycle das Capabilities, mas não administra o Control Plane nem os módulos centrais. EKS compartilhado e seu upgrade/provisioning são governados pelo Platform Team (DevOps Federado pode propor).

**Dois tipos de Capabilities:** **Infrastructure Capabilities** (Crossplane especializado, consumido pelo DevOps Federado) e **Application Capabilities** (padrões de runtime/build/deploy/telemetria, consumidos pelos desenvolvedores; **não implicam XRD Crossplane**). Um **Golden Path / Starter Kit** combina padrões e automações para uma jornada específica. Backstage, CLI e Git são canais alternativos; Backstage não é gateway obrigatório.

**Aplicação mínima da POC (preferencialmente Python + Flask):** apenas uma carga de teste. **Após criar EKS e habilitar Argo CD local (Lab 10B)**, fazer smoke test de deploy, conectividade e health check, sem depender de Developer Experience ou Starter Kit. **Após observabilidade e governança**, montar a **Developer Experience (Lab 15)** e **por último** reutilizar o mesmo app na demonstração de Golden Path/integração end-to-end (Lab 16), com CI/CD e telemetria. A POC não pretende desenvolver aplicação complexa ou gerador completo de templates. Infraestrutura Database/S3/Messaging permanece com DevOps Federado. Detalhes em [ADR-0007](adr/0007-personas-and-developer-experience.md).

## 4. Quatro camadas de ownership

| Camada | Quem solicita | Quem controla execução e lifecycle | Recursos / motor |
| :--- | :--- | :--- | :--- |
| **Foundation** | Plataforma central; federado pode propor PR | Plataforma central | Conta, landing zone, rede, IAM; **Terraform** e configurações por Entity/Environment |
| **Entity Core Infrastructure** | DevOps federado nos limites autorizados | Plataforma governa componentes críticos | Três EKS e recursos comuns por Environment; **Terraform para EKS** |
| **Domain Infrastructure** | DevOps Federado (solicita/gerencia intenção no Git) | DevOps Federado sob guardrails da plataforma; Crossplane central reconcilia | Capabilities Database, ObjectStorage, Messaging via **XRDs** e contratos aprovados |
| **Domain Applications** | Application Team | Squad mantém código e manifests no repo da aplicação; GitHub Actions CI/CD + Argo CD local | Código, testes, build/scan/publicação, manifests de aplicação, Deployments/Services/HPA e runtime autorizado |

Um recurso externo tem **um único owner de reconciliação**. Terraform administra Foundation/EKS; Crossplane as Capabilities que lhe forem atribuídas. Core futuro pode conter Capability compartilhada, desde que tenha owner explícito e não seja excluída pela remoção de um Domain consumidor.

## 5. Organização do monorepo

```text
platform-engineering-poc/
├── platform/hub/                        # bootstrap kind
├── control-plane/                       # Crossplane, Capabilities e políticas
├── workflow-plane/                      # GitOps central, Workflows, Terraform
├── observability-plane/
├── foundation/                          # central, a criar no lab
│   ├── modules/
│   └── config/entity-a/
│       ├── uat/environment.yaml
│       └── prod/environment.yaml
├── data-plane/
│   └── entity-a/
│       ├── core-infrastructure/         # compartilhado por ambiente
│       │   ├── uat/
│       │   │   ├── management-eks/
│       │   │   ├── eks-pci/
│       │   │   └── eks-non-pci/
│       │   └── prod/
│       │       ├── management-eks/
│       │       ├── eks-pci/
│       │       └── eks-non-pci/
│       └── domains/
│           ├── payments/
│           │   ├── domain.yaml          # identidade única
│           │   ├── uat/
│           │   │   └── infrastructure/  # XRs · DevOps Federado
│           │   └── prod/
│           │       └── infrastructure/  # XRs · DevOps Federado
│           └── orders/
├── docs/
├── labs/
└── tests/
```

**Separado deste monorepo:** o repositório da aplicação de teste, previsto para um lab posterior, pertence às **Squads de Produto (Application Team)** e inclui `src/`, `tests/`, `Dockerfile`, `.github/workflows/` e `k8s/{base,overlays/uat,overlays/prod}`. O **Argo CD local** sincroniza esses manifests diretamente do repositório da aplicação. A pasta `applications/` que aparecia na árvore anterior do monorepo deixa de ser o destino-alvo para código/manifests de aplicação.

**Esta é a árvore-alvo; o primeiro lab cria somente a configuração executável do Hub `kind` e um template declarativo de conta UAT.** Diretórios de Domain/Foundation serão adicionados somente com conteúdo real dos labs. Os README de infraestrutura preexistentes podem permanecer como documentação temporária até a implementação; não são fontes de configuração runtime. A [estrutura](repository-structure.md) especifica essa transição.

O monorepo demonstra autonomia via PR, revisão de paths e eventualmente CODEOWNERS. **CODEOWNERS sem regras de proteção não basta;** também são necessários RBAC Kubernetes, admission, IAM por conta e AppProjects Argo CD. Terraform state fica fora do Git.


## 6. Domain Bootstrap: proposta anterior em reavaliação

> [!NOTE]
> Esta seção preserva um desenho anterior para rastreabilidade. **Não implementar nem assumir Domain Bootstrap como obrigatório** neste ciclo. A necessidade do gerador será reavaliada; ele pode não existir na arquitetura final.

O **Domain é uma identidade lógica única na Entity**. **No desenho anterior, a criação de um Domain usaria o Domain Bootstrap** como fluxo de inicialização: escolhe e valida um perfil versionado, gera sua estrutura e as declarações dos recursos previstos nesse perfil. O processo é iniciado **pela criação do Domain**, não como uma obrigação de executar um Bootstrap independente para cada Environment. O Domain pode solicitar Capabilities adicionais depois. **Cadastrar um Domain não equivale a concluir seu Bootstrap**.

### O que o perfil `standard` cria (exemplo da POC)

| Recurso inicial definido pelo perfil | Motor de entrega | Destino / ownership |
| :--- | :--- | :--- |
| `Database` | Manifest de Capability → Argo CD **central** → Crossplane → AWS | Domain, no Environment |
| `ObjectStorage` | Manifest de Capability → Argo CD **central** → Crossplane → AWS | Domain, no Environment |
| Namespace(s) nos destinos Kubernetes permitidos | Manifest Kubernetes → Argo CD **local** | Domain, por cluster/destino |
| Quotas, políticas de rede/segurança, permissões e labels/tags aplicáveis | GitOps local para Kubernetes; políticas de Capability para AWS | Governança da plataforma, instanciada para Domain |

**Nota de fases:** o perfil `standard` de referência cria banco, bucket, namespace e políticas como padrão de nascimento do Domain. `ObjectStorage` continua sendo o primeiro **produto técnico implementado** no lab; `Database` será o segundo, com **Aurora PostgreSQL**, conforme validação de rede, custo, permissões e estratégia de remoção. Um perfil de laboratório reduzido pode demonstrar o mecanismo antes disso. **Não declarar um Domain com perfil `standard` como `Ready` enquanto os recursos previstos por esse perfil não estiverem provisionados e verificados**.

### Processo de criação

```mermaid
flowchart TB
    R["Criar Domain<br/>identidade + perfil + destinos selecionados"] --> V["Validação de identidade, escopo,<br/>permissões e perfil"]
    V --> B["Domain Bootstrap<br/>renderização declarativa idempotente"]
    B --> I["Manifests de Capabilities<br/>por Domain + Environment"]
    B --> K["Manifests Kubernetes<br/>por Domain + Environment"]
    I --> C["Argo CD central → Kubernetes API → Crossplane → AWS"]
    K --> L["Argo CD local → EKS autorizado"]
    C --> S["Verificação das dependências obrigatórias"]
    L --> S
    S --> READY["Domain Environment Ready"]
```

**Local da definição:** `data-plane/<entity>/domains/<domain>/domain.yaml` (identidade, perfil e versão desejados, Environments selecionados). O **perfil versionado**, com template e requisitos, pertence ao **Control Plane** (catálogo/padrões da plataforma), enquanto o **mecanismo de execução do Bootstrap** pertence ao **Workflow Plane**. Os manifests gerados ficam sob `domains/<domain>/<env>/{infrastructure,applications}/`, revisados e versionados no **mesmo monorepo**. **A aplicação é feita somente pelo GitOps**; o Bootstrap **não cria recursos diretamente** na AWS ou nos EKS, evitando dois escritores.

**Disparo proposto para a POC:** solicitação de **criação de Domain** com perfil e ambientes de destino → geração idempotente da estrutura e dos manifests iniciais → validação/revisão/merge → reconciliadores existentes. **Um único processo de Bootstrap do Domain** pode preparar recursos para UAT, PROD ou ambos, de acordo com a solicitação e suas autorizações. A posterior inclusão de um novo Environment é uma **evolução do Domain** e pode reutilizar o mesmo mecanismo de geração, sem representar um novo nascimento. Não supor que gravar o YAML dispara um controller: gatilho de CI/workflow e gerador são **componentes ainda a implementar**. O Bootstrap não executa `apply` AWS só por aprovação documental.

**Prontidão e falhas:** após sincronização, uma verificação independente consolida evidências de execução **para o Domain e seus destinos (Domain + Environment)** (XRs `Ready`, Argo CD Sync/Health, namespace/políticas). Estado como `Pending`, `Provisioning`, `Ready` ou `Failed` é **derivado de evidências**, não `status` magicamente escrito no `domain.yaml`. Falhas permitem reexecução idempotente sem duplicar buckets, bancos ou namespaces; se faltar um recurso do padrão num destino, o Bootstrap daquele destino não está concluído. A checagem por Environment é apenas uma dimensão operacional do mesmo Domain.

**Evolução:** perfis possuem versão imutável/referenciada; novo perfil **não altera automaticamente** Domains existentes. Migração exige PR/avaliação de mudanças, especialmente recursos stateful e PROD. A remoção de um Domain não dispara automaticamente `destroy`: exige descomissionamento explícito, políticas de retenção, dependências e gates. Recursos de **Entity Core Infrastructure** nunca são excluídos com o Domain.

## 7. Três fluxos sem sobreposição

### Foundation / Entity Core: Terraform orquestrado

```text
foundation/config/ ou core-infrastructure/ (Git)
  → validação de intenção + PR/aprovação
  → Argo CD central → Kubernetes API / TerraformStack CR
  → TerraformStack Controller → Argo Workflows
  → Terraform Executor: plan → Policy Engine
  → auto apply / aprovação / bloqueio
  → verificação e status → AWS Foundation / EKS
  → nova verificação periódica de drift (sem apply incondicional)
```

O Bootstrap Terraform central é reutilizado por configuração de Entity/Environment, com **state segregado**. Inicialmente pode rodar em um lab manual e auditável. **Argo CD não executa Terraform**. Criar conta AWS automaticamente é possibilidade futura, **não exige `AWSEnvironment` CRD/controller próprio agora**.

### Evolução planejada do TerraformStack

1. **Declarative Execution:** intenção Git → Argo CD central → `TerraformStack` → controller → Workflow/plan → aplicação condicionada à política. Começar com executor simulado.
2. **Drift Detection:** controller agenda plans periódicos e atualiza condições/status, **sem remediação automática nesta fase**. Distinguir drift externo de alteração declarada.
3. **Policy-Based Reconciliation:** decisão versionada e auditável (**auto apply / require approval / deny**) por recurso, atributos, Environment, criticidade e origem; apply somente de plan autorizado, com isolamento e state seguro.

A escolha da implementação do Policy Engine e o schema final do CRD ficam para os labs. **A criação do CR Kubernetes é responsabilidade do Argo CD central**, não do Argo Workflows. O Argo Workflows é o executor do plan/apply.

### Domain Infrastructure: Argo CD central + Crossplane

```mermaid
flowchart LR
  D["DevOps federado"] --> G["PR: Domain/Environment/infrastructure"]
  G --> V["Validações e merge"]
  V --> CD["Argo CD central"]
  CD --> API["Kubernetes API do Hub"]
  API --> XP["Crossplane central"]
  XP --> AWS["AWS validada"]
  XP -. "status / conditions" .-> API
```

Neste fluxo de **Domain Infrastructure**, o Argo CD central sincroniza CRs Kubernetes de **Capability**; na arquitetura completa também sincronizará `TerraformStack`, para Foundation/Entity Core. Os registros de Domain/Environment informam ownership/destino após validação, **não são CRs**; **Crossplane não observa Git diretamente**. Em cada Capability validar schema, escopo, credenciais e política de deletion.

### Domain Applications: GitHub Actions + GitOps local

```text
repositorio-da-aplicacao/k8s/overlays/<env>/ (Git da squad)
  → Argo CD no Management EKS daquele Environment
  → EKS PCI ou Non-PCI permitido
  → Namespaces / quotas / policies / Deployments / Services
```

O código, as pipelines GitHub Actions **e os manifests Kubernetes da aplicação** ficarão **juntos no repositório separado da squad** (não neste monorepo). A pipeline publica imagem e propõe mudanças GitOps; o Argo CD local aplica os manifests aprovados. **Manifests de Infrastructure Capabilities** continuam no monorepo sob gestão do DevOps Federado. Uma Composition Crossplane central não deve possuir esses mesmos namespaces. O controle PCI requer avaliação das permissões de acesso e eventual segregação do GitOps local.

## 8. Segurança, exclusão, drift e falhas

- **AWS real:** primeiro conta preexistente; STS comprova conta/região, credenciais temporárias, IAM mínimo, orçamento/custos e aprovação explícita **a cada execução**. Segunda etapa: UAT/PROD em contas distintas; terceira etapa opcional: automação de criação de contas.
- **Isolamento:** validar negativa para tentativas de Domain A acessar Domain B ou UAT selecionar role/conta PROD. Paths Git e namespaces não equivalem a isolamento de contas AWS.
- **Exclusão:** conta AWS **nunca** é apagada por excluir YAML; EKS só com gate e análise de dependências. Capabilities devem ter políticas de retenção/deletion testadas. Finalizer **não é aprovação humana**.
- **Drift:** Crossplane reconcilia Capabilities continuamente enquanto o Hub estiver disponível; Terraform não reconcilia sem execução; Argo CD local continua operando aplicações se seu acesso ao Git e clusters permanecer válido.
- **Observabilidade:** correlacionar revisão Git, aplicação Argo CD, execução Terraform, XR, condições `Ready` e eventos AWS.

## 9. Nascimento de uma entidade versus ordem pedagógica

**Fluxo operacional alvo, sem Domain Bootstrap obrigatório:** Foundation cria/adota contas e rede → Entity Core prepara EKS e Argo CD local → manifests de Capabilities aprovados são reconciliados por Crossplane central → workloads de Domain Applications são entregues por Argo CD local. **O fluxo Domain Bootstrap descrito na seção 6 está adiado e em reavaliação.**

**POC pedagógica:** baseline → Hub `kind` → Argo CD central/Argo Workflows → Crossplane local sem AWS → configuração UAT e autenticação AWS (a definir) → `ObjectStorage` e `Database` Aurora PostgreSQL → registro Domain/Environment + GitOps central → Terraform/Argo Workflows → contrato declarativo Terraform (proposto) → Core/EKS incremental → Argo CD local → **smoke test de app mínima (Lab 10B)** → UAT/PROD segregados → Observability Plane → segurança/governança → extensões opcionais → **Developer Experience (Lab 15)** → **Golden Path integrado (Lab 16, último passo)**. **Sem Bootstrap obrigatório.** **Não criar seis EKS simultaneamente**, nem implementar um controller de lifecycle apenas para manter uma ideia anterior.

## 10. Decisões vigentes e pontos abertos

**Atualização:** TerraformStack com controller e reconciliação orientada por políticas é um contrato alvo ainda não implementado; GitHub Actions das aplicações será externo a este monorepo; OpenTelemetry/Grafana/Loki/Tempo são a stack-alvo; Operators locais ainda serão selecionados; Domain Bootstrap está em reavaliação e adiado.

**Fechado:** Domain como unidade lógica de consumo/ownership, Environment como destino AWS, Entity como organização, Capability como XRD especializado, quatro camadas (incluindo **Domain Applications**), monorepo, Crossplane central sem cascata, Terraform para Foundation/EKS, GitOps central de infra e local de aplicações, AWS real gradual e Auto Mode como preferência. **Criação padronizada via Domain Bootstrap não está fechada: em reavaliação.**

**A validar nos labs:** autenticação de UAT sem credenciais permanentes no Git, arquitetura de rede e custo do Aurora PostgreSQL, decisão futura sobre manter ou descartar Domain Bootstrap (sem implementá-lo agora), schema dos YAMLs, transformação segura de registros lógicos em escopo/credenciais, autenticação Crossplane a partir do `kind`, distribuição de namespaces/quotas ao GitOps local, gates, observabilidade, PCI e necessidade eventual de uma API/controller de ciclo de vida de contas.

[Decisões](adr/README.md) · [Monorepo](repository-structure.md) · [Roteiro](learning-roadmap.md)

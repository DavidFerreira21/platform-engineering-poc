# Platform Engineering POC

> Plataforma **Kubernetes-native** construída em labs incrementais, com governança central e autonomia federada por **Domain**.

**`Kubernetes API` · `Crossplane` · `Argo CD` · `Argo Workflows` · `Terraform` · `AWS`**

> [!NOTE]
> **Estado atual:** arquitetura documentada, sem deploy AWS/Kubernetes realizado por esta revisão. As pastas são organizacionais ou exemplos de implementação futura.

## 🧭 Navegação

| Página | Conteúdo |
| :--- | :--- |
| [🏗️ Arquitetura](docs/architecture.md) | Modelo, topologia, ownership, contratos e fluxo completo |
| [🗂️ Monorepo](docs/repository-structure.md) | Responsabilidades e localização das configurações |
| [📝 Decisões vigentes](docs/adr/README.md) | ADRs atuais, sem versões contraditórias |
| [🧪 Roteiro prático](docs/learning-roadmap.md) | Fases de laboratório, AWS real e validações |
| [▶️ Lab 01: Hub kind](labs/01-platform-management-cluster.md) | Criar, verificar e remover o Platform Management Cluster local |
| [📖 Fundamentos](docs/fundamentos/README.md) | 24 aulas conceituais pt-BR |
| [🤖 Orientações para agentes](AGENTS.md) | Padrões de implementação e segurança |

## 🧱 Visão arquitetural

```mermaid
flowchart TB
  E["Experience Layer<br/>Backstage · CLI · GitOps · AI/MCP"]
  subgraph HUB["Platform Management Cluster · kind inicial"]
    C["Control Plane<br/>Kubernetes API · Crossplane central"]
    W["Workflow Plane<br/>Argo CD central · Workflows · Terraform"]
    O["Observability Plane<br/>métricas · logs · traces"]
  end
  subgraph D["Data Plane · Entity A"]
    CORE["Core Infrastructure<br/>UAT/PROD · 3 EKS por conta"]
    DOM["Domains<br/>identidade única · infra e apps por ambiente"]
  end
  E --> C
  C --> W
  C -. "Capabilities AWS" .-> DOM
  W -. "Foundation / EKS" .-> CORE
```

**Environment é o contexto UAT/PROD associado à conta AWS inteira, não apenas ao Core.** O Data Plane contém **Core Infrastructure compartilhada** e **Domains** de negócio/aplicação. Um Domain pode possuir Capabilities, namespaces e apps em ambos os ambientes. **A criação do Domain segue um Domain Bootstrap com perfil versionado**: o perfil `standard` prevê Database, ObjectStorage, namespace e políticas nos destinos selecionados. O Bootstrap gera manifests no Git; Crossplane central e Argo CD local entregam os recursos. Nada é marcado pronto antes de verificar todas as dependências obrigatórias.

## 🧩 Quatro conceitos fundamentais

| Conceito | Papel | Representação inicial |
| :--- | :--- | :--- |
| **Entity** | Organização / responsáveis | Git |
| **Environment** | Conta AWS UAT/PROD, região e perfis | `environment.yaml`, não CRD |
| **Domain** | Identidade lógica, ownership, consumo, infraestrutura e aplicações | Um `domain.yaml` por Domain, não XRD |
| **Capability** | Produto técnico entregue pela plataforma | XRD + Composition + Provider Crossplane |

`ObjectStorage` é a primeira Capability e **`Database` (Aurora PostgreSQL)** vem em seguida; `Messaging` e `ComputeInstance` ficam planejados. Sem MSK inicial ou `CapabilityClaim` genérico.

## 🔐 Quatro camadas de ownership

| Camada | Gestão | Motor |
| :--- | :--- | :--- |
| **Foundation** | Central | Terraform / configuração por Entity e Environment |
| **Entity Core Infrastructure** | Federado propõe; mudanças críticas sob governança | Terraform para EKS |
| **Domain Infrastructure** | Federado / Domain | Argo CD central → Crossplane → AWS |
| **Applications** | Local | Argo CD local → EKS autorizado |

**Um único monorepo** simula a organização por Entity/Domain. O Crossplane é central, sem cascata, e Argo CD local administra aplicações. A definição das contas e o provisioning de Foundation não dependem de criar `AWSEnvironment` ou Environment Controller próprio nesta fase.

## 🧰 Domain Bootstrap

**Ao criar um Domain**, a plataforma executa o **Domain Bootstrap** como fluxo padrão: o perfil versionado define a estrutura e os recursos iniciais, considerando os Environments escolhidos. Não existe um Bootstrap independente obrigatório por Environment. O gerador produz os manifests de Capabilities (`infrastructure/`) e de namespaces/policies (`applications/`) para revisão em PR. O **Argo CD central** aplica os CRs de Capability, o **Crossplane** entrega AWS e o **Argo CD local** configura Kubernetes. A conclusão é verificada para o Domain, com detalhamento por Environment; criar em UAT não seleciona PROD automaticamente. [Detalhamento](docs/architecture.md#6-domain-bootstrap-criação-padronizada-do-domain).

## ☁️ Estratégia AWS

Primeiro Hub `kind` com [Lab 01 executável](labs/01-platform-management-cluster.md), depois Crossplane no Hub; em seguida preparar a **conta UAT existente** e a autenticação AWS, ainda a definir, para `ObjectStorage` real e `Database` com **Amazon Aurora PostgreSQL**. Depois contas UAT/PROD distintas, Terraform Foundation/Core, um EKS Auto Mode de cada vez e GitOps local. A criação automatizada de contas é evolução opcional, justificada em lab futuro.

> [!IMPORTANT]
> Executar AWS somente com conta/região verificadas via STS, IAM mínimo, custo, autorização explícita e plano de cleanup. Não publicar credenciais, kubeconfigs ou Terraform state. Aprovar a documentação **não** autoriza `apply`/`destroy`.

[Arquitetura completa](docs/architecture.md) · [Roteiro](docs/learning-roadmap.md)

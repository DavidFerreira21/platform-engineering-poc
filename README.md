# Platform Engineering POC

Esta POC é um exercício prático para aprender **Engenharia de Plataforma** construindo uma plataforma interna de ponta a ponta. O percurso começa pelos fundamentos, passa pelas decisões de arquitetura, automação e governança e chega à entrega e à operação de uma aplicação por um time de produto.

Kubernetes, Terraform, Crossplane e Argo são os meios para percorrer esse caminho. Cada ferramenta é explorada na medida necessária para entender como os componentes se conectam, como as responsabilidades são distribuídas e o que a plataforma oferece a quem entrega software.

## Por onde começar

1. Leia os [Fundamentos](docs/fundamentos/README.md). As 24 aulas apresentam Engenharia de Plataforma, produto interno, self-service, Golden Paths, governança e os conceitos usados na POC.
2. Leia a [Arquitetura](docs/architecture.md). Use o diagrama de referência para entender a topologia, as três personas, a responsabilidade de cada plano e como Entity, Environment, Domain e Capability se relacionam.
3. Siga o roteiro para acompanhar a construção da plataforma. Os guias mostram por que cada decisão foi tomada, como reproduzi-la e o que observar no sistema. [Roteiro prático](docs/learning-roadmap.md) · [Guias das etapas](labs/README.md)

Para consultar detalhes, veja a [estrutura do monorepo](docs/repository-structure.md), as [decisões arquiteturais](docs/adr/README.md) e as [orientações para agentes](AGENTS.md).

## Cenário de demonstração

Uma transportadora global fictícia recebe e distribui encomendas em centros de distribuição de vários países. Cada centro tem suas próprias instalações, equipamentos e rotinas de trabalho. Nos horários de pico, vários centros registram recebimentos ao mesmo tempo.

A POC representa a transportadora como **uma única Entity**, a `entity-a`, com ambientes **UAT** e **PROD**. Dentro dela, as squads de produto cuidam de Domains com responsabilidades distintas.

| Domain | Responsabilidade da squad |
| :--- | :--- |
| **Rastreamento de Encomendas** | Registrar os eventos do percurso e permitir a consulta ao histórico de uma encomenda. |
| **Estoque** | Controlar a disponibilidade de materiais de embalagem e insumos em cada centro. |
| **Planejamento de Rotas** | Planejar as transferências entre centros e as rotas de entrega. |

O exercício implementa uma fatia do **Rastreamento de Encomendas**. Um operador informa o **Código da encomenda**, clica em **Registrar recebimento** e confere o código e o horário na lista de recebimentos. A preferência é usar Python + Flask. Na **Etapa 12**, uma versão HTTP mínima valida deploy, health check e rollback. No **Golden Path da Etapa 24**, a mesma aplicação grava e consulta os recebimentos no Aurora PostgreSQL, solicitado pelo DevOps Federado como Capability. [ADR-0007](docs/adr/0007-personas-and-developer-experience.md)

## 🧱 Visão arquitetural

![Arquitetura de referência da Platform Engineering POC — Hub, Spokes, personas e camadas de recursos](docs/arquitetura%20Plataform.png)

[**Ver diagrama em tamanho completo**](docs/arquitetura%20Plataform.png) · [Topologia lógica e fluxos detalhados](docs/architecture.md#2-visão-geral)

> **Nota de arquitetura:** o diagrama oficial está em `docs/arquitetura Plataform.png` e representa a **topologia-alvo**. Para contratos e decisões atualizadas, consultar os [ADRs](docs/adr/README.md). O nome vigente do contrato declarativo proposto é `TerraformStack` (CRD + controller); `TerraformRun` fica reservado para uma execução individual, se necessário.

**Environment é o contexto UAT/PROD associado à conta AWS inteira, não apenas ao Core.** O Data Plane contém **Core Infrastructure compartilhada** e **Domains** de negócio/aplicação. Um Domain pode possuir Capabilities, namespaces e apps em ambos os ambientes. Capabilities continuam entregues por Crossplane central e aplicações por Argo CD local.

## 👥 Personas e experiência da plataforma

| Persona | Responsabilidade | Nível de abstração |
| :--- | :--- | :--- |
| **Platform Team** | Constrói/opera a plataforma, módulos, contratos, guardrails e Golden Paths | Conhece os motores internos |
| **DevOps Federado (Federated DevOps Team)** | Gerencia o lifecycle de **Domain Infrastructure** com Capabilities aprovadas; propõe mudanças de EKS, mas não administra o Core compartilhado | Conhece contratos, catálogo e processos de infraestrutura |
| **Squads de Produto (Application Team)** | Mantêm código, testes, CI/CD e manifests Kubernetes em repositório próprio | Consomem **Golden Paths / Starter Kits**, sem conhecer os motores internos |

A infraestrutura do Domain é simulada em **pastas do monorepo**, sob ownership do **DevOps Federado**. Já **código, GitHub Actions e manifests Kubernetes da aplicação** pertencem ao **repositório separado da squad**. O Argo CD local entrega os manifests do repo da aplicação nos EKS autorizados.

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
| **Domain Applications** | Application Team (repo próprio com código **e manifests Kubernetes**) | GitHub Actions → Argo CD local → EKS autorizado |

**Um único monorepo** simula a organização por Entity/Domain. O Crossplane é central, sem cascata, e Argo CD local administra aplicações. A definição das contas e o provisioning de Foundation não dependem de criar `AWSEnvironment` ou Environment Controller próprio nesta fase.

## 🔧 Decisões arquiteturais

- **TerraformStack.** O CRD da Etapa 05 descreve a infraestrutura desejada e chega à API do Hub via Argo CD central. O controller começa com execução simulada na Etapa 06 e aplica a primeira stack após aprovação humana na Etapa 09. Detecção de drift e reconciliação por políticas entram nas Etapas 18 e 19. `TerraformRun` fica como nome opcional de uma execução individual. [ADR-0006](docs/adr/0006-terraformrun-and-application-ci.md)
- **Domain Applications:** código e workflows CI/CD **GitHub Actions em um repositório de aplicação de teste separado**, a integrar em etapa posterior; deploy GitOps nos EKS via Argo CD local.
- **Management EKS:** Argo CD local, Operators **a escolher** e OpenTelemetry Collector/gateway para logs e traces; backend central com **Grafana + Loki + Tempo**. Estratégia de métricas/Prometheus em validação.
- **Ownership:** Foundation → Entity Core Infrastructure → Domain Infrastructure → Domain Applications.

[Decisão de personas e Golden Paths](docs/adr/0007-personas-and-developer-experience.md)

## ☁️ Estratégia AWS

A POC prepara o Hub `kind` e testa o `TerraformStack` com execução simulada. Para a primeira execução real, a equipe registra UAT, valida a identidade do executor e prepara o backend de estado. Terraform aplica a rede e, em seguida, provisiona Management EKS e EKS Non-PCI, um cluster por vez. Após o smoke test, o Crossplane reconcilia `ObjectStorage` e `Database` com Amazon Aurora PostgreSQL. O guia introduz PROD nos testes de isolamento. [Roteiro prático](docs/learning-roadmap.md)

> [!IMPORTANT]
> Executar AWS somente com conta/região verificadas via STS, IAM mínimo, custo, autorização explícita e plano de cleanup. Não publicar credenciais, kubeconfigs ou Terraform state. Aprovar a documentação **não** autoriza `apply`/`destroy`.

[Arquitetura completa](docs/architecture.md) · [Roteiro](docs/learning-roadmap.md)

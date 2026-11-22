# ADR-0007 — Personas, ownership federado e Developer Experience por Golden Paths

**Estado:** decisão arquitetural vigente; implementação de labs pendente · **Data:** 21/11/2026

## Contexto e objetivo

A plataforma é um produto destinado a **três personas com níveis distintos de conhecimento e autonomia**. O objetivo é reduzir a carga cognitiva dos desenvolvedores, preservar a autonomia da squad no ciclo de vida da aplicação e manter ownership técnico claro sobre infraestrutura e plataformas.

A arquitetura conserva as **quatro camadas**: **Foundation → Entity Core Infrastructure → Domain Infrastructure → Domain Applications**. Elas **não são quatro equipes**: as responsabilidades são distribuídas entre as três personas.

## Decisão 1 — Três personas

| Persona | Foco | Conhecimento esperado da plataforma | Responsabilidades |
| :--- | :--- | :--- | :--- |
| **Platform Team** | **Construir e operar a plataforma como produto** | Profundo: Kubernetes API, Control/Workflow/Observability Planes, Crossplane, Terraform, Argo, contratos, segurança | Contratos, catálogo de Capabilities, módulos/Compositions, guardrails, DX/Golden Paths, foundations, Core compartilhado e governança |
| **Federated DevOps Team** | **Ponte técnica entre plataforma e squad** | Maior que o desenvolvimento: catálogo, contratos, ambientes, mudanças de infraestrutura, operação e políticas | Propor, configurar, evoluir e acompanhar o lifecycle de **Domain Infrastructure** (Database, ObjectStorage, Messaging etc.) por manifests Git; solicitar à plataforma Capabilities ausentes |
| **Application Team / Developers** | **Construir e evoluir o produto de negócio** | Padrões de consumo da plataforma e interface dos Starter Kits; não precisa conhecer motores internos | Código, testes, build, imagens, configuração de deploy, manifests de aplicação, CI/CD e evolução de **Domain Applications** |

O DevOps Federado não implementa, por padrão, módulos Terraform, XRDs ou Controllers da plataforma, nem possui credenciais administrativas irrestritas na AWS. Ele **consome e administra a intenção** dos produtos de infraestrutura aprovados; a plataforma implementa e governa seus motores. Colaboração e solicitações de evolução são permitidas.

O Application Team não deve ser obrigado a aprender ou manter internals de Crossplane, TerraformStack/Run ou Argo Workflows para publicar uma aplicação.

## Decisão 2 — Matriz de responsabilidade por camada

| Camada | Owner principal | Quem solicita/propõe | Motor e fronteira |
| :--- | :--- | :--- | :--- |
| **Foundation** | Platform Team | Plataforma e propostas sob governança | Terraform + workflows/policies de Foundation |
| **Entity Core Infrastructure** | Platform Team (infra compartilhada) | Federated DevOps pode solicitar mudanças, não criar EKS arbitrariamente | Terraform e gates centrais; cluster upgrade/provisioning de EKS sob controle da plataforma |
| **Domain Infrastructure** | Federated DevOps Team (intenção e lifecycle da infraestrutura do Domain) | DevOps Federado a partir da necessidade da squad | Manifests de Capability (Crossplane central) para Database, ObjectStorage, Messaging; infraestrutura técnica, não código de aplicação |
| **Domain Applications** | Application Team | Desenvolvedores e responsáveis pelo produto | Código/manifests no repositório da aplicação, GitHub Actions CI/CD e Argo CD local para deploy |

**Exemplo:** para adicionar/atualizar banco, bucket ou recurso de mensageria do Domain, o DevOps Federado altera o contrato correspondente e acompanha status/upgrade. Para upgrade do **EKS compartilhado**, o DevOps Federado pode propor ou coordenar a demanda, mas **não possui ownership do cluster**; a mudança é governada pelo Platform Team.

As mudanças precisam de permissões por path/repo, revisão, RBAC/admission e IAM — rótulos de owner não são controles técnicos suficientes. Aprovação da mudança Git e aprovação da execução de infraestrutura podem ser gates diferentes.

## Decisão 3 — Fronteiras dos repositórios

**POC monorepo — plataforma e Domain Infrastructure**

```text
platform-engineering-poc/
├── foundation/                        # Platform Team
├── platform/                          # Platform Team
├── control-plane/                     # Platform Team
├── workflow-plane/                    # Platform Team
├── observability-plane/               # Platform Team
└── data-plane/
    └── entity-a/
        ├── core-infrastructure/        # Platform Team
        └── domains/
            └── payments/
                ├── domain.yaml        # identidade lógica/owner
                ├── uat/infrastructure/ # Federated DevOps
                └── prod/infrastructure/# Federated DevOps
```

Em produção, a infraestrutura por Domain **pode residir em repositórios separados**. No laboratório, **pastas do monorepo** são suficientes para simular autonomia federada sem criar repositórios extras.

**Repositório separado da aplicação de teste — Domain Applications**

```text
payments-app/                          # nome ilustrativo, ainda não criado
├── src/
├── tests/
├── Dockerfile
├── .github/workflows/                 # GitHub Actions da aplicação
└── k8s/                                # manifests de aplicação
    ├── base/
    └── overlays/{uat,prod}/
```

**O código e os manifests Kubernetes da aplicação pertencem ao repositório da aplicação**, não às pastas `data-plane/<entity>/domains/<domain>/<env>/applications/` do monorepo. Esses paths eram exemplos anteriores da arquitetura-alvo e **não devem ser implementados como destino obrigatório da Application Team**.

Os manifests de **infraestrutura/Capability**, criados e mantidos pelo DevOps Federado, **permanecem separados** do código e dos manifests de Deployments/Services/HPA mantidos pelo time de desenvolvimento.

**Argo CD central** sincroniza Capabilities do Domain Infrastructure com a API do Hub, onde Crossplane central reconcilia serviços AWS. **Argo CD local** do Management EKS observa manifests GitOps da aplicação, no repositório separado, e entrega aos clusters PCI/Non-PCI autorizados. A pipeline GitHub Actions realiza build, testes, scans e publicação, seguindo revisões/policies para atualização declarativa da imagem. Não fazer deploy imperativo por padrão.

## Decisão 4 — Catálogo com dois níveis de consumo

- **Infrastructure Capability:** produtos técnicos como `Database`, `ObjectStorage`, `Messaging`, expostos como contratos especializados para o **Federated DevOps Team**, que gerencia alterações e versões.
- **Application Capability:** padrões reutilizáveis de runtime, build/deploy, telemetria, deployment, segurança e operação consumidos pela **Application Team**. **Não implica criar um XRD Crossplane para cada recurso de aplicação**.
- **Golden Path / Starter Kit:** experiência guiada que combina templates, convenções, automações e Application Capabilities para entregar um tipo de aplicação com baixo esforço e padrões padronizados.

O Developer pode consumir os padrões por **Backstage, CLI ou Git/Template**. **Backstage não é um gateway obrigatório** e não deve inventar API paralela aos contratos estabelecidos.

## Decisão 5 — Primeiro Golden Path: Java Service

A primeira experiência completa de aplicação será um **Java Service Starter Kit**, usando **Spring Boot como opção inicial a validar no lab**. O kit deve gerar um repositório de aplicação com:

1. Código e estrutura de projeto Java; convenções de build e testes.
2. Dockerfile/build de imagem e GitHub Actions para lint/testes/scan/publicação.
3. Manifests Kubernetes (por exemplo Kustomize base/overlays UAT e PROD), com requests/limits, health probes, configuração de runtime e identidade.
4. Observabilidade padronizada com OpenTelemetry, logs e traces e integração com os backends disponíveis.
5. Security/quality gates e políticas padrão; adaptações e escape hatches **governados**, sem exigir conhecimento dos mecanismos internos.

**Jornada:** escolher Java Service → fornecer metadados mínimos → gerar repo → desenvolver código → commit/PR → CI/CD → GitOps Argo CD local → aplicação no cluster autorizado.

O Starter Kit **não deve criar RDS, S3 ou outros recursos de Domain Infrastructure sem a participação/autorização do DevOps Federado**. O desenvolvimento registra uma necessidade; o DevOps Federado consome a Capability apropriada.

## Consequências e decisões pendentes para os labs

- **Modelo de permissões:** CODEOWNERS/path rules nos exemplos; RBAC/admission/AppProjects/IAM efetivos para impedir mudanças fora do ownership.
- **Integração CI → GitOps:** definir como a pipeline da aplicação atualiza os manifests versionados, aprovações e promoção entre UAT/PROD.
- **Golden Path:** validar ferramentas de geração (Backstage Software Templates, CLI/template Git) e alternativas Helm/Kustomize; **não escolher tecnologia sem lab**.
- **Produto e métricas DX:** medir esforço para criar serviço, necessidade de intervenção da plataforma, caminhos de exceção, tempo de entrega e feedback de persona.
- **Não implementar Domain Bootstrap agora:** ADR-0005 continua adiado; a proposta de Starter Kit de aplicação é **distinta** do Bootstrap de Domain e não exige esse gerador.
- **Status:** estes contratos e exemplos descrevem o desenho pretendido, não repositório da aplicação já fornecido ou componentes AWS/EKS implantados.

[Arquitetura](../architecture.md) · [Estrutura](../repository-structure.md) · [Roadmap](../learning-roadmap.md) · [Índice ADRs](README.md)

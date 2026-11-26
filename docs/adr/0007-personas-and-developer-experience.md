# ADR-0007 — Personas, ownership federado e Developer Experience por Golden Paths

**Estado:** decisão arquitetural vigente; implementação das etapas pendente · **Data:** 21/11/2026

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
            └── rastreamento-encomendas/
                ├── domain.yaml        # identidade lógica/owner
                ├── uat/infrastructure/ # Federated DevOps
                └── prod/infrastructure/# Federated DevOps
```

Em produção, a infraestrutura por Domain **pode residir em repositórios separados**. No laboratório, **pastas do monorepo** são suficientes para simular autonomia federada sem criar repositórios extras.

**Repositório separado da aplicação de teste — Domain Applications**

```text
rastreamento-encomendas-app/                          # nome ilustrativo, ainda não criado
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

## Decisão 5 — Aplicação de referência e caminho de entrega

A POC tem como produto principal **a plataforma**, não a aplicação. O cenário de **Rastreamento de Encomendas** continua intencionalmente pequeno: uma interface registra o código de uma encomenda, a API grava o recebimento e a tela consulta o histórico.

A **stack escolhida** é **Python + FastAPI** para API/backend HTTP, **Jinja2 + HTML/CSS** para a interface server-side, e um contêiner entregue no EKS Non-PCI. **Não utilizar Django, Flask em paralelo, SPA JavaScript ou múltiplos serviços** sem necessidade. A API grava código/timestamp no **Aurora PostgreSQL** da Capability `Database` e um JSON de evidência no bucket **S3** da Capability `ObjectStorage`. Logs operacionais serão observados pela stack de telemetria, não pelo S3.

**CI/CD padronizado, com ownership da squad:**
- **Application Team:** código, testes, Dockerfile, arquivo de configuração da aplicação e manifests Kubernetes (no repositório separado da squad); decide a versão e solicita a promoção.
- **Platform Team:** template/Workflow de CI, ambiente de publicação no **Amazon ECR**, identidade IAM temporária via GitHub Actions OIDC, defaults Helm/Kustomize, políticas e Argo CD local já preparado.
- **DevOps Federado:** solicita e evolui Capabilities AWS de domínio, **não** é intermediário obrigatório para cada deploy de nova imagem.
- **Pipeline GitHub Actions:** testa, constrói, verifica segurança e publica uma imagem versionada por digest; propõe uma mudança no Git com o digest desejado, sem deploy imperativo por `kubectl apply`.
- **Argo CD local:** entrega a revisão GitOps aprovada nos destinos Kubernetes autorizados. O código/manifests podem compartilhar o mesmo repositório na POC; separar repositórios é uma evolução, não requisito.

**Entregas em momentos diferentes do roadmap:**

1. **Etapa 18 — smoke test técnico:** sincronizar e testar um workload HTTP temporário no EKS Non-PCI após o Argo CD local existir, sem exigir aplicação de negócio ou CI completa.
2. **Etapas 22–25 — primeira aplicação e CI/CD:** construir a interface/API, conectar Aurora e S3, publicar imagem no ECR e efetuar deploy pelo Argo CD local. A etapa entrega **software funcional e sua esteira**, não Developer Experience final.
3. **Etapas 32–33 — Experience Layer:** estruturar contratos, ownership, permissões e Starter Kit consumível por Git/CLI; Backstage não é obrigatório.
4. **Etapas 34–35 — Golden Path:** DevOps Federado solicita Capabilities do Domain, e a squad reutiliza a aplicação e pipeline existentes para demonstrar o percurso completo. Não criar um segundo app nem uma segunda esteira.

Não é objetivo implementar arquitetura de software complexa, frontend sofisticado, vários serviços, catálogo multilíngue nem uma plataforma proprietária de builds. A aplicação e a sua esteira existem para validar **contratos, integração e entrega de ponta a ponta**.

## Consequências e decisões pendentes para as etapas

- **Modelo de permissões:** CODEOWNERS/path rules nos exemplos; RBAC/admission/AppProjects/IAM efetivos para impedir mudanças fora do ownership.
- **Integração CI → GitOps:** na Etapa 25, implementar atualização versionada de digest por PR; na Etapa 35, demonstrar o fluxo como jornada da squad. Promoção real para PROD requer destino autorizado.
- **Golden Paths:** validar na **Etapa 35, após a Experience Layer na Etapa 33**, reutilizando a aplicação e a CI/CD das Etapas 22–25. Backstage segue opcional; **não exigir um gerador completo**.
- **Produto e métricas DX:** demonstrar o fluxo da plataforma e a redução da carga cognitiva com a aplicação mínima; métricas de geração de novos serviços são opcionais.
- **Status:** estes contratos e exemplos descrevem o desenho pretendido, não repositório da aplicação já fornecido ou componentes AWS/EKS implantados.

[Arquitetura](../architecture.md) · [Estrutura](../repository-structure.md) · [Roadmap](../learning-roadmap.md) · [Índice ADRs](README.md)

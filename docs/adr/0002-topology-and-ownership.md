# ADR-0002 — Topologia, ownership e organização do Data Plane

**Estado:** decisão vigente, implementação pendente · **Data:** 19/11/2026

**Decisão:** a taxonomia tem **Experience Layer, Control Plane, Workflow Plane, Observability Plane e Data Plane**. O Hub (`kind` inicialmente) hospeda os planos centrais. Uma **Entity** possui UAT/PROD (contas AWS distintas no alvo), cada um com três EKS — Management, PCI, Non-PCI — e AWS Runtime. Essa topologia é replicável; sete clusters é apenas a contagem de uma entidade mais Hub. Hub-and-Spoke não significa rede AWS.

As **três personas** são **Platform Team**, **Federated DevOps Team** e **Application Team**. Seus níveis de abstração diferem: a plataforma constrói os contratos e motores, o federado opera as intenções de infraestrutura do Domain e o desenvolvimento concentra-se no código e no ciclo da aplicação. Ver [ADR-0007](0007-personas-and-developer-experience.md).

As quatro camadas de ownership são **Foundation** (equipe central, Terraform), **Entity Core Infrastructure** (federado solicita sob governança, EKS via Terraform), **Domain Infrastructure** (federado, Capabilities Crossplane) e **Domain Applications** (squads, CI/CD em repositório GitHub Actions separado e Argo CD local). São responsabilidades, **não planos novos**.

**Organização Git (decisão atual):** monorepo com `data-plane/<entity>/core-infrastructure/{uat,prod}` para Core e `data-plane/<entity>/domains/<domain>/{uat,prod}/infrastructure` para Capabilities **mantidas pelo DevOps Federado**. Código, GitHub Actions e manifests de Domain Applications ficam em **repositório separado da aplicação**, sob a squad. O antigo diretório `applications/` no monorepo era apenas exemplo de estrutura-alvo, agora substituído. A identidade `domain.yaml` aparece uma vez. Foundation/config fica no escopo central. Environment representa a conta inteira, não somente o Core.

**Complemento (21/11/2026):** o Management EKS de cada Environment contempla Argo CD local, **Operators ainda a escolher** e **OpenTelemetry Collector/gateway**. Grafana, Loki e Tempo fazem parte da observabilidade centralizada. O AWS Runtime cobre Core compartilhado e recursos dos Domains, com ownership exclusivo por motor. Ver [ADR-0006](0006-terraformrun-and-application-ci.md).

**Consequências:** CODEOWNERS/paths simulam papéis, mas não substituem rulesets, RBAC e IAM. MSK fora do escopo inicial; recurso compartilhado tem lifecycle próprio. Não criar diretórios vazios prematuramente.

[Arquitetura](../architecture.md) · [Índice](README.md)

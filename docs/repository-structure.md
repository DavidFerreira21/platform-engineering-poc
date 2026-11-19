# 🗂️ Estrutura do monorepo

[← Início](../README.md) · [Arquitetura](architecture.md) · [Roadmap](learning-roadmap.md)

A POC utiliza **um único repositório** para demonstrar responsabilidades que poderiam ser separadas por equipes/repositórios reais. A estrutura-alvo está detalhada na [arquitetura](architecture.md#5-organização-do-monorepo).

## Divisão por ownership

| Caminho-alvo | Camada e gestão |
| :--- | :--- |
| `foundation/modules/` + `foundation/config/<entity>/<env>/` | **Foundation** · equipe central, Terraform reutilizável |
| `data-plane/<entity>/core-infrastructure/<env>/` | **Entity Core** · compartilhado por ambiente, EKS por Terraform |
| `data-plane/<entity>/domains/<domain>/domain.yaml` | **Domain** · registro lógico único, não CRD |
| `data-plane/<entity>/domains/<domain>/<env>/infrastructure/` | **Domain Infra** · Capabilities por Crossplane central |
| `data-plane/<entity>/domains/<domain>/<env>/applications/` | **Applications** · Argo CD local |
| `control-plane/`, `workflow-plane/`, `observability-plane/`, `platform/hub/` | Código e configuração da plataforma central |

O **Domain Bootstrap** usa um perfil versionado definido pela plataforma e gera, mediante PR/validação, manifests de Capabilities obrigatórias sob `infrastructure/` e namespaces/policies sob `applications/` (por ambiente); o Argo CD central e o local são os únicos aplicadores. O gerador é um componente do **Workflow Plane** a implementar. A identidade do Domain permanece única. [Fluxo detalhado](architecture.md#6-domain-bootstrap-criação-padronizada-do-domain).

**Environment é a conta AWS UAT/PROD inteira**, não a pasta de Core. O Core é a parte compartilhada e o Domain possui os recursos de negócio. Ambos ficam no **Data Plane**. O registro de Environment fica na Foundation por ownership central, mas descreve o destino de todos os recursos do Environment.

### Situação de migração

Os marcadores originais `data-plane/uat/` e `data-plane/prod/` são **READMEs legados, sem infraestrutura implantada**. A estrutura final será `data-plane/entity-a/core-infrastructure/{uat,prod}/`, além de `data-plane/entity-a/domains/`. Criar diretórios/manifests de infraestrutura apenas durante os labs, sem duplicar recursos ou mover infraestrutura ativa. Não criar repositórios Git separados para Entity ou Domain.

O desenho técnico da plataforma é mantido em `control-plane/` e `workflow-plane/`; o fato de um manifesto de Capability morar no Data Plane **não** move o Crossplane para lá.

[Arquitetura](architecture.md) · [Decisões vigentes](adr/README.md)

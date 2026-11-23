# 🗂️ Estrutura do monorepo

[← Início](../README.md) · [Arquitetura](architecture.md) · [Roadmap](learning-roadmap.md)

A POC utiliza **um único repositório** para demonstrar responsabilidades que poderiam ser separadas por equipes/repositórios reais. [Diagrama visual da arquitetura e das três personas](arquitetura%20Plataform.png). A estrutura-alvo está detalhada na [arquitetura](architecture.md#5-organização-do-monorepo).

## Divisão por ownership

| Caminho-alvo | Camada e gestão |
| :--- | :--- |
| `foundation/modules/` + `foundation/config/<entity>/<env>/` | **Foundation** · equipe central, Terraform reutilizável |
| `data-plane/<entity>/core-infrastructure/<env>/` | **Entity Core** · compartilhado por ambiente, EKS por Terraform |
| `data-plane/<entity>/domains/<domain>/domain.yaml` | **Domain** · registro lógico único, não CRD |
| `data-plane/<entity>/domains/<domain>/<env>/infrastructure/` | **Domain Infrastructure** · manifests de Capabilities sob gestão do DevOps Federado; Crossplane central é reconciliador |
| **Repositório próprio da aplicação (etapa futura)**: `src/`, `tests/`, `.github/workflows/`, `k8s/` | **Domain Applications** · Application Team (código, manifests K8s, CI/CD) + Argo CD local |
| `control-plane/`, `workflow-plane/`, `observability-plane/`, `platform/hub/` | Código e configuração da plataforma central |

Os registros de Domain ficam no monorepo. O DevOps Federado mantém os manifests de infraestrutura em `infrastructure/` para cada destino, e a squad mantém os manifests da aplicação no próprio repositório. Ver [ADR-0007](adr/0007-personas-and-developer-experience.md).

**Environment é a conta AWS UAT/PROD inteira**, não a pasta de Core. O Core é a parte compartilhada e o Domain possui os recursos de negócio. Ambos ficam no **Data Plane**. O registro de Environment fica na Foundation por ownership central, mas descreve o destino de todos os recursos do Environment.

### Atualizações de arquitetura-alvo (a implementar)

- **TerraformStack (CRD + controller):** contrato de infraestrutura desejada e drift em `control-plane/`, sincronizado via Argo CD central, evoluindo em 11A (deploy aprovado), 11B (detecção sem remediação) e 11C (reconciliação por políticas), com Argo Workflows + Terraform Executor em `workflow-plane/`. `TerraformRun` é nome opcional de execução individual, sem segundo CRD obrigatório. Não confundir com XRDs do Crossplane.
- **GitHub Actions das aplicações:** será mantido **em repositório de aplicação de teste separado**, a integrar durante o lab correspondente. Este monorepo não incluirá código/CI da aplicação.
- **Management EKS:** Argo CD local, Operators a escolher e OpenTelemetry Collector/gateway para observabilidade central (Grafana/Loki/Tempo).

### Três personas e a fronteira entre repositórios

**Platform Team:** mantém plataforma, Foundation, Core compartilhado, Capabilities e Golden Paths. **Federated DevOps Team:** gerencia a intenção e o lifecycle de **Domain Infrastructure** no monorepo. **Application Team:** mantém código, testes, imagens, GitHub Actions e manifests Kubernetes no **mesmo repositório de aplicação**. Os paths `applications/` no monorepo eram apenas exemplos anteriores; **não criá-los** para Application Team neste lab. Backstage/CLI/Git são canais de consumo do Starter Kit, não imposição de um único frontend. Ver [ADR-0007](adr/0007-personas-and-developer-experience.md).

### Primeira implementação

- `platform/hub/kind-config.yaml` e `platform/hub/scripts/`: criação, inspeção e remoção **apenas do cluster local kind**.
- [Guias das etapas](../labs/README.md) em `labs/`, começando pelo Hub local na Etapa 01.
- `foundation/config/entity-a/{uat,prod}/environment.yaml`: registros versionados das contas UAT e PROD com IDs distintos, região, baseline de rede e clusters EKS previstos. A [Etapa 05](../labs/05-registro-ambientes-aws.md) valida os destinos via IAM Identity Center.

### Situação de migração

Os marcadores originais `data-plane/uat/` e `data-plane/prod/` são **READMEs legados, sem infraestrutura implantada**. A estrutura final será `data-plane/entity-a/core-infrastructure/{uat,prod}/`, além de `data-plane/entity-a/domains/`. Criar diretórios/manifests de infraestrutura apenas nas etapas correspondentes, sem duplicar recursos ou mover infraestrutura ativa. Não criar repositórios Git separados para Entity ou Domain.

O desenho técnico da plataforma é mantido em `control-plane/` e `workflow-plane/`; o fato de um manifesto de Capability morar no Data Plane **não** move o Crossplane para lá.

[Arquitetura](architecture.md) · [Decisões vigentes](adr/README.md)

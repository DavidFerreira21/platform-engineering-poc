# 🗂️ Estrutura do monorepo

[← Início](../README.md) · [Arquitetura](architecture.md) · [Roadmap](learning-roadmap.md)

A POC utiliza **um único repositório** para demonstrar responsabilidades que poderiam ser separadas por equipes/repositórios reais. [Diagrama visual da arquitetura e das três personas](arquitetura%20Plataform.png). A estrutura-alvo está detalhada na [arquitetura](architecture.md#5-organização-do-monorepo).

O código é organizado pelo componente que a plataforma instala e mantém. A numeração dos guias indica a ordem de montagem. Overlays registram diferenças de destino e reutilizam os manifests base. O Hub EKS usa as raízes Terraform e os blocos organizados por componente.

## Onde encontrar cada parte da POC

| Tema | Caminho |
| :--- | :--- |
| Topologia e decisões | [Arquitetura](architecture.md) e [ADRs](adr/README.md) |
| Fluxos e contratos | [GitOps](flows/gitops.md), [TerraformStack](flows/terraformstack.md), [GitOps Bridge](flows/gitops-bridge.md) e [observabilidade](flows/observability.md) |
| Evidências e reprodução | [Estado da POC](learning-roadmap.md#estado-da-poc) e [guias das etapas](../labs/README.md) |

## Divisão por ownership

| Caminho-alvo | Camada e gestão |
| :--- | :--- |
| `foundation/config/` | YAMLs de valores por stack, da camada global à Entity e ao Environment, e registros de destino `environment.yaml` |
| `foundation/stacks/` + `foundation/modules/` | Raízes `simulator`, `network`, `eks` e `platform-hub`, com módulos Terraform reutilizáveis, incluindo `configuration` |
| `data-plane/<entity>/core-infrastructure/<env>/` | **Entity Core** · compartilhado por ambiente, EKS por Terraform |
| `data-plane/<entity>/domains/<domain>/domain.yaml` | **Domain** · registro lógico único, não CRD |
| `data-plane/<entity>/domains/<domain>/<env>/infrastructure/` | **Domain Infrastructure** · manifests de Capabilities sob gestão do DevOps Federado; Crossplane central é reconciliador |
| **Repositório próprio da aplicação (etapa futura)**: `src/`, `tests/`, `.github/workflows/`, `k8s/` | **Domain Applications** · Application Team (código, manifests K8s, CI/CD) + Argo CD local |
| `control-plane/`, `workflow-plane/`, `platform/hub/` | Código e configuração dos componentes centrais |
| [`observability-plane/`](../observability-plane/README.md) | Documentação do Observability Plane sob gestão do Platform Team. A topologia prevê consulta central e coleta e armazenamento por Environment |

Os registros de Domain ficam no monorepo. O DevOps Federado mantém os manifests de infraestrutura em `infrastructure/` para cada destino, e a squad mantém os manifests da aplicação no próprio repositório. Ver [ADR-0007](adr/0007-personas-and-developer-experience.md).

**Environment é a conta AWS UAT/PROD inteira**, não a pasta de Core. O Core é a parte compartilhada e o Domain possui os recursos de negócio. Ambos ficam no **Data Plane**. O registro de Environment fica na Foundation por ownership central, mas descreve o destino de todos os recursos do Environment.

O repositório mantém os valores em `foundation/config/`. As stacks com configuração em YAML informam a raiz de configuração e a hierarquia ao módulo `configuration`, que compõe os arquivos na ordem global, Entity e Environment. A pipeline do Simulator usa o mesmo commit para código e YAMLs de valores e registra a referência efetiva em `status.resolvedSource.reference`. Os registros `environment.yaml` descrevem o destino AWS autorizado. [Decisão sobre configuração por entidade e ambiente](adr/0006-terraformrun-and-application-ci.md#configuração-por-entidade-e-ambiente)

### Componentes do TerraformStack

- `control-plane/terraformstack/` reúne CRD, controller Go, configuração das integrações Argo, exemplos e overlay do Hub local. Em `controller/internal/`, `controller` concentra reconciliação; `execution` define o contrato e `execution/argo` implementa o adapter; `observability` expõe logs, Events e métricas de saúde sem controlar o lifecycle. O CRD e o controller têm Applications próprias. [Estrutura do controller](../control-plane/terraformstack/controller/README.md).
- `workflow-plane/argo-workflows/terraformstack/` contém o WorkflowTemplate de referência, scripts e perfil do Simulator. Integrações alternativas podem cadastrar seus próprios templates sem alterar o núcleo do controller. O runner e o Policy Engine são blocos do Workflow Plane. [Execution Contract](flows/terraformstack-execution-contract.md) · [Policy Engine](../workflow-plane/policy-engine/README.md).
- `foundation/stacks/simulator/` cria o ConfigMap do ensaio local. A raiz `platform-hub` compõe rede, EKS, Argo CD e o bucket S3 para futuras execuções do TerraformStack no bootstrap manual da Etapa 01, com state local. Nas spokes, o primeiro apply aprovado da raiz `network` está previsto na Etapa 13. [Roadmap](learning-roadmap.md).

### Três personas e a fronteira entre repositórios

**Platform Team:** mantém plataforma, Foundation, Core compartilhado, Capabilities e Golden Paths. **Federated DevOps Team:** gerencia a intenção e o lifecycle de **Domain Infrastructure** no monorepo. **Application Team:** mantém código, testes, imagens, GitHub Actions e manifests Kubernetes no **mesmo repositório de aplicação**. Os paths `applications/` no monorepo eram apenas exemplos anteriores; **não criá-los** para Application Team neste lab. Backstage/CLI/Git são canais de consumo do Starter Kit, não imposição de um único frontend. Ver [ADR-0007](adr/0007-personas-and-developer-experience.md).

### Implementação atual e aceites pendentes

- `platform/hub/addons/` reúne o bootstrap dos AppProjects e do `ApplicationSet`, além de uma entrada por add-on do Hub. O primeiro sync no EKS é validado na [Etapa 02](../labs/02-argocd-central.md).
- [Guias das etapas](../labs/README.md) em `labs/`, começando pelo Hub EKS na Etapa 01.
- `foundation/config/entity-a/{uat,prod}/environment.yaml` contém os registros de UAT e PROD com IDs distintos, região, baseline de rede e clusters EKS previstos. O roteiro usa UAT pela primeira vez na [Etapa 11](learning-roadmap.md#etapa-11--registro-de-uat). PROD entra na Etapa 31 para validar isolamento.

### Situação de migração

Os marcadores originais `data-plane/uat/` e `data-plane/prod/` são **READMEs legados, sem infraestrutura implantada**. A estrutura final será `data-plane/entity-a/core-infrastructure/{uat,prod}/`, além de `data-plane/entity-a/domains/`. Criar diretórios/manifests de infraestrutura apenas nas etapas correspondentes, sem duplicar recursos ou mover infraestrutura ativa. Não criar repositórios Git separados para Entity ou Domain.

O desenho técnico da plataforma é mantido em `control-plane/` e `workflow-plane/`; o fato de um manifesto de Capability morar no Data Plane **não** move o Crossplane para lá.

[Arquitetura](architecture.md) · [Decisões vigentes](adr/README.md)

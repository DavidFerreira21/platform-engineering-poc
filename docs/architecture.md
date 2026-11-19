# Arquitetura da POC de Platform Engineering

**Situação:** arquitetura-alvo, não totalmente implementada. **Escopo:** dois ambientes — UAT e PROD — sem países, entidades ou referências corporativas.

## Camadas arquiteturais

```text
EXPERIENCE LAYER
  Backstage | CLI | GitOps | AI/MCP
                    |
PLATFORM MANAGEMENT CLUSTER
  CONTROL PLANE
    Kubernetes API Server
    CRDs: AWSEnvironment | Domain | Capabilities especializadas
    Platform Controllers, RBAC, admissão, reconciliação e status
  WORKFLOW PLANE
    Argo CD | Argo Workflows | Terraform Executor
    Gates manuais | Crossplane (introdução progressiva)
  OBSERVABILITY PLANE
    OpenTelemetry | métricas | logs | traces | auditoria
                    |
DATA PLANE
  UAT — conta AWS                PROD — conta AWS
    Environment Management EKS     Environment Management EKS
      Argo CD / Operators / OTel     Argo CD / Operators / OTel
    EKS PCI                        EKS PCI
    EKS Non-PCI                    EKS Non-PCI
    AWS Runtime                    AWS Runtime
      RDS, S3, EC2, VPC etc.         RDS, S3, EC2, VPC etc.
```

**Control Plane, Workflow Plane, Observability Plane e Data Plane são nomes de responsabilidade arquitetural.** Os três primeiros compartilham um Platform Management Cluster no alvo. Cada ambiente do Data Plane mantém seu gerenciamento local e seus workloads. A nomenclatura Hub-and-Spoke é uma **interpretação complementar da topologia de gerenciamento**, não o nome de uma camada nem uma configuração obrigatória de rede AWS.

A arquitetura física-alvo prevê **sete clusters Kubernetes** (um central, três UAT e três PROD). Não provisionar os sete apenas para iniciar o lab: começar com `kind`, mocks e representação progressiva dos papéis.

## Contratos Kubernetes-native
O **Kubernetes API Server** é a API oficial da plataforma, com três famílias de contratos:
- **`AWSEnvironment`:** representa uma conta AWS e seu ciclo de vida: criação/adoção, bootstrap, rede, segurança e clusters. Uma conta por ambiente; exclusão do CR nunca implica exclusão automática da conta.
- **`Domain`:** representa o ecossistema de aplicações/serviços e recursos compartilhados de um domínio de negócio. Vários Domains podem existir no mesmo Environment.
- **Capabilities especializadas:** `Database`, `ObjectStorage`, `ComputeInstance`, `Messaging` e outros tipos tipados. Cada CR especializado representa a própria instância consumida. **Não utilizar `CapabilityClaim` genérico na versão inicial.**

Os CRDs são definições de tipos; Custom Resources são instâncias. Relações por `environmentRef` e `domainRef` são validadas por controllers/políticas — Kubernetes não garante integridade referencial automaticamente.

Exemplo **ilustrativo, ainda não instalado**:

```yaml
apiVersion: platform.lab.example/v1alpha1
kind: AWSEnvironment
metadata:
  name: uat
spec:
  stage: uat
  profile: standard
  clusters:
    management: true
    pci: true
    nonPCI: true
```

## Experience Layer e GitOps
O Git é fonte da verdade para **estado desejado declarativo**. O Argo CD sincroniza recursos com a Kubernetes API; `status` e resultados operacionais permanecem no Kubernetes e nos sistemas de execução. Backstage, CLI e AI/MCP devem preferencialmente produzir PRs para mudanças persistentes. Consultas podem usar a API diretamente com RBAC. Evitar concorrência de escrita sobre o `spec` entre GitOps e `kubectl`.

## Fluxo de provisionamento AWSEnvironment
1. O Environment Controller observa o recurso e valida autenticação, autorização, parâmetros e identidade da conta.
2. Aciona um **Argo Workflow versionado**, mantendo as etapas técnicas fora do contrato público.
3. O workflow cria/adota a conta, aguarda IPs quando necessário, executa Terraform para bootstrap/rede, valida subnets, rotas e conectividade e aguarda certificados ou aprovação humana.
4. Gates manuais usam suspensão/retomada com autorização, identidade, justificativa e evidência. Abrir um chamado não comprova execução nem substitui validações técnicas.
5. O fluxo prepara os clusters e atualiza condições como `AccountReady`, `NetworkReady`, `SecurityReady`, `ClustersReady` e `Ready`. Terraform state e logs não vão para o Git nem para o etcd.
6. Exclusão de conta requer processo deliberado e separado de descomissionamento.

## Operação de cada ambiente
O Environment Management EKS é candidato a hospedar Argo CD local, controllers/operators aptos a reconciliação remota e gateway de observabilidade. O Argo CD local gerencia EKS PCI e EKS Non-PCI **somente se o isolamento administrativo PCI permitir**. Operators/agents que exigem acesso local permanecem nos clusters de workload. O AWS Runtime (RDS, S3, EC2, networking etc.) pertence à mesma conta AWS do respectivo ambiente.

## Limites do desenho
A indisponibilidade do Hub não deve automaticamente interromper workloads já existentes. O bootstrap e a recuperação do gerenciamento exigem plano próprio. Namespaces não oferecem isolamento equivalente a contas/clusters. Crossplane e Terraform não devem possuir simultaneamente o mesmo recurso AWS. Esta POC não implementa nem integra chamados reais automaticamente.

Leitura complementar: [ADRs](adr/README.md), [estrutura do monorepo](repository-structure.md) e [roteiro de aprendizado](learning-roadmap.md).

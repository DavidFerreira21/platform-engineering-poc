# AGENTS.md — Instruções aos agentes

## Idioma
**Português brasileiro (pt-BR) é obrigatório** para README, ADRs, arquitetura, roadmap, tutoriais, descrições de PR e explicações. Preservar nomes técnicos/identificadores: Kubernetes, Argo CD, `AWSEnvironment`, `spec`, `status`, `kind`, nomes de pastas e comandos.

## Fonte de verdade
Ler [arquitetura](docs/architecture.md), [ADRs](docs/adr/README.md) e [estrutura do repositório](docs/repository-structure.md) antes de implementar. Esta baseline substitui os ADRs antigos baseados em FastAPI como API principal e em apenas dois clusters kind.

## Arquitetura obrigatória
- **Experience Layer:** Backstage, CLI, GitOps e AI/MCP. Para alterações persistentes, preferir PR no Git; leitura de estado pode acessar Kubernetes API sob RBAC.
- **Control Plane:** Kubernetes API Server como API da plataforma, CRD nativo `AWSEnvironment` com Environment Controller próprio, XRDs `Domain` e Capabilities com Crossplane Core/Compositions/Providers, validação e status.
- **Famílias de CRDs:** `AWSEnvironment` (uma conta AWS + foundation), `Domain` (ecossistema de aplicações) e Capabilities especializadas (por exemplo `Database`, `ObjectStorage`, `ComputeInstance`). Capability é categoria, não necessariamente um recurso de kind Capability. Não criar Claim genérico na primeira versão.
- **Workflow Plane:** Argo Workflows, Terraform Executor, Argo CD e gates manuais; Crossplane pertence ao Control Plane, não ao Workflow Plane. Não usar Argo CD para executar etapas imperativas Terraform.
- **Observability Plane:** telemetria, métricas, logs, traces, eventos e auditoria transversal.
- **Data Plane:** dois ambientes, UAT e PROD, cada qual uma conta AWS com Environment Management EKS, EKS PCI e EKS Non-PCI usando **Amazon EKS Auto Mode por padrão**, além de AWS Runtime. O Hub continua em kind local na primeira fase. Não incluir países nem entidades.
- **Platform Management Cluster:** hospeda Control, Workflow e Observability Planes; estes são agrupamentos lógicos, não três clusters.
- **Hub-and-Spoke:** apenas conceito de topologia de gerenciamento, documentado sem substituir os nomes das camadas. Não implica topologia de rede ou Transit Gateway.
- Operadores podem exigir execução no próprio cluster de workload. Compartilhar controle entre PCI e Non-PCI depende da avaliação de segurança.

## Responsabilidades do monorepo
- `control-plane/`: CRD/controller AWSEnvironment, XRDs/Compositions/Providers Crossplane para Domain/Capabilities e políticas.
- `workflow-plane/`: Argo Workflows, Terraform Executor, gates e GitOps, sem Crossplane.
- `observability-plane/`: configurações transversais de telemetria.
- `platform/hub/`: bootstrap do Platform Management Cluster.
- `data-plane/uat/`, `data-plane/prod/`: configuração específica do ambiente e de seus clusters e AWS Runtime.
- `docs/`, `labs/`, `tests/`: documentação, aprendizado e validação.

## Qualidade e segurança
Implementar progressivamente, com testes observáveis, falhas controladas e limpeza. Usar mocks para testes automatizados e validação inicial, mas implementar **provisionamento AWS real** nos labs autorizados. Hub local pode usar kind. Confirmar conta e região com STS; usar sessões/roles temporárias, least privilege, plano, estimativa de custos, autorização por execução, state protegido e cleanup. Nunca publicar credenciais, IDs reais, certificados privados, estado Terraform ou dados corporativos. Mudanças no Git devem ocorrer via branches e PRs. Reconciliação idempotente, status confiável, aprovações autorizadas, auditoria e proteção contra exclusão destrutiva são requisitos. Distinguir claramente **planejado**, **mockado**, **implementado** e **adiado**.

## Fases aprovadas para a arquitetura
1. Conta AWS real preexistente e autenticada para Capabilities, sem criar conta.
2. UAT e PROD em contas reais preexistentes e separadas.
3. Automação de criação de contas com AWSEnvironment como evolução futura.

O escopo real de cada execução AWS deve ser aprovado antes de qualquer apply. Crossplane é escolhido desde já para Domain/Capabilities; não manter referência a Crossplane como motor opcional do Workflow Plane. Ler [ADR-0009](docs/adr/0009-crossplane-e-reconciliacao.md) e [ADR-0010](docs/adr/0010-aws-real-e-ambientes.md).

## Padrão de clusters EKS
Usar **Amazon EKS Auto Mode** como padrão para clusters reais nos diretórios `data-plane/{uat,prod}/{management-eks,eks-pci,eks-non-pci}`. Não instalar Karpenter autogerenciado, AWS Load Balancer Controller ou drivers de storage equivalentes indiscriminadamente quando essas funções forem gerenciadas pelo Auto Mode. Validar compatibilidade de operators, DaemonSets, Pod Identity/IAM, políticas, requisitos PCI, custos e controles de nó antes de executar. O Hub local em `kind` não é EKS Auto Mode. Ver [ADR-0011](docs/adr/0011-eks-auto-mode.md).

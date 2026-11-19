# 📝 Architecture Decision Records

Os ADRs registram **por que** a arquitetura foi escolhida, quais alternativas e consequências foram consideradas e o que ainda exige validação.

> **Situação:** decisões de arquitetura-alvo. Um ADR aceito não significa componente instalado ou recurso AWS provisionado.

## 📚 Índice de decisões

| ADR | Tema |
| :--- | :--- |
| [**0001**](0001-kubernetes-native-platform-api.md) | Kubernetes API como API oficial |
| [**0002**](0002-contract-families.md) | Três famílias de contratos |
| [**0003**](0003-hub-spoke-topology.md) | Camadas e topologia Hub-and-Spoke |
| [**0004**](0004-workflow-and-gitops.md) | GitOps e workflows com responsabilidades distintas |
| [**0005**](0005-long-running-environment-lifecycle.md) | Lifecycle de AWSEnvironment e gates |
| [**0006**](0006-observability-and-local-operators.md) | Observabilidade e operators locais |
| [**0007**](0007-security-and-isolation.md) | Segurança, isolamento e PCI |
| [**0008**](0008-monorepo-and-learning.md) | Monorepo e aprendizado incremental |
| [**0009**](0009-crossplane-e-reconciliacao.md) | Crossplane e ownership de recursos |
| [**0010**](0010-aws-real-e-ambientes.md) | AWS real em três etapas |
| [**0011**](0011-eks-auto-mode.md) | EKS Auto Mode nos Data Planes |
| [**0012**](0012-federated-ownership-and-gitops.md) | Quatro camadas, ownership federado, Crossplane central e GitOps local |

## 🔎 Como consultar

- **Contratos e reconciliação:** [0001](0001-kubernetes-native-platform-api.md), [0002](0002-contract-families.md), [0009](0009-crossplane-e-reconciliacao.md)
- **Execução e lifecycle:** [0004](0004-workflow-and-gitops.md), [0005](0005-long-running-environment-lifecycle.md), [0010](0010-aws-real-e-ambientes.md)
- **Topologia e operação:** [0003](0003-hub-spoke-topology.md), [0006](0006-observability-and-local-operators.md), [0007](0007-security-and-isolation.md), [0011](0011-eks-auto-mode.md)
- **Organização:** [0008](0008-monorepo-and-learning.md), [0012](0012-federated-ownership-and-gitops.md)

## 📌 Convenções

Textos em **pt-BR**; nomes técnicos preservados. A taxonomia de referência é **Experience Layer, Control Plane, Workflow Plane, Observability Plane e Data Plane**. Hub-and-Spoke representa a topologia de gerenciamento, **não uma camada adicional** nem uma topologia de rede obrigatória.

[← Arquitetura](../architecture.md) · [Estrutura](../repository-structure.md) · [Página inicial](../../README.md)

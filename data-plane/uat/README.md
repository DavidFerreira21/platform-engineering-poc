# 🧪 UAT · Entity Core (marcador legado)

Marcador de documentação do **Core Infrastructure** do Environment UAT, associado à sua própria conta AWS. A estrutura abaixo está planejada, ainda não provisionada. No modelo vigente, Core ficará em `data-plane/entity-a/core-infrastructure/uat/`, e Domains estarão em `data-plane/entity-a/domains/`. Environment inclui ambos os escopos.

| Componente | Papel |
| :--- | :--- |
| [🏗️ Environment Management EKS](management-eks/README.md) | Argo CD local, operators e telemetria |
| [🔒 EKS PCI](eks-pci/README.md) | Workloads PCI |
| [📦 EKS Non-PCI](eks-non-pci/README.md) | Workloads não PCI |
| [☁️ AWS Runtime](aws-runtime/README.md) | S3, RDS, VPC e serviços gerenciados |

**EKS Auto Mode** será o padrão dos três clusters Kubernetes. Avaliar o controle administrativo entre PCI e Non-PCI.

> [!NOTE]
> **Situação:** estrutura de documentação e planejamento. Nenhum componente desta pasta está implantado por padrão.

---

[← Página inicial](../../README.md) · [Arquitetura](../../docs/architecture.md) · [Roteiro prático](../../docs/learning-roadmap.md)

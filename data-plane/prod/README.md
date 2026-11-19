# 🚀 PROD · Conta AWS

Ambiente **PROD**, representado por uma conta AWS independente de UAT. A estrutura abaixo é o **alvo**, ainda não provisionado.

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

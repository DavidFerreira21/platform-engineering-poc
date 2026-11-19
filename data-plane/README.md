# ☁️ Data Plane

O Data Plane reúne **dois ambientes isolados**, cada um associado a uma **conta AWS própria** na arquitetura-alvo. A implementação é gradual; nenhum ambiente está provisionado automaticamente por esta documentação.

| Ambiente | Gerenciamento local | Workloads | Serviços AWS |
| :--- | :--- | :--- | :--- |
| [**UAT**](uat/README.md) | [Management EKS](uat/management-eks/README.md) | [PCI](uat/eks-pci/README.md) · [Non-PCI](uat/eks-non-pci/README.md) | [AWS Runtime](uat/aws-runtime/README.md) |
| [**PROD**](prod/README.md) | [Management EKS](prod/management-eks/README.md) | [PCI](prod/eks-pci/README.md) · [Non-PCI](prod/eks-non-pci/README.md) | [AWS Runtime](prod/aws-runtime/README.md) |

Os **seis clusters EKS** previstos nos dois ambientes adotam **EKS Auto Mode** inicialmente, sujeitos a custo, compatibilidade e requisitos de segurança.

> [!NOTE]
> **Situação:** estrutura de documentação e planejamento. Nenhum componente desta pasta está implantado por padrão.

---

[← Página inicial](../README.md) · [Arquitetura](../docs/architecture.md) · [Roteiro prático](../docs/learning-roadmap.md)

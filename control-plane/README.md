# 🧩 Control Plane

O **Kubernetes API Server** é a API oficial; os reconciliadores implementam os contratos sem criar APIs paralelas.

| Contrato | Responsabilidade |
| :--- | :--- |
| `AWSEnvironment` | CRD nativo + Environment Controller próprio (a implementar) |
| `Domain` | XRD + Crossplane Composition (a implementar) |
| `Database`, `ObjectStorage`, `ComputeInstance`, `Messaging` | XRDs + Crossplane Compositions/Providers (a implementar) |

**Atenção:** Domain remoto e permissões exigem validação. Criação automática de conta AWS é etapa futura.

> [!NOTE]
> **Situação:** estrutura de documentação e planejamento. Nenhum componente desta pasta está implantado por padrão.

---

[← Página inicial](../README.md) · [Arquitetura](../docs/architecture.md) · [Roteiro prático](../docs/learning-roadmap.md)

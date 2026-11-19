# ⚙️ Workflow Plane

Coordena **fluxos, aprovações e entrega GitOps**, sem substituir a reconciliação nativa do Crossplane.

| Componente | Papel |
| :--- | :--- |
| Argo Workflows | Etapas, esperas e gates autorizados |
| Terraform Executor | `plan`/`apply` controlados e state protegido |
| Argo CD | Sincronização declarativa de Git para Kubernetes |

O Environment Controller pode iniciar workflows, mas pertence ao **Control Plane**. Crossplane também pertence ao Control Plane.

> [!NOTE]
> **Situação:** estrutura de documentação e planejamento. Nenhum componente desta pasta está implantado por padrão.

---

[← Página inicial](../README.md) · [Arquitetura](../docs/architecture.md) · [Roteiro prático](../docs/learning-roadmap.md)

# 🔒 UAT · EKS PCI

Cluster de **workloads PCI**, com controle de acesso e isolamento próprios. Integrações com ferramentas centralizadas dependem de validação de segurança.

| Atributo | Definição |
| :--- | :--- |
| Ambiente | **UAT** — conta AWS independente |
| Situação | **Planejado; não provisionado** |
| Tecnologia | Amazon EKS Auto Mode (padrão inicial) |
| Validação pendente | Políticas de segurança, DaemonSets, compliance e custo |

Nenhuma execução AWS é autorizada apenas por este documento. Recursos devem ser provisionados com conta/região confirmadas, aprovação e plano de limpeza.

---

[← UAT](../README.md) · [Data Plane](../../README.md) · [Arquitetura](../../../docs/architecture.md) · [ADR EKS Auto Mode](../../../docs/adr/0011-eks-auto-mode.md)

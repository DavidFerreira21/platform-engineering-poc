# ☁️ PROD · AWS Runtime

Serviços **AWS Runtime** da conta, como VPC, S3, RDS/Aurora e mensageria, com ownership explícito entre Terraform e Crossplane.

| Atributo | Definição |
| :--- | :--- |
| Ambiente | **PROD** — conta AWS independente |
| Situação | **Planejado; não provisionado** |
| Tecnologia | Serviços AWS gerenciados |
| Validação pendente | Políticas, ownership, custos, exclusão e drift |

Nenhuma execução AWS é autorizada apenas por este documento. Recursos devem ser provisionados com conta/região confirmadas, aprovação e plano de limpeza.

---

[← PROD](../README.md) · [Data Plane](../../README.md) · [Arquitetura](../../../docs/architecture.md) · [ADR EKS Auto Mode](../../../docs/adr/0011-eks-auto-mode.md)

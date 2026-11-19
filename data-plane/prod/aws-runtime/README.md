# ☁️ PROD · AWS Runtime

Inventário **conceitual** de recursos AWS no Environment. **Foundation** (por exemplo VPC/rede) é central; **Core** contém recursos comuns; **Domain Infrastructure** possui Capabilities como S3/RDS. Terraform e Crossplane possuem recursos distintos, sem disputa.

| Atributo | Definição |
| :--- | :--- |
| Ambiente | **PROD** — conta AWS independente |
| Situação | **Planejado; não provisionado** |
| Tecnologia | Serviços AWS gerenciados |
| Validação pendente | Políticas, ownership, custos, exclusão e drift |

Nenhuma execução AWS é autorizada apenas por este documento. Recursos devem ser provisionados com conta/região confirmadas, aprovação e plano de limpeza.

---

[← PROD](../README.md) · [Data Plane](../../README.md) · [Arquitetura](../../../docs/architecture.md)

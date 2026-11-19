# 📦 UAT · EKS Non-PCI

Cluster de **workloads não PCI**, separado do cluster PCI e sujeito a políticas próprias de acesso, operação e observabilidade.

| Atributo | Definição |
| :--- | :--- |
| Ambiente | **UAT** — conta AWS independente |
| Situação | **Planejado; não provisionado** |
| Tecnologia | Amazon EKS Auto Mode (padrão inicial) |
| Validação pendente | Identidade, rede, políticas e custo |

Nenhuma execução AWS é autorizada apenas por este documento. Recursos devem ser provisionados com conta/região confirmadas, aprovação e plano de limpeza.

---

[← UAT](../README.md) · [Data Plane](../../README.md) · [Arquitetura](../../../docs/architecture.md) · [ADR AWS e EKS Auto Mode](../../../docs/adr/0004-aws-security-and-roadmap.md)

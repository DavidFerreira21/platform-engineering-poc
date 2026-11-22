# 📊 Observability Plane

## Stack-alvo da POC

- **OpenTelemetry** — instrumentação, coleta, normalização e encaminhamento de telemetria.
- **Grafana** — visualização, correlação e dashboards centralizados.
- **Loki** — backend central de logs.
- **Tempo** — backend central de traces.
- **Métricas / Prometheus** — parte do desenho de referência, com arquitetura operacional e retenção a validar nos labs.
- **Eventos e auditoria** — correlacionar Git, Argo CD central/local, Crossplane, TerraformRun/Argo Workflows e AWS.

Cada **Management EKS** de UAT/PROD deverá ter um **OpenTelemetry Collector/gateway** (ou mecanismo equivalente a validar) para encaminhar a telemetria. **Operators locais permanecem a definir**. O plano central terá visibilidade das entidades, preservando segregação e permissões de ambientes e PCI/Non-PCI. Isso não exige executar Loki/Tempo dentro de cada Management EKS.

## Etapas de aprendizado

Validar primeiro pipelines e coleta locais em `kind`, depois instrumentar os spokes, autenticação, retenção, falhas, dashboards e custos na AWS. Os componentes acima representam **arquitetura-alvo**, não stack já implantada.

[← Página inicial](../README.md) · [Arquitetura](../docs/architecture.md) · [Roteiro](../docs/learning-roadmap.md)

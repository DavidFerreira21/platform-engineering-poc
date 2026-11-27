# 📊 Observability Plane

## Stack-alvo da POC

- **OpenTelemetry** — instrumentação, coleta, normalização e encaminhamento de telemetria.
- **Grafana** — visualização, correlação e dashboards centralizados.
- **Loki** — backend central de logs.
- **Tempo** — backend central de traces.
- **Métricas / Prometheus** — parte do desenho de referência, com arquitetura operacional e retenção a validar nos labs.
- **Eventos e auditoria** — correlacionar revisão Git, Argo CD, `TerraformStack.status`, Workflow, executores, Crossplane e eventos AWS.

Cada **Management EKS** de UAT/PROD deverá ter um **OpenTelemetry Collector/gateway** (ou mecanismo equivalente a validar) para encaminhar a telemetria. **Operators locais permanecem a definir**. O plano central terá visibilidade das entidades, preservando segregação e permissões de ambientes e PCI/Non-PCI. Isso não exige executar Loki/Tempo dentro de cada Management EKS.

## Etapas de aprendizado

No Hub `kind`, o Simulator correlaciona revisão Git, Workflow, condições, horários de drift e auditoria da aprovação na API Kubernetes. A instalação dos Collectors, dos backends centrais e a instrumentação dos spokes entram nas Etapas 26 e 27. [Evidências locais](../labs/06-terraformstack-simulator.md#evidências-de-aceite).

[← Página inicial](../README.md) · [Arquitetura](../docs/architecture.md) · [Roteiro](../docs/learning-roadmap.md)

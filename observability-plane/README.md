# 📊 Observability Plane

## Stack-alvo da POC

- **OpenTelemetry** — instrumentação, coleta, normalização e encaminhamento de telemetria.
- **Grafana** — visualização, correlação e dashboards centralizados.
- **Loki** — backend central de logs.
- **Tempo** — backend central de traces.
- **Métricas / Prometheus** — parte do desenho de referência, com arquitetura operacional e retenção a validar nos labs.
- **Eventos e auditoria** — correlacionar revisão Git, Argo CD, `TerraformStack.status`, Workflow, executores, Crossplane e eventos AWS.

Cada **Management EKS** de UAT/PROD deverá ter um **OpenTelemetry Collector/gateway** (ou mecanismo equivalente a validar) para encaminhar a telemetria. **Operators locais permanecem a definir**. O plano central terá visibilidade das entidades, preservando segregação e permissões de ambientes e PCI/Non-PCI. Isso não exige executar Loki/Tempo dentro de cada Management EKS.

## Sinais já disponíveis no TerraformStack

O controller publica logs estruturados JSON em stdout, Events na API Kubernetes e métricas nativas/customizadas em `:8080/metrics`, incluindo condição, fase, drift pendente, horários e tentativas por CR. Os gauges são derivados do cache de status; contadores de transições são locais ao processo. O controlador não instala coletores, backends, ServiceMonitor, dashboards ou regras de alerta. A ausência de sinal de drift não significa convergência; falhas internas antes de persistir status exigem também logs e métricas do runtime. O operador decide como coletar e proteger o endpoint. [Catálogo de métricas e diagnóstico](../control-plane/terraformstack/controller/README.md#observabilidade-operacional).

## Etapas de aprendizado

No Hub `kind`, o Simulator correlaciona revisão Git, Workflow, condições, horários de drift e auditoria da aprovação na API Kubernetes. A integração da coleta dos sinais existentes, a instalação dos Collectors, dos backends centrais e a instrumentação dos spokes entram nas Etapas 26 e 27. [Evidências locais](../labs/06-terraformstack-simulator.md#evidências-de-aceite).

[← Página inicial](../README.md) · [Arquitetura](../docs/architecture.md) · [Roteiro](../docs/learning-roadmap.md)

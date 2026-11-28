# 📊 Observability Plane

[← Início](../README.md) · [Arquitetura](../docs/architecture.md) · [Fluxo de observabilidade](../docs/flows/observability.md) · [ADR-0010](../docs/adr/0010-observability-topology.md)

## Topologia-alvo

A conta da Plataforma oferece consulta central com **Amazon Managed Grafana** e usa **CloudWatch** para os sinais essenciais do Hub. Cada conta UAT ou PROD tem seus próprios destinos de armazenamento. Os EKS Management, PCI e Non-PCI coletam a telemetria local com OpenTelemetry. A ativação do perfil PCI depende da validação da fronteira de isolamento.

| Sinal | Destino no Environment |
| :--- | :--- |
| Métricas | Cada OTel Collector envia diretamente ao **Amazon Managed Service for Prometheus (AMP)** do ambiente. |
| Logs | Gateway no Management EKS → **Loki** no mesmo cluster → **S3** do ambiente. |
| Traces | Gateway no Management EKS → **Tempo** no mesmo cluster → **S3** do ambiente. |

O Management EKS também gera telemetria. Seu Collector coleta métricas do Argo CD local e de outros add-ons e envia ao AMP. O Gateway recebe logs e traces locais e dos clusters de workloads. O Grafana consulta AMP via IAM cross-account e Loki/Tempo via PrivateLink.

## Sinais já disponíveis no TerraformStack

O controller publica logs estruturados JSON em stdout, Events na API Kubernetes e métricas nativas/customizadas em `:8080/metrics`, incluindo condição, fase, drift pendente, horários e tentativas por CR. Os gauges são derivados do cache de status e os contadores de transições são locais ao processo. Correlacionar ausência de sinal de drift com status, logs e métricas de runtime para diagnosticar falhas anteriores à persistência do status. [Catálogo de métricas e diagnóstico](../control-plane/terraformstack/controller/README.md#observabilidade-operacional)

## Etapas de aprendizado

No Hub `kind`, o Simulator correlaciona revisão Git, Workflow, condições, horários de drift e auditoria da aprovação na API Kubernetes. Na AWS, as Etapas 26–27 integram coletores, backends, sinais do Hub e diagnóstico de perda de ingestão em UAT. As Etapas 30–31 validam as fronteiras PCI e PROD quando esses destinos forem autorizados e provisionados. [Evidências locais](../labs/06-terraformstack-simulator.md#evidências-de-aceite) · [Roteiro](../docs/learning-roadmap.md#fase-6--observabilidade-operação-e-governança)

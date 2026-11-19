# ADR-0006: Observabilidade transversal e controllers distribuídos

- **Situação:** Aceito como arquitetura-alvo da POC
- **Data:** 18/11/2026

## Contexto
Telemetria e automações locais podem depender do isolamento e da conectividade de cada ambiente.

## Decisão
O Observability Plane abrange Hub e Data Plane, com OpenTelemetry, métricas, logs, traces, alertas e auditoria. O Environment Management EKS pode hospedar Argo CD, gateways e operators compatíveis com gerenciamento remoto; operators que exigem execução local permanecem nos clusters PCI/Non-PCI.

## Consequências e pontos de atenção
Validar compatibilidade, permissões e regras PCI de cada operator; separar coleta, processamento e armazenamento de dados de telemetria.

## Validação
Validar a decisão em labs incrementais e atualizar a documentação se surgir um requisito incompatível. Não executar mudanças reais na AWS sem autorização.

# ADR-0005: Ciclo longo de AWSEnvironment com aprovações

- **Situação:** Aceito como arquitetura-alvo da POC
- **Data:** 18/11/2026

## Contexto
Criar uma conta e deixá-la pronta exige etapas como IPs, redes, rotas, firewall e certificados, inclusive por equipes externas sem integração automatizada.

## Decisão
AWSEnvironment declara o resultado desejado; um controller invoca workflow versionado no Argo Workflows. Etapas automáticas usam Terraform e validações técnicas com timeout; etapas humanas usam suspend/resume e autorização, justificativa e evidência. O controller expõe AccountReady, NetworkReady, SecurityReady, ClustersReady e Ready. Abrir chamado não equivale a validá-lo.

## Consequências e pontos de atenção
Exigir idempotência, recuperação, auditoria, proteção de exclusão e distinção entre solicitação registrada e dependência atendida. Integração real com sistemas de chamados não faz parte do início da POC.

## Validação
Validar a decisão em labs incrementais e atualizar a documentação se surgir um requisito incompatível. Não executar mudanças reais na AWS sem autorização.

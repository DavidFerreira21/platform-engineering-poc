# ADR-0001: Kubernetes API Server como API da plataforma

- **Situação:** Aceito como arquitetura-alvo da POC
- **Data:** 18/11/2026

## Contexto
A baseline anterior previa uma REST API própria baseada em FastAPI, mas a nova arquitetura pretende aproveitar contratos nativos do Kubernetes.

## Decisão
O Kubernetes API Server do Platform Management Cluster é a API oficial. CRDs, RBAC, admission e Platform Controllers implementam contratos declarativos, validação e reconciliação. Clientes autorizados acessam a API diretamente ou por adaptadores, sem uma REST API proprietária obrigatória.

## Consequências e pontos de atenção
É necessário versionar schemas, controlar permissões, definir condições de status e políticas de upgrade. O Kubernetes não oferece automaticamente lógica de negócio para os CRDs.

## Validação
Validar a decisão em labs incrementais e atualizar a documentação se surgir um requisito incompatível. Não executar mudanças reais na AWS sem autorização.

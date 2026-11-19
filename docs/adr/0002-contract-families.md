# ADR-0002: Três famílias de contratos Kubernetes

- **Situação:** Aceito como arquitetura-alvo da POC
- **Data:** 18/11/2026

## Contexto
Precisamos representar ambientes AWS, ecossistemas de negócio e produtos da plataforma sem acoplar todas as solicitações em um CRD genérico.

## Decisão
As três famílias são AWSEnvironment (uma conta AWS e sua fundação), Domain (ecossistema de aplicações e serviços) e Capabilities especializadas (Database, ObjectStorage, ComputeInstance, Messaging). A própria instância do CRD especializado representa o consumo; não usar CapabilityClaim genérico inicialmente. Referências entre objetos são explícitas e validadas pelos controllers.

## Consequências e pontos de atenção
São necessários schemas próprios por Capability, regras claras de lifecycle e validação de referências. Uma família lógica não precisa corresponder a um único CRD.

## Validação
Validar a decisão em labs incrementais e atualizar a documentação se surgir um requisito incompatível. Não executar mudanças reais na AWS sem autorização.

## Evolução da implementação (ADR-0009)

`AWSEnvironment` será CRD Kubernetes nativo com Environment Controller próprio; `Domain` e cada Capability serão XRDs Crossplane que expõem recursos pela **mesma Kubernetes API**. Os recursos Crossplane serão implementados via Compositions e Providers. A criação de contas será posterior ao suporte a contas AWS preexistentes. Consultar [ADR-0009](0009-crossplane-e-reconciliacao.md) e [ADR-0010](0010-aws-real-e-ambientes.md).

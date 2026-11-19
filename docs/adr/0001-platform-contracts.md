# ADR-0001 — API e modelo de contratos

**Estado:** decisão vigente, implementação pendente · **Data:** 19/11/2026

**Decisão:** Kubernetes API Server é a API oficial dos contratos técnicos Kubernetes-native; não exigir uma REST API paralela. **Entity** é identidade organizacional; **Environment** é destino UAT/PROD com conta AWS; **Domain** é unidade lógica de ownership, negócio e consumo de infra/aplicações. Os três começam como registros YAML validados no Git, **não CRDs/XRDs**. **Capability** é produto técnico com XRD/Composition/Provider especializado em Crossplane, começando por `ObjectStorage`. Não criar `CapabilityClaim` genérico.

**Por quê:** negócio/organização não exige reconciliador próprio. Namespace é materialização local do Domain; Domain não provisiona automaticamente todas as suas aplicações ou bancos. Um mesmo Domain pode consumir Capabilities em UAT e PROD.

**Consequências:** o pipeline precisa validar Domain/Environment e resolver conta/IAM autorizada antes da aplicação de uma Capability. YAML de registro não é manifest Kubernetes; Argo CD central sincroniza somente CRs de produtos. Um eventual contrato/controller para lifecycle de contas deve ser justificado futuramente.

[Arquitetura](../architecture.md) · [Índice](README.md)

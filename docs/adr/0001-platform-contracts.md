# ADR-0001 — API e modelo de contratos

**Estado:** decisão vigente, implementação pendente · **Data:** 19/11/2026

**Decisão:** Kubernetes API Server é a API oficial dos contratos técnicos Kubernetes-native; não exigir uma REST API paralela. **Entity** é identidade organizacional; **Environment** é destino UAT/PROD com conta AWS; **Domain** é unidade lógica de ownership, negócio e consumo de infra/aplicações. Os três começam como registros YAML validados no Git, **não CRDs/XRDs**. **Capability** é produto técnico com XRD/Composition/Provider especializado em Crossplane, começando por `ObjectStorage`. Não criar `CapabilityClaim` genérico.

**Por quê.** O registro lógico define identidade e ownership. O namespace materializa o Domain no cluster autorizado. Um mesmo Domain pode solicitar Capabilities em UAT e PROD. Cada solicitação de infraestrutura e entrega de aplicação segue seu fluxo de revisão e reconciliação.

**Consequências.** O pipeline valida o Environment e resolve a conta e a identidade autorizadas antes de aplicar uma Capability. Solicitações federadas também validam o Domain e seu ownership. Os registros YAML ficam no Git para consulta. O Argo CD central sincroniza os CRs Kubernetes. Um controller de lifecycle de contas depende de uma necessidade futura.

**Sequência.** O `TerraformStack` usa dados fictícios nas Etapas 04 e 05. A referência lógica a Entity e Environment passa a ser resolvida pelo executor na Etapa 12, a partir do registro de UAT validado na Etapa 11. As primeiras Capabilities são exercícios do Platform Team em UAT. O contrato de Domain, incluindo ownership e Environments permitidos, entra na Etapa 34, antes da primeira solicitação federada. [Roteiro](../learning-roadmap.md)

[Arquitetura](../architecture.md) · [Índice](README.md)

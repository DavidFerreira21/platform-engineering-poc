# ADR-0001 — API e modelo de contratos

**Estado:** decisão vigente, implementação pendente · **Data:** 19/11/2026

**Decisão:** Kubernetes API Server é a API oficial dos contratos técnicos Kubernetes-native; não exigir uma REST API paralela. **Entity** é identidade organizacional; **Environment** é destino UAT/PROD com conta AWS; **Domain** é unidade lógica de ownership, negócio e consumo de infra/aplicações. Os três começam como registros YAML validados no Git, **não CRDs/XRDs**. **Capability** é produto técnico com XRD/Composition/Provider especializado em Crossplane, começando por `ObjectStorage`. Não criar `CapabilityClaim` genérico.

**Atualização (21/11/2026):** Domain Bootstrap permanece apenas como uma proposta antiga em reavaliação, sem implementação ou pré-requisito nos próximos labs (ADR-0005). A organização de contratos Entity/Environment/Domain/Capability não depende desse gerador. A camada final de aplicações chama-se **Domain Applications**.

**Por quê:** negócio/organização não exige reconciliador próprio. Namespace é materialização local do Domain; **A proposta anterior previa que criar um Domain seguiria o Domain Bootstrap**, um fluxo padrão versionado que gera sua estrutura e os recursos iniciais previstos no perfil (Database, ObjectStorage, namespace, políticas), para os Environments escolhidos. O Bootstrap gera manifests; Crossplane e Argo CD local provisionam. Aplicações e Capabilities opcionais permanecem solicitações separadas. Um mesmo Domain pode consumir Capabilities em UAT e PROD.

**Consequências:** o pipeline precisa validar Domain/Environment e resolver conta/IAM autorizada antes da aplicação de uma Capability. YAML de registro não é manifest Kubernetes; Argo CD central sincroniza somente CRs de produtos. Um eventual contrato/controller para lifecycle de contas deve ser justificado futuramente.

[Arquitetura](../architecture.md) · [Índice](README.md)

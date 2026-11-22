# ADR-0005 — Domain Bootstrap como padrão de criação do Domain

**Estado:** proposta preservada em reavaliação, implementação adiada · **Data:** 19/11/2026; revisão 21/11/2026

> [!NOTE]
> Domain Bootstrap permanece documentado para referência, mas **não está previsto nas próximas implementações** e pode ser descartado após nova avaliação. Não criar gerador ou controller nem torná-lo pré-requisito agora.

## Contexto (desenho histórico, sem adoção vigente)

O Domain é uma **unidade lógica de negócio e ownership**, não um XRD. **Na proposta anterior**, ao criar um Domain, a plataforma **seguiria** um processo de Bootstrap para preparar a estrutura e os recursos previstos no perfil: por exemplo Database, bucket, namespace, quotas e políticas. O padrão nasce com o Domain; **Bootstrap não é um processo separado obrigatório para cada Environment**. Precisamos preservar a separação entre provisionamento AWS e objetos Kubernetes remotos.

## Decisão proposta anteriormente (suspensa)

> Os itens abaixo documentam a hipótese histórica; **não são tarefas aprovadas nem requisitos dos labs atuais**. A decisão vigente mantém Domain Infrastructure no monorepo sob DevOps Federado e Domain Applications em repositório separado da squad (ADR-0007).

1. **Manter Domain como registro único** `data-plane/<entity>/domains/<domain>/domain.yaml`, com owners, políticas, perfil/versão e Environments habilitados. **Não criar Claim ou XRD de Domain para o primeiro laboratório.**
2. **Criar um Domain Bootstrap** como automação do **Workflow Plane**. Na **criação do Domain**, ele valida identidade, perfil e Environments selecionados, e **gera sua estrutura e manifests iniciais declarativos idempotentes no monorepo**, submetidos a PR e aprovação. Uma única criação pode configurar vários destinos; um novo Environment adicionado depois é uma evolução do Domain, com possível reutilização do gerador. A definição do perfil versionado pertence à governança da plataforma (Control Plane).
3. **Perfil completo `standard` da POC:** requer pelo menos `Database`, `ObjectStorage`, namespace(s) em clusters permitidos, quotas e políticas. **Argo CD central → Crossplane central** reconcilia as Capabilities AWS; **Argo CD local** aplica os objetos Kubernetes. Não criar dois owners do mesmo recurso.
4. **Conclusão do Bootstrap do Domain**, com verificação detalhada por destino (`Domain + Environment`), a partir de `Ready` das Capabilities e sync/health de Argo CD local, incluindo existência de objetos/policies. Não interpretar um `domain.yaml` aprovado como `Ready`. Falhas devem suportar reexecução idempotente, sem duplicar recursos.
5. **Evolução segura:** perfis devem ter versões controladas; novo perfil não altera Domains existentes sem migração revisada. Selecionar UAT na criação não implica selecionar PROD; recursos stateful precisam de análise específica de exclusão/upgrade. Apagar o Domain não implica destruição automática de bancos, buckets ou Core.

## Escolha técnica e faseamento (somente se a proposta for reativada)

O **Bootstrap não é um Claim Crossplane**: a Composition descreve um produto técnico (Capability), enquanto o Bootstrap coordena a entrega de produtos AWS e configurações Kubernetes por mecanismos distintos. Futuramente pode ser avaliado um XR de Domain, mas não é pré-requisito.

Os primeiros labs implementam `ObjectStorage` e, em seguida, `Database` como **Aurora PostgreSQL**, sempre de forma autorizada e progressiva. Um **perfil reduzido de laboratório** demonstra a geração de manifests; o **perfil standard completo só estará `Ready` após o laboratório implementar Database e GitOps local**. O mecanismo de disparo PR/CI, renderização e agregação de estado são **a implementar**, não recursos já ativos.

## Consequências (caso reativado)

Uma automação extra **passaria** a fazer parte da POC. Deve existir schema validado para Domain/profile/Environment, separação de acesso ao Git e à AWS, suporte a revisão, detecção de drift e auditoria. A organização `Data Plane → Entity → Core Infrastructure / Domains` continua inalterada.

[Arquitetura](../architecture.md) · [ADRs](README.md)

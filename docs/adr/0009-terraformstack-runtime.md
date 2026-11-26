# ADR-0009 — Runtime e contrato de execução do TerraformStack

[← ADRs](README.md) · [Fluxo](../flows/terraformstack.md)

**Status** — decisão vigente com runtime Go versionado. A Etapa 06 registra o aceite ponta a ponta no Hub `kind`. [Evidências](../../labs/06-terraformstack-simulator.md#evidências-de-aceite)

## Contexto

O TerraformStack mantém a intenção declarativa de stacks completas. A execução de Terraform, avaliação de políticas e aprovação precisam de um workflow governado, com identidade e estado segregados. Não duplicar o motor de execução do Argo Workflows.

## Decisão

- Usar Go e controller-runtime, com watches, work queue, cache/index de dependências, workers, Leader Election, métricas e probes. Drift e retry usam RequeueAfter por stack.
- Manter a reconciliação, lifecycle e autorização em `internal/controller`. A interface em `internal/execution` define criação, leitura normalizada, cancelamento, validação e snapshot. O adapter em `internal/execution/argo` administra objetos, parâmetros e eventos Argo. Argo Workflows é a única integração suportada.
- Manter o CRD atual, o catálogo confiável e o protocolo de autorização. A leitura direta da API antecede decisões e o status é escrito com resourceVersion. O claim do runner marca o início do apply; o apply já iniciado não é interrompido por mudança de intenção.
- O Platform Team controla o arquivo [controller-config.yaml](../../control-plane/terraformstack/controller/controller-config.yaml), catálogo, WorkflowTemplate e permissões. A configuração valida nomes e conserva um snapshot por tentativa; drenar execuções antes de mudanças incompatíveis. Não aceitar escolha arbitrária de template pelo CR.
- WorkflowTemplate, runner, scripts e módulos/perfis OPA são a implementação de referência da POC. O controller não executa Terraform nem OPA diretamente. Uma pipeline substituta confiável deve cumprir a semântica do protocolo e as mesmas garantias de autorização, integridade e auditoria; não basta renomear parâmetros. Alterar namespaces ou templates requer adequar scripts, RBAC, admissão e GitOps.

## Consequências e aceite

O runtime pode ser atualizado sem alterar o contrato declarativo, desde que as execuções sejam drenadas e os componentes compatíveis. Testes Go, race detector, envtest, build e validações de manifests protegem o contrato; testes OPA e empacotamento validam a implementação de referência. O [Lab 05](../../labs/05-terraformstack-runtime.md) verifica instalação, identidade e disponibilidade. O [Lab 06](../../labs/06-terraformstack-simulator.md) comprova o fluxo completo, incluindo falha fechada, aprovação, claim, drift, retry, dependências e destroy. O CI não substitui esse aceite.

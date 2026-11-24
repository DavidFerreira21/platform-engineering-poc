# ADR-0004 — AWS real, segurança e entrega incremental

**Estado:** decisão vigente, implementação pendente · **Data:** 19/11/2026

**Decisão.** A POC começa no Hub `kind` com demonstrações locais de Crossplane e `TerraformStack`. Para a primeira execução AWS, a equipe valida o registro de UAT via IAM Identity Center, prepara a identidade do executor e o backend, e aplica a rede com aprovação humana. Em seguida, avalia EKS Auto Mode **um cluster de cada vez** e executa um smoke test. Depois dos clusters, valida a identidade dedicada do Provider AWS no Hub e usa `ObjectStorage` e `Database` em UAT. A implementação de `Database` usa **Amazon Aurora PostgreSQL**, com avaliação de custo, sub-redes, IAM e teardown. PROD é validado como destino separado no exercício de isolamento. A criação automatizada de contas e um controller de Environment ficam para avaliação futura. [Guia de UAT](../../labs/07-registro-ambientes-aws.md) · [Roteiro](../learning-roadmap.md)

**Governança:** cada operação AWS exige STS para confirmar identidade/conta, região, IAM mínimo, credenciais temporárias, custo, aprovação explícita e plano de cleanup. `apply` e `destroy` nunca estão autorizados por documentação. Terraform state separado por escopo e fora do Git; sem segredos/certificados/kubeconfig reais em commits. Remoção de registro YAML não deleta conta AWS. EKS destrutivo exige gate. Capabilities precisam de política de retenção/deletion testada; finalizers não são aprovação humana.

**Riscos:** validar operators/agentes, custos e requisitos PCI antes de Auto Mode. Um Argo CD não deve administrar PCI/Non-PCI sem revisão dos limites administrativos. Comprovar tentativas indevidas negadas entre Domain e conta/Environment. Observar drift, logs, eventos e trilha de auditoria.

[Arquitetura](../architecture.md) · [Roadmap](../learning-roadmap.md)

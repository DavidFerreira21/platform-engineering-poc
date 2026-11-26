# ADR-0004 — AWS real, segurança e entrega incremental

**Estado:** decisão vigente, implementação pendente · **Data:** 19/11/2026

**Decisão.** A plataforma conclui o TerraformStack e depois o Crossplane com recursos Kubernetes no Hub `kind`. As Etapas 09–10 preparam Foundation, módulos e acesso da conta de plataforma. O bootstrap externo provisiona o Hub EKS Auto Mode. Os componentes centrais são reinstalados do Git e os dois aceites são repetidos antes de preparar as spokes. [Roteiro](../learning-roadmap.md)

**Identidade.** A Etapa 11 valida o registro UAT e o acesso humano via IAM Identity Center. A Etapa 12 comprova a sessão temporária de origem do Pod no Hub EKS, a assunção da role autorizada de UAT e o backend S3 segregado. A Etapa 13 aplica a Network da spoke com aprovação vinculada ao plano. Management, Non-PCI, GitOps e Bridge seguem nas Etapas 14–18. [Registro UAT](../learning-roadmap.md#etapa-11--registro-de-uat) · [Identidade do executor](../learning-roadmap.md#etapa-12--identidade-do-executor-e-state)

**Governança:** cada operação AWS exige STS para confirmar identidade/conta, região, IAM mínimo, credenciais temporárias, custo, aprovação explícita e plano de cleanup. `apply` e `destroy` nunca estão autorizados por documentação. Terraform state separado por escopo e fora do Git; sem segredos/certificados/kubeconfig reais em commits. Remoção de registro YAML não deleta conta AWS. EKS destrutivo exige gate. Capabilities precisam de política de retenção/deletion testada; finalizers não são aprovação humana.

**Riscos:** validar operators/agentes, custos e requisitos PCI antes de Auto Mode. Um Argo CD não deve administrar PCI/Non-PCI sem revisão dos limites administrativos. Comprovar tentativas indevidas negadas entre Domain e conta/Environment. Observar drift, logs, eventos e trilha de auditoria.

[Arquitetura](../architecture.md) · [Roadmap](../learning-roadmap.md)

# ADR-0004 — AWS real, segurança e entrega incremental

**Estado:** decisão vigente, implementação pendente · **Data:** 19/11/2026

**Decisão:** a POC começa com **conta AWS real existente**, Hub `kind` e `ObjectStorage` barato/reversível. Depois evolui para contas preexistentes UAT/PROD separadas e EKS Auto Mode **um cluster de cada vez**. A criação automatizada de contas fica fora da primeira versão, assim como qualquer obrigação de CRD/controller próprio de Environment.

**Governança:** cada operação AWS exige STS para confirmar identidade/conta, região, IAM mínimo, credenciais temporárias, custo, aprovação explícita e plano de cleanup. `apply` e `destroy` nunca estão autorizados por documentação. Terraform state separado por escopo e fora do Git; sem segredos/certificados/kubeconfig reais em commits. Remoção de registro YAML não deleta conta AWS. EKS destrutivo exige gate. Capabilities precisam de política de retenção/deletion testada; finalizers não são aprovação humana.

**Riscos:** validar operators/agentes, custos e requisitos PCI antes de Auto Mode. Um Argo CD não deve administrar PCI/Non-PCI sem revisão dos limites administrativos. Comprovar tentativas indevidas negadas entre Domain e conta/Environment. Observar drift, logs, eventos e trilha de auditoria.

[Arquitetura](../architecture.md) · [Roadmap](../learning-roadmap.md)

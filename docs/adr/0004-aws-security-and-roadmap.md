# ADR-0004 — AWS real, segurança e entrega incremental

**Estado:** decisão vigente, implementação pendente · **Data:** 19/11/2026

**Decisão:** a POC começa com **Hub `kind` e Crossplane em modo local, sem AWS**. O [Lab 03B](../../labs/03b-uat-e-autenticacao-aws.md) registra **contas UAT e PROD distintas** por referência lógica no Git e valida os destinos por IAM Identity Center. As primeiras Capabilities AWS serão testadas na UAT. Após validar a identidade de workload do Provider AWS no Hub, a primeira Capability real será `ObjectStorage`, seguida de `Database` implementada por **cluster Amazon Aurora PostgreSQL**, condicionada à análise de custos, VPC/subnets, IAM e teardown. A POC valida o isolamento entre UAT e PROD conforme avança o provisionamento e avalia EKS Auto Mode **um cluster de cada vez**. A criação automatizada de contas fica fora da primeira versão, assim como qualquer obrigação de CRD/controller próprio de Environment.

**Governança:** cada operação AWS exige STS para confirmar identidade/conta, região, IAM mínimo, credenciais temporárias, custo, aprovação explícita e plano de cleanup. `apply` e `destroy` nunca estão autorizados por documentação. Terraform state separado por escopo e fora do Git; sem segredos/certificados/kubeconfig reais em commits. Remoção de registro YAML não deleta conta AWS. EKS destrutivo exige gate. Capabilities precisam de política de retenção/deletion testada; finalizers não são aprovação humana.

**Riscos:** validar operators/agentes, custos e requisitos PCI antes de Auto Mode. Um Argo CD não deve administrar PCI/Non-PCI sem revisão dos limites administrativos. Comprovar tentativas indevidas negadas entre Domain e conta/Environment. Observar drift, logs, eventos e trilha de auditoria.

[Arquitetura](../architecture.md) · [Roadmap](../learning-roadmap.md)

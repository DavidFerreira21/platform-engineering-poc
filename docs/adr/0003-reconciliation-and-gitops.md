# ADR-0003 — Execução e reconciliação sem sobreposição

**Estado:** decisão vigente, implementação pendente · **Data:** 19/11/2026

**Decisão:** **um Crossplane central** no Control Plane com XRDs/Compositions/Providers para Capabilities AWS, sem Crossplane em cascata. **Argo CD central** sincroniza CRs de Capability do Git com a Kubernetes API do Hub; não observa YAMLs lógicos de Domain/Environment como CRs. **Argo CD local** do Management EKS administra configuração, namespaces, políticas e workloads em clusters autorizados. **Terraform + Argo Workflows** coordena Foundation/EKS, plan, gates, apply e verificação; Argo CD não executa Terraform.

**Domain Bootstrap:** um gerador/workflow validado por PR cria manifests obrigatórios conforme perfil versionado para cada Domain/Environment. Argo CD central sincroniza Capabilities, e Argo CD local sincroniza namespace/policies; o Bootstrap não aplica recursos diretamente. A prontidão é verificada por evidências de ambos, e o perfil completo não fica `Ready` sem banco/bucket/namespace/políticas. Não reexecutar destruição apenas pela remoção do registro.

**Por quê:** evita controllers concorrendo pelos mesmos recursos, isola operação AWS dos Deployments e preserva autonomia federada via Git.

**Consequências:** Crossplane não monitora Git, apenas a API. Credenciais IAM precisam de associação autorizada por Environment e validação de Domain. O Argo CD local pode precisar de segregação de acesso PCI. Terraform só converge quando é executado; indisponibilidade do Hub interrompe reconciliação das Capabilities centrais, não necessariamente os workloads.

[Arquitetura](../architecture.md) · [Índice](README.md)

# 🧩 Control Plane

O **Kubernetes API Server** é a API da plataforma para produtos técnicos Kubernetes-native. **Crossplane central** implementará XRDs, Compositions e Providers para **Capabilities** AWS.

| Conceito | Forma na POC |
| :--- | :--- |
| Entity | Registro de organização no Git |
| Environment | Registro YAML de UAT/PROD e conta/região AWS, não CRD |
| Domain | Registro lógico de ownership/consumo, não XRD |
| `ObjectStorage` | Primeira Capability via XRD + Composition + Provider |
| `Database` | Segunda Capability: **Aurora PostgreSQL**, dependente de conta UAT autorizada, VPC e controle de custo |
| `Messaging`, `ComputeInstance` | Capabilities posteriores |

Crossplane **não observa diretamente Git**. O **Argo CD central sincroniza CRs de Capability e, futuramente, `TerraformStack`** com a Kubernetes API do Hub. Crossplane reconcilia exclusivamente as Capabilities AWS de seu ownership; o TerraformStack Controller coordena os Workflows de Foundation/Entity Core. Os registros lógicos Domain/Environment não são CRs e exigem resolução confiável de destino e IAM; o Domain não escolhe uma role arbitrária. **Domain Bootstrap é uma proposta anterior adiada e em reavaliação (ADR-0005):** seus perfis/gerador não são pré-requisito e o Workflow Plane **não gera manifests de Bootstrap atualmente**. O Argo CD local mantém objetos Kubernetes de aplicações nos destinos autorizados.

## TerraformStack (proposta)

O Control Plane prevê **`TerraformStack` CRD + controller próprios** para expressar a **infraestrutura desejada** de Foundation/Entity Core, semelhante à experiência declarativa do Crossplane, porém provisionada por Terraform. O **Argo CD central** aplica o CR Kubernetes a partir do Git; o controller executa/verifica `terraform plan` via **Argo Workflows + Terraform Executor**, detecta drift e consulta políticas para **auto apply, aprovação ou bloqueio**. O controller atualiza condições/status, sem realizar `apply` em todo reconcile. `TerraformRun` é nome opcional para uma execução isolada, **sem segundo CRD obrigatório**.

**Admission / Policies:** Kyverno/OPA são tecnologias a avaliar. A Etapa 07 define o CRD TerraformStack, e as Etapas 10 e 11 tratam do controller e das políticas de execução. [ADR-0006](../docs/adr/0006-terraformrun-and-application-ci.md).

A [Etapa 04](../labs/04-crossplane-core.md) instala o Crossplane Core no Hub via Argo CD e valida um XRD namespaced que compõe um ConfigMap local.

[← Início](../README.md) · [Arquitetura](../docs/architecture.md)

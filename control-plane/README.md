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

Crossplane **não observa diretamente Git**. O **Argo CD central sincroniza CRs de Capability e, futuramente, `TerraformStack`** com a Kubernetes API do Hub. Crossplane reconcilia exclusivamente as Capabilities AWS de seu ownership; o TerraformStack Controller coordena os Workflows de Foundation/Entity Core. Os registros lógicos Domain/Environment não são CRs e exigem resolução confiável de destino e IAM; o Domain não escolhe uma role arbitrária. O Argo CD local mantém objetos Kubernetes de aplicações nos destinos autorizados.

## TerraformStack

O Control Plane publica o CRD **`TerraformStack`** para expressar a **infraestrutura desejada** de Foundation e Entity Core. A [Etapa 05](../labs/05-terraformstack-runtime.md) exercita o controller com execução simulada. Na arquitetura-alvo, o Argo CD central sincroniza a solicitação, e o controller coordena os planos via Argo Workflows e Terraform Executor. Drift e reconciliação por políticas entram nas Etapas 22 e 23. `TerraformRun` é o nome opcional de uma execução isolada. [Guia da Etapa 04](../labs/04-terraformstack-crd.md)

**Admission e políticas.** O CRD entra na Etapa 04. O controller e a pipeline entram na Etapa 05. O Simulator valida OPA e a reconciliação na Etapa 06, antes da validação AWS nas Etapas 28–29. [ADR-0006](../docs/adr/0006-terraformrun-and-application-ci.md)

A Etapa 07 instala o Crossplane Core no Hub via Argo CD e valida um XRD namespaced que compõe um ConfigMap local. As Capabilities AWS são exercitadas nas Etapas 18 e 19, após os clusters e a identidade do Provider AWS. [Guia da Etapa 07](../labs/07-crossplane-core.md)

[← Início](../README.md) · [Arquitetura](../docs/architecture.md)

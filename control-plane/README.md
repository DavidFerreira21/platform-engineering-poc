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

O Argo CD central sincroniza os CRs de Capability e `TerraformStack` com a Kubernetes API do Hub. Crossplane reconcilia as Capabilities de seu ownership. O TerraformStack Controller coordena os Workflows de Foundation e Entity Core. A plataforma valida destino e IAM a partir dos registros lógicos Domain e Environment. O Argo CD local mantém objetos Kubernetes de aplicações nos destinos autorizados.

## TerraformStack

O Control Plane publica o CRD `TerraformStack` na Etapa 04. Na Etapa 05, instala o controller que valida a intenção declarada, agenda os Workflows e autoriza a execução. A pipeline resolve Git, contexto e state. Cada tentativa é representada por um Workflow, e o controller atualiza o status do CR a partir dos outputs tipados. A Etapa 06 comprova o fluxo no Simulator. [Fluxo do TerraformStack](../docs/flows/terraformstack.md) · [Guia do Simulator](../labs/06-terraformstack-simulator.md)

O primeiro apply de Network em UAT está previsto para a Etapa 13, com aprovação humana. As Etapas 28 e 29 validam drift e reconciliação por políticas na AWS. [Roadmap](../docs/learning-roadmap.md) · [ADR-0006](../docs/adr/0006-terraformrun-and-application-ci.md)

A Etapa 07 instala o Crossplane Core no Hub via Argo CD e valida um XRD namespaced que compõe um ConfigMap local. A Etapa 19 valida a identidade do Provider AWS. As Etapas 20 e 21 exercitam `ObjectStorage` e `Database` na AWS. [Guia da Etapa 07](../labs/07-crossplane-core.md)

[← Início](../README.md) · [Arquitetura](../docs/architecture.md)

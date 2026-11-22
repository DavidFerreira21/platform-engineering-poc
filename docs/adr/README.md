# 📝 Decisões arquiteturais vigentes

[← Início](../../README.md) · [Arquitetura](../architecture.md)

Estes documentos contêm **somente a arquitetura atual** e suas justificativas. As propostas anteriores permanecem recuperáveis no **histórico Git**, mas não poluem a documentação da POC com decisões contraditórias. Nenhum ADR implica recurso já implantado.

| Decisão | Tema |
| :--- | :--- |
| [**0001**](0001-platform-contracts.md) | API Kubernetes e conceitos Entity, Environment, Domain, Capability |
| [**0002**](0002-topology-and-ownership.md) | Planos, Data Plane, Core/Domains, quatro camadas e monorepo |
| [**0003**](0003-reconciliation-and-gitops.md) | Crossplane central, Terraform e GitOps central/local |
| [**0004**](0004-aws-security-and-roadmap.md) | AWS real, EKS Auto Mode, isolamento, lifecycle e labs |
| [**0005**](0005-domain-bootstrap.md) | Domain Bootstrap — proposta preservada, em reavaliação / adiada |
| [**0006**](0006-terraformrun-and-application-ci.md) | TerraformStack CRD/controller, drift, policies de reconciliação e CI/CD de Domain Applications em repo separado |

Documento principal: [Arquitetura](../architecture.md) · [Estrutura](../repository-structure.md) · [Roadmap](../learning-roadmap.md)

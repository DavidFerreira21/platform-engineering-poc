# 📝 Decisões arquiteturais vigentes

[← Início](../../README.md) · [Arquitetura](../architecture.md)

Os ADRs registram as **decisões arquiteturais vigentes** e, quando necessário, propostas históricas identificadas como adiadas ou em reavaliação. Decisões substituídas também podem ser consultadas no histórico Git. Nenhum ADR comprova implantação de recursos.

| Decisão | Tema |
| :--- | :--- |
| [**0001**](0001-platform-contracts.md) | API Kubernetes e conceitos Entity, Environment, Domain, Capability |
| [**0002**](0002-topology-and-ownership.md) | Planos, Data Plane, Core/Domains, quatro camadas e monorepo |
| [**0003**](0003-reconciliation-and-gitops.md) | Crossplane central, Terraform e GitOps central/local |
| [**0004**](0004-aws-security-and-roadmap.md) | AWS real, EKS Auto Mode, isolamento, lifecycle e labs |
| [**0006**](0006-terraformrun-and-application-ci.md) | TerraformStack CRD/controller, drift e reconciliação por políticas; CI/CD de aplicações em repo separado |
| [**0007**](0007-personas-and-developer-experience.md) | Três personas, ownership federado, repositórios e app mínima Python/FastAPI para validar DX no fim da POC |
| [**0008**](0008-managed-argocd-connectivity.md) | Argo CD central gerenciado e acesso GitOps entre VPCs isoladas |
| [**0009**](0009-terraformstack-runtime.md) | Runtime Go, integrações Argo selecionáveis, autorização opcional, exclusividade por CR e emissão de telemetria |
| [**0010**](0010-observability-topology.md) | Topologia-alvo de observabilidade: coleta local nos EKS, AMP e backends por Environment, consulta central e isolamento |

Diagrama visual: [Arquitetura de referência com as três personas](../arquitetura%20Plataform.png).

Documento principal: [Arquitetura](../architecture.md) · [Estrutura](../repository-structure.md) · [Roadmap](../learning-roadmap.md)

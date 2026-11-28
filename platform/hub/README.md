# 🏗️ Platform Management Cluster

O **Platform Management Cluster** hospeda os componentes de **Control Plane** e **Workflow Plane**. Na topologia AWS, a conta da Plataforma usa Managed Grafana para a consulta central e CloudWatch para os sinais essenciais do Hub. Os caminhos de ingestão dos workloads ficam nos respectivos Environments. [Fluxo de observabilidade](../../docs/flows/observability.md)

## Hub EKS

- **Etapa 01.** A raiz Terraform `platform-hub` provisiona rede, EKS Auto Mode, EKS Capability for Argo CD e bucket de state do executor. O acesso à API, os NodePools e a capability já foram verificados. O workload de teste validou agendamento, DNS e HTTP. [Guia da Etapa 01](../../labs/01-platform-management-cluster.md)
- **Etapa 02.** O primeiro sync a partir da `main` foi validado com destino por ARN e permissões de escrita limitadas ao exercício. [Guia da Etapa 02](../../labs/02-argocd-central.md)
- **Etapas 03–08.** O catálogo em `addons/catalog/` organiza uma Application por componente. O Argo Workflows, o runtime do TerraformStack, o Simulator e o Crossplane têm evidências no EKS. A instalação completa pelo catálogo segue com aceite próprio. [Guias das etapas](../../labs/README.md)

## Ciclo das Applications

Cada componente com sincronização própria tem uma Application. Uma etapa pode observar ou reutilizar as Applications compartilhadas. As Applications `terraformstack-crd`, `terraformstack-workflows`, `terraformstack-controller` e `terraformstack-simulator` têm sincronização manual e independente. [Guia da Etapa 05](../../labs/05-terraformstack-runtime.md) · [Guia da Etapa 06](../../labs/06-terraformstack-simulator.md) · [ADR-0003](../../docs/adr/0003-reconciliation-and-gitops.md#applications-e-sincronização-no-hub)

---

[← Página inicial](../../README.md) · [Arquitetura](../../docs/architecture.md) · [Roteiro prático](../../docs/learning-roadmap.md)

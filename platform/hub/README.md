# 🏗️ Platform Management Cluster

O **Platform Management Cluster** hospeda os componentes de **Control Plane** e **Workflow Plane**. Na topologia AWS, a conta da Plataforma usa Managed Grafana para a consulta central e CloudWatch para os sinais essenciais do Hub. Os caminhos de ingestão dos workloads ficam nos respectivos Environments. [Fluxo de observabilidade](../../docs/flows/observability.md)

## Hub EKS

- **Etapa 01.** A raiz Terraform `platform-hub` provisiona rede, EKS Auto Mode, EKS Capability for Argo CD e bucket de state do executor. O acesso à API, os NodePools e a capability já foram verificados. O teste com workload segue pendente. [Guia da Etapa 01](../../labs/01-platform-management-cluster.md)
- **Etapa 02.** Registrar o Hub por ARN, limitar permissões e validar o primeiro sync a partir da `main` publicada. [Guia da Etapa 02](../../labs/02-argocd-central.md)
- **Etapas 03–08.** O catálogo em `addons/catalog/` organiza uma Application por componente. Os AppProjects, destinos e configurações dos add-ons precisam ser adequados à capability para reproduzir os aceites no EKS. [Guias das etapas](../../labs/README.md)

## Lab local em kind

- **Base local.** `bash platform/hub/scripts/create.sh` → `bash platform/hub/scripts/install-traefik.sh` → `bash platform/hub/scripts/verify.sh`. O Traefik expõe `http://demo.localhost:8080`, `https://argocd.localhost:8443` e `https://workflows.localhost:8443`.
- **GitOps local.** Instalar o Argo CD autogerenciado com `bash platform/hub/scripts/install-argocd.sh` e aplicar `addons/bootstrap.yaml` no contexto `kind-platform-management`. O ApplicationSet `hub-addons` lê o catálogo da `main` publicada. Os AppProjects limitam origem, destino e tipos de recurso. O sync é manual e a remoção de uma Application gerada preserva seus recursos.
- **Ciclo das Applications.** Cada componente com sincronização própria tem uma Application. Uma etapa pode observar ou reutilizar as Applications compartilhadas. As Applications `terraformstack-crd`, `terraformstack-workflows`, `terraformstack-controller` e `terraformstack-simulator` têm sincronização manual e independente. [Guia da Etapa 05](../../labs/05-terraformstack-runtime.md) · [Guia da Etapa 06](../../labs/06-terraformstack-simulator.md) · [ADR-0003](../../docs/adr/0003-reconciliation-and-gitops.md#applications-e-sincronização-no-hub)

## Limpeza do lab local

- `bash platform/hub/scripts/delete.sh --confirm` remove o cluster `kind` e todos os objetos locais, incluindo as configurações de etapas posteriores. Trate recursos externos em suas operações de remoção, quando aplicável.
- Se nenhuma Application depender mais do Argo CD central, `helm --kube-context kind-platform-management -n argocd uninstall argocd` remove o release. O chart preserva os CRDs. A exclusão de um CRD é uma operação separada e remove as instâncias do tipo no cluster.

---

[← Página inicial](../../README.md) · [Arquitetura](../../docs/architecture.md) · [Roteiro prático](../../docs/learning-roadmap.md)

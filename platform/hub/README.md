# 🏗️ Platform Management Cluster

O **Platform Management Cluster** hospeda os componentes de **Control Plane** e **Workflow Plane**. Na topologia AWS, a conta da Plataforma usa Managed Grafana para a consulta central e CloudWatch para os sinais essenciais do Hub. Os caminhos de ingestão dos workloads ficam nos respectivos Environments. [Fluxo de observabilidade](../../docs/flows/observability.md)

- **Etapa 01 — cluster base e networking:** [roteiro](../../labs/01-platform-management-cluster.md), [kind-config.yaml](kind-config.yaml), [namespaces.yaml](namespaces.yaml), Traefik instalado em etapa separada via Helm, Ingress HTTP de demonstração.
- **Passos locais:** `bash platform/hub/scripts/create.sh` → `bash platform/hub/scripts/install-traefik.sh` → `bash platform/hub/scripts/verify.sh`.
- **Etapa 02 — Argo CD central e add-ons.** Instalar Argo CD com `bash platform/hub/scripts/install-argocd.sh` e aplicar `addons/bootstrap.yaml` no Hub. O `ApplicationSet` `hub-addons` lê um arquivo por add-on em `addons/catalog/` na `main` publicada. Os AppProjects limitam origem, destino e tipos de recurso. [Guia da Etapa 02](../../labs/02-argocd-central.md) · [Bootstrap](addons/bootstrap.yaml)
- **Rede.** O Traefik recebe `http://demo.localhost:8080` na Etapa 01, `https://argocd.localhost:8443` na Etapa 02 e `https://workflows.localhost:8443` na Etapa 03.
- **Segurança.** O bootstrap usa o contexto `kind-platform-management`. Alterações no catálogo e nos AppProjects são mudanças do Platform Team e passam por revisão via PR. Nas etapas locais, o sync dos add-ons é manual. A remoção de uma Application gerada preserva seus recursos no Hub.
- **Ciclo das Applications.** Cada componente com sincronização própria tem uma Application. Uma etapa pode observar ou reutilizar as Applications compartilhadas. As Applications `terraformstack-crd`, `terraformstack-workflows`, `terraformstack-controller` e `terraformstack-simulator` têm sincronização manual e independente. [Guia da Etapa 05](../../labs/05-terraformstack-runtime.md) · [Guia da Etapa 06](../../labs/06-terraformstack-simulator.md) · [ADR-0003](../../docs/adr/0003-reconciliation-and-gitops.md#applications-e-sincronização-no-hub)
- **Objetivo:** construir o contrato de plataforma e validar integração com AWS real.
- **Não confundir:** Hub `kind` local com os EKS Auto Mode dos Data Planes.

## Manutenção do Hub

- `bash platform/hub/scripts/delete.sh --confirm` remove o cluster `kind` e todos os objetos locais, incluindo as configurações de etapas posteriores. Trate recursos externos em suas operações de remoção, quando aplicável.
- Se nenhuma Application depender mais do Argo CD central, `helm --kube-context kind-platform-management -n argocd uninstall argocd` remove o release. O chart preserva os CRDs. A exclusão de um CRD é uma operação separada e remove as instâncias do tipo no cluster.

---

[← Página inicial](../../README.md) · [Arquitetura](../../docs/architecture.md) · [Roteiro prático](../../docs/learning-roadmap.md)

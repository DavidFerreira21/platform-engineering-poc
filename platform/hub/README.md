# 🏗️ Platform Management Cluster

O **Platform Management Cluster** hospeda os componentes lógicos de **Control Plane**, **Workflow Plane** e **Observability Plane**.

- **Etapa 01 — cluster base e networking:** [roteiro](../../labs/01-platform-management-cluster.md), [kind-config.yaml](kind-config.yaml), [namespaces.yaml](namespaces.yaml), Traefik instalado em etapa separada via Helm, Ingress HTTP de demonstração.
- **Passos locais:** `bash platform/hub/scripts/create.sh` → `bash platform/hub/scripts/install-traefik.sh` → `bash platform/hub/scripts/verify.sh`. Remoção **destrutiva**: `bash platform/hub/scripts/delete.sh --confirm`.
- **Etapa 02 — Argo CD central e add-ons.** Instalar Argo CD com `bash platform/hub/scripts/install-argocd.sh` e aplicar `addons/bootstrap.yaml` no Hub. O `ApplicationSet` `hub-addons` lê um arquivo por add-on em `addons/catalog/` na `main` publicada. Os AppProjects limitam origem, destino e tipos de recurso. [Guia da Etapa 02](../../labs/02-argocd-central.md) · [Bootstrap](addons/bootstrap.yaml)
- **Rede:** acesso somente pelo host `127.0.0.1:8080` (HTTP) e `127.0.0.1:8443` (HTTPS reservado); roteamento por `demo.localhost`.
- **Segurança.** O bootstrap usa o contexto `kind-platform-management`. Alterações no catálogo e nos AppProjects são mudanças do Platform Team e passam por revisão via PR. Nas etapas locais, o sync dos add-ons é manual. A remoção de uma Application gerada preserva seus recursos no Hub.
- **Objetivo:** construir o contrato de plataforma e validar integração com AWS real.
- **Não confundir:** Hub `kind` local com os EKS Auto Mode dos Data Planes.

---

[← Página inicial](../../README.md) · [Arquitetura](../../docs/architecture.md) · [Roteiro prático](../../docs/learning-roadmap.md)

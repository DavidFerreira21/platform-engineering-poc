# 🏗️ Platform Management Cluster

O **Platform Management Cluster** hospeda os componentes lógicos de **Control Plane**, **Workflow Plane** e **Observability Plane**.

- **Etapa 01 — cluster base e networking:** [roteiro](../../labs/01-platform-management-cluster.md), [kind-config.yaml](kind-config.yaml), [namespaces.yaml](namespaces.yaml), Traefik instalado em etapa separada via Helm, Ingress HTTP de demonstração.
- **Passos locais:** `bash platform/hub/scripts/create.sh` → `bash platform/hub/scripts/install-traefik.sh` → `bash platform/hub/scripts/verify.sh`. Remoção **destrutiva**: `bash platform/hub/scripts/delete.sh --confirm`.
- **Etapa 02 — Argo CD central:** [roteiro](../../labs/02-argocd-central.md) e bootstrap via `bash platform/hub/scripts/install-argocd.sh`.
- **Rede:** acesso somente pelo host `127.0.0.1:8080` (HTTP) e `127.0.0.1:8443` (HTTPS reservado); roteamento por `demo.localhost`.
- **Segurança:** os scripts só atuam no contexto `kind-platform-management`; nenhuma ação AWS, sem Crossplane instalado neste lab.
- **Objetivo:** construir o contrato de plataforma e validar integração com AWS real.
- **Não confundir:** Hub `kind` local com os EKS Auto Mode dos Data Planes.

---

[← Página inicial](../../README.md) · [Arquitetura](../../docs/architecture.md) · [Roteiro prático](../../docs/learning-roadmap.md)

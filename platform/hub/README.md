# 🏗️ Platform Management Cluster

O **Platform Management Cluster** hospeda os componentes lógicos de **Control Plane**, **Workflow Plane** e **Observability Plane**.

- **Lab 01 — cluster base e networking:** [roteiro](../../labs/01-platform-management-cluster.md), [kind-config.yaml](kind-config.yaml), [namespaces.yaml](namespaces.yaml), Traefik instalado em etapa separada via Helm, Ingress HTTP de demonstração.
- **Passos locais:** `bash platform/hub/scripts/create.sh` → `bash platform/hub/scripts/install-traefik.sh` → `bash platform/hub/scripts/verify.sh`. Remoção **destrutiva**: `bash platform/hub/scripts/delete.sh --confirm`.
- **Rede:** acesso somente pelo host `127.0.0.1:8080` (HTTP) e `127.0.0.1:8443` (HTTPS reservado); roteamento por `demo.localhost`.
- **Segurança:** os scripts só atuam no contexto `kind-platform-management`; nenhuma ação AWS, sem Crossplane instalado neste lab.
- **Objetivo:** construir o contrato de plataforma e validar integração com AWS real.
- **Não confundir:** Hub `kind` local com os EKS Auto Mode dos Data Planes.

> [!NOTE]
> **Situação:** Lab 01 preparado em arquivos e scripts, sem execução real do cluster neste PR. Argo CD (Lab 02) e Crossplane (Lab 03) serão entregues separadamente.

---

[← Página inicial](../../README.md) · [Arquitetura](../../docs/architecture.md) · [Roteiro prático](../../docs/learning-roadmap.md)

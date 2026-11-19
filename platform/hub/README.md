# 🏗️ Platform Management Cluster

O **Platform Management Cluster** hospeda os componentes lógicos de **Control Plane**, **Workflow Plane** e **Observability Plane**.

- **Primeira implementação:** cluster local `kind` com [Lab 01](../../labs/01-platform-management-cluster.md) e `kind-config.yaml`.
- **Passos locais:** `bash platform/hub/scripts/create.sh` → `bash platform/hub/scripts/verify.sh` → `bash platform/hub/scripts/delete.sh`.
- **Segurança:** os scripts só atuam no contexto `kind-platform-management`; nenhuma ação AWS, sem Crossplane instalado neste lab.
- **Objetivo:** construir o contrato de plataforma e validar integração com AWS real.
- **Não confundir:** Hub `kind` local com os EKS Auto Mode dos Data Planes.

> [!NOTE]
> **Situação:** primeira automação local de laboratório adicionada; nenhum cluster é criado automaticamente por fazer merge.

---

[← Página inicial](../../README.md) · [Arquitetura](../../docs/architecture.md) · [Roteiro prático](../../docs/learning-roadmap.md)

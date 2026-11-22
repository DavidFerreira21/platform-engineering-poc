# ADR-0006 — TerraformRun declarativo e CI/CD de Domain Applications

**Estado:** decisão de arquitetura-alvo; implementação pendente · **Data:** 21/11/2026

## Contexto e ownership

A POC usa Kubernetes API como contrato técnico e quatro camadas de ownership: **Foundation → Entity Core Infrastructure → Domain Infrastructure → Domain Applications**.

- **Foundation / Entity Core:** Terraform/OpenTofu para rede, IAM, dependências, e os três EKS por Environment (Management, PCI e Non-PCI).
- **Domain Infrastructure:** Crossplane v2 central com XRDs/Compositions/Providers específicos para recursos AWS como ObjectStorage e Database.
- **Domain Applications:** Application Team mantém **código, GitHub Actions e manifests Kubernetes no mesmo repositório de aplicação de teste separado**, ainda não fornecido; Argo CD **local no Management EKS** lê esses manifests GitOps e entrega em EKS PCI/Non-PCI autorizados. O DevOps Federado mantém somente os manifests de Domain Infrastructure.

## Decisão: TerraformRun CRD + Controller

No Control Plane do Hub, criar futuramente um **CRD Kubernetes próprio** de nome de trabalho `TerraformRun` (`platform.example.io/v1alpha1`) e um **controller dedicado**. Não é XRD de Capability nem Crossplane executando Terraform.

Fluxo planejado:

```text
PR/validação Git → Argo CD central → Kubernetes API / TerraformRun
→ Controller TerraformRun → Argo Workflows → Terraform Executor
→ plan → gate/aprovação → apply → verificação/status
→ AWS Foundation / Entity Core
```

O CRD referenciará um **módulo/versionamento autorizado**, ação e destino Entity/Environment resolvido por configuração confiável. Não permitir que consumidores definam livremente role, account ID, backend ou state. O controller valida RBAC/escopo e registra a referência do Workflow e seu status.

**Terraform não é reconciliação contínua**: não executar `apply` novamente a cada evento de reconciliação. É necessário definir idempotência por solicitação/generation, revisão imutável do plano aprovado, limites de concorrência por state, retries controlados, política de exclusão, gates para PROD e operações destrutivas. Terraform state fica fora do Git e segregado. O Lab 08 valida o executor; um Lab 08B validará CRD + controller.

## GitHub Actions e GitOps

O código, a pipeline CI/CD **e os manifests Kubernetes da aplicação de exemplo** **não ficarão neste repositório**; o usuário fornecerá um repositório separado posteriormente. GitHub Actions realizará build, tests, scans e publicação de imagem; a entrega será por alteração declarativa GitOps e **Argo CD local**, e não por deploy imperativo padrão da pipeline ao EKS. O workflow `.github/workflows/validate.yml` atual da POC continua exclusivo para validação de documentação e labs. As três personas e o modelo de Golden Paths são definidos no [ADR-0007](0007-personas-and-developer-experience.md).

## Observabilidade e Spokes

O Management EKS de cada UAT/PROD inclui **Argo CD local**, espaço para **Operators ainda a estudar** e **OpenTelemetry Collector/gateway** para encaminhar telemetria. O Observability Plane central usa **OpenTelemetry, Grafana, Loki (logs) e Tempo (traces)**; métricas/Prometheus, isolamento PCI, autenticação e retenção serão validados nos labs. Visibilidade central não concede acesso administrativo irrestrito às contas.

## Fora do escopo imediato

O **Domain Bootstrap** do ADR-0005 permanece documentado para rastreabilidade, mas está **em reavaliação e adiado**. Não é pré-requisito para Foundation, Capabilities ou Domain Applications.

Nenhum CRD/controller TerraformRun, pipeline de aplicação ou recurso AWS foi implantado por esta decisão.

[Arquitetura](../architecture.md) · [ADRs](README.md) · [Roadmap](../learning-roadmap.md)

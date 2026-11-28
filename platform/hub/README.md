# 🏗️ Platform Management Cluster

O **Platform Management Cluster** hospeda os componentes de **Control Plane** e **Workflow Plane**. Na topologia AWS, a conta da Plataforma usa Managed Grafana para a consulta central e CloudWatch para os sinais essenciais do Hub. Os caminhos de ingestão dos workloads ficam nos respectivos Environments. [Fluxo de observabilidade](../../docs/flows/observability.md)

## Hub EKS

- **Etapa 01.** A raiz Terraform `platform-hub` provisiona rede, EKS Auto Mode, EKS Capability for Argo CD e bucket de state do executor. O acesso à API, os NodePools e a capability já foram verificados. O workload de teste validou agendamento, DNS e HTTP. [Guia da Etapa 01](../../labs/01-platform-management-cluster.md)
- **Etapa 02.** O primeiro sync a partir da `main` foi validado com destino por ARN e permissões de escrita limitadas ao exercício. [Guia da Etapa 02](../../labs/02-argocd-central.md)
- **Etapas 03–08.** O catálogo em `addons/catalog/` organiza Applications explícitas por Argo Workflows, Crossplane e TerraformStack. O Argo Workflows, o runtime do TerraformStack, o Simulator e o Crossplane têm evidências no EKS. A instalação completa pelo catálogo segue com aceite próprio. [Guias das etapas](../../labs/README.md)

## Capacidade do laboratório

O perfil [`lab-graviton`](compute/lab-graviton.yaml) restringe um NodePool de teste a `t4g.medium` On-Demand, arquitetura `arm64`, com limite de 4 vCPUs e consolidação de nós vazios. Ele reutiliza a NodeClass `default`; pelo menos um NodePool gerenciado precisa continuar habilitado para a AWS manter essa NodeClass. O limite de CPU controla a capacidade do pool. As cobranças do control plane, do NAT Gateway e do armazenamento têm ciclo próprio. [Instâncias suportadas](https://docs.aws.amazon.com/eks/latest/userguide/automode-learn-instances.html) · [NodePools gerenciados](https://docs.aws.amazon.com/eks/latest/userguide/set-builtin-node-pools.html)

Para o primeiro teste, usar um Pod descartável com imagem ARM64, seletor `karpenter.sh/nodepool: lab-graviton` e toleration para `platform.example.io/lab-graviton=true:NoSchedule`. Conferir a arquitetura e o tipo do nó alocado. Para migrar componentes, confirmar antes o suporte ARM64 das imagens principais, init containers e providers. As imagens do TerraformStack usadas no primeiro ensaio foram publicadas para `linux/amd64`. [Runtime do TerraformStack](../../labs/05-terraformstack-runtime.md#imagens-e-armazenamento-no-hub-eks)

## Ciclo das Applications

Cada componente com sincronização própria tem uma Application. O bootstrap reúne os AppProjects. Cada arquivo de `addons/catalog/<componente>/` é um manifesto de Application com sync manual. Preparar uma versão local com `HUB_EKS_ARN` substituído pelo ARN já conferido na Etapa 02. Para o controller e o WorkflowTemplate do TerraformStack, substituir também `HUB_CONTROLLER_IMAGE` ou `HUB_RUNNER_IMAGE` pela referência privada por digest. Inspecionar projeto, revisão, destino e diff antes do sync. A distribuição por ApplicationSet nas spokes está prevista na Etapa 16.

Por exemplo, a partir da raiz do monorepo, revisar o bootstrap e uma Application por vez com a função `hub` da Etapa 02:

```bash
sed "s|HUB_EKS_ARN|${hub_arn:?Confirme o ARN do Hub}|g" \
  platform/hub/addons/bootstrap.yaml | hub diff -f -

sed "s|HUB_EKS_ARN|${hub_arn:?Confirme o ARN do Hub}|g" \
  platform/hub/addons/catalog/crossplane/crossplane-demo.yaml | hub diff -f -
```

Depois de revisar, aplicar os manifests renderizados na mesma ordem e sincronizar a Application manualmente. Em um Hub já provisionado, conferir o owner dos componentes instalados por Helm antes de adotar o mesmo recurso no Argo CD. Uma etapa pode observar ou reutilizar as Applications compartilhadas. As Applications `terraformstack-crd`, `terraformstack-workflows`, `terraformstack-controller` e `terraformstack-simulator` têm sincronização manual e independente. [Guia da Etapa 05](../../labs/05-terraformstack-runtime.md) · [Guia da Etapa 06](../../labs/06-terraformstack-simulator.md) · [ADR-0003](../../docs/adr/0003-reconciliation-and-gitops.md#applications-e-sincronização-no-hub)

---

[← Página inicial](../../README.md) · [Arquitetura](../../docs/architecture.md) · [Roteiro prático](../../docs/learning-roadmap.md)

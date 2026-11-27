# ADR-0003 — Execução e reconciliação sem sobreposição

**Estado:** decisão vigente, com TerraformStack validado no Hub local · **Data:** 19/11/2026

**Decisão:** **um Crossplane central** no Control Plane com XRDs/Compositions/Providers para Capabilities AWS, sem Crossplane em cascata. **Argo CD central** sincroniza CRs de Capability do Git com a Kubernetes API do Hub; não observa YAMLs lógicos de Domain/Environment como CRs. **Argo CD local** do Management EKS administra configuração, namespaces, políticas e workloads em clusters autorizados. **Terraform + Argo Workflows** coordena Foundation/EKS, plan, gates, apply e verificação; Argo CD não executa Terraform.

**Complemento sobre personas e GitOps (ADR-0007):** **DevOps Federado** gerencia os manifests de **Domain Infrastructure** no monorepo de laboratório; o **Application Team** possui o repositório separado com código, GitHub Actions e **manifests Kubernetes da aplicação**. Argo CD **central** aplica CRs Crossplane do repositório de infraestrutura; Argo CD **local** observa manifests do **repositório da aplicação** e entrega os workloads. O **Platform Team** mantém Foundation, Entity Core e os padrões de ambos. O antigo caminho `data-plane/<entity>/domains/<domain>/<env>/applications/` não é mais o destino-alvo da aplicação.

**Por quê:** evita controllers concorrendo pelos mesmos recursos, isola operação AWS dos Deployments e preserva autonomia federada via Git.

**TerraformStack.** O CRD declara a infraestrutura desejada. O Argo CD central sincroniza o CR e o controller agenda tentativas em Argo Workflows. A pipeline resolve a fonte e o destino, gera o plano e consulta OPA. O apply usa o plano autorizado depois do claim da execução. O Simulator valida aplicação automática, aprovação, bloqueio e drift no Hub local. [ADR-0006](0006-terraformrun-and-application-ci.md) · [ADR-0009](0009-terraformstack-runtime.md).

**Telemetria.** O desenho AWS prevê OpenTelemetry Collector nos Management EKS com destino central em Grafana, Loki e Tempo.

**Consequências:** Crossplane não monitora Git, apenas a API. Credenciais IAM precisam de associação autorizada por Environment e validação de Domain. O Argo CD local pode precisar de segregação de acesso PCI. Terraform só converge quando é executado; indisponibilidade do Hub interrompe reconciliação das Capabilities centrais, não necessariamente os workloads.

## Applications e sincronização no Hub

O catálogo do Hub declara uma `Application` por componente que precisa ser sincronizado e verificado de forma independente. Cada recurso tem uma `Application` responsável. Argo Workflows e Crossplane Core são compartilhados pelas etapas do laboratório. As etapas locais usam sync manual para permitir a inspeção da revisão e do diff antes de aplicar a mudança.

O `TerraformStack` segue essa divisão. A Application `terraformstack-crd` publica o contrato na Etapa 04, permitindo validá-lo na API antes de ligar o controller. As Applications `terraformstack-controller` e `terraformstack-workflows` entregam o controller e o WorkflowTemplate na Etapa 05. O controller cria cada Workflow em resposta à solicitação. O Simulator da Etapa 06 validou a pausa para aprovação dentro do Workflow. A Etapa 13 usa o mesmo contrato de autorização no primeiro apply da rede UAT. [ADR-0006](0006-terraformrun-and-application-ci.md)

[Arquitetura](../architecture.md) · [Índice](README.md)

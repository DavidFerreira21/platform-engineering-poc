# ADR-0003 — Execução e reconciliação sem sobreposição

**Estado:** decisão vigente, implementação pendente · **Data:** 19/11/2026

**Decisão:** **um Crossplane central** no Control Plane com XRDs/Compositions/Providers para Capabilities AWS, sem Crossplane em cascata. **Argo CD central** sincroniza CRs de Capability do Git com a Kubernetes API do Hub; não observa YAMLs lógicos de Domain/Environment como CRs. **Argo CD local** do Management EKS administra configuração, namespaces, políticas e workloads em clusters autorizados. **Terraform + Argo Workflows** coordena Foundation/EKS, plan, gates, apply e verificação; Argo CD não executa Terraform.

**Complemento sobre personas e GitOps (ADR-0007):** **DevOps Federado** gerencia os manifests de **Domain Infrastructure** no monorepo de laboratório; o **Application Team** possui o repositório separado com código, GitHub Actions e **manifests Kubernetes da aplicação**. Argo CD **central** aplica CRs Crossplane do repositório de infraestrutura; Argo CD **local** observa manifests do **repositório da aplicação** e entrega os workloads. O **Platform Team** mantém Foundation, Entity Core e os padrões de ambos. O antigo caminho `data-plane/<entity>/domains/<domain>/<env>/applications/` não é mais o destino-alvo da aplicação.

**Por quê:** evita controllers concorrendo pelos mesmos recursos, isola operação AWS dos Deployments e preserva autonomia federada via Git.

**Complemento (21/11/2026):** o `TerraformStack` CRD + controller é o contrato declarativo **proposto de infraestrutura desejada**, com plan inicial/periódico, **detecção de drift** e **Policy Engine** para auto apply, aprovação ou bloqueio. O Argo CD central aplica o CR; Argo Workflows realiza plan/gate/apply com status, sem `apply` contínuo em toda reconciliação. `TerraformRun` pode ser o nome de uma execução isolada, se necessário. O **GitHub Actions das Domain Applications ficará em repositório de aplicação separado**, enquanto o Argo CD local mantém a entrega por GitOps. A coleta nos Management EKS será por **OpenTelemetry Collector/gateway**, para visualização central no Grafana com Loki/Tempo. Ver [ADR-0006](0006-terraformrun-and-application-ci.md).

**Consequências:** Crossplane não monitora Git, apenas a API. Credenciais IAM precisam de associação autorizada por Environment e validação de Domain. O Argo CD local pode precisar de segregação de acesso PCI. Terraform só converge quando é executado; indisponibilidade do Hub interrompe reconciliação das Capabilities centrais, não necessariamente os workloads.

## Applications e sincronização no Hub

O catálogo do Hub declara uma `Application` por componente que precisa ser sincronizado e verificado de forma independente. Cada recurso tem uma `Application` responsável. Argo Workflows e Crossplane Core são compartilhados pelas etapas do laboratório. As etapas locais usam sync manual para permitir a inspeção da revisão e do diff antes de aplicar a mudança.

O `TerraformStack` segue essa divisão. A Application `terraformstack-crd` publica o contrato na Etapa 05, permitindo validá-lo na API antes de ligar o controller. A Application `terraformstack-controller` entrega o controller e o `WorkflowTemplate` na Etapa 06. O controller cria cada Workflow em resposta à solicitação. A pausa para aprovação do plano ocorre dentro do Workflow na Etapa 09. [ADR-0006](0006-terraformrun-and-application-ci.md)

[Arquitetura](../architecture.md) · [Índice](README.md)

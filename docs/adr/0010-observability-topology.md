# ADR-0010 — Observabilidade por ambiente com consulta central

[← Decisões](README.md) · [Arquitetura](../architecture.md) · [Fluxo de observabilidade](../flows/observability.md)

**Estado:** decisão aceita, implementação pendente · **Data:** 28/11/2026

## Contexto

A POC tem uma conta AWS da Plataforma e contas distintas para UAT e PROD. Cada Environment prevê um Management EKS, um EKS PCI e um EKS Non-PCI. O Management EKS executa o Argo CD local e concentra serviços de gestão do ambiente. A equipe precisa investigar aplicações, GitOps e provisionamento durante falhas do próprio Management EKS.

Este ADR consolida a topologia de observabilidade descrita nos [ADR-0002](0002-topology-and-ownership.md), [ADR-0003](0003-reconciliation-and-gitops.md) e [ADR-0006](0006-terraformrun-and-application-ci.md). O [ADR-0008](0008-managed-argocd-connectivity.md) define a conectividade para GitOps.

## Decisão

- **Cada cluster coleta sua própria telemetria.** Management EKS, PCI e Non-PCI têm OTel Collectors locais no desenho-alvo. A coleta do Management EKS inclui as métricas Prometheus do Argo CD local e os sinais disponíveis dos add-ons.
- **Métricas chegam diretamente ao AMP do ambiente.** Cada coletor envia via Prometheus `remote_write`, autenticado com IAM/SigV4. Um workspace do Amazon Managed Service for Prometheus (AMP) por Environment concentra as séries autorizadas. O Gateway do Management EKS fica fora do caminho de métricas dos workloads.
- **Logs e traces são processados no ambiente.** Os coletores enviam ao OTel Gateway do Management EKS do próprio Environment. O Gateway encaminha logs ao Loki e traces ao Tempo. Os dois backends usam armazenamento S3 na mesma conta AWS.
- **Amazon Managed Grafana no Hub é o ponto central de consulta.** O Grafana consulta os workspaces AMP via IAM cross-account e as APIs de leitura do Loki e Tempo por AWS PrivateLink, com conexão à VPC do Hub e um proxy autenticado no ambiente. Essa conexão é dedicada à observabilidade e preserva VPCs sem peering ou Transit Gateway.
- **O Hub tem um caminho próprio.** CloudWatch recebe os sinais operacionais essenciais do Hub conforme a interface de cada componente. A investigação correlaciona logs, métricas e Events do TerraformStack com revisão Git, Workflow e condições do Crossplane. Para serviços AWS, usar compartilhamento via CloudWatch cross-account dentro da mesma região quando aplicável. Um check externo ao Management EKS sinaliza perda de ingestão.

O Platform Team mantém a instrumentação-base, os add-ons e os acessos. Terraform provisiona os recursos AWS do plano; o Argo CD central distribui os componentes Kubernetes da plataforma via GitOps segundo o perfil de cada cluster. O Argo CD local entrega as aplicações do Domain. O mesmo recurso deve ter um único reconciliador.

## Alternativas avaliadas

| Alternativa | Avaliação |
| :--- | :--- |
| Roteamento dos três sinais pelo Management EKS | Simplifica o encaminhamento. Uma falha do Gateway interrompe também a ingestão de métricas dos workloads. |
| Loki e Tempo compartilhados no Hub | Reduz o número de instalações. Amplia a dependência do Hub durante falhas e concentra tráfego e dados de ambientes distintos. |
| AMP direto por ambiente e backends de logs/traces no Management EKS | Mantém a ingestão de métricas independente do Gateway e permite exercitar Loki e Tempo na POC com os clusters já planejados. |

## Consequências e validação

- Quando o Management EKS fica indisponível, os workloads que continuam em execução seguem enviando métricas ao AMP. O caminho de logs e traces fica indisponível até a recuperação do Gateway e dos backends. Medir filas, retries e perda. Testar a persistência da fila onde a recuperação após reinício for necessária.
- Tratar o uso da mesma ingestão e dos mesmos backends por PCI e Non-PCI como hipótese do desenho. Antes de habilitar telemetria PCI, validar acesso de escrita e leitura, retenção, conteúdo permitido e a fronteira exigida. Se o requisito incluir isolamento de ingestão, armazenamento ou falhas, usar componentes separados para PCI. Permissões de painel do Grafana organizam a visualização; os controles de consulta devem ser aplicados à fonte de dados e ao backend.
- Para uma operação de produção com requisitos mais fortes de disponibilidade, a ingestão e a consulta necessárias ao diagnóstico devem sobreviver à perda do cluster de gestão. Serviços gerenciados ou infraestrutura de observabilidade com ciclo de vida próprio são opções para esse cenário.
- O aceite começa em UAT com Management e Non-PCI nas Etapas 26–27. PCI e PROD entram quando autorizados nas Etapas 30–31. A disposição dos componentes está decidida. Dimensionamento, retenção, custo e evidências de recuperação permanecem para as validações previstas. [Roteiro](../learning-roadmap.md#fase-6--observabilidade-operação-e-governança)

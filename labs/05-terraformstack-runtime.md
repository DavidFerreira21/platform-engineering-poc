# Etapa 05 — Controller e pipeline do TerraformStack

## Objetivo

Instalar o controller e a pipeline canônica que executa stacks Terraform a partir da intenção sincronizada pelo Argo CD.

## O que será feito

Construir a imagem do controller e o runner genérico. Publicar o WorkflowTemplate, seus scripts, o perfil do Simulator e a política OPA em ConfigMaps imutáveis. Instalar o controller com a configuração de integração Argo.

## Por que fizemos assim

O controller Go utiliza Controller Runtime no padrão Kubebuilder, com watches, cache, work queue e Leader Election. Dois Pods disputam uma Lease; apenas o líder reconcilia. Drift e retry usam temporizadores por stack. [Implementação](../control-plane/terraformstack/controller/README.md).

O controller registra a intenção, verifica dependências, seleciona o alias cadastrado e preserva uma tentativa ativa por CR. A autorização é habilitada por padrão e configurável por integração; o lab mantém a referência com `true`. A pipeline de referência resolve Git, contexto, PVC, backend e configuração a partir do perfil aprovado. Os steps Terraform usam a imagem comum e executam o script montado. O OPA roda em sua imagem oficial e monta o pacote versionado. O step `verify` pode executar `verify.sh` na raiz da stack. [Fluxo da pipeline](../docs/flows/terraformstack.md#3-pipeline-e-aprovação) · [Perfil do Simulator](../workflow-plane/argo-workflows/terraformstack/simulator.json)

> A instalação e o aceite desta etapa no Hub EKS estão pendentes. Os critérios abaixo descrevem a entrega prevista.

## Pré-requisitos

Concluir as Etapas 02–04 no Hub EKS. Preparar as imagens do controller e do runner em registry acessível ao Hub e registrar seus digests nos manifests da revisão que será sincronizada.

## Aceite no Hub EKS

- Adequar os destinos por ARN, AppProjects, ServiceAccounts, RBAC e admissão. Validar o perfil do Simulator e o storage autorizado para artefatos e state de teste.
- Inspecionar destino, revisão e diff de cada Application. Sincronizar WorkflowTemplate e configuração da pipeline antes de iniciar solicitações do Simulator.
- Validar duas réplicas saudáveis e uma Lease com líder ativo. Conferir logs JSON, Events e o endpoint de métricas do controller.
- Em atualizações de contrato, drenar as tentativas e suspender novos agendamentos pelo GitOps. Revisar imagem, CRD, configuração e admissão na mesma revisão.

## O que observar

O controller inicia com a integração de referência e preserva uma tentativa ativa por CR. A Etapa 06 confirma o comportamento executando o Simulator. [Contrato de execução](../docs/flows/terraformstack-execution-contract.md)

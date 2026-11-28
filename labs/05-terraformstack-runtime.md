# Etapa 05 — Controller e pipeline do TerraformStack

## Objetivo

Instalar o controller e a pipeline canônica que executa stacks Terraform a partir da intenção sincronizada pelo Argo CD.

## O que será feito

Construir a imagem do controller e o runner genérico. Publicar o WorkflowTemplate, seus scripts, o perfil do Simulator e a política OPA em ConfigMaps imutáveis. Instalar o controller com a configuração de integração Argo.

## Por que fizemos assim

O controller Go utiliza Controller Runtime no padrão Kubebuilder, com watches, cache, work queue e Leader Election. Dois Pods disputam uma Lease; apenas o líder reconcilia. Drift e retry usam temporizadores por stack. [Implementação](../control-plane/terraformstack/controller/README.md).

O controller registra a intenção, verifica dependências, seleciona o alias cadastrado e preserva uma tentativa ativa por CR. A autorização é habilitada por padrão e configurável por integração; o lab mantém a referência com `true`. A pipeline de referência resolve Git, contexto, PVC, backend e configuração a partir do perfil aprovado. Os steps Terraform usam a imagem comum e executam o script montado. O OPA roda em sua imagem oficial e monta o pacote versionado. O step `verify` pode executar `verify.sh` na raiz da stack. [Fluxo da pipeline](../docs/flows/terraformstack.md#3-pipeline-e-aprovação) · [Perfil do Simulator](../workflow-plane/argo-workflows/terraformstack/simulator.json)

> O smoke test no Hub EKS está registrado abaixo. A instalação pelo catálogo GitOps segue com aceite próprio.

## Pré-requisitos

Concluir as Etapas 02–04 no Hub EKS. Preparar as imagens do controller e do runner em registry acessível ao Hub e registrar seus digests nos manifests da revisão que será sincronizada.

## Aceite no Hub EKS

- Adequar os destinos por ARN, AppProjects, ServiceAccounts, RBAC e admissão. Validar o perfil do Simulator e o storage autorizado para artefatos e state de teste.
- Inspecionar destino, revisão e diff de cada Application. Sincronizar WorkflowTemplate e configuração da pipeline antes de iniciar solicitações do Simulator.
- Validar duas réplicas saudáveis e uma Lease com líder ativo. Conferir logs JSON, Events e o endpoint de métricas do controller.
- Em atualizações de contrato, drenar as tentativas e suspender novos agendamentos pelo GitOps. Revisar imagem, CRD, configuração e admissão na mesma revisão.

## O que observar

O controller inicia com a integração de referência e preserva uma tentativa ativa por CR. A Etapa 06 confirma o comportamento executando o Simulator. [Contrato de execução](../docs/flows/terraformstack-execution-contract.md)

## Imagens e armazenamento no Hub EKS

Construir as imagens do controller e do runner para a arquitetura dos nós e publicá-las em repositórios privados do ECR autorizados. Registrar a referência imutável por digest na configuração de implantação. Os nós do Auto Mode precisam alcançar o registry e ter permissão de pull. O build na máquina do operador produz as imagens que serão publicadas. Para atender aos dois tipos de nó pela mesma referência, publicar um índice multiarch com `docker buildx build --platform linux/amd64,linux/arm64 --tag "$image_tag" --push <contexto>` e confirmar as duas plataformas com `docker buildx imagetools inspect "$image_tag"`. O ensaio no NodePool Graviton depende desse build e da verificação das imagens e providers do Workflow.

O [overlay EKS](../control-plane/terraformstack/overlays/eks/kustomization.yaml) usa a StorageClass `terraformstack-simulator-eks` para os dois PVCs do Simulator. Ela usa volumes gp3 criptografados pelo provisionador do EKS Auto Mode. Com `WaitForFirstConsumer`, um PVC fica `Pending` até a criação do primeiro Pod que o utiliza. [StorageClass do Auto Mode](https://docs.aws.amazon.com/eks/latest/userguide/create-storage-class.html)

Para repetir o smoke test com aplicação direta, usar a função `hub` da Etapa 02. Definir `controller_image` e `runner_image` com as referências completas por digest no ECR privado. Os valores são resolvidos na sessão do operador. Conferir o diff antes de aplicar.

```bash
kubectl kustomize workflow-plane/argo-workflows/terraformstack \
  | sed "s|terraformstack-runner:local|${runner_image:?Informe o digest do runner}|g" \
  | hub apply -f -

kubectl kustomize control-plane/terraformstack/overlays/eks \
  | sed "s|terraformstack-controller:local|${controller_image:?Informe o digest do controller}|g" \
  | hub apply -f -

hub -n platform-system rollout status deployment/terraformstack-controller
hub -n platform-system get leases
```

## Evidência no Hub EKS

Em 28/11/2026, aplicamos os manifests renderizados do WorkflowTemplate e do controller para testar o runtime. O EKS baixou as imagens do ECR privado por digest. O controller ficou com duas réplicas prontas e uma Lease. Consultamos as métricas por port-forward local e verificamos os Events do TerraformStack na API. A Etapa 06 usa essas instalações para verificar o efeito do Workflow no ConfigMap de teste.

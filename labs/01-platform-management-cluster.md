# Etapa 01 — Bootstrap do Hub EKS Auto Mode na conta de plataforma

[← Guias das etapas](README.md) · [Arquitetura](../docs/architecture.md) · [Roteiro](../docs/learning-roadmap.md)

## Objetivo

Provisionar a base do Hub `platform-hub` na conta de plataforma e confirmar que o operador consegue acessar a API Kubernetes. A Etapa 02 usa essa base para validar a entrega pelo Argo CD.

## O que será feito

- Configurar o perfil AWS `platform-hub` com IAM Identity Center e validar a conta de destino.
- Executar a stack Terraform [`platform-hub`](../foundation/stacks/platform-hub/main.tf) para criar rede, EKS Auto Mode, capability gerenciada do Argo CD e bucket S3 para os futuros states do TerraformStack.
- Confirmar a conectividade com a API, a prontidão dos NodePools e o estado da capability. Usar um workload curto para validar o provisionamento de nós, DNS e HTTP.

## Por que fizemos assim

O bootstrap roda manualmente porque estabelece a infraestrutura que vai hospedar o próprio TerraformStack. A stack mantém seu state local, fora do checkout, e usa referências entre os módulos para criar a rede, o cluster e o Argo CD em ordem. O bucket S3 com versionamento e locking por arquivo atende aos states das execuções posteriores.

A VPC do Hub tem três subnets privadas de aplicação e três de database, uma de cada por AZ. O NAT Gateway regional dá saída às subnets de aplicação pelo Internet Gateway. As subnets de database mantêm a rota local. O EKS Auto Mode usa as subnets de aplicação e os NodePools gerenciados `system` e `general-purpose`. [Módulo Network](../foundation/modules/network/main.tf) · [Módulo EKS](../foundation/modules/eks/main.tf)

Durante o bootstrap, o operador pode usar o endpoint público da API, restrito ao IP de saída do seu terminal. O endpoint privado atende ao tráfego interno. A Access Entry dá à role SSO autorizada as permissões Kubernetes. O Argo CD gerenciado tem URL própria e usa o IAM Identity Center da organização para o login.

## 1. Preparar a sessão e os arquivos locais

Usar AWS CLI v2, Terraform 1.10 ou superior e `kubectl`. No assistente SSO, selecionar a conta de plataforma, a role autorizada para o bootstrap e `us-east-1` como região padrão do perfil.

```bash
aws configure sso --profile platform-hub
aws sso login --profile platform-hub
aws sts get-caller-identity --profile platform-hub --no-cli-pager

aws ec2 describe-availability-zones --profile platform-hub --region us-east-1 \
  --filters Name=state,Values=available Name=zone-type,Values=availability-zone \
  --query 'AvailabilityZones[].ZoneName' --output table --no-cli-pager
```

Conferir no STS se a sessão está na conta de plataforma esperada.

Entrar em `foundation/stacks/platform-hub` e preencher `hub.auto.tfvars` a partir do [modelo](../foundation/stacks/platform-hub/hub.tfvars.example). Fixar a conta esperada em `account_id`. A configuração padrão usa o CIDR `10.80.0.0/16`, as AZs `us-east-1a/b/c` e a versão Kubernetes mais recente em suporte padrão. Informar a instância IAM Identity Center, sua região primária e o ID do grupo Argo CD autorizado em `identity_center`. Para acesso do terminal pela internet, preencher `api_public_access_cidrs` com o IP público de saída no formato `/32`.

Configurar `backend_override.tf` com o caminho absoluto do state local, conforme o [guia da stack](../foundation/stacks/README.md#preparar-o-acesso-e-as-entradas). Em uma retomada, usar o mesmo arquivo de state. O Git ignora o override e o `hub.auto.tfvars`. Manter o state e seu backup em diretório protegido.

## 2. Provisionar a base do Hub

Na pasta `foundation/stacks/platform-hub`, inicializar o Terraform, validar a configuração e gerar o plano:

```bash
terraform init
terraform validate
terraform plan
```

Revisar conta, região, custo e alterações. O plano inicial deve incluir seis subnets privadas, NAT Gateway regional, Internet Gateway, endpoint gateway S3, EKS Auto Mode, capability Argo CD e bucket de state do executor. Conferir que `compute_config` inclui os NodePools e a role IAM dos nós. Se habilitado, o acesso público deve mostrar apenas os CIDRs autorizados.

Após aprovar a operação, executar na mesma pasta:

```bash
terraform apply
```

Revisar o plano apresentado pelo `apply` e confirmar apenas se as alterações corresponderem ao que foi aprovado. Aguardar a mensagem `Apply complete!` antes de verificar o EKS e o Argo CD.

Em caso de falha parcial, usar `terraform state list` para identificar os recursos já registrados e gerar um novo plano a partir do mesmo state. Se `CreateCluster` rejeitar NodePools com `nodeRoleArn` vazio, verificar se o módulo EKS está atualizado para a versão fixada no repositório e rodar `terraform init` para resolver a dependência.

## 3. Verificar EKS e Argo CD

```bash
aws eks update-kubeconfig --profile platform-hub --region us-east-1 --name platform-hub
hub_context="$(aws eks describe-cluster --profile platform-hub --region us-east-1 \
  --name platform-hub --query 'cluster.arn' --output text)"
hub() { kubectl --context "$hub_context" "$@"; }

hub get --raw=/readyz
hub get nodepools
hub get nodes
aws eks describe-capability --profile platform-hub --region us-east-1 \
  --cluster-name platform-hub --capability-name argocd \
  --query 'capability.{Status:status,Issues:health.issues,URL:configuration.argoCd.serverUrl}'
```

A API deve responder ao `kubectl`. Os NodePools podem aparecer com zero nós até haver Pods para agendar. A capability deve ficar `ACTIVE`. Se o `kubectl` expirar pelo WSL, verificar o IP de saída e a lista `api_public_access_cidrs`; se retornar `Unauthorized`, conferir a sessão SSO e a Access Entry. [Diagnóstico do EKS Auto Mode](https://docs.aws.amazon.com/eks/latest/userguide/auto-troubleshoot.html)

## 4. Exercitar o primeiro workload

Com o Hub e o contexto conferidos, criar um Pod descartável para testar que a imagem é baixada, que o Pod é agendado e que o DNS do cluster e a saída HTTP funcionam:

```bash
hub -n default run hub-network-smoke --image=busybox:1.37.0 --restart=Never --command -- \
  sh -ec 'nslookup kubernetes.default.svc.cluster.local; wget -q -O /dev/null http://example.com'
hub -n default wait --for=jsonpath='{.status.phase}'=Succeeded pod/hub-network-smoke --timeout=8m
hub -n default logs hub-network-smoke
hub get nodes
```

Se o Pod ficar em `Pending`, consultar `hub -n default describe pod hub-network-smoke` e `hub get nodeclaims`. Se o teste falhar depois de agendado, conferir os logs do Pod. Ao terminar, remover apenas o Pod de teste:

```bash
hub -n default delete pod hub-network-smoke --ignore-not-found
```

## O que observar

- A stack registra a infraestrutura no mesmo state local entre execuções.
- O operador consulta a API Kubernetes a partir do WSL. Os dois NodePools aparecem em `Ready`.
- O workload de teste provoca a criação de um nó e termina com sucesso após as consultas de DNS e HTTP.
- A capability Argo CD aparece `ACTIVE` e fornece a URL de acesso.

Na execução inicial, foram confirmados o acesso pelo WSL, os dois NodePools em `Ready` e a capability `ACTIVE` sem alertas. O teste com workload e o primeiro sync GitOps ainda precisam ser executados. [Etapa 02 — Primeiro sync no Argo CD](02-argocd-central.md)

# Etapa 01 — Bootstrap do Hub EKS Auto Mode na conta de plataforma

[← Guias das etapas](README.md) · [Arquitetura](../docs/architecture.md) · [Roteiro](../docs/learning-roadmap.md)

## Requisitos

- Uma conta AWS para a plataforma e acesso administrativo autorizado para o bootstrap. Use uma role via IAM Identity Center com as permissões necessárias para provisionar a stack e acessar o cluster.
- Uma instância do IAM Identity Center configurada, com a região primária e o ID do grupo autorizado a entrar no Argo CD.
- Checkout do repositório e **Terraform 1.10 ou superior**, **AWS CLI v2** e **kubectl** instalados na máquina.
- Conectividade à API do Hub a partir do terminal e um diretório protegido, fora do checkout, para guardar o state local e seu backup.

## Objetivo

Provisionar a base do Hub `platform-hub` na conta de plataforma e confirmar que o operador consegue acessar a API Kubernetes. A Etapa 02 usa essa base para validar a entrega pelo Argo CD.

## O que será feito

1. **Preparar o Platform Hub.** Configurar o perfil AWS `platform-hub` com IAM Identity Center, conferir a conta com STS e preencher os arquivos locais de parâmetros e backend.
2. **Provisionar a base da plataforma.** A stack [`platform-hub`](../foundation/stacks/platform-hub/main.tf) é o código Terraform que combina a rede padronizada, o cluster EKS em Auto Mode, a EKS Capability for Argo CD e o bucket S3 para os futuros states do TerraformStack. Gerar o plano, revisar as alterações e aplicar depois da aprovação.
3. **Verificar a infraestrutura criada.** Conferir o state local e os outputs da stack. Atualizar o kubeconfig, testar o acesso à API Kubernetes e conferir o estado dos NodePools e da capability Argo CD.
4. **Exercitar o cluster.** Rodar um workload descartável para verificar o agendamento do Pod, o download da imagem, a resolução DNS e a saída HTTP. Remover o Pod de teste ao terminar.

## Por que fizemos assim

Como este é o primeiro cluster da POC, executamos o Terraform na nossa máquina, com a sessão temporária do IAM Identity Center e o state do bootstrap guardado localmente. A partir daí, os componentes da plataforma rodam na AWS. As execuções posteriores de infraestrutura coordenadas pelo TerraformStack usarão o Hub EKS e terão state no backend S3, com chaves segregadas por destino e stack. A Etapa 12 valida a identidade do executor e o acesso a esse backend. [Roteiro da Etapa 12](../docs/learning-roadmap.md#etapa-12--identidade-do-executor-e-state)

A VPC do Hub tem três subnets privadas de aplicação e três de database, uma de cada por AZ. O NAT Gateway regional dá saída às subnets de aplicação pelo Internet Gateway. As subnets de database mantêm a rota local. O EKS Auto Mode usa as subnets de aplicação e os NodePools gerenciados `system` e `general-purpose`. [Módulo Network](../foundation/modules/network/main.tf) · [Módulo EKS](../foundation/modules/eks/main.tf)

Durante o bootstrap, o operador pode usar o endpoint público da API, restrito ao IP de saída do seu terminal. O endpoint privado atende ao tráfego interno. A Access Entry dá à role SSO autorizada as permissões Kubernetes. O Argo CD gerenciado tem URL própria e usa o IAM Identity Center da organização para o login.

## 1. Preparar a sessão e os arquivos locais

No assistente SSO, selecionar a conta de plataforma, a role autorizada para o bootstrap e `us-east-1` como região padrão do perfil.

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

Na pasta da stack, conferir os recursos registrados e os destinos apresentados nos outputs:

```bash
terraform state list
terraform output
```

Atualizar o kubeconfig e consultar a API do Hub, os NodePools e a capability:

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

Na execução inicial, foram confirmados o acesso pelo WSL, os dois NodePools em `Ready` e a capability `ACTIVE` sem alertas. Em 28/11/2026, repetimos o bootstrap e o Pod de teste terminou em `Succeeded`, confirmando provisionamento do nó, download da imagem, DNS e saída HTTP. A primeira consulta ao log retornou erro TLS do kubelet. As consultas de log funcionaram durante o [smoke test de Argo Workflows](03-argo-workflows.md#evidência-no-hub-eks). [Etapa 02 — Primeiro sync no Argo CD](02-argocd-central.md)

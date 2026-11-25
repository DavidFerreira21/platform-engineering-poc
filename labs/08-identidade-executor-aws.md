# Etapa 08 — Integração AWS do executor em UAT com STS e state S3

[← Etapa 07](07-registro-ambientes-aws.md) · [Roadmap](../docs/learning-roadmap.md#fase-2--terraformstack-e-primeira-execução-na-aws) · [IAM](../foundation/iam/lab08/README.md)

## Objetivo

Validar o acesso do Pod à conta UAT e ao backend S3. O Workflow deste exercício roda diretamente uma stack de validação que lê os YAMLs de Network e EKS com Configuration e produz um plan de teste. O bootstrap pode criar o bucket de state. A rede e os clusters são provisionados em etapas posteriores.

O [marco local 06B](../docs/learning-roadmap.md#etapa-06b--terraformstack-de-ponta-a-ponta-no-kind) cobre a execução iniciada pelo CR, OPA, aprovação, drift, destroy e status no `kind`. Seu guia executável está planejado.

## O que será feito

1. Confirmar a conta UAT via IAM Identity Center (Lab 07) e roles de bootstrap/execução autorizadas.
2. Criar/configurar o bucket S3 de state com criptografia SSE-S3, versionamento, Block Public Access e **`use_lockfile = true`**.
3. Executar `Configuration` com `foundation/config/network.yaml`, `foundation/config/eks.yaml` e overrides de Entity/Environment.
4. Construir e carregar a imagem do Terraform Executor no `kind`.
5. Preparar STS temporário, executar um Job de identidade e um **Argo Workflow** com `terraform init`, `validate` e `plan` de uma stack **sem recursos de infraestrutura**.
6. Remover Pods/Workflows/Secret após o teste; **preservar o bucket de state** para o Lab 09.

## Por que fizemos assim

O Hub `kind` não dispõe de EKS Pod Identity. Para a POC supervisionada, IAM Identity Center fornece a sessão humana, que assume uma role de bootstrap autorizada. O Pod recebe **temporariamente** as credenciais dessa role por Kubernetes Secret e usa STS para assumir a role de execução da UAT.

O state é segregado por Entity/Environment/stack no S3. O locking nativo `use_lockfile=true` evita introduzir uma tabela DynamoDB apenas para state locking (Terraform 1.10+). O S3 precisa estar disponível **antes** do primeiro init com backend remoto.

Configuration lê os YAMLs por hierarquia: valores globais → Entity → Environment, com sobrescrita de valores mais específicos. O executor prepara a estrutura e verifica sem provisionar networking — isso pertence à Etapa 09.

**Atenção:** o script de bootstrap não cria IAM roles, e o script de bucket só cria recursos após `--execute`. Se a credencial STS expirar enquanto um Workflow aguarda aprovação em etapas futuras, outro bootstrap será necessário. Não há renovação automática.

## Pré-requisitos

- Mac/PC com `docker`, `kind`, `kubectl`, Python 3, AWS CLI v2 e `terraform >=1.10` para testes locais opcionais.
- Repositório atualizado e Hub criado pelas Etapas 01–06, com namespace `argo-workflows` e Argo Workflows operacionais.
- IAM Identity Center `platform-poc-uat` validado na Etapa 07.
- Duas roles autorizadas: **Bootstrap** (somente STS para UAT) e **Execution UAT** (STS/S3 backend); veja [IAM e policies](../foundation/iam/lab08/README.md).
- Permissão SSO para criar/configurar o bucket S3. Isso provoca **custos de armazenamento/operações S3**, ainda que o lab não crie VPC, EKS ou EC2.

## 1. Conferir destino e infraestrutura local

Na raiz do repositório:

```bash
git pull --ff-only
aws sso login --profile platform-poc-uat
read -r -p 'Account ID UAT: ' uat_account
read -r -p 'Region UAT (ex.: us-east-1): ' uat_region
bash foundation/scripts/verify-environment-identity.sh \
  platform-poc-uat "$uat_account" "$uat_region"

kubectl config current-context
kubectl --context kind-platform-management -n argo-workflows get deployments
```

O script valida conta e região contra o que você informou. **Confira antes o registro versionado** em [`foundation/config/entity-a/uat/environment.yaml`](../foundation/config/entity-a/uat/environment.yaml). Não use IDs/roles de outros ambientes.

## 2. Criar e verificar o backend remoto de state

Escolha um nome S3 globalmente único, por exemplo com prefixo `terraform-state-poc-` e sufixo do número da conta UAT. **O bucket tem custo e permanece para os labs seguintes.**

```bash
read -r -p 'Bucket S3 exclusivo para state: ' state_bucket

# Preview, sem criação.
python3 foundation/scripts/prepare-state-backend.py \
  --profile platform-poc-uat \
  --account "$uat_account" --region "$uat_region" \
  --bucket "$state_bucket"

# Somente após revisar e autorizar o bucket.
python3 foundation/scripts/prepare-state-backend.py \
  --profile platform-poc-uat \
  --account "$uat_account" --region "$uat_region" \
  --bucket "$state_bucket" --execute
```

O script verifica conta/região, habilita **S3 Block Public Access, SSE-S3 e versionamento**. Não gera/commita credenciais. Em `us-east-1`, cria bucket sem `LocationConstraint`. Uma colisão de nome ou erro de permissão aborta a operação; não contorne permissões.

A **role de execução UAT** precisa de S3 List/Get/Put e `DeleteObject` **restrito ao `.tflock`**, de acordo com o prefixo utilizado. O prefixo da verificação é `platform/entity-a/uat/lab08/`. A stack real da Etapa 09 terá sua própria key. [Política orientativa](../foundation/iam/lab08/README.md#permissões-para-backend-s3).

## 3. Validar Configuration (sem AWS)

A stack de teste está em [`foundation/lab08/validation-stack/`](../foundation/lab08/validation-stack/main.tf), usando [Configuration](../foundation/modules/configuration/main.tf). Dados de referência: `network.yaml` e `eks.yaml`; Environment UAT sobrescreve campos específicos sem descartar outras opções.

Opcionalmente, no computador:

```bash
terraform -chdir=foundation/lab08/validation-stack init -backend=false
terraform -chdir=foundation/lab08/validation-stack validate
```

O teste local não faz `terraform apply`. O Workflow testará também o backend real.

## 4. Construir e carregar a imagem do Terraform Executor

```bash
docker build \
  -f workflow-plane/terraform-executor/Dockerfile \
  -t terraformstack-executor:lab08 .

kind load docker-image terraformstack-executor:lab08 \
  --name platform-management

kubectl --context kind-platform-management -n argo-workflows \
  apply -f workflow-plane/argo-workflows/lab08/workflowtemplate.yaml
```

A imagem contém Terraform, AWS CLI, o módulo Configuration e a stack de **validação**. Não contém credenciais. A imagem é local à POC e precisa ser reconstruída para novos valores de YAML ou código. A imagem de produção será evoluída para obter artefatos `sourceRef` versionados; isso **não está implementado** nesta etapa.

## 5. Preparar credenciais temporárias

```bash
read -r -p 'Role ARN Bootstrap (pode estar na UAT): ' bootstrap_role
read -r -p 'Role ARN Execution UAT: ' uat_role

python3 foundation/scripts/bootstrap-workflow-credentials.py prepare \
  --profile platform-poc-uat \
  --uat-account "$uat_account" \
  --region "$uat_region" \
  --bootstrap-role-arn "$bootstrap_role" \
  --target-role-arn "$uat_role"
```

O script testa o encadeamento STS **antes** de criar o Secret de credenciais temporárias e o ConfigMap de configuração do profile; recusa substituir objetos já existentes. O Secret fica em `argo-workflows`. **Não faça `kubectl get secret -o yaml` nem `kubectl describe pod` contendo env para compartilhar com terceiros.**

Se ocorrer `AccessDenied`, revise trust policy e permissões STS das roles sem ampliá-las genericamente.

## 6. Testar STS dentro de um Pod

```bash
kubectl --context kind-platform-management -n argo-workflows create \
  -f foundation/iam/lab08/verify-uat-executor-job.yaml

kubectl --context kind-platform-management -n argo-workflows wait \
  --for=condition=complete --timeout=180s \
  job/terraform-executor-identity-check

kubectl --context kind-platform-management -n argo-workflows logs \
  job/terraform-executor-identity-check
```

O `Account` retornado precisa ser **exatamente** o UAT. Não avance se a role não corresponder ao destino.

## 7. Executar a validação pelo Argo Workflows

```bash
python3 foundation/scripts/submit-lab08-workflow.py \
  --bucket "$state_bucket" \
  --account "$uat_account" \
  --region "$uat_region" --stack network
```

O comando imprime o nome do Workflow gerado. Acompanhe pela Argo Workflows UI da Etapa 03 ou por:

```bash
kubectl --context kind-platform-management -n argo-workflows \
  get workflows -o wide

kubectl --context kind-platform-management -n argo-workflows \
  logs -l workflows.argoproj.io/workflow -c main --tail=100
```

O container verifica o `GetCallerIdentity` para UAT e realiza `terraform init` no bucket/state key separados, `terraform validate` e **`terraform plan` sem aplicar**. Na saída, conferir `configuration_loaded = true` e o CIDR de UAT `10.81.0.0/16`. Repetir com `--stack eks` para comprovar leitura de outro YAML.

**Limitação:** o workflow não executa o controller `TerraformStack` real e não materializa `sourceRef` do Git/Registry. Também não incorpora aprovação humana neste teste: a aprovação do plano imutável é implementada no Lab 09. Não confundir este smoke test do executor com o motor completo.

## O que observar

- A identidade humana SSO, a role temporária de bootstrap e a role de execução UAT são distintas.
- Pod consegue ler state no S3 autorizado usando role UAT e locking nativo, sem credenciais permanentes.
- A estrutura `Global → Entity → Environment` conserva valores vizinhos no override do `network.vpc.cidr`.
- O Argo Workflow termina `Succeeded` após `init/validate/plan`. Nenhum `terraform apply` ou provisionamento de VPC/EKS ocorre.
- O Secret de bootstrap tem expiração curta e **não** é renovado pela plataforma.

## 8. Limpeza

```bash
kubectl --context kind-platform-management -n argo-workflows \
  delete job terraform-executor-identity-check --ignore-not-found

kubectl --context kind-platform-management -n argo-workflows \
  delete workflows -l workflows.argoproj.io/workflow-template=lab08-terraform-executor-check \
  --ignore-not-found

kubectl --context kind-platform-management -n argo-workflows \
  delete -f workflow-plane/argo-workflows/lab08/workflowtemplate.yaml \
  --ignore-not-found

python3 foundation/scripts/bootstrap-workflow-credentials.py cleanup
unset bootstrap_role uat_role uat_account uat_region state_bucket
```

Se não houver label do WorkflowTemplate na versão do Argo utilizada, remova **somente os Workflows `lab08-terraform-check-*`** identificados na lista; não execute deleção genérica de todos os Workflows.

**Não excluir o bucket S3 de state durante a limpeza.** A Etapa 09 utilizará state separado por stack, na mesma estrutura protegida.

[Próximo passo — Etapa 09: Stack Network](../docs/learning-roadmap.md#fase-2--terraformstack-e-primeira-execução-na-aws)

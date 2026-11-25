# Etapa 08 — Credenciais temporárias do Terraform Executor (kind → UAT)

[← Etapa 07](07-registro-ambientes-aws.md) · [Roadmap](../docs/learning-roadmap.md#fase-2--terraformstack-e-primeira-execução-na-aws) · [IAM do Lab 08](../foundation/iam/lab08/README.md)

## Objetivo

Comprovar que **um Pod no Hub kind** obtém identidade AWS temporária de bootstrap e assume **somente a role autorizada da conta UAT**, sem access keys permanentes ou segredos no Git. Este é o laboratório de identidade; a stack Network, o state S3 e o `terraform apply` real ficam para a continuação desta fase.

## O que será feito

- Reutilizar o login IAM Identity Center validado na [Etapa 07](07-registro-ambientes-aws.md).
- Com duas roles previamente autorizadas (bootstrap na conta de plataforma e execução em UAT), verificar o encadeamento STS.
- Criar **Secret temporário** e ConfigMap exclusivamente no namespace `argo-workflows` do Hub `kind`.
- Rodar Job de teste no Pod com `aws --profile uat-executor sts get-caller-identity` e depois limpar.
- Preparar as entradas do executor real, **sem criar state nem provisionar AWS nesta etapa**.

## Por que fizemos assim

No Hub local ainda não existe EKS Pod Identity; a sessão de IAM Identity Center fica no computador. O script assume uma **role intermediária com permissão mínima** e compartilha com o Pod apenas essa sessão STS efêmera. Dentro do Pod, AWS CLI usa um profile com `credential_source = Environment` para assumir a **role de UAT**. É o mesmo limite de confiança que pretendemos preservar quando o Hub usar EKS Pod Identity.

**Isto não é autenticação autônoma:** a credencial de origem expira, Kubernetes Secret exige RBAC/admission e o operador precisará renovar o bootstrap. O Secret **não** deve ficar em cluster compartilhado ou exposto a Workflows arbitrários.

## Pré-requisitos

- Etapa 07 concluída: IAM Identity Center e `platform-poc-uat` validados por STS, conta/região do registro UAT confirmadas.
- Hub `kind-platform-management` recriado e `argo-workflows` instalado; `kubectl config current-context` deve ser `kind-platform-management`. O script recusa outros contextos.
- AWS CLI v2, `kubectl`, Docker e Python 3 disponíveis localmente.
- Duas roles **já criadas por administrador autorizado**, com trust e políticas mínimas descritas em [IAM do laboratório](../foundation/iam/lab08/README.md). A role de bootstrap deve conseguir `sts:AssumeRole` somente na role UAT; a role UAT tem permissões da etapa futura, a conceder somente quando necessárias.
- Testar num cluster local isolado, com acesso administrativo somente do operador. Nada sensível em commits.

## 1. Autenticar e conferir UAT

```bash
aws sso login --profile platform-poc-uat
read -r -p 'Account ID de UAT no environment.yaml: ' uat_account
read -r -p 'Region de UAT no environment.yaml: ' uat_region
bash foundation/scripts/verify-environment-identity.sh \
  platform-poc-uat "$uat_account" "$uat_region"
kubectl config current-context
kubectl --context kind-platform-management get ns argo-workflows
```

A etapa não cria a role de UAT, não altera permission sets e não muda o cadastro do Environment.

## 2. Preparar uma sessão temporária para o Pod

Defina **apenas os ARNs**, nunca access keys. O script verifica o perfil, assume a role de bootstrap por STS e testa a role UAT **antes** de criar o Secret. Se já houver Secret anterior, é necessário limpar primeiro.

```bash
read -r -p 'ARN da role de bootstrap na conta da plataforma: ' bootstrap_role
read -r -p 'ARN da role de execução na conta UAT: ' uat_role

python3 foundation/scripts/bootstrap-workflow-credentials.py prepare \
  --profile platform-poc-uat \
  --uat-account "$uat_account" \
  --region "$uat_region" \
  --bootstrap-role-arn "$bootstrap_role" \
  --target-role-arn "$uat_role"
```

A saída mostra **somente ARNs/contas e expiração**; nunca imprima ou faça `kubectl get secret -o yaml`. O Secret `terraform-executor-bootstrap` contém **apenas credenciais de origem temporárias**, não as credenciais finais de UAT. O ConfigMap `terraform-executor-aws-config` contém role de destino e configuração do profile. O script usa `kubectl create`, **não** `kubectl apply`, para não inserir credenciais no annotation `last-applied-configuration`.

Se falhar com `AccessDenied`, conferir **ambos** os lados do STS: a role de origem precisa permitir `sts:AssumeRole` e a trust policy do destino precisa confiar na origem. Não afrouxar a trust policy para conseguir passar no Lab.

## 3. Comprovar a identidade do Pod

```bash
kubectl --context kind-platform-management -n argo-workflows create \
  -f foundation/iam/lab08/verify-uat-executor-job.yaml

kubectl --context kind-platform-management -n argo-workflows wait \
  --for=condition=complete --timeout=180s job/terraform-executor-identity-check

kubectl --context kind-platform-management -n argo-workflows logs \
  job/terraform-executor-identity-check
```

Compare `Account` com o ID de UAT. O campo `Arn` deve representar a role **de execução UAT**, e **não** a role SSO humana ou a de bootstrap. Este Job só chama STS; não executa Terraform. Se houver erro, confira `kubectl describe job` e `kubectl logs`; não exiba variáveis de ambiente ou credenciais.

## O que observar

- Perfil SSO humano acessa a conta UAT prevista na Foundation.
- Role de bootstrap consegue **somente** assumir a role de execução UAT permitida.
- Pod no Hub obtém role da UAT sem access keys fixas e sem URL/role arbitrária no `TerraformStack.spec`.
- Sessão expira; não há renovação automática. Para a Etapa 09, cada `plan` e `apply` deverá receber uma sessão válida, inclusive após o `suspend` de aprovação; um Secret expirado deve causar falha fechada.

## 4. Limpeza obrigatória

```bash
kubectl --context kind-platform-management -n argo-workflows delete \
  job terraform-executor-identity-check --ignore-not-found
python3 foundation/scripts/bootstrap-workflow-credentials.py cleanup
unset bootstrap_role uat_role uat_account uat_region
```

Limpar o Secret **antes de compartilhar o cluster ou terminar a sessão do Lab**. Não versionar Secret, token, cache SSO, arquivos AWS locais nem outputs sensíveis. O state/backend AWS permanece para implementação separada e expressamente autorizada da Etapa 08–09.

[Próxima etapa: Stack Network e primeiro plan/apply AWS](../docs/learning-roadmap.md#fase-2--terraformstack-e-primeira-execução-na-aws)

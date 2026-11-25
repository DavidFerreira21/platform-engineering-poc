# Identidade temporária para Terraform Executor no Hub kind

Este diretório contém o **desenho de IAM** da Etapa 08. Nenhuma role é criada automaticamente aqui. Provisionar as roles por um administrador autorizado, em contas determinadas pelo cadastro de Environment, antes de executar o bootstrap.

## Papéis separados

1. **Identidade humana:** acesso com IAM Identity Center ao perfil `platform-poc-uat`. Usado apenas no computador para assumir uma role inicial temporária.
2. **Role de bootstrap na conta de plataforma (ou na própria UAT quando a POC possui apenas uma conta AWS):** exemplo `PlatformPocTerraformBootstrap`. Trust policy limitada ao principal/role de SSO autorizado; permissões **somente `sts:AssumeRole`** para a role UAT aprovada. Não conceder permissões diretamente de EC2/EKS a essa role.
3. **Role de execução na conta UAT:** exemplo `PlatformPocTerraformExecution`. Trust policy permite que a role de bootstrap (ARN exato) seja assumida. A role de execução contém apenas as permissões necessárias às stacks e backend aprovados, a detalhar antes do primeiro apply.

> A conta de plataforma **pode ser igual à UAT** na POC pessoal, desde que as duas roles sejam distintas e o privilégio de bootstrap continue restrito a `sts:AssumeRole`. Não colocar nomes de contas ou ARNs como entradas livres do `TerraformStack`: o controller deverá resolver `Entity + Environment` a partir do registro confiável. Na Etapa 08, os ARNs são parâmetros de bootstrap operacional, não parte do CR.

### Exemplo de política de permissão da role de bootstrap (ajustar placeholders)

```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "AssumeApprovedUatExecutor",
    "Effect": "Allow",
    "Action": ["sts:AssumeRole"],
    "Resource": "arn:aws:iam::<UAT_ACCOUNT_ID>:role/PlatformPocTerraformExecution"
  }]
}
```

### Exemplo de trust policy da role UAT

```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "TrustOnlyPlatformBootstrap",
    "Effect": "Allow",
    "Principal": {
      "AWS": "arn:aws:iam::<PLATFORM_ACCOUNT_ID>:role/PlatformPocTerraformBootstrap"
    },
    "Action": "sts:AssumeRole"
  }]
}
```

A role de bootstrap também precisa de uma trust policy para o principal SSO autorizado e o permission set SSO precisa permitir `sts:AssumeRole` nessa role. Não usar `Principal: "*"`, wildcard de roles ou permissões `AdministratorAccess` como atalho.

## Como funciona o laboratório

O script `foundation/scripts/bootstrap-workflow-credentials.py` assume a role de **bootstrap**, valida que ela pode assumir UAT e armazena somente sua sessão STS **temporária** em um Secret no namespace `argo-workflows` do Hub local. Um ConfigMap separado contém apenas os dados não secretos do profile `uat-executor` para que o AWS CLI/SDK do Pod encadeie `AssumeRole` usando `credential_source = Environment`.

O Job `foundation/iam/lab08/verify-uat-executor-job.yaml` comprova que o **Pod**, usando a role de bootstrap, consegue assumir a role UAT e obter o `GetCallerIdentity` correto. Não executa Terraform, nem cria recursos AWS.

**Limitações intencionais:** Kubernetes Secrets não são um cofre, o bootstrap depende de login humano e a sessão expira. Não há renovação automática, nem EKS Pod Identity em `kind`. Proteger acesso a `Secret`, `Pod`, `Workflow`, `ConfigMap`, `exec` e logs por RBAC/admission; este laboratório é isolado e supervisionado. O primeiro apply exigirá controle de permissão do Pod executor e renovação da sessão antes do gate de aprovação.

[Guia 07](../../../labs/07-registro-ambientes-aws.md) · [Guia 08](../../../labs/08-identidade-executor-aws.md)

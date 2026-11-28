# Identidade temporária e backend do Terraform Executor

A Etapa 12 valida o executor a partir do Hub EKS. Os exemplos abaixo descrevem o desenho de IAM e um utilitário supervisionado de diagnóstico. Nenhuma role é criada automaticamente aqui. Provisionar as roles por um administrador autorizado, em contas determinadas pelo cadastro de Environment, antes de executar o bootstrap.

## Diagnóstico local versus identidade-alvo AWS

Os exemplos abaixo registram um diagnóstico inicial supervisionado com SSO e Secret temporário. A Etapa 12 usa a identidade de Pod do Hub EKS. Na topologia-alvo, Platform, UAT e PROD são contas distintas. Na Etapa 12, validar a identidade de Pod do executor no Hub EKS e a assunção da role UAT por essa identidade; não transportar sessões humanas em Secrets como padrão de operação AWS. O controller não assume a role de provisionamento.

## Papéis do diagnóstico local

1. **Identidade humana:** acesso com IAM Identity Center ao perfil `platform-poc-uat`. Usado apenas no computador para assumir uma role inicial temporária.
2. **Role de bootstrap na conta de plataforma (ou na própria UAT em um diagnóstico isolado de conta única, que não representa a topologia-alvo):** exemplo `PlatformPocTerraformBootstrap`. Trust policy limitada ao principal/role de SSO autorizado; permissões **somente `sts:AssumeRole`** para a role UAT aprovada. Não conceder permissões diretamente de EC2/EKS a essa role.
3. **Role de execução na conta UAT:** exemplo `PlatformPocTerraformExecution`. Trust policy permite que a role de bootstrap (ARN exato) seja assumida. A role de execução contém apenas as permissões necessárias às stacks e backend aprovados, a detalhar antes do primeiro apply.

Na integração AWS prevista, o destino e as roles autorizadas virão do registro confiável que a pipeline consulta a partir de `spec.context`. O Simulator atual não implementa essa resolução AWS. No utilitário supervisionado, os ARNs são parâmetros de bootstrap operacional. [Contrato de execução](../../../docs/flows/terraformstack-execution-contract.md)

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

## Backend do Hub EKS

A Etapa 01 provisiona, na conta da Plataforma, o bucket de state do executor e publica sua configuração no output `executor_backend`. A Etapa 12 valida o acesso da identidade autorizada ao backend, com key e lock segregados por destino e stack. A pipeline ainda precisa ligar o registro do destino, a identidade de execução e o vínculo do backend antes do primeiro apply na spoke. [Stack do Hub](../../stacks/README.md) · [Roteiro da Etapa 12](../../../docs/learning-roadmap.md#etapa-12--identidade-do-executor-e-state)

## Backend S3 do diagnóstico local

O diagnóstico abaixo usa uma key exclusiva de preflight, segregada dos states das stacks:

```text
s3://<STATE_BUCKET>/platform/entity-a/uat/preflight/validation.tfstate
s3://<STATE_BUCKET>/platform/entity-a/uat/preflight/validation.tfstate.tflock
```

A role `PlatformPocTerraformExecution` (na UAT) deve receber uma policy **escopada** a esse bucket/prefixo. Exemplo orientativo, substituindo os placeholders:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ListOnlyPreflightStatePrefix",
      "Effect": "Allow",
      "Action": ["s3:ListBucket"],
      "Resource": "arn:aws:s3:::<STATE_BUCKET>",
      "Condition": {
        "StringLike": {"s3:prefix": ["platform/entity-a/uat/preflight/*"]}
      }
    },
    {
      "Sid": "ReadWriteState",
      "Effect": "Allow",
      "Action": ["s3:GetObject", "s3:PutObject"],
      "Resource": "arn:aws:s3:::<STATE_BUCKET>/platform/entity-a/uat/preflight/validation.tfstate"
    },
    {
      "Sid": "LockObject",
      "Effect": "Allow",
      "Action": ["s3:GetObject", "s3:PutObject", "s3:DeleteObject"],
      "Resource": "arn:aws:s3:::<STATE_BUCKET>/platform/entity-a/uat/preflight/validation.tfstate.tflock"
    }
  ]
}
```

A role de bootstrap **não** recebe permissões de S3: somente a role de execução UAT administra o state. O bucket terá versionamento, SSE-S3 e Block Public Access. O uso de `use_lockfile = true` requer Terraform 1.10+; não usamos DynamoDB para locking. Nunca conceder `s3:DeleteObject` sobre o próprio state como atalho.

Para o diagnóstico local, um operador com IAM Identity Center pode preparar um bucket de preflight com `prepare-state-backend.py --execute`. O backend do Hub EKS é provisionado pela raiz `platform-hub`. O namespace dos artefatos do Argo e o state do Terraform têm responsabilidades distintas.

## Como funciona o laboratório

O script `foundation/scripts/bootstrap-workflow-credentials.py` assume a role de **bootstrap**, valida que ela pode assumir UAT e armazena somente sua sessão STS **temporária** em um Secret no namespace `argo-workflows` do Hub local. Um ConfigMap separado contém apenas os dados não secretos do profile `uat-executor` para que o AWS CLI/SDK do Pod encadeie `AssumeRole` usando `credential_source = Environment`.

O [Job de diagnóstico](verify-uat-executor-job.yaml) consulta `GetCallerIdentity` com credenciais temporárias. O fluxo da Etapa 12 valida a identidade de Pod no Hub EKS antes de integrar a execução real de Network.

**Limitações intencionais:** Kubernetes Secrets não são um cofre, o bootstrap depende de login humano e a sessão expira. A sessão temporária exige renovação manual no diagnóstico. Proteger acesso a `Secret`, `Pod`, `Workflow`, `ConfigMap`, `exec` e logs por RBAC/admission; este laboratório é isolado e supervisionado. No diagnóstico local, conferir a validade da sessão antes de executar o Job. O primeiro apply na spoke depende do aceite da identidade de Pod do executor no Hub EKS.

[Registro UAT](../../../docs/learning-roadmap.md#etapa-11--registro-de-uat) · [Identidade e state](../../../docs/learning-roadmap.md#etapa-12--identidade-do-executor-e-state)

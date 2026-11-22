# Lab 03B — Registro das contas UAT e PROD no Git

[← Lab 03](03-crossplane-core.md) · [Índice dos labs](README.md)

O Platform Team registra duas contas AWS distintas para a `entity-a`. O entregável são dois arquivos `environment.yaml` versionados na Foundation. A validação deste exercício usa consultas de identidade e tem custo AWS zero.

## 1. Criar os registros de conta

Use uma branch de trabalho a partir da raiz do monorepo. Para criar um registro, copie o [modelo de Environment](../foundation/config/entity-a/uat/environment.example.yaml) e preencha os campos conforme a tabela. Se os registros já estiverem presentes, revise os arquivos versionados.

| Environment | Registro | `aws.accountRef` | `aws.region` |
| --- | --- | --- | --- |
| `uat` | [UAT](../foundation/config/entity-a/uat/environment.yaml) | `platform-poc-uat` | `us-east-1` |
| `prod` | [PROD](../foundation/config/entity-a/prod/environment.yaml) | `platform-poc-prod` | `us-east-1` |

Os dois registros usam `entity: entity-a`, `foundation.profile: standard` e `core.profile: eks-standard`. O campo `environment` corresponde ao diretório e a `accountRef` identifica a conta aprovada pelo Platform Team. A referência lógica e a região ficam no Git. IDs numéricos reais e credenciais ficam fora do repositório.

## 2. Validar as duas contas

Tenha a AWS CLI v2 e acesso à UAT e à PROD pelo IAM Identity Center. Se precisar configurar os perfis, execute `aws configure sso --profile platform-poc-uat` e `aws configure sso --profile platform-poc-prod`. No assistente de cada perfil, selecione a conta correspondente, o permission set atribuído e a região do YAML. Quando a sessão expirar, renove com `aws sso login --profile <perfil>`. [IAM Identity Center na AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sso.html)

No portal de acesso AWS, confirme os IDs das contas pelas referências lógicas da tabela. Informe os valores apenas no terminal:

```bash
read -r -p 'ID da conta UAT conferido no portal AWS: ' uat_account
read -r -p 'ID da conta PROD conferido no portal AWS: ' prod_account
if [[ "$uat_account" == "$prod_account" ]]; then
  echo 'ERRO: UAT e PROD precisam apontar para contas distintas.'
else
  bash foundation/scripts/verify-environment-identity.sh platform-poc-uat "$uat_account" us-east-1 &&
    bash foundation/scripts/verify-environment-identity.sh platform-poc-prod "$prod_account" us-east-1
fi
```

O resultado esperado é um `OK` para cada conta. O verificador confere a conta e a região configuradas no perfil e a conta e a role retornadas pelo STS.

### Falha controlada e recuperação

Use o perfil de UAT com o ID esperado da PROD. O verificador deve recusar a divergência antes da chamada ao STS. Em seguida, use o perfil correto de PROD e confirme o resultado `OK`.

```bash
bash foundation/scripts/verify-environment-identity.sh platform-poc-uat "$prod_account" us-east-1
bash foundation/scripts/verify-environment-identity.sh platform-poc-prod "$prod_account" us-east-1
```

## 3. Revisar e publicar

```bash
git add foundation/config/entity-a/{uat,prod}/environment.yaml
git diff --cached --check
git diff --cached -- foundation/config/entity-a/{uat,prod}/environment.yaml
```

Confira referências diferentes, a correspondência dos Environments com o caminho de cada arquivo e a ausência de identificadores numéricos. Publique os registros por PR. O aceite do lab exige os dois YAMLs versionados e a identidade validada para cada conta. Ao finalizar, limpe as variáveis do terminal com `unset uat_account prod_account`.

## Antes do Lab 04

O primeiro ObjectStorage será validado na UAT. Antes de provisioná-lo, o Provider AWS no Hub precisa de uma identidade de workload dedicada. A opção proposta é `WebIdentity` com issuer HTTPS verificável pela AWS, audiência `sts.amazonaws.com` e confiança limitada à ServiceAccount prevista. O aceite inclui STS no Hub, renovação da sessão e rejeição de outra identidade. [OIDC no IAM](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc.html) · [Autenticação do Provider AWS](https://github.com/crossplane-contrib/provider-upjet-aws/blob/main/AUTHENTICATION.md)

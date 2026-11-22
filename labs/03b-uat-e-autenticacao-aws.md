# Lab 03B — Registro da conta UAT no Git

[← Lab 03](03-crossplane-core.md) · [Índice dos labs](README.md)

O Platform Team registra uma conta AWS existente como destino do Environment `uat` da `entity-a`. O entregável é o `environment.yaml` versionado na Foundation. O exercício usa consultas de identidade e tem custo AWS zero.

## 1. Criar o registro da UAT

Use uma branch de trabalho a partir da raiz do monorepo. Para criar o registro pela primeira vez, copie o [modelo](../foundation/config/entity-a/uat/environment.example.yaml) para `foundation/config/entity-a/uat/environment.yaml` e preencha os valores abaixo. Se o arquivo já estiver presente, revise o [registro versionado](../foundation/config/entity-a/uat/environment.yaml).

```yaml
entity: entity-a
environment: uat
aws:
  accountRef: platform-poc-uat
  region: us-east-1
foundation:
  profile: standard
core:
  profile: eks-standard
```

- `entity` e `environment` devem corresponder ao caminho da Foundation.
- `aws.accountRef` é a referência lógica pela qual o Platform Team identifica a conta aprovada. O vínculo com o ID numérico é conferido no portal AWS e na validação via STS. O repositório guarda a referência lógica e a região. IDs numéricos reais e credenciais ficam fora do Git.
- `foundation.profile` e `core.profile` identificam os perfis de referência das duas camadas. São campos declarativos do registro de Environment.

## 2. Validar o destino no terminal

Para o teste, tenha a AWS CLI v2 e um perfil do IAM Identity Center apontando para a conta UAT. Se precisar configurar o perfil, use `aws configure sso --profile platform-poc-uat` e selecione a conta UAT, o permission set atribuído e a região do registro. Use `aws sso login --profile platform-poc-uat` quando a sessão expirar. [Configuração do IAM Identity Center na AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sso.html)

Confira no portal AWS que o nome da conta corresponde a `aws.accountRef`, informe o ID confirmado apenas no terminal e use a região do YAML.

```bash
read -r -p 'ID da conta UAT conferido no portal AWS: ' uat_account
uat_region=us-east-1
bash foundation/scripts/verify-uat-identity.sh platform-poc-uat "$uat_account" "$uat_region"
```

O resultado esperado é `OK`. O STS confirma a conta e a role da sessão. O script também confere se a região foi configurada no perfil.

### Falha controlada e recuperação

Repita a verificação com uma região diferente da declarada. O primeiro comando deve recusar a divergência antes de chamar o STS. O segundo deve retornar `OK`.

```bash
wrong_region=us-west-2
if [[ "$uat_region" == "$wrong_region" ]]; then wrong_region=us-east-1; fi
bash foundation/scripts/verify-uat-identity.sh platform-poc-uat "$uat_account" "$wrong_region"
bash foundation/scripts/verify-uat-identity.sh platform-poc-uat "$uat_account" "$uat_region"
```

## 3. Revisar e publicar

```bash
git add foundation/config/entity-a/uat/environment.yaml
git diff --cached --check
git diff --cached -- foundation/config/entity-a/uat/environment.yaml
```

Revise a referência da conta, a região e o caminho do Environment. Publique o registro por PR na branch de trabalho. O aceite do lab é o YAML versionado e a verificação do destino concluída. Ao finalizar, use `unset uat_account uat_region wrong_region` para limpar os valores do terminal.

## Antes do Lab 04

A primeira Capability AWS exige uma identidade de workload dedicada para o Provider no Hub. A opção proposta é `WebIdentity` com issuer HTTPS verificável pela AWS, audiência `sts.amazonaws.com` e confiança limitada à ServiceAccount prevista. Teste STS no Hub, renovação da sessão e rejeição de outra identidade antes de autorizar ObjectStorage. [OIDC no IAM](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc.html) · [Autenticação do Provider AWS](https://github.com/crossplane-contrib/provider-upjet-aws/blob/main/AUTHENTICATION.md)

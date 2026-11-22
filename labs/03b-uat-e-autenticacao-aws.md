# Lab 03B — Registro das contas UAT e PROD no Git

[← Lab 03](03-crossplane-core.md) · [Índice dos labs](README.md)

## Objetivo

Registrar UAT e PROD como destinos AWS distintos da `entity-a` e verificar que cada perfil aponta para a conta declarada.

## O que será feito

- Criar ou revisar os dois arquivos `environment.yaml` na Foundation. Eles declaram o ID da conta, a região, o baseline de rede e os tipos de EKS previstos.
- Conferir as identidades das duas contas via IAM Identity Center e STS, provocar uma troca de perfis e verificar que o script rejeita a divergência.
- Revisar os YAMLs e publicar os registros por PR. As contas são pré-requisitos do exercício. As consultas ao STS têm custo AWS zero.

## Pré-requisitos

- AWS CLI v2 e acesso por IAM Identity Center a duas contas AWS existentes, uma para UAT e outra para PROD.
- Checkout do monorepo com Git disponível e permissão para abrir um PR no repositório usado para o exercício.

## 1. Criar os registros de conta

Use uma branch de trabalho a partir da raiz do monorepo. Os registros abaixo descrevem as contas desta POC. Para reproduzir o lab em outra organização, copie o [modelo de Environment](../foundation/config/entity-a/uat/environment.example.yaml) em uma branch própria e preencha os IDs e as regiões das suas contas.

| Environment | Registro | Conta nesta POC |
| --- | --- | --- |
| `uat` | [UAT](../foundation/config/entity-a/uat/environment.yaml) | `platform-poc-uat` |
| `prod` | [PROD](../foundation/config/entity-a/prod/environment.yaml) | `platform-poc-prod` |

O registro identifica a Entity e o Environment, define o destino AWS e declara o que a plataforma prevê implantar.

| Campo | Significado |
| --- | --- |
| `entity` e `environment` | Identidade lógica e ambiente. Os valores correspondem ao caminho na Foundation. |
| `aws.accountId` e `aws.region` | ID de 12 dígitos entre aspas e região de destino. |
| `foundation.networkBaseline` | Baseline de rede da Foundation. |
| `core.eksManagement`, `core.eksNonPci` e `core.eksPci` | Clusters EKS Management, Non-PCI e PCI previstos para o Environment. |

Uma flag `true` inclui o componente no escopo desejado. O provisionamento segue as etapas, aprovações e controles do roteiro. Alterar o YAML apenas registra a intenção.

## 2. Validar as duas contas

Se precisar configurar os perfis, execute `aws configure sso --profile platform-poc-uat` e `aws configure sso --profile platform-poc-prod`. No assistente de cada perfil, selecione a conta correspondente, o permission set atribuído e a região do YAML. Quando a sessão expirar, renove com `aws sso login --profile <perfil>`. [IAM Identity Center na AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sso.html)

No portal de acesso AWS, confira cada conta com `aws.accountId` e use a região de seu YAML ao executar o verificador. Informe os valores de cada registro nos comandos abaixo.

```bash
read -r -p 'aws.accountId do registro UAT: ' uat_account
read -r -p 'aws.region do registro UAT: ' uat_region
read -r -p 'aws.accountId do registro PROD: ' prod_account
read -r -p 'aws.region do registro PROD: ' prod_region
if [[ "$uat_account" == "$prod_account" ]]; then
  echo 'ERRO: UAT e PROD precisam apontar para contas distintas.'
else
  bash foundation/scripts/verify-environment-identity.sh platform-poc-uat "$uat_account" "$uat_region" &&
    bash foundation/scripts/verify-environment-identity.sh platform-poc-prod "$prod_account" "$prod_region"
fi
```

O resultado esperado é um `OK` para cada conta. O verificador confere a conta e a região configuradas no perfil e a conta e a role retornadas pelo STS.

### Falha controlada e recuperação

Use o perfil de UAT com o ID esperado da PROD. O verificador deve recusar a divergência antes da chamada ao STS. Em seguida, use o perfil correto de PROD e confirme o resultado `OK`.

```bash
bash foundation/scripts/verify-environment-identity.sh platform-poc-uat "$prod_account" "$prod_region"
bash foundation/scripts/verify-environment-identity.sh platform-poc-prod "$prod_account" "$prod_region"
```

## 3. Revisar e publicar

```bash
git add foundation/config/entity-a/{uat,prod}/environment.yaml
git diff --cached --check
git diff --cached -- foundation/config/entity-a/{uat,prod}/environment.yaml
```

Confira IDs diferentes, a região, a correspondência dos Environments com o caminho de cada arquivo e as flags esperadas. Mantenha as credenciais no fluxo do IAM Identity Center. Publique os registros por PR. O aceite do lab exige os dois YAMLs versionados e a identidade validada para cada conta. Ao finalizar, limpe as variáveis do terminal com `unset uat_account uat_region prod_account prod_region`.

## Antes do Lab 04

O primeiro ObjectStorage será validado na UAT. Antes de provisioná-lo, o Provider AWS no Hub precisa de uma identidade de workload dedicada. A opção proposta é `WebIdentity` com issuer HTTPS verificável pela AWS, audiência `sts.amazonaws.com` e confiança limitada à ServiceAccount prevista. O aceite inclui STS no Hub, renovação da sessão e rejeição de outra identidade. [OIDC no IAM](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc.html) · [Autenticação do Provider AWS](https://github.com/crossplane-contrib/provider-upjet-aws/blob/main/AUTHENTICATION.md)

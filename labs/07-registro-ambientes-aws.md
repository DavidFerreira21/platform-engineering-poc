# Etapa 07 — Registro e validação de UAT

[← Roteiro da Fase 2](../docs/learning-roadmap.md#fase-2--terraformstack-e-primeira-execução-na-aws) · [Guias das etapas](README.md)

## Objetivo

Registrar UAT como o primeiro destino AWS do `TerraformStack`. Confirmar a identidade do operador na conta declarada para `entity-a/uat` e conferir a região do perfil.

## O que será feito

- Criar ou revisar o `environment.yaml` de UAT na Foundation. Ele declara o ID da conta, a região, o baseline de rede e os tipos de EKS previstos.
- Validar o perfil do operador via IAM Identity Center e STS. Provocar e corrigir uma divergência de região.
- Versionar o registro por PR. A conta UAT é um pré-requisito e a consulta ao STS tem custo AWS zero.

## Pré-requisitos

- Ter exercitado o CRD `TerraformStack` e o Workflow com executor simulado nas Etapas 05 e 06.
- AWS CLI v2 e acesso via IAM Identity Center à conta UAT.
- Checkout do monorepo e permissão para abrir um PR no repositório usado no exercício.

## 1. Revisar o registro UAT

Use uma branch de trabalho a partir da raiz do monorepo. Abra o [registro UAT da POC](../foundation/config/entity-a/uat/environment.yaml). Para reproduzir o exercício com sua própria conta, copie o [modelo de Environment](../foundation/environment.example.yaml) para `foundation/config/entity-a/uat/environment.yaml` na sua branch e preencha a conta e a região aprovadas.

| Campo | Significado |
| --- | --- |
| `entity` e `environment` | Identidade lógica e ambiente. Os valores correspondem ao caminho na Foundation. |
| `aws.accountId` e `aws.region` | ID de 12 dígitos entre aspas e região de destino. |
| `foundation.networkBaseline` | Baseline de rede da Foundation. |
| `core.eksManagement`, `core.eksNonPci` e `core.eksPci` | Clusters EKS Management, Non-PCI e PCI previstos para o Environment. |

Uma flag `true` inclui o componente no escopo desejado. Cada provisionamento segue as etapas e aprovações do roteiro.

## 2. Validar o acesso humano a UAT

Se precisar configurar o perfil, execute `aws configure sso --profile platform-poc-uat`. Selecione a conta UAT, o permission set atribuído e a região do YAML. Renove a sessão quando necessário com `aws sso login --profile platform-poc-uat`. [IAM Identity Center na AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sso.html)

Confira o ID no portal de acesso AWS e informe os valores do registro.

```bash
read -r -p 'aws.accountId do registro UAT: ' uat_account
read -r -p 'aws.region do registro UAT: ' uat_region
bash foundation/scripts/verify-environment-identity.sh platform-poc-uat "$uat_account" "$uat_region"
```

O resultado esperado é `OK`. O script compara a conta e a região do perfil com os valores esperados e confere a conta e a role retornadas pelo STS.

### Falha controlada e recuperação

Passe uma região diferente da declarada. O script deve interromper a validação antes de chamar o STS. Confira a região do perfil e repita com a região correta.

```bash
if [[ "$uat_region" == us-west-2 ]]; then
  wrong_region=us-east-1
else
  wrong_region=us-west-2
fi
bash foundation/scripts/verify-environment-identity.sh platform-poc-uat "$uat_account" "$wrong_region"
aws configure get region --profile platform-poc-uat
bash foundation/scripts/verify-environment-identity.sh platform-poc-uat "$uat_account" "$uat_region"
```

## 3. Publicar e limpar

```bash
git add foundation/config/entity-a/uat/environment.yaml
git diff --cached --check
git diff --cached -- foundation/config/entity-a/uat/environment.yaml
unset uat_account uat_region wrong_region
```

Se você criou ou alterou o registro, publique a mudança por PR. O aceite da etapa é o YAML de UAT versionado e o resultado `OK` para o perfil do operador. A Etapa 08 prepara a identidade temporária do executor e o backend de estado.

## Uso do registro PROD na Etapa 20

Quando for testar o isolamento entre contas, use o [registro PROD da POC](../foundation/config/entity-a/prod/environment.yaml) ou o modelo de Environment com a conta e a região aprovadas para PROD. Configure o perfil `platform-poc-prod` no IAM Identity Center para essa conta. Valide os dois destinos com IDs distintos antes dos testes de isolamento.

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
unset uat_account uat_region prod_account prod_region
```

# Etapa 07 — Registro e validação de UAT

[← Etapa 06B no roadmap](../docs/learning-roadmap.md#etapa-06b--terraformstack-de-ponta-a-ponta-no-kind) · [Guias das etapas](README.md) · [Etapa 08 →](08-identidade-executor-aws.md)

## Objetivo

A trilha AWS começa aqui, após o marco local 06B previsto no roteiro. Registrar UAT como o primeiro destino AWS do `TerraformStack`. Confirmar a identidade do operador na conta declarada para `entity-a/uat` e conferir a região do perfil.

## O que será feito

- Preparar perfil SSO local, registrar conta/região e comprovar acesso humano, **sem distribuir credenciais ao Kubernetes**.
- Conferir os pré-requisitos de IAM da Etapa 08 para montar a identidade temporária do Executor em uma sessão posterior.
- Criar ou revisar o `environment.yaml` de UAT na Foundation. Ele declara o ID da conta, a região, o baseline de rede e os tipos de EKS previstos.
- Validar o perfil do operador via IAM Identity Center e STS.
- Versionar o registro por PR. A conta UAT é um pré-requisito e a consulta ao STS tem custo AWS zero.

## Por que fizemos assim

O primeiro contato com a AWS começa pela conferência do destino e da identidade do operador. O registro na Foundation documenta a conta e a região esperadas antes de a etapa seguinte configurar a identidade temporária do executor.

## Pré-requisitos

- Ter exercitado o CRD `TerraformStack` e o Workflow com executor simulado nas Etapas 05 e 06.
- AWS CLI v2 e acesso via IAM Identity Center à conta UAT. Tenha à mão o portal de acesso, o ID de UAT e o permission set atribuído. Não é necessário usar access keys.
- Não é necessário ter o Hub kind ativo para esta etapa: o STS é consultado a partir do computador. O bootstrap para Pods será feito **somente na Etapa 08**.
- Checkout do monorepo e permissão para abrir um PR no repositório usado no exercício.

## Checklist de início (para executar a Etapa 07)

```bash
aws --version
git status --short
aws configure list-profiles
```

A conta e a região UAT já estão registradas no arquivo `foundation/config/entity-a/uat/environment.yaml`. Confirme se são **as da conta AWS que você realmente controla**, em vez de substituir sem conferir. Não cole tokens, `AWS_SECRET_ACCESS_KEY` ou respostas completas do STS em commits, issues ou prints. Se o perfil `platform-poc-uat` ainda não existir, configure-o no passo 2.

**Critério de aceite da Etapa 07:** o perfil SSO é renovável e `verify-environment-identity.sh` retorna `OK` para a conta e a região do registro. Isso **não comprova** que o Workflow pode assumir a role da UAT; essa prova pertence à Etapa 08.

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

Se precisar configurar o perfil, execute `aws configure sso --profile platform-poc-uat`. Selecione a conta UAT, o permission set atribuído e a região do YAML. Renove a sessão quando necessário com `aws sso login --profile platform-poc-uat`. Não é necessário conceder permissão administrativa para um simples `GetCallerIdentity`. [IAM Identity Center na AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sso.html)

Confira o ID no portal de acesso AWS e informe os valores do registro.

```bash
read -r -p 'aws.accountId do registro UAT: ' uat_account
read -r -p 'aws.region do registro UAT: ' uat_region
bash foundation/scripts/verify-environment-identity.sh platform-poc-uat "$uat_account" "$uat_region"
```

O resultado esperado é `OK`. O script compara a conta e a região do perfil com os valores esperados e confere a conta e a role retornadas pelo STS. Ele **não** exporta credenciais para variáveis de ambiente nem as cria no cluster.

Se houver uma divergência, confira a conta escolhida no IAM Identity Center e a região do perfil com `aws configure get region --profile platform-poc-uat`. Compare esses valores com o registro UAT antes de repetir a verificação.

## O que observar

- A conferência com o perfil correto retorna `OK` e comprova o acesso humano à conta declarada. A identidade do executor é verificada na etapa seguinte.

## Preparação para a Etapa 08 — identidade do Pod

Depois de obter `OK`, o próximo exercício será [Etapa 08 — Terraform Executor no kind](08-identidade-executor-aws.md). Nele o fluxo será:

```text
AWS CLI + SSO no computador
  → STS AssumeRole (role bootstrap com privilégio mínimo)
  → Secret STS temporário no namespace argo-workflows do kind
  → Pod do Job / Terraform Executor
  → STS AssumeRole na role aprovada da conta UAT
  → aws sts get-caller-identity
```

**Não criar o Secret nesta Etapa 07.** Antes do passo seguinte, um administrador deve disponibilizar as duas roles com trust policies restritas, como descrito no [desenho IAM do Lab 08](../foundation/iam/lab08/README.md). Para uma POC com **apenas uma conta AWS**, as duas roles podem existir na própria UAT, com nomes e políticas diferentes; não há necessidade de abrir uma conta de plataforma só para essa validação. A role de bootstrap não recebe permissões diretas de EC2/EKS.

No `kind`, a credencial STS de bootstrap expira e precisará ser renovada manualmente. Não é uma solução permanente de identidade de workload; EKS Pod Identity/IRSA será avaliado quando o Hub migrar para EKS. Na Etapa 08, um Job verificará a identidade do Pod e haverá um script de limpeza do Secret.

## 3. Publicar e limpar

```bash
git add foundation/config/entity-a/uat/environment.yaml
git diff --cached --check
git diff --cached -- foundation/config/entity-a/uat/environment.yaml
unset uat_account uat_region
```

Se você criou ou alterou o registro, publique a mudança por PR. O aceite da etapa é o YAML de UAT versionado e o resultado `OK` para o perfil do operador. A [Etapa 08](08-identidade-executor-aws.md) prepara a identidade temporária do executor com verificação em Pod; o backend de state será implementado antes do primeiro plan/apply AWS da Etapa 09.

## Uso do registro PROD na Etapa 27

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

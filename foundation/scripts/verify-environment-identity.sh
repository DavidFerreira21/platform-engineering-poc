#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 3 ]]; then
  echo "Uso: bash $0 <perfil-sso> <conta-esperada> <região-esperada>" >&2
  exit 2
fi

profile="$1"
expected_account="$2"
expected_region="$3"

if [[ -z "$profile" || ! "$expected_account" =~ ^[0-9]{12}$ || ! "$expected_region" =~ ^[a-z]{2}(-[a-z0-9]+)+-[0-9]+$ ]]; then
  echo "ERRO: informe um perfil, o ID de 12 dígitos da conta esperada e a região AWS." >&2
  exit 2
fi

if ! command -v aws >/dev/null 2>&1; then
  echo "ERRO: AWS CLI não encontrada." >&2
  exit 1
fi

if ! aws configure get sso_session --profile "$profile" >/dev/null ||
   ! permission_set="$(aws configure get sso_role_name --profile "$profile")"; then
  echo "ERRO: o perfil precisa de uma sessão do IAM Identity Center e um permission set." >&2
  exit 1
fi

if [[ "$(aws configure get sso_account_id --profile "$profile")" != "$expected_account" ]]; then
  echo "ERRO: a conta configurada no perfil difere do registro do Environment." >&2
  exit 1
fi

if [[ "$(aws configure get region --profile "$profile")" != "$expected_region" ]]; then
  echo "ERRO: a região configurada no perfil difere do registro do Environment." >&2
  exit 1
fi

if ! identity="$(AWS_PAGER= AWS_IGNORE_CONFIGURED_ENDPOINT_URLS=true   aws sts get-caller-identity --profile "$profile" --region "$expected_region"     --query '[Account, Arn]' --output text)"; then
  echo "ERRO: falha ao consultar o STS. Confira a sessão do IAM Identity Center." >&2
  exit 1
fi

read -r caller_account caller_arn <<< "$identity"
if [[ "$caller_account" != "$expected_account" ||
      "$caller_arn" != "arn:aws:sts::$expected_account:assumed-role/AWSReservedSSO_${permission_set}_"*/* ]]; then
  echo "ERRO: a identidade retornada pelo STS difere da role esperada para o Environment." >&2
  exit 1
fi

printf 'OK: STS confirmou a role do IAM Identity Center na conta esperada. Região selecionada: %s.\n' "$expected_region"

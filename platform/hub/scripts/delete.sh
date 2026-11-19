#!/usr/bin/env bash
set -euo pipefail

CLUSTER="platform-management"
command -v kind >/dev/null || { echo "ERRO: instale kind." >&2; exit 1; }
if ! kind get clusters | grep -Fxq "$CLUSTER"; then
  echo "Cluster '$CLUSTER' não existe; nenhuma alteração."
  exit 0
fi
echo "ATENÇÃO: este comando remove o cluster local kind '$CLUSTER' e seus dados."
if [[ "${1:-}" != "--confirm" ]]; then
  echo "Para remover, execute: bash platform/hub/scripts/delete.sh --confirm"
  exit 2
fi
kind delete cluster --name "$CLUSTER"
echo "Cluster local removido."

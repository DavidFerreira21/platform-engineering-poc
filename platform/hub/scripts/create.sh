#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
CLUSTER="platform-management"
CONTEXT="kind-${CLUSTER}"

for cmd in kind kubectl docker; do
  command -v "$cmd" >/dev/null 2>&1 || { echo "ERRO: '$cmd' não encontrado." >&2; exit 1; }
done
docker info >/dev/null 2>&1 || { echo "ERRO: Docker não está acessível." >&2; exit 1; }

if kind get clusters | grep -Fxq "$CLUSTER"; then
  echo "O cluster '$CLUSTER' já existe; não será recriado automaticamente."
else
  kind create cluster --name "$CLUSTER" --config "$ROOT/platform/hub/kind-config.yaml" --wait 120s
fi

kubectl --context "$CONTEXT" wait --for=condition=Ready nodes --all --timeout=120s
echo "Cluster local pronto no contexto $CONTEXT."
kubectl --context "$CONTEXT" get nodes -o wide

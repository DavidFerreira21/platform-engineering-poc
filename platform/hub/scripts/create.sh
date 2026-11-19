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
  echo "Cluster '$CLUSTER' já existe; validando sem alterar suas portas."
else
  kind create cluster --name "$CLUSTER" --config "$ROOT/platform/hub/kind-config.yaml" --wait 120s
fi
kubectl --context "$CONTEXT" wait --for=condition=Ready nodes --all --timeout=120s
kubectl --context "$CONTEXT" apply -f "$ROOT/platform/hub/namespaces.yaml"
echo "Hub e namespaces preparados no contexto $CONTEXT."
kubectl --context "$CONTEXT" get nodes -o wide
kubectl --context "$CONTEXT" get namespaces
echo "Próxima etapa: bash platform/hub/scripts/install-traefik.sh"

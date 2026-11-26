#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
CLUSTER="${HUB_KIND_NAME:-platform-management}"
CONTEXT="kind-$CLUSTER"
CHART_VERSION="$(tr -d '[:space:]' < "$ROOT/platform/hub/traefik/chart-version.txt")"

for cmd in kind kubectl helm; do
  command -v "$cmd" >/dev/null 2>&1 || { echo "ERRO: '$cmd' não encontrado." >&2; exit 1; }
done
kind get clusters | grep -Fxq "$CLUSTER" || { echo "ERRO: execute create.sh primeiro." >&2; exit 1; }
kubectl --context "$CONTEXT" get namespace traefik >/dev/null
# A versão do chart é fixa; não instalar a versão latest involuntariamente.
helm repo add traefik https://traefik.github.io/charts --force-update >/dev/null
helm repo update traefik
helm --kube-context "$CONTEXT" upgrade --install traefik traefik/traefik \
  --namespace traefik --version "$CHART_VERSION" \
  --values "$ROOT/platform/hub/traefik/values.yaml" \
  --wait --timeout 5m
kubectl --context "$CONTEXT" -n traefik rollout status deploy/traefik --timeout=180s
kubectl --context "$CONTEXT" apply -f "$ROOT/platform/hub/traefik/demo.yaml"
kubectl --context "$CONTEXT" -n platform-system rollout status deploy/echo --timeout=180s
echo "Traefik instalado. Confira HTTP: curl -fsS --max-time 10 http://demo.localhost:${HUB_HTTP_PORT:-8080}/"

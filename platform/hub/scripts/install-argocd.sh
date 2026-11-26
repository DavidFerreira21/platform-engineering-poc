#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
CLUSTER="${HUB_KIND_NAME:-platform-management}"
CONTEXT="kind-$CLUSTER"
CONFIG="$ROOT/workflow-plane/argocd"
CHART_VERSION="$(tr -d '[:space:]' < "$CONFIG/chart-version.txt")"

for cmd in kind kubectl helm; do
  command -v "$cmd" >/dev/null 2>&1 || { echo "ERRO: '$cmd' não encontrado." >&2; exit 1; }
done
kind get clusters | grep -Fxq "$CLUSTER" || { echo "ERRO: conclua a Etapa 01 primeiro." >&2; exit 1; }
kubectl --context "$CONTEXT" get namespace argocd >/dev/null
kubectl --context "$CONTEXT" wait --for=condition=Established --timeout=60s \
  crd/ingressroutetcps.traefik.io || {
  echo "ERRO: atualize o Traefik com bash platform/hub/scripts/install-traefik.sh." >&2
  exit 1
}

helm repo add argo https://argoproj.github.io/argo-helm --force-update >/dev/null
helm repo update argo
helm --kube-context "$CONTEXT" upgrade --install argocd argo/argo-cd \
  --namespace argocd --version "$CHART_VERSION" --values "$CONFIG/values.yaml" \
  --wait --timeout 5m
kubectl --context "$CONTEXT" wait --for=condition=Established --timeout=60s \
  crd/applications.argoproj.io crd/appprojects.argoproj.io crd/applicationsets.argoproj.io
kubectl --context "$CONTEXT" -n argocd rollout status \
  deployment/argocd-applicationset-controller --timeout=120s
echo "Argo CD pronto. Abra https://argocd.localhost:${HUB_HTTPS_PORT:-8443}."

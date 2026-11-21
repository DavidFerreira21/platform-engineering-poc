#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
CONTEXT="kind-platform-management"
CONFIG="$ROOT/workflow-plane/argocd"
CHART_VERSION="$(tr -d '[:space:]' < "$CONFIG/chart-version.txt")"

for cmd in kind kubectl helm; do
  command -v "$cmd" >/dev/null 2>&1 || { echo "ERRO: '$cmd' não encontrado." >&2; exit 1; }
done
kind get clusters | grep -Fxq platform-management || { echo "ERRO: conclua o Lab 01 primeiro." >&2; exit 1; }
kubectl --context "$CONTEXT" get namespace argocd >/dev/null

helm repo add argo https://argoproj.github.io/argo-helm --force-update >/dev/null
helm repo update argo
helm --kube-context "$CONTEXT" upgrade --install argocd argo/argo-cd \
  --namespace argocd --version "$CHART_VERSION" --values "$CONFIG/values.yaml" \
  --wait --timeout 5m
kubectl --context "$CONTEXT" wait --for=condition=Established --timeout=60s \
  crd/applications.argoproj.io crd/appprojects.argoproj.io
echo "Argo CD pronto. Consulte labs/02-argocd-central.md para acessar a interface local."

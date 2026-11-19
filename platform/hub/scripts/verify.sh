#!/usr/bin/env bash
set -euo pipefail

CLUSTER="platform-management"
CONTEXT="kind-${CLUSTER}"
command -v kind >/dev/null || { echo "ERRO: instale kind." >&2; exit 1; }
command -v kubectl >/dev/null || { echo "ERRO: instale kubectl." >&2; exit 1; }
kind get clusters | grep -Fxq "$CLUSTER" || { echo "ERRO: cluster não existe." >&2; exit 1; }
kubectl --context "$CONTEXT" cluster-info
kubectl --context "$CONTEXT" wait --for=condition=Ready nodes --all --timeout=60s
kubectl --context "$CONTEXT" get nodes
kubectl --context "$CONTEXT" get namespaces
kubectl --context "$CONTEXT" get pods -n kube-system
echo "Smoke test do Hub concluído: API acessível e nós Ready."

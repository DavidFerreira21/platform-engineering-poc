#!/usr/bin/env bash
set -euo pipefail
CLUSTER="platform-management"
CONTEXT="kind-${CLUSTER}"
for cmd in kind kubectl curl; do
  command -v "$cmd" >/dev/null 2>&1 || { echo "ERRO: '$cmd' não encontrado." >&2; exit 1; }
done
kind get clusters | grep -Fxq "$CLUSTER" || { echo "ERRO: cluster não existe." >&2; exit 1; }
kubectl --context "$CONTEXT" cluster-info
kubectl --context "$CONTEXT" wait --for=condition=Ready nodes --all --timeout=60s
for ns in platform-system traefik argocd crossplane-system; do
  kubectl --context "$CONTEXT" get namespace "$ns" >/dev/null
done
kubectl --context "$CONTEXT" -n kube-system rollout status deployment/coredns --timeout=120s
kubectl --context "$CONTEXT" -n traefik rollout status deployment/traefik --timeout=120s
kubectl --context "$CONTEXT" -n platform-system rollout status deployment/echo --timeout=120s
kubectl --context "$CONTEXT" -n traefik get service traefik
kubectl --context "$CONTEXT" -n platform-system get ingress echo
response="$(curl --noproxy '*' --fail --silent --show-error --max-time 15 --resolve demo.localhost:8080:127.0.0.1 http://demo.localhost:8080/)"
if [[ "$response" != "platform-management-ok" ]]; then
  echo "ERRO: rota HTTP incorreta: $response" >&2
  exit 1
fi
echo "OK: kind, namespaces, CoreDNS, Traefik, echo e http://demo.localhost:8080/"
echo "TLS em 8443 fica reservado para serviços futuros; certificados locais não são configurados no Lab 01."

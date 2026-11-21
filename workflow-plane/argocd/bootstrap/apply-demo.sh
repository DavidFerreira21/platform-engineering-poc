#!/usr/bin/env bash
set -euo pipefail
BOOTSTRAP="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONTEXT="kind-platform-management"
REVISION="${1:-}"

if [[ $# -ne 1 || ! "$REVISION" =~ ^[0-9a-f]{40}$ ]]; then
  echo "Uso: bash $0 <SHA completo de um commit publicado no monorepo>" >&2
  exit 2
fi

kubectl --context "$CONTEXT" -n argocd apply -f "$BOOTSTRAP/project.yaml"
sed "s/__LAB02_GIT_SHA__/$REVISION/" "$BOOTSTRAP/application.yaml" |
  kubectl --context "$CONTEXT" -n argocd apply -f -

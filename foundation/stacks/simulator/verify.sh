#!/usr/bin/env bash
set -euo pipefail

namespace="$(jq -er .variables.namespace <<< "$EXECUTION_CONTEXT")"
slot="$(jq -er .variables.slot <<< "$EXECUTION_CONTEXT")"
resource="$(kubectl -n "$namespace" get configmap "$slot" --ignore-not-found -o json)"

if [[ "$(jq -er .intent.operation <<< "$INTENT")" == destroy ]]; then
  if [[ -n "$resource" ]]; then
    echo "O ConfigMap do Simulator permanece após destroy." >&2
    exit 1
  fi
else
  observed="$(jq -ceS '.data["network.json"] | fromjson' <<< "$resource")"
  expected="$("$IAC_BIN" output -json network | jq -ceS .)"
  if [[ "$observed" != "$expected" ]]; then
    echo "O ConfigMap do Simulator diverge do output Terraform." >&2
    exit 1
  fi
fi

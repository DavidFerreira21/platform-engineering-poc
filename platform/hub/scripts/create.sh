#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
cluster="${HUB_KIND_NAME:-platform-management}"
node="$cluster-control-plane"
for cmd in kind kubectl docker; do
  command -v "$cmd" >/dev/null || { echo "ERRO: '$cmd' não encontrado." >&2; exit 1; }
done
docker info >/dev/null 2>&1 || { echo "ERRO: Docker não está acessível." >&2; exit 1; }

create_hub() (
  config_dir="$(mktemp -d)"
  trap 'rm -f "$config_dir/kind.yaml" "$config_dir/policy.yaml"; rmdir "$config_dir"' EXIT
  cat > "$config_dir/policy.yaml" <<'YAML'
apiVersion: audit.k8s.io/v1
kind: Policy
omitStages: [RequestReceived]
rules:
  - level: RequestResponse
    verbs: [update, patch]
    namespaces: [argo-workflows]
    userGroups: [system:serviceaccounts:terraformstack-simulator]
    resources:
      - group: argoproj.io
        resources: [workflows]
  - level: None
YAML
  sed -e "s/hostPort: 8080$/hostPort: ${HUB_HTTP_PORT:-8080}/" \
      -e "s/hostPort: 8443$/hostPort: ${HUB_HTTPS_PORT:-8443}/" \
      "$root/platform/hub/kind-config.yaml" > "$config_dir/kind.yaml"
  cat >> "$config_dir/kind.yaml" <<'YAML'
kubeadmConfigPatches:
  - |
    kind: ClusterConfiguration
    apiServer:
      extraArgs:
        audit-policy-file: /var/local/terraformstack-audit/policy.yaml
        audit-log-path: /var/local/terraformstack-audit/requests.jsonl
        audit-log-maxage: "1"
        audit-log-maxsize: "10"
      extraVolumes:
        - name: terraformstack-audit
          hostPath: /var/local/terraformstack-audit
          mountPath: /var/local/terraformstack-audit
          pathType: DirectoryOrCreate
YAML

  # A política fica no nó kind, inclusive quando Docker Desktop roda em outra VM.
  kind create cluster --name "$cluster" --config "$config_dir/kind.yaml" --wait 120s &
  create_pid=$!
  while kill -0 "$create_pid" 2>/dev/null; do
    if docker exec "$node" mkdir -p /var/local/terraformstack-audit 2>/dev/null &&
       docker cp "$config_dir/policy.yaml" "$node:/var/local/terraformstack-audit/policy.yaml"; then
      break
    fi
    sleep 1
  done
  wait "$create_pid"
)

clusters="$(kind get clusters)"
if [[ $'\n'$clusters$'\n' == *$'\n'"$cluster"$'\n'* ]]; then
  echo "Hub $cluster já existe; conferindo o contexto."
else
  create_hub
fi
kubectl --context "kind-$cluster" wait --for=condition=Ready nodes --all --timeout=120s
kubectl --context "kind-$cluster" apply -f "$root/platform/hub/namespaces.yaml"
echo "Hub e namespaces preparados no contexto kind-$cluster."

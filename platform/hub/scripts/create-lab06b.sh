#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
cluster=platform-management
node="$cluster-control-plane"
clusters="$(kind get clusters)"
if [[ $'\n'$clusters$'\n' == *$'\n'"$cluster"$'\n'* ]]; then
  echo "Já existe um Hub $cluster. Preserve-o e use um Hub descartável para o 06B." >&2
  exit 1
fi

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
    userGroups: [system:serviceaccounts:terraformstack-lab06b]
    resources:
      - group: argoproj.io
        resources: [workflows]
  - level: None
YAML
cat "$root/platform/hub/kind-config.yaml" > "$config_dir/kind.yaml"
cat >> "$config_dir/kind.yaml" <<'YAML'
kubeadmConfigPatches:
  - |
    kind: ClusterConfiguration
    apiServer:
      extraArgs:
        audit-policy-file: /var/local/lab06b-audit/policy.yaml
        audit-log-path: /var/local/lab06b-audit/requests.jsonl
        audit-log-maxage: "1"
        audit-log-maxsize: "10"
      extraVolumes:
        - name: lab06b-audit
          hostPath: /var/local/lab06b-audit
          mountPath: /var/local/lab06b-audit
          pathType: DirectoryOrCreate
YAML

# A política fica no nó kind, inclusive quando Docker Desktop roda em outra VM.
kind create cluster --config "$config_dir/kind.yaml" --wait 120s &
create_pid=$!
while kill -0 "$create_pid" 2>/dev/null; do
  if docker exec "$node" mkdir -p /var/local/lab06b-audit 2>/dev/null &&
     docker cp "$config_dir/policy.yaml" "$node:/var/local/lab06b-audit/policy.yaml"; then
    break
  fi
  sleep 1
done
wait "$create_pid"
kubectl --context "kind-$cluster" apply -f "$root/platform/hub/namespaces.yaml"

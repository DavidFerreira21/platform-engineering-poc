#!/usr/bin/env bash
set -euo pipefail
umask 077

value() { jq -er "$1" <<<"$INTENT"; }
digest() { sha256sum "$1" | cut -d ' ' -f 1; }

workspace="/state/$WORKFLOW"
root="$workspace/source/$(value .path)"
operation="$(value .operation)"
saved="$workspace/plan.bin"

checkout() {
  local destination="$1" revision="$2"
  git init -q "$destination"
  git -C "$destination" fetch -q --depth=1 "$(value .repository)" "$revision"
  git -C "$destination" checkout -q --detach FETCH_HEAD
  [[ "$(git -C "$destination" rev-parse HEAD)" == "$revision" ]]
}

plan() {
  local result=0 path="$1"
  local args=(-input=false -no-color -detailed-exitcode -lock-timeout=30s
              "-var-file=$workspace/variables.tfvars.json" "-out=$path")
  if [[ "$operation" == destroy ]]; then
    args+=(-destroy)
  fi
  terraform -chdir="$root" plan "${args[@]}" > "$workspace/terraform-plan.log" 2>&1 || result=$?
  if [[ "$result" != 0 && "$result" != 2 ]]; then
    echo "terraform plan retornou $result." >&2
    return 1
  fi
  printf '%s\n' "$result"
}

signature() {
  terraform -chdir="$root" show -json "$1" |
    jq -cS '{resource_changes, resource_drift, output_changes}'
}

claim() {
  local object version patch
  local stack="$(value .stack)"
  local args=(-n platform-system terraformstack "$stack")
  for attempt in {1..30}; do
    object="$(kubectl get "${args[@]}" -o json)" || return 1
    if ! jq -e --argjson desired "$INTENT" --arg workflow "$WORKFLOW" '
      .metadata.uid == $desired.uid and .metadata.generation == $desired.generation and
      (.metadata.annotations["platform.example.io/revision"] // "") == $desired.revisionAnnotation and
      (.metadata.deletionTimestamp != null) == $desired.deleting and
      (.spec.suspend // false | not) and
      .status.execution.workflow == $workflow and .status.execution.intent == $desired and
      (.status.execution.cancelling // false | not) and (.status.execution.applying // false | not) and
      (if $desired.deleting then .status.deletionIntent == $desired else true end)
    ' <<< "$object" >/dev/null; then
      return 1
    fi
    if [[ "$(jq -r '.status.execution.authorization // ""' <<< "$object")" == "$PLAN_HASH" ]]; then
      version="$(jq -er '.metadata.resourceVersion' <<< "$object")"
      patch="$(jq -cn --arg version "$version" '[
        {op:"test",path:"/metadata/resourceVersion",value:$version},
        {op:"add",path:"/status/execution/applying",value:true}
      ]')"
      if kubectl patch "${args[@]}" --subresource=status --type=json -p "$patch" >/dev/null; then
        return 0
      fi
    fi
    sleep 1
  done
  return 1
}

case "$1" in
  prepare)
    mkdir "$workspace"
    revision="$(value .revision)"
    configuration="$(value .configuration)"
    checkout "$workspace/source" "$revision"
    config_checkout="$workspace/source"
    if [[ "$configuration" != "$revision" ]]; then
      config_checkout="$workspace/configuration"
      checkout "$config_checkout" "$configuration"
    fi
    jq -e --arg path "$config_checkout/$(value .configurationPath)"       '.variables + {config_root: $path}' <<< "$INTENT" > "$workspace/variables.tfvars.json"
    ;;
  init)
    mapfile -t backend < <(jq -r '.backend | to_entries[] | "-backend-config=\(.key)=\(.value)"' <<< "$INTENT")
    terraform -chdir="$root" init -input=false -no-color "${backend[@]}"
    ;;
  validate)
    terraform -chdir="$root" validate -no-color
    ;;
  plan)
    result="$(plan "$saved")"
    terraform -chdir="$root" show -json "$saved" > "$workspace/plan.json"
    jq -c --argjson exitCode "$result" '{operation, target, variables, exitCode: $exitCode}' \
      <<< "$INTENT" > "$workspace/policy-context.json"
    jq -cn --rawfile document "$workspace/plan.json" --rawfile context "$workspace/policy-context.json" \
      '{document:$document, context:$context}' > "$workspace/policy-input.json"
    digest "$saved" > /tmp/plan-hash
    digest "$workspace/plan.json" > /tmp/document-hash
    digest "$workspace/policy-context.json" > /tmp/context-hash
    printf '%s\n' "$result" > /tmp/plan-exit
    ;;
  decision)
    if ! jq -ce --arg doc "$DOCUMENT_HASH" --arg context "$CONTEXT_HASH" \
      --arg op "$operation" --argjson code "$PLAN_EXIT" '
      select(.documentHash == $doc and .contextHash == $context and
             (.reason | type == "string" and length > 0) and
             (.decision as $decision | ["AutoApply","RequireApproval","Deny","NoChanges"] | index($decision)) and
             (if $op == "destroy" then .decision == "RequireApproval" or .decision == "Deny" else true end) and
             (if .decision == "NoChanges" then $code == 0 else true end))
      | del(.documentHash, .contextHash)
    ' <<< "$POLICY_RESULT" >/tmp/policy-result.json; then
      printf '%s\n' '{"decision":"Error","reason":"Decisão OPA inválida ou incompatível com o plano"}' >/tmp/policy-result.json
    fi
    jq -c --arg hash "$PLAN_HASH" --arg policyDigest "$(digest /platform/policy/policy.rego)" \
      '. + {hash:$hash, policyDigest:$policyDigest}' /tmp/policy-result.json > /tmp/decision-record
    jq -r .decision /tmp/decision-record > /tmp/decision
    ;;
  authorize)
    printf '%s\n' none > /tmp/apply-request
    printf '%s\n' Skipped > /tmp/approval-result
    case "$DECISION" in
      AutoApply) [[ "$operation" == apply ]] ;;
      RequireApproval)
        if [[ "$APPROVAL" != Approve || "$APPROVED_HASH" != "$PLAN_HASH" ]]; then
          printf '%s\n' Rejected > /tmp/approval-result
          exit 0
        fi
        ;;
      *) exit 0 ;;
    esac
    printf '%s\n' "$PLAN_HASH" > /tmp/apply-request
    ;;
  apply)
    approved=/tmp/approved-plan.bin
    cp "$saved" "$approved"
    if [[ "$(digest "$approved")" != "$PLAN_HASH" ]]; then
      echo "Hash do plano salvo diverge." >&2
      exit 1
    fi
    plan "$workspace/revalidate.bin" >/dev/null
    signature "$approved" > /tmp/approved.signature
    signature "$workspace/revalidate.bin" > /tmp/current.signature
    if ! cmp -s /tmp/approved.signature /tmp/current.signature; then
      echo "O destino ou o state mudou desde o plan." >&2
      exit 1
    fi
    if ! claim; then
      printf '%s\n' Cancelled > "$workspace/apply-status"
      exit 0
    fi
    if ! terraform -chdir="$root" apply -input=false -no-color -lock-timeout=30s "$approved" > "$workspace/apply.log" 2>&1; then
      echo "terraform apply falhou. Consulte o diagnóstico no volume da execução." >&2
      exit 1
    fi
    printf '%s\n' Applied > "$workspace/apply-status"
    ;;
  verify)
    if [[ "$(<"$workspace/apply-status")" == Cancelled ]]; then
      printf '%s\n' Cancelled > /tmp/apply-result
      exit 0
    fi
    if [[ "$(plan "$workspace/verify.bin")" != 0 ]]; then
      echo "Plan pós-apply ainda apresenta mudanças." >&2
      exit 1
    fi
    printf '%s\n' Verified > /tmp/apply-result
    ;;
  *) exit 2 ;;
esac

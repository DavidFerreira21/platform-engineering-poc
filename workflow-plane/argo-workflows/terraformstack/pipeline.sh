#!/usr/bin/env bash
set -euo pipefail
umask 077

value() { jq -er "$1" <<<"$INTENT"; }
bound() { jq -er "$1" <<<"$EXECUTION_CONTEXT"; }
digest() { sha256sum "$1" | cut -d ' ' -f 1; }
policy_digest() { (cd /platform/policy && LC_ALL=C sha256sum *.rego *.yaml) | sha256sum | cut -d ' ' -f 1; }

envelope() {
  jq -c --arg workflow "$WORKFLOW" --arg hash "$intent_hash" \
    '. + {workflow:$workflow, intentHash:$hash}'
}

result() {
  jq -cn --arg outcome "$1" --arg reason "$2" --arg message "$3" \
    '{outcome:$outcome, reason:$reason, message:$message}' | envelope > /tmp/execution-result
}

block() {
  result Blocked "$1" "$2"
  printf '%s\n' false > /tmp/allowed
}

resolve() {
  local profile=/platform/profile/simulator.json slot repository path reference revision binding resolved
  : > /tmp/execution-result
  : > /tmp/resolved-source
  : > /tmp/execution-context
  : > /tmp/execution-volume
  slot="$(value .intent.stack)"
  repository="$(value .intent.sourceRef.repository)"
  path="$(value .intent.sourceRef.path)"
  if ! jq -e --arg repo "$repository" --arg path "$path" \
      '.repository == $repo and .path == $path' "$profile" >/dev/null; then
    block SourceNotAllowed 'O repositório ou caminho não está autorizado pelo perfil.'
    return
  fi
  if ! jq -e --arg slot "$slot" --argjson intent "$(value .intent)" \
      '.context == $intent.context and .stackNamespace == $intent.namespace and (.slots | index($slot) != null)' \
      "$profile" >/dev/null; then
    block ContextNotAllowed 'O contexto e a identidade da stack não estão autorizados pelo perfil.'
    return
  fi
  resolved="$(jq -ceS --arg slot "$slot" '
    {repository, path, configurationPath, backend, policyProfile, iac, context,
     stackNamespace, volume:$slot, target:.context,
     variables:{slot:$slot, namespace:.resourceNamespace}}' "$profile")"
  binding="$(printf '%s' "$resolved" | sha256sum | cut -d ' ' -f 1)"
  if [[ -n "$(value '.binding // ""')" && "$(value .binding)" != "$binding" ]]; then
    block BindingChanged 'O perfil efetivo diverge do vínculo registrado para a stack.'
    return
  fi
  reference="$(value '.pinnedReference // .intent.sourceRef.version')"
  if [[ "$reference" == latest ]]; then
    reference=HEAD
    if [[ "$repository" == "$(jq -er .gitOpsRepository "$profile")" ]]; then
      local annotation
      annotation="$(value .intent.revisionAnnotation)"
      if [[ -n "$annotation" && ! "$annotation" =~ ^[a-f0-9]{40}$ ]]; then
        block InvalidGitOpsRevision 'A revisão GitOps deve ser um commit Git completo.'
        return
      fi
      reference=${annotation:-HEAD}
    fi
  fi
  if [[ ! "$reference" =~ ^[a-zA-Z0-9][a-zA-Z0-9._/-]*$ ]] || ! git check-ref-format --allow-onelevel "$reference"; then
    block InvalidReference 'A referência Git tem formato inválido.'
    return
  fi
  git init -q /tmp/source
  git -C /tmp/source fetch -q --depth=1 -- "$repository" "$reference"
  revision="$(git -C /tmp/source rev-parse 'FETCH_HEAD^{commit}')"
  [[ "$revision" =~ ^[a-f0-9]{40}$ ]]
  if [[ "$reference" =~ ^[a-f0-9]{40}$ && "$revision" != "$reference" ]]; then
    block ReferenceChanged 'O commit encontrado diverge da revisão solicitada.'
    return
  fi
  jq -c --arg revision "$revision" --arg binding "$binding" \
    '. + {revision:$revision, binding:$binding}' <<<"$resolved" > /tmp/execution-context
  jq -cn --arg ref "$revision" --arg binding "$binding" \
    '{reference:$ref, binding:$binding}' | envelope > /tmp/resolved-source
  printf '%s\n' "$slot" > /tmp/execution-volume
  printf '%s\n' true > /tmp/allowed
}

intent_hash="$(value .intentHash)"
[[ "$(value .workflow)" == "$WORKFLOW" && "$intent_hash" =~ ^[a-f0-9]{64}$ ]]
if [[ "$1" == resolve ]]; then
  resolve
  exit 0
fi

workspace="/state/$WORKFLOW"
root="$workspace/source/$(bound .path)"
operation="$(value .intent.operation)"
saved="$workspace/plan.bin"
IAC_BIN="$(bound .iac)"
case "$IAC_BIN" in terraform|tofu) ;; *) exit 2 ;; esac
export IAC_BIN

checkout() {
  local destination="$1" revision="$2"
  git init -q "$destination"
  git -C "$destination" fetch -q --depth=1 -- "$(bound .repository)" "$revision"
  git -C "$destination" checkout -q --detach FETCH_HEAD
  [[ "$(git -C "$destination" rev-parse HEAD)" == "$revision" ]]
}

bind_state() {
  local marker=/state/.terraformstack-binding pending
  if [[ ! -e "$marker" ]]; then
    if [[ -e /state/terraform.tfstate || -e /state/terraform.tfstate.backup ]]; then
      return 1
    fi
    pending="$(mktemp /state/.binding.XXXXXX)"
    bound .binding > "$pending"
    ln "$pending" "$marker" 2>/dev/null || true
    rm "$pending"
  fi
  [[ ! -L "$marker" && -f "$marker" && "$(<"$marker")" == "$(bound .binding)" ]]
}

plan() {
  local result=0 path="$1"
  local args=(-input=false -no-color -detailed-exitcode -lock-timeout=30s
              "-var-file=$workspace/variables.tfvars.json" "-out=$path")
  if [[ "$operation" == destroy ]]; then
    args+=(-destroy)
  fi
  "$IAC_BIN" -chdir="$root" plan "${args[@]}" > "$workspace/terraform-plan.log" 2>&1 || result=$?
  if [[ "$result" != 0 && "$result" != 2 ]]; then
    echo "$IAC_BIN plan retornou $result." >&2
    return 1
  fi
  printf '%s\n' "$result"
}

signature() {
  "$IAC_BIN" -chdir="$root" show -json "$1" |
    jq -cS '{resource_changes, resource_drift, output_changes}'
}

claim() {
  local object version patch stack namespace
  stack="$(value .intent.stack)"
  namespace="$(value .intent.namespace)"
  local args=(-n "$namespace" terraformstack "$stack")
  for attempt in {1..30}; do
    object="$(kubectl get "${args[@]}" -o json)" || return 1
    if ! jq -e --argjson request "$INTENT" '
      $request.intent as $desired |
      .metadata.uid == $desired.uid and .metadata.generation == $desired.generation and
      (.metadata.annotations["platform.example.io/revision"] // "") == $desired.revisionAnnotation and
      (.metadata.deletionTimestamp != null) == $desired.deleting and
      (.spec.suspend // false | not) and
      .status.execution.workflow == $request.workflow and .status.execution.intent == $desired and
      .status.execution.intentHash == $request.intentHash and
      (.status.execution.cancelling // false | not) and (.status.claim == null) and
      (if $desired.deleting then .status.deletionIntent == $desired else true end)
    ' <<< "$object" >/dev/null; then
      return 1
    fi
    if [[ "$(jq -r '.status.execution.authorization // ""' <<< "$object")" == "$PLAN_HASH" ]]; then
      version="$(jq -er '.metadata.resourceVersion' <<< "$object")"
      patch="$(jq -cn --arg version "$version" --arg wf "$WORKFLOW" --arg hash "$intent_hash" --arg token "$PLAN_HASH" '[
        {op:"test",path:"/metadata/resourceVersion",value:$version},
        {op:"add",path:"/status/claim",value:{workflow:$wf,intentHash:$hash,authorization:$token}}
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
    : > /tmp/execution-result
    if ! bind_state; then
      block StateBindingMismatch 'O volume já tem state sem vínculo ou pertence a outro destino.'
      exit 0
    fi
    mkdir "$workspace"
    checkout "$workspace/source" "$(bound .revision)"
    real_root="$(realpath -e "$root")"
    config_root="$(realpath -e "$workspace/source/$(bound .configurationPath)")"
    [[ "$real_root" == "$workspace/source/"* && "$config_root" == "$workspace/source/"* ]]
    jq -e --arg path "$config_root" '.variables + {config_root: $path}' \
      <<< "$EXECUTION_CONTEXT" > "$workspace/variables.tfvars.json"
    printf '%s\n' true > /tmp/allowed
    ;;
  init)
    mapfile -t backend < <(jq -r '.backend | to_entries[] | "-backend-config=\(.key)=\(.value)"' <<< "$EXECUTION_CONTEXT")
    "$IAC_BIN" -chdir="$root" init -input=false -no-color "${backend[@]}"
    ;;
  validate)
    "$IAC_BIN" -chdir="$root" validate -no-color
    ;;
  plan)
    code="$(plan "$saved")"
    "$IAC_BIN" -chdir="$root" show -json "$saved" > "$workspace/plan.json"
    jq -c --arg operation "$operation" --argjson exitCode "$code" \
      '{operation:$operation, target, variables, policyProfile, exitCode:$exitCode}' \
      <<< "$EXECUTION_CONTEXT" > "$workspace/policy-context.json"
    jq -cn --rawfile document "$workspace/plan.json" --rawfile context "$workspace/policy-context.json" \
      '{document:$document, context:$context}' > "$workspace/policy-input.json"
    digest "$saved" > /tmp/plan-hash
    digest "$workspace/plan.json" > /tmp/document-hash
    digest "$workspace/policy-context.json" > /tmp/context-hash
    printf '%s\n' "$code" > /tmp/plan-exit
    ;;
  decision)
    if ! jq -cers --arg doc "$DOCUMENT_HASH" --arg context "$CONTEXT_HASH" \
      --arg op "$operation" --arg profile "$(bound .policyProfile)" --argjson code "$PLAN_EXIT" '
      select(length == 1) | .[0] |
      select(.documentHash == $doc and .contextHash == $context and
             (.reason | type == "string" and length > 0) and
             .profile == $profile and (.modules | type == "array") and
             (.decision as $decision | ["AutoApply","RequireApproval","Deny","NoChanges"] | index($decision)) and
             (if $op == "destroy" then .decision == "RequireApproval" or .decision == "Deny" else true end) and
             (if .decision == "NoChanges" then $code == 0 else true end))
      | .decision
    ' <<< "$POLICY_RESULT" >/tmp/decision; then
      printf '%s\n' Error > /tmp/decision
    fi
    jq -cn --arg decision "$(< /tmp/decision)" --arg hash "$PLAN_HASH" --arg package "$(policy_digest)" \
      --argjson drift "$(jq '[.resource_drift[]? | select(.change.actions != ["no-op"])] | length > 0' "$workspace/plan.json")" '
      {phase:(if $decision == "RequireApproval" then "Waiting" else "Running" end),
       driftDetected:$drift, details:{policyDecision:$decision, planHash:$hash, policyDigest:$package}}
    ' | envelope > /tmp/execution-state
    ;;
  authorize)
    printf '%s\n' none > /tmp/apply-request
    : > /tmp/authorization-request
    : > /tmp/execution-result
    case "$DECISION" in
      AutoApply) [[ "$operation" == apply ]] ;;
      RequireApproval)
        if [[ "$APPROVAL" != Approve || "$APPROVED_HASH" != "$PLAN_HASH" ]]; then
          result Blocked ApprovalRejected 'A aprovação foi rejeitada ou o hash informado diverge.'
          exit 0
        fi
        ;;
      NoChanges) result NoChanges PlanNoChanges 'O plano não contém alterações.'; exit 0 ;;
      Deny) result Blocked PolicyDenied 'A política do Simulator bloqueou a alteração.'; exit 0 ;;
      *) result Blocked PolicyError 'Decisão OPA inválida ou incompatível com o plano.'; exit 0 ;;
    esac
    printf '%s\n' "$PLAN_HASH" > /tmp/apply-request
    jq -cn --arg token "$PLAN_HASH" '{token:$token}' | envelope > /tmp/authorization-request
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
    if ! "$IAC_BIN" -chdir="$root" apply -input=false -no-color -lock-timeout=30s "$approved" > "$workspace/apply.log" 2>&1; then
      echo "$IAC_BIN apply falhou. Consulte o diagnóstico no volume da execução." >&2
      exit 1
    fi
    printf '%s\n' Applied > "$workspace/apply-status"
    ;;
  verify)
    if [[ "$(<"$workspace/apply-status")" == Cancelled ]]; then
      result Cancelled ClaimUnavailable 'A execução não obteve o claim do plano.'
      exit 0
    fi
    if [[ -f "$root/verify.sh" ]]; then
      (cd "$root" && bash ./verify.sh)
    fi
    if [[ "$(plan "$workspace/verify.bin")" != 0 ]]; then
      echo "Plan pós-apply ainda apresenta mudanças." >&2
      exit 1
    fi
    result Verified PlanVerified 'O plano aplicado foi verificado.'
    ;;
  *) exit 2 ;;
esac

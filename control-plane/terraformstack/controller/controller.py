"""Reconciliação TerraformStack por intenção e execução persistidas na API."""

import hashlib
import json
import logging
import os
import re
from datetime import datetime, timedelta, timezone
from pathlib import Path
from time import sleep

from kubernetes import client, config
from kubernetes.client.exceptions import ApiException
from urllib3.exceptions import HTTPError

STACK = ("platform.example.io", "v1alpha1", "platform-system", "terraformstacks")
WORKFLOW = ("argoproj.io", "v1alpha1", "argo-workflows", "workflows")
REVISION = "platform.example.io/revision"
FINALIZER = "platform.example.io/terraformstack-cleanup"
TERMINAL = {"Succeeded", "Failed", "Error"}
TIMEOUT = 10
SHA = re.compile(r"[0-9a-f]{40}")


def read(api, resource, name):
    try:
        return api.get_namespaced_custom_object(*resource, name, _request_timeout=TIMEOUT)
    except ApiException as error:
        if error.status != 404:
            raise
        return None


def patch(api, stack, values, status=True):
    metadata = stack["metadata"]
    resource_version = {"resourceVersion": metadata["resourceVersion"]}
    if not status:
        return api.patch_namespaced_custom_object(
            *STACK, metadata["name"], {"metadata": {**resource_version, **values}},
            _request_timeout=TIMEOUT,
        )
    return api.replace_namespaced_custom_object_status(
        *STACK, metadata["name"], {
            "apiVersion": stack["apiVersion"], "kind": stack["kind"],
            "metadata": {"name": metadata["name"], "namespace": metadata["namespace"], **resource_version},
            "status": {**stack.get("status", {}), **values},
        }, _request_timeout=TIMEOUT,
    )


def now():
    return datetime.now(timezone.utc)


def stamp():
    return now().isoformat(timespec="seconds").replace("+00:00", "Z")


def seconds(value):
    result = sum(int(amount) * {"s": 1, "m": 60, "h": 3600}[unit]
                 for amount, unit in re.findall(r"(\d+)([smh])", value))
    if not 0 < result <= 86400:
        raise ValueError("A duração deve estar entre 1 segundo e 24 horas.")
    return result


def intent(stack, catalog):
    metadata, spec = stack["metadata"], stack["spec"]
    deleting = bool(metadata.get("deletionTimestamp"))
    lifecycle = spec.get("lifecycle", {})
    identity = {"stack": metadata["name"], "uid": metadata["uid"],
                "generation": metadata["generation"], "deleting": deleting}
    if deleting and not lifecycle.get("destroyResourcesOnDeletion"):
        return {**identity, "operation": "retain"}

    source_ref = spec["sourceRef"]
    source = catalog["sources"].get(source_ref["name"], {})
    target_ref = spec["target"]
    target = catalog["targets"].get(f"{target_ref['entity']}/{target_ref['environment']}", {})
    slot = target.get("stacks", {}).get(metadata["name"])
    if source_ref["path"] not in source.get("paths", []) or not slot or slot["source"] != source_ref["name"]:
        raise ValueError("Fonte ou destino fora do catálogo.")
    if slot["path"] != source_ref["path"]:
        raise ValueError("Raiz Terraform não autorizada para esta stack.")

    # A escolha do perfil pertence ao catálogo do Platform Team, não ao CR.
    policy_profile = slot.get("policyProfile")
    if not isinstance(policy_profile, str) or not re.fullmatch(r"[a-z][a-z0-9-]*", policy_profile):
        raise ValueError("Policy Profile ausente ou inválido no catálogo.")

    pinned = metadata.get("annotations", {}).get(REVISION, "")
    revision = source_ref.get("version", "latest")
    configuration = spec.get("configuration", {}).get("version", "latest")
    revision = pinned if revision == "latest" else revision
    configuration = pinned if configuration == "latest" else configuration
    if not SHA.fullmatch(revision) or not SHA.fullmatch(configuration):
        raise ValueError("Fonte e configuração exigem SHAs Git completos.")
    return {
        **identity, "revision": revision, "configuration": configuration,
        "revisionAnnotation": pinned, "source": source_ref["name"],
        "repository": source["repository"], "path": source_ref["path"],
        "configurationPath": slot.get("configurationPath"), "target": target_ref,
        "backend": slot["backend"], "volume": slot["volume"],
        "policyProfile": policy_profile,
        "variables": slot["variables"], "serviceAccount": slot["serviceAccount"],
        "operation": "destroy" if deleting or lifecycle.get("destroy") else "apply",
    }


def report(api, stack, condition, reason, message, **values):
    status = stack.get("status", {})
    generation = stack["metadata"]["generation"]
    # Cada atualização apresenta a condição atual; drift também permanece no registro da execução.
    previous = next(iter(status.get("conditions", [])), {})
    item = {
        "type": condition, "status": "True", "reason": reason, "message": message,
        "observedGeneration": generation,
        "lastTransitionTime": (previous["lastTransitionTime"]
                               if previous.get("type") == condition else stamp()),
    }
    values.update(observedGeneration=generation, conditions=[item])
    if any(status.get(key) != value for key, value in values.items()):
        patch(api, stack, values)


def workflow_body(run):
    data = run["intent"]
    return {
        "apiVersion": "argoproj.io/v1alpha1", "kind": "Workflow",
        "metadata": {"name": run["workflow"], "namespace": WORKFLOW[2],
                     "labels": {"platform.example.io/workflow": "terraformstack",
                                "platform.example.io/stack-uid": data["uid"]}},
        "spec": {
            "serviceAccountName": data["serviceAccount"],
            "workflowTemplateRef": {"name": "terraformstack"},
            "arguments": {"parameters": [
                {"name": "intent", "value": json.dumps(data, separators=(",", ":"))},
                {"name": "volume", "value": data["volume"]},
            ]},
        },
    }


def dependencies_ready(stacks, stack, data, catalog):
    spec = stack["spec"]
    if data["operation"] == "apply":
        for dependency in spec.get("dependsOn", []):
            name = dependency["name"]
            other = stacks.get(name) if name != data["stack"] else None
            if not other or other["metadata"].get("deletionTimestamp") or other["spec"]["target"] != spec["target"]:
                return False
            status = other.get("status", {})
            try:
                desired = intent(other, catalog)
            except ValueError:
                return False
            if (status.get("observedGeneration") != other["metadata"]["generation"]
                    or status.get("execution", {}).get("intent") != desired
                    or not any(c["type"] == "Ready" for c in status.get("conditions", []))):
                return False
    if data["operation"] == "destroy":
        for other in stacks.values():
            if other["spec"]["target"] != spec["target"]:
                continue
            if any(d["name"] == data["stack"] for d in other["spec"].get("dependsOn", [])):
                status = other.get("status", {})
                if not any(c["type"] == "Destroyed" for c in status.get("conditions", [])
                           if c.get("observedGeneration") == other["metadata"]["generation"]):
                    return False
    return True


def result_of(workflow):
    status = workflow.get("status", {})
    outputs = {p["name"]: p.get("value", "") for p in status.get("outputs", {}).get("parameters", [])}
    raw = outputs.get("decision-record")
    result = json.loads(raw) if raw else {}
    for node in status.get("nodes", {}).values():
        if node.get("templateName") == "opa" and node.get("phase") in {"Error", "Failed"}:
            result.update(decision="Error", reason="OPA falhou ao avaliar o plano")
        if node.get("type") == "Suspend" and node.get("templateName") == "approval":
            result["approval"] = node.get("phase", "")
            actor = workflow["metadata"].get("labels", {}).get("workflows.argoproj.io/actor")
            if actor:
                result["approvalActor"] = actor
            if node.get("message"):
                result["approvalMessage"] = node["message"][:512]
    if outputs.get("apply-request", "none") != "none":
        result["applyRequest"] = outputs["apply-request"]
    if outputs.get("approval-result") == "Rejected":
        result["outcome"] = "Rejected"
    if outputs.get("apply-result"):
        result["outcome"] = outputs["apply-result"]
    return result


def retry_due(settings, run, age):
    policy = settings.get("retryPolicy", {})
    strategy = policy.get("strategy", "ExponentialBackoff")
    attempt = run.get("attempt", 0)
    if strategy == "None" or attempt >= policy.get("maxRetries", 0):
        return False
    interval = seconds(policy.get("interval", "10s"))
    if strategy == "ExponentialBackoff":
        interval = min(interval * 2 ** min(attempt, 17), seconds(policy.get("maxInterval", "5m")))
    return age >= interval


def reconcile(api, stack, stacks, catalog):
    metadata, spec = stack["metadata"], stack["spec"]
    status = stack.get("status", {})
    run = status.get("execution") or {}
    finalizers = metadata.get("finalizers", [])
    deleting = bool(metadata.get("deletionTimestamp"))
    if deleting and FINALIZER not in finalizers:
        return
    try:
        data = intent(stack, catalog)
        settings = spec.get("reconciliation", {}) if data["operation"] != "retain" else {}
        drift_interval = seconds(settings["driftCheckInterval"]) if settings.get("driftCheckInterval") else None
        eligible = dependencies_ready(stacks, stack, data, catalog)
        if deleting:
            snapshot = status.get("deletionIntent")
            if snapshot is None:
                patch(api, stack, {"deletionIntent": data})
                return
            if snapshot != data:
                raise ValueError("A intenção de exclusão diverge do snapshot registrado.")
    except ValueError as error:
        if run and not run.get("applying") and not run.get("finishedAt"):
            workflow = read(api, WORKFLOW, run["workflow"])
            if workflow and workflow.get("status", {}).get("phase") not in TERMINAL:
                if workflow["spec"].get("shutdown") != "Terminate":
                    api.patch_namespaced_custom_object(
                        *WORKFLOW, run["workflow"], {"spec": {"shutdown": "Terminate"}}, _request_timeout=TIMEOUT,
                    )
        report(api, stack, "ReconcileBlocked", "InvalidIntent", str(error))
        return

    if not deleting and FINALIZER not in finalizers:
        patch(api, stack, {"finalizers": [*finalizers, FINALIZER]}, status=False)
        return

    previous = run.get("intent")
    workflow = read(api, WORKFLOW, run["workflow"]) if run else None
    phase = workflow.get("status", {}).get("phase", "") if workflow else ""
    changed = data != previous
    suspended = spec.get("suspend", False)
    drift = False
    attempt = 0

    if run and not run.get("finishedAt"):
        if not workflow:
            if run.get("created") or run.get("applying"):
                report(api, stack, "Failed", "WorkflowMissing", "Execução removida; investigar Pods e state.")
                return
            if not changed and not suspended and eligible:
                api.create_namespaced_custom_object(*WORKFLOW, workflow_body(run), _request_timeout=TIMEOUT)
                return
        elif phase not in TERMINAL:
            if (changed or suspended or not eligible) and not run.get("applying"):
                if not run.get("cancelling"):
                    report(api, stack, "ReconcilePending", "Cancelling", run["workflow"],
                           execution={**run, "created": True, "cancelling": True})
                    return
                if workflow["spec"].get("shutdown") != "Terminate":
                    api.patch_namespaced_custom_object(*WORKFLOW, run["workflow"],
                        {"spec": {"shutdown": "Terminate"}}, _request_timeout=TIMEOUT)
                return
            result = result_of(workflow)
            if result.get("drift") and not run.get("drift"):
                report(api, stack, "DriftDetected", "ExternalChange", run["workflow"],
                       execution={**run, "created": True, "drift": True}, policy=result,
                       lastDriftTime=stamp())
                return
            execution = {**run, "created": True}
            request = result.get("applyRequest")
            if (not changed and not suspended and eligible and not run.get("cancelling")
                    and request and request == result.get("hash")
                    and result.get("decision") in {"AutoApply", "RequireApproval"}):
                execution["authorization"] = request
            condition, reason = "Reconciling", "Planning"
            if result.get("approval") == "Running":
                condition, reason = "ApprovalRequired", "AwaitingApproval"
            elif run.get("applying"):
                reason = "Applying"
            report(api, stack, condition, reason, run["workflow"],
                   execution=execution, policy=result)
            return

        result = result_of(workflow) if workflow else {}
        if run.get("cancelling") or not workflow or result.get("outcome") == "Cancelled":
            condition, reason = "ReconcileBlocked", "Cancelled"
        elif result.get("outcome") == "Rejected":
            condition, reason = "ReconcileBlocked", "ApprovalRejected"
        elif result.get("decision") in {"Deny", "Error"}:
            condition = "ReconcileBlocked"
            reason = "PolicyDenied" if result["decision"] == "Deny" else "PolicyError"
        elif phase == "Succeeded" and result.get("decision") == "NoChanges" and previous["operation"] == "apply":
            condition, reason = "Ready", "NoChanges"
        elif phase == "Succeeded" and result.get("outcome") == "Verified":
            condition, reason = ("Destroyed" if previous["operation"] == "destroy" else "Ready"), "Verified"
        else:
            condition, reason = "Failed", "WorkflowFailed"
        drift_status = {"lastDriftTime": stamp()} if result.get("drift") and not run.get("drift") else {}
        completed = {**run, "finishedAt": stamp(), "result": condition, "reason": reason,
                     "drift": bool(result.get("drift"))}
        if changed:
            # O resultado anterior não torna a nova intenção Ready.
            condition, reason = "ReconcilePending", "PreviousExecutionFinished"
        report(api, stack, condition, reason, run["workflow"],
               execution=completed, policy=result, lastCheckTime=stamp(), **drift_status)
        return

    if data["operation"] == "retain":
        patch(api, stack, {"finalizers": [value for value in finalizers if value != FINALIZER]}, status=False)
        return
    if suspended:
        report(api, stack, "ReconcilePending", "Suspended", "Novas execuções suspensas.")
        return
    if not eligible:
        report(api, stack, "ReconcilePending", "DependencyBlocked", "Dependência pendente ou ainda ativa.")
        return

    if not changed:
        result = run["result"]
        if deleting and result == "Destroyed":
            patch(api, stack, {"finalizers": [value for value in finalizers if value != FINALIZER]}, status=False)
            return
        age = (now() - datetime.fromisoformat(run["finishedAt"])).total_seconds()
        if result == "Ready" and drift_interval and age >= drift_interval:
            drift = True
        elif result == "Failed" and retry_due(settings, run, age):
            attempt = run.get("attempt", 0) + 1
        else:
            return

    sequence = run.get("sequence", 0) + 1
    key = hashlib.sha256(json.dumps(data, sort_keys=True).encode()).hexdigest()[:12]
    execution = {"intent": data, "workflow": f"tf-{data['uid']}-{key}-{sequence}",
                 "sequence": sequence, "attempt": attempt,
                 "trigger": "DriftCheck" if drift else "Retry" if attempt else "GitOps"}
    report(api, stack, "Reconciling", "Scheduling", execution["workflow"], execution=execution,
           resolvedSource={"name": data["source"], "version": data["revision"],
                           "configurationVersion": data["configuration"], "digest": data["revision"]},
           policy={})


def main():
    logging.basicConfig(level=logging.INFO)
    config.load_incluster_config()
    api = client.CustomObjectsApi()
    catalog_path = Path(os.getenv("TERRAFORMSTACK_CATALOG", "/etc/terraformstack/catalog.json"))
    catalog = json.loads(catalog_path.read_text(encoding="utf-8"))
    while True:
        try:
            snapshot = api.list_namespaced_custom_object(*STACK, _request_timeout=TIMEOUT)
        except (ApiException, HTTPError) as error:
            logging.warning("Listagem de TerraformStacks: %s", getattr(error, "status", str(error)))
            sleep(2)
            continue
        stacks = {item["metadata"]["name"]: item for item in snapshot["items"]}
        for name, stack in stacks.items():
            try:
                reconcile(api, stack, stacks, catalog)
            except (ApiException, HTTPError, ValueError) as error:
                logging.warning("Reconciliação de %s: %s", name, getattr(error, "status", str(error)))
        sleep(2)


if __name__ == "__main__":
    main()

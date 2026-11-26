"""Reconciliação TerraformStack por intenção e execução persistidas na API."""

import hashlib
import json
import logging
import re
from datetime import datetime, timedelta, timezone
from time import sleep

from kubernetes import client, config
from kubernetes.client.exceptions import ApiException
from urllib3.exceptions import HTTPError

STACK = ("platform.example.io", "v1alpha1", "platform-system", "terraformstacks")
WORKFLOW = ("argoproj.io", "v1alpha1", "argo-workflows", "workflows")
REPOSITORY = "https://github.com/DavidFerreira21/platform-engineering-poc.git"
REVISION = "platform.example.io/revision"
FINALIZER = "platform.example.io/terraformstack-cleanup"
SLOTS = {"lab06b-network", "lab06b-dependent"}
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


def intent(stack):
    metadata, spec = stack["metadata"], stack["spec"]
    deleting = bool(metadata.get("deletionTimestamp"))
    lifecycle = spec.get("lifecycle", {})
    identity = {"stack": metadata["name"], "uid": metadata["uid"],
                "generation": metadata["generation"], "deleting": deleting}
    if deleting and not lifecycle.get("destroyResourcesOnDeletion"):
        return {**identity, "operation": "retain"}
    source = spec["sourceRef"]
    if (metadata["name"] not in SLOTS or spec["target"] != {"entity": "entity-a", "environment": "uat"}
            or source["name"] != "foundation-stacks" or source["path"] != "foundation/lab06b/network"):
        raise ValueError("Fonte ou destino fora do catálogo local.")
    pinned = metadata.get("annotations", {}).get(REVISION, "")
    revision = source.get("version", "latest")
    configuration = spec.get("configuration", {}).get("version", "latest")
    revision = pinned if revision == "latest" else revision
    configuration = pinned if configuration == "latest" else configuration
    if not SHA.fullmatch(revision) or not SHA.fullmatch(configuration):
        raise ValueError("Fonte e configuração exigem SHAs Git completos.")
    return {
        **identity, "revision": revision, "configuration": configuration,
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
                     "labels": {"lab.platform.example/exercise": "lab06b",
                                "lab.platform.example/stack-uid": data["uid"]}},
        "spec": {
            "workflowTemplateRef": {"name": "terraformstack-local"},
            "arguments": {"parameters": [
                {"name": "intent", "value": json.dumps(data, separators=(",", ":"))},
                {"name": "slot", "value": data["stack"]},
            ]},
        },
    }


def dependencies_ready(api, stack, data):
    spec = stack["spec"]
    for dependency in spec.get("dependsOn", []) if data["operation"] == "apply" else []:
        name = dependency["name"]
        other = read(api, STACK, name) if name in SLOTS - {data["stack"]} else None
        if not other or other["metadata"].get("deletionTimestamp") or other["spec"]["target"] != spec["target"]:
            return False
        status = other.get("status", {})
        if (status.get("observedGeneration") != other["metadata"]["generation"]
                or status.get("execution", {}).get("intent") != intent(other)
                or not any(c["type"] == "Ready" for c in status.get("conditions", []))):
            return False
    if data["operation"] == "destroy":
        for name in SLOTS - {data["stack"]}:
            other = read(api, STACK, name)
            if other and any(d["name"] == data["stack"] for d in other["spec"].get("dependsOn", [])):
                if not any(c["type"] == "Destroyed" for c in other.get("status", {}).get("conditions", [])
                           if c.get("observedGeneration") == other["metadata"]["generation"]):
                    return False
    return True


def result_of(workflow):
    status = workflow.get("status", {})
    outputs = {p["name"]: p.get("value", "") for p in status.get("outputs", {}).get("parameters", [])}
    raw = outputs.get("decision-record")
    result = json.loads(raw) if raw else {}
    for node in status.get("nodes", {}).values():
        if node.get("type") == "Suspend" and node.get("templateName") == "approval":
            result["approval"] = node.get("phase", "")
            actor = workflow["metadata"].get("labels", {}).get("workflows.argoproj.io/actor")
            if actor:
                result["approvalActor"] = actor
            if node.get("message"):
                result["approvalMessage"] = node["message"][:512]
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


def reconcile(api, stack):
    metadata, spec = stack["metadata"], stack["spec"]
    status = stack.get("status", {})
    run = status.get("execution") or {}
    finalizers = metadata.get("finalizers", [])
    deleting = bool(metadata.get("deletionTimestamp"))
    if deleting and FINALIZER not in finalizers:
        return
    if not deleting and FINALIZER not in finalizers:
        patch(api, stack, {"finalizers": [*finalizers, FINALIZER]}, status=False)
        return

    try:
        data = intent(stack)
        settings = spec.get("reconciliation", {}) if data["operation"] != "retain" else {}
        drift_interval = seconds(settings["driftCheckInterval"]) if settings.get("driftCheckInterval") else None
        eligible = dependencies_ready(api, stack, data)
        if deleting:
            snapshot = status.get("deletionIntent")
            if snapshot is None:
                patch(api, stack, {"deletionIntent": data})
                return
            if snapshot != data:
                raise ValueError("A intenção de exclusão diverge do snapshot registrado.")
    except ValueError as error:
        report(api, stack, "ReconcileBlocked", "InvalidIntent", str(error))
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
            condition, reason = "Reconciling", "Planning"
            if result.get("approval") == "Running":
                condition, reason = "ApprovalRequired", "AwaitingApproval"
            elif run.get("applying"):
                reason = "Applying"
            report(api, stack, condition, reason, run["workflow"],
                   execution={**run, "created": True}, policy=result)
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
    execution = {"intent": data, "workflow": f"lab06b-{data['uid']}-{key}-{sequence}",
                 "sequence": sequence, "attempt": attempt,
                 "trigger": "DriftCheck" if drift else "Retry" if attempt else "GitOps"}
    report(api, stack, "Reconciling", "Scheduling", execution["workflow"], execution=execution,
           resolvedSource={"name": "foundation-stacks", "version": data["revision"],
                           "configurationVersion": data["configuration"], "digest": data["revision"]},
           policy={})


def main():
    logging.basicConfig(level=logging.INFO)
    config.load_incluster_config()
    api = client.CustomObjectsApi()
    while True:
        for name in sorted(SLOTS):
            try:
                stack = read(api, STACK, name)
                if stack:
                    reconcile(api, stack)
            except (ApiException, HTTPError, ValueError) as error:
                logging.warning("Reconciliação de %s: %s", name, getattr(error, "status", str(error)))
        sleep(2)


if __name__ == "__main__":
    main()

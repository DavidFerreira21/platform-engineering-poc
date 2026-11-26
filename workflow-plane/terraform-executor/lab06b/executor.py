"""Planeja, avalia e aplica o plano salvo com identidade Kubernetes local."""

import hashlib
import ipaddress
import json
import os
import shutil
import subprocess
import sys
from pathlib import Path
from time import sleep

from kubernetes import client, config
from kubernetes.client.exceptions import ApiException

from real import REPOSITORY, STACK, dependencies_ready, intent, patch, read

POLICY = Path("/app/policy.rego")
OUTPUT = Path("/tmp")
NAMESPACE = "terraformstack-lab06b"


def command(arguments, *, cwd, allowed=(0,), capture=False):
    result = subprocess.run(arguments, cwd=cwd, text=True, capture_output=True, check=False)
    if result.returncode not in allowed:
        # Planos e estados ficam no volume. Expor no log apenas comando e código.
        raise RuntimeError(f"{arguments[0]} {arguments[1]} retornou {result.returncode}")
    return result.stdout if capture else result.returncode


def digest(path):
    with path.open("rb") as handle:
        return hashlib.file_digest(handle, "sha256").hexdigest()


def checkout(directory, revision):
    directory.mkdir()
    command(["git", "init", "-q"], cwd=directory)
    command(["git", "fetch", "-q", "--depth=1", REPOSITORY, revision], cwd=directory)
    command(["git", "checkout", "-q", "--detach", "FETCH_HEAD"], cwd=directory)
    actual = command(["git", "rev-parse", "HEAD"], cwd=directory, capture=True).strip()
    if actual != revision:
        raise RuntimeError("O commit obtido difere do SHA solicitado.")


def changes(plan):
    return [item for item in plan.get("resource_changes", [])
            if item["change"]["actions"] != ["no-op"]]


def planned_cidrs(document):
    for resource in changes(document):
        if resource["address"] == "kubernetes_config_map_v1.network":
            result = {}
            for phase in ("before", "after"):
                try:
                    snapshot = resource["change"][phase]
                    network = json.loads(snapshot["data"]["network.json"])
                    result[phase] = str(ipaddress.ip_network(network["vpc"]["cidr"]))
                except (KeyError, TypeError, ValueError):
                    continue
            return result
    return {}


def drift_detected(plan):
    for resource in plan.get("resource_drift", []):
        if resource["address"] == "kubernetes_config_map_v1.network":
            difference = resource["change"]
            before, after = difference["before"], difference["after"]
            if before is None or after is None:
                return True
            # O provider também atualiza metadados calculados pela API.
            if before.get("data") != after.get("data"):
                return True
    return False


def plan_json(root, path):
    return json.loads(command(["terraform", "show", "-json", str(path)], cwd=root, capture=True))


def signature(plan):
    return {key: plan.get(key) for key in ("resource_changes", "resource_drift", "output_changes")}


def terraform_plan(root, config_root, data, path):
    args = ["terraform", "plan", "-no-color", "-input=false", "-lock-timeout=30s",
            "-detailed-exitcode", f"-out={path}", f"-var=slot={data['stack']}",
            f"-var=config_root={config_root}"]
    if data["operation"] == "destroy":
        args.append("-destroy")
    return command(args, cwd=root, allowed=(0, 2))


def evaluate(document, data, directory):
    input_path = directory / "policy-input.json"
    input_path.write_text(json.dumps({"changes": changes(document), "operation": data["operation"],
                                      "slot": data["stack"], "drift": drift_detected(document)}))
    try:
        raw = command(["opa", "eval", "--format=raw", "--data", str(POLICY), "--input",
                       str(input_path), "data.terraformstack.local.decision"], cwd=directory, capture=True)
        decision = json.loads(raw)
        valid_reason = isinstance(decision["reason"], str) and decision["reason"].strip()
        if decision["decision"] not in {"Deny", "AutoApply", "RequireApproval"} or not valid_reason:
            raise ValueError("Decisão inválida.")
        if data["operation"] == "destroy" and decision["decision"] == "AutoApply":
            raise ValueError("Destroy exige gate.")
    except (RuntimeError, ValueError, KeyError, TypeError):
        return {"decision": "Error", "reason": "OPA falhou ou retornou decisão inválida"}
    return decision


def claim(data, workflow, drift):
    config.load_incluster_config()
    api = client.CustomObjectsApi()
    for _ in range(15):
        stack = read(api, STACK, data["stack"])
        if not stack or intent(stack) != data or stack["spec"].get("suspend"):
            return False
        run = stack.get("status", {}).get("execution", {})
        if (run.get("workflow") != workflow or run.get("intent") != data or run.get("cancelling")
                or run.get("applying") or not dependencies_ready(api, stack, data)
                or (data["deleting"] and stack["status"].get("deletionIntent") != data)):
            return False
        if drift and not run.get("drift"):
            # O controller publica DriftDetected antes de liberar a correção.
            sleep(1)
            continue
        try:
            patch(api, stack, {"execution": {**run, "applying": True}})
            return True
        except ApiException as error:
            if error.status != 409:
                raise
    return False


def verify(root, data):
    api = client.CoreV1Api()
    try:
        resource = api.read_namespaced_config_map(data["stack"], NAMESPACE)
    except ApiException as error:
        if error.status == 404 and data["operation"] == "destroy":
            return
        raise
    if data["operation"] == "destroy":
        raise RuntimeError("ConfigMap permanece após destroy.")
    actual = json.loads(resource.data["network.json"])
    expected = json.loads(command(["terraform", "output", "-json", "network"], cwd=root, capture=True))
    if actual != expected:
        raise RuntimeError("ConfigMap diverge do output Network.")


def main():
    os.umask(0o077)
    stage = sys.argv[1]
    data = json.loads(os.environ["INTENT"])
    workflow = os.environ["WORKFLOW"]
    directory = Path("/state") / workflow
    root = directory / "source" / "foundation" / "lab06b" / "network"
    config_root = directory / "configuration" / "foundation" / "config"
    saved = directory / "plan.bin"

    if stage == "plan":
        directory.mkdir()
        checkout(directory / "source", data["revision"])
        checkout(directory / "configuration", data["configuration"])
        command(["terraform", "init", "-no-color", "-input=false",
                 "-backend-config=path=/state/terraform.tfstate"], cwd=root)
        command(["terraform", "validate", "-no-color"], cwd=root)
        code = terraform_plan(root, config_root, data, saved)
        document = plan_json(root, saved)
        decision = ({"decision": "NoChanges", "reason": "Plan sem mudanças"}
                    if code == 0 and data["operation"] == "apply"
                    else evaluate(document, data, directory))
        record = {**decision, "hash": digest(saved), "policyDigest": digest(POLICY),
                  "drift": drift_detected(document),
                  "actions": [item["change"]["actions"] for item in changes(document)],
                  "cidrs": planned_cidrs(document)}
        (OUTPUT / "decision-record").write_text(json.dumps(record))
        (OUTPUT / "decision").write_text(record["decision"])
        (OUTPUT / "plan-hash").write_text(record["hash"])
        print(f"{record['decision']} · {record['reason']} · SHA-256 {record['hash']}")
        return

    hash_expected = os.environ["PLAN_HASH"]
    approved_plan = OUTPUT / "approved-plan.bin"
    shutil.copyfile(saved, approved_plan)
    if digest(approved_plan) != hash_expected:
        raise RuntimeError("Hash do plano salvo diverge.")
    if os.environ["DECISION"] == "RequireApproval":
        supplied_hash = os.environ["APPROVED_HASH"]
        response = os.environ["APPROVAL"]
        if response != "Approve" or supplied_hash != hash_expected:
            (OUTPUT / "apply-result").write_text("Rejected")
            return
    document = plan_json(root, approved_plan)
    if evaluate(document, data, directory)["decision"] != os.environ["DECISION"]:
        raise RuntimeError("A política mudou desde o plan.")
    check = directory / "revalidate.bin"
    terraform_plan(root, config_root, data, check)
    if signature(document) != signature(plan_json(root, check)):
        raise RuntimeError("O destino ou o state mudou desde o plan.")
    if not claim(data, workflow, drift_detected(document)):
        (OUTPUT / "apply-result").write_text("Cancelled")
        return
    command(["terraform", "apply", "-no-color", "-input=false", "-lock-timeout=30s",
             str(approved_plan)], cwd=root)
    verify(root, data)
    if terraform_plan(root, config_root, data, directory / "verify.bin") != 0:
        raise RuntimeError("Plan pós-apply ainda apresenta mudanças.")
    (OUTPUT / "apply-result").write_text("Verified")


if __name__ == "__main__":
    try:
        main()
    except (RuntimeError, ValueError, OSError, ApiException) as error:
        print(f"Executor: {type(error).__name__}: {getattr(error, 'status', str(error))}", file=sys.stderr)
        sys.exit(1)

"""API simulada com cópias, resourceVersion e conflitos otimistas."""
import importlib.util
import json
from copy import deepcopy
from datetime import datetime, timezone
from pathlib import Path
from unittest.mock import Mock

import pytest
from kubernetes.client.exceptions import ApiException

ROOT = Path(__file__).resolve().parents[2]
BASE = ROOT / "control-plane/terraformstack"
spec = importlib.util.spec_from_file_location("terraformstack_controller", BASE / "controller/controller.py")
controller = importlib.util.module_from_spec(spec)
spec.loader.exec_module(controller)
SHA = "a" * 40
NOW = datetime(2026, 1, 1, tzinfo=timezone.utc)


class Cluster:
    def __init__(self):
        self.stacks = {}
        self.workflows = {}
        self.api = Mock()
        self.api.get_namespaced_custom_object.side_effect = self.get
        self.api.replace_namespaced_custom_object_status.side_effect = self.status
        self.api.patch_namespaced_custom_object.side_effect = self.patch
        self.api.create_namespaced_custom_object.side_effect = self.create
        self.catalog = json.loads((BASE / "overlays/kind/catalog.json").read_text())

    def add(self, name="simulator-network", **changes):
        stack = {
            "apiVersion": "platform.example.io/v1alpha1", "kind": "TerraformStack",
            "metadata": {"name": name, "namespace": "platform-system", "uid": name,
                         "generation": 1, "resourceVersion": "1",
                         "annotations": {controller.REVISION: SHA}},
            "spec": {"sourceRef": {"name": "foundation-stacks", "path": "foundation/stacks/simulator"},
                     "target": {"entity": "entity-a", "environment": "simulator"}, **changes},
        }
        self.stacks[name] = stack
        return stack

    def get(self, group, version, namespace, plural, name, **kwargs):
        objects = self.workflows if plural == "workflows" else self.stacks
        if name not in objects:
            raise ApiException(status=404)
        return deepcopy(objects[name])

    def cas(self, name, metadata):
        current = self.stacks[name]["metadata"]
        if metadata["resourceVersion"] != current["resourceVersion"]:
            raise ApiException(status=409)
        current["resourceVersion"] = str(int(current["resourceVersion"]) + 1)

    def status(self, group, version, namespace, plural, name, body, **kwargs):
        self.cas(name, body["metadata"])
        self.stacks[name]["status"] = deepcopy(body["status"])

    def patch(self, group, version, namespace, plural, name, body, **kwargs):
        if plural == "workflows":
            self.workflows[name]["spec"].update(deepcopy(body["spec"]))
        else:
            self.cas(name, body["metadata"])
            self.stacks[name]["metadata"].update(
                {k: deepcopy(v) for k, v in body["metadata"].items() if k != "resourceVersion"}
            )

    def create(self, group, version, namespace, plural, body, **kwargs):
        name = body["metadata"]["name"]
        if name in self.workflows:
            raise ApiException(status=409)
        self.workflows[name] = deepcopy(body)

    def tick(self, name="simulator-network", count=1):
        for _ in range(count):
            snapshot = deepcopy(self.stacks)
            controller.reconcile(self.api, snapshot[name], snapshot, self.catalog)

    def start(self, name="simulator-network"):
        self.tick(name, 4)
        return self.workflows[self.run(name)["workflow"]]

    def run(self, name="simulator-network"):
        return self.stacks[name]["status"]["execution"]

    def condition(self, name="simulator-network"):
        return self.stacks[name]["status"]["conditions"][0]

    def change(self, name="simulator-network", **changes):
        stack = self.stacks[name]
        stack["spec"].update(changes)
        stack["metadata"]["generation"] += 1
        stack["metadata"]["resourceVersion"] = str(int(stack["metadata"]["resourceVersion"]) + 1)

    def outcome(self, phase="Succeeded", decision="NoChanges", outcome=None, request=None, name="simulator-network", **extra):
        wf = self.workflows[self.run(name)["workflow"]]
        outputs = [{"name": "decision-record", "value": json.dumps({"decision": decision, **extra})}]
        if outcome:
            outputs.append({"name": "apply-result", "value": outcome})
        if request:
            outputs.append({"name": "apply-request", "value": request})
        wf["status"] = {"phase": phase, "outputs": {"parameters": outputs}}
        return wf


@pytest.fixture
def cluster(monkeypatch):
    monkeypatch.setattr(controller, "now", lambda: NOW)
    return Cluster()

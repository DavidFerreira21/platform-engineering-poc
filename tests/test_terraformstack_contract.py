#!/usr/bin/env python3
"""Verifica o contrato v1alpha1 sem acesso a Kubernetes/AWS."""

import ast
from pathlib import Path
import unittest

import yaml
from jsonschema import Draft7Validator, ValidationError

ROOT = Path(__file__).resolve().parents[1]
CRD = ROOT / "control-plane/terraformstack/crd/terraformstack.yaml"
DEMOS = [
    ROOT / "control-plane/terraformstack/demo/terraformstack.yaml",
    ROOT / "control-plane/terraformstack/demo/lab06/terraformstack.yaml",
]
CONTROLLER = ROOT / "control-plane/terraformstack/controller/controller.py"
WORKFLOW = ROOT / "workflow-plane/argo-workflows/terraformstack/workflowtemplate.yaml"


class TerraformStackContractTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        crd = yaml.safe_load(CRD.read_text(encoding="utf-8"))
        cls.schema = crd["spec"]["versions"][0]["schema"]["openAPIV3Schema"]
        Draft7Validator.check_schema(cls.schema)
        cls.validator = Draft7Validator(cls.schema)

    def example(self, **spec_changes):
        spec = {
            "target": {"entity": "entity-a", "environment": "uat"},
            "sourceRef": {"name": "foundation-stacks", "path": "stacks/network"},
        }
        spec.update(spec_changes)
        return {"spec": spec}

    def test_both_demo_manifests_validate(self):
        for path in DEMOS:
            with self.subTest(path=path):
                self.validator.validate(yaml.safe_load(path.read_text(encoding="utf-8")))

    def test_latest_defaults_are_in_schema(self):
        properties = self.schema["properties"]["spec"]["properties"]
        self.assertEqual(properties["sourceRef"]["properties"]["version"]["default"], "latest")
        self.assertEqual(properties["configuration"]["properties"]["revision"]["default"], "latest")
        self.assertFalse(properties["lifecycle"]["properties"]["destroy"]["default"])
        self.assertFalse(properties["lifecycle"]["properties"]["destroyResourcesOnDeletion"]["default"])

    def test_version_can_be_omitted_or_specified(self):
        self.validator.validate(self.example())
        self.validator.validate(self.example(sourceRef={
            "name": "foundation-stacks", "path": "stacks/network", "version": "1.2.3",
        }))
        self.validator.validate(self.example(configuration={"revision": "latest"}))

    def test_legacy_module_still_works(self):
        self.validator.validate({"spec": {
            "module": {"name": "lab-network", "version": "1.0.0"},
            "target": {"entity": "entity-a", "environment": "uat"},
        }})
        with self.assertRaises(ValidationError):
            self.validator.validate({"spec": {
                **self.example()["spec"],
                "module": {"name": "lab-network", "version": "1.0.0"},
            }})

    def test_rejects_unknown_source_selector_and_invalid_controls(self):
        for spec in (
            {"target": {"entity": "entity-a", "environment": "uat"}},
            self.example(sourceRef={"name": "foundation-stacks"}),
            self.example(reconciliation={"driftCheckInterval": "garbage"}),
            self.example(lifecycle={"destroy": "true"}),
            self.example(timeout={"apply": "tomorrow"}),
            self.example(reconciliation={"retryPolicy": {"strategy": "Always"}}),
        ):
            with self.subTest(spec=spec), self.assertRaises(ValidationError):
                self.validator.validate(spec if "spec" in spec else {"spec": spec})

    def test_extended_fields_and_workflow_parameters(self):
        self.validator.validate(self.example(
            dependsOn=[{"name": "network-uat"}], suspend=True,
            reconciliation={"driftCheckInterval": "6h", "retryPolicy": {
                "strategy": "ExponentialBackoff", "interval": "5m", "maxInterval": "1h",
            }},
            lifecycle={"destroy": False, "destroyResourcesOnDeletion": False},
            timeout={"plan": "15m", "apply": "1h"},
            healthChecks=[{"name": "vpc", "type": "TerraformOutput", "target": "vpc_id"}],
        ))
        template = yaml.safe_load(WORKFLOW.read_text(encoding="utf-8"))
        names = {p["name"] for p in template["spec"]["arguments"]["parameters"]}
        self.assertTrue({"source", "path", "module", "version", "config-revision",
                         "entity", "environment"}.issubset(names))

    def test_controller_python_syntax(self):
        ast.parse(CONTROLLER.read_text(encoding="utf-8"))


if __name__ == "__main__":
    unittest.main()

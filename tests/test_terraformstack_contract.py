#!/usr/bin/env python3
"""Verifica o contrato v1alpha1 e o transporte da pipeline."""

import json
from pathlib import Path
import unittest

import yaml
from jsonschema import Draft7Validator, ValidationError

ROOT = Path(__file__).resolve().parents[1]
TERRAFORMSTACK = ROOT / "control-plane/terraformstack"
CRD = TERRAFORMSTACK / "crd/terraformstack.yaml"
SIMULATOR = TERRAFORMSTACK / "examples/simulator/stacks.yaml"
CONTROLLER = TERRAFORMSTACK / "controller"
PIPELINE = ROOT / "workflow-plane/argo-workflows/terraformstack"


class TerraformStackContractTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        crd = yaml.safe_load(CRD.read_text(encoding="utf-8"))
        cls.schema = crd["spec"]["versions"][0]["schema"]["openAPIV3Schema"]
        Draft7Validator.check_schema(cls.schema)
        spec = cls.schema["properties"]["spec"]
        for fields in (spec, spec["properties"]["sourceRef"]):
            fields["additionalProperties"] = False
        cls.validator = Draft7Validator(cls.schema)

    @staticmethod
    def example(**changes):
        spec = {"sourceRef": {"repository": "https://example.org/stacks.git", "path": "stacks/network"}}
        spec.update(changes)
        return {"spec": spec}

    def test_simulator_manifests_match_pipeline_profile(self):
        manifests = list(yaml.safe_load_all(SIMULATOR.read_text(encoding="utf-8")))
        profile = json.loads((PIPELINE / "simulator.json").read_text(encoding="utf-8"))
        self.assertEqual({item["metadata"]["name"] for item in manifests}, set(profile["slots"]))
        for manifest in manifests:
            with self.subTest(stack=manifest["metadata"]["name"]):
                self.validator.validate(manifest)
                self.assertEqual(manifest["metadata"]["namespace"], profile["stackNamespace"])
                self.assertEqual(manifest["spec"]["sourceRef"]["path"], profile["path"])
                self.assertEqual(manifest["spec"]["sourceRef"]["repository"], profile["repository"])
                self.assertEqual(manifest["spec"]["context"], profile["context"])

    def test_optional_context_and_version(self):
        self.validator.validate(self.example())
        self.validator.validate(self.example(context={"anything": "any value"}))
        self.validator.validate(self.example(workflowRef={"name": "custom-workflow"}))
        self.validator.validate(self.example(sourceRef={
            "repository": "opaque-repository", "path": "stacks/network", "version": "v1.2.3",
        }))
        properties = self.schema["properties"]["spec"]["properties"]
        self.assertEqual(properties["sourceRef"]["properties"]["version"]["default"], "latest")
        self.assertFalse(properties["lifecycle"]["properties"]["destroy"]["default"])
        self.assertFalse(properties["lifecycle"]["properties"]["destroyResourcesOnDeletion"]["default"])

    def test_previous_fields_and_invalid_types_are_rejected(self):
        for request in (
            {"spec": {}},
            self.example(sourceRef={"name": "foundation-stacks", "path": "a"}),
            self.example(sourceRef={"repository": "repo", "path": "a", "name": "old"}),
            self.example(sourceRef={"repository": "repo", "path": "/absolute"}),
            self.example(context={"scope": ["one", "two"]}),
            self.example(workflowRef={"name": "../custom"}),
            self.example(workflowRef={}),
            self.example(target={"entity": "entity-a", "environment": "uat"}),
            self.example(configuration={"version": "latest"}),
            self.example(timeout={"apply": "15m"}),
            self.example(healthChecks=[]),
            self.example(module={"name": "old"}),
            self.example(reconciliation={"driftCheckInterval": "garbage"}),
            self.example(lifecycle={"destroy": "true"}),
            self.example(reconciliation={"retryPolicy": {"strategy": "Always"}}),
        ):
            with self.subTest(request=request), self.assertRaises(ValidationError):
                self.validator.validate(request)

    def test_controls_and_workflow_contract(self):
        self.validator.validate(self.example(
            dependsOn=[{"name": "network-uat"}], suspend=True,
            reconciliation={"driftCheckInterval": "6h", "retryPolicy": {
                "strategy": "ExponentialBackoff", "interval": "5m", "maxInterval": "1h",
            }},
            lifecycle={"destroy": False, "destroyResourcesOnDeletion": False},
        ))
        template = yaml.safe_load((PIPELINE / "workflowtemplate.yaml").read_text(encoding="utf-8"))
        config = yaml.safe_load((CONTROLLER / "controller-config.yaml").read_text(encoding="utf-8"))
        transport = next(w["parameters"] for w in config["workflows"] if w["name"] == config["defaultWorkflow"])
        inputs = {p["name"] for p in template["spec"]["arguments"]["parameters"]}
        self.assertEqual(inputs, {transport["input"]["intent"]})
        templates = {item["name"]: item for item in template["spec"]["templates"]}
        steps = lambda name: [item["name"] for group in templates[name]["steps"] for item in group]
        self.assertEqual(steps("reconcile"), ["resolve", "prepare", "execute", "cleanup"])
        self.assertEqual(steps("execute"), ["init", "validate", "plan", "opa", "decision",
                                            "approval", "authorize", "apply", "verify"])
        outputs = {p["globalName"] for t in templates.values()
                   for p in t.get("outputs", {}).get("parameters", []) if "globalName" in p}
        self.assertTrue(set(transport["output"].values()).issubset(outputs))
        self.assertEqual(templates["opa"]["container"]["command"], ["/opa"])
        self.assertNotIn("volumes", templates["resolve"])
        cleanup_step = template["spec"]["templates"][0]["steps"][-1][0]
        self.assertEqual(cleanup_step["continueOn"], {"failed": True, "error": True})
        self.assertEqual(templates["cleanup"]["container"]["args"], ["cleanup"])
        confirmed = {p.get("globalName") for p in templates["prepare"]["outputs"]["parameters"]}
        self.assertIn(transport["output"]["resolvedSource"], confirmed)
        for name in ["resolve", "prepare", "init", "validate", "plan", "apply", "verify", "cleanup"]:
            self.assertEqual(templates[name]["container"]["image"], "terraformstack-runner:local")


if __name__ == "__main__":
    unittest.main()

#!/usr/bin/env python3
"""Exercita a resolução Git no mesmo runner usado pelo Workflow."""

import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[1]
PIPELINE = ROOT / "workflow-plane/argo-workflows/terraformstack"
RUNNER = os.environ.get("TERRAFORMSTACK_RUNNER", "terraformstack-runner:local")
REPOSITORY = "file:///fixture/source"


class TerraformStackResolutionTest(unittest.TestCase):
    def setUp(self):
        tmp = tempfile.TemporaryDirectory()
        self.addCleanup(tmp.cleanup)
        self.fixture = Path(tmp.name)
        self.source = self.fixture / "source"
        self.source.mkdir()
        self.git("init", "--quiet", "--initial-branch=main")
        self.git("commit", "--quiet", "--allow-empty", "-m", "Primeira revisão")
        self.first = self.git("rev-parse", "HEAD")
        self.git("tag", "-a", "v1.0.0", "-m", "Revisão anotada")
        self.git("commit", "--quiet", "--allow-empty", "-m", "Segunda revisão")
        self.head = self.git("rev-parse", "HEAD")
        self.profile = json.loads((PIPELINE / "simulator.json").read_text())
        self.profile.update(repository=REPOSITORY, gitOpsRepository=REPOSITORY)

    def git(self, *args):
        return subprocess.check_output([
            "git", "-C", str(self.source), "-c", "user.name=Simulator",
            "-c", "user.email=simulator@example.invalid", *args,
        ], text=True).strip()

    def resolve(self, *, version="latest", annotation="", context=None, **controls):
        profile = self.fixture / "simulator.json"
        profile.write_text(json.dumps(self.profile))
        request = {
            "workflow": "fixture",
            "intentHash": "a" * 64,
            "intent": {
                "stack": "simulator-network",
                "namespace": self.profile["stackNamespace"],
                "sourceRef": {"repository": REPOSITORY, "path": self.profile["path"], "version": version},
                "revisionAnnotation": annotation,
                "context": self.profile["context"] if context is None else context,
            },
            **controls,
        }
        output = subprocess.check_output([
            "docker", "run", "--rm", "--network=none", "--user", f"{os.getuid()}:{os.getgid()}",
            "-v", f"{self.fixture}:/fixture:ro",
            "-v", f"{profile}:/platform/profile/simulator.json:ro",
            "-v", f"{PIPELINE / 'pipeline.sh'}:/pipeline.sh:ro",
            "-e", f"INTENT={json.dumps(request)}", "-e", "WORKFLOW=fixture", RUNNER,
            "-ec", 'bash /pipeline.sh resolve; if grep -qx true /tmp/allowed; then '
            'cat /tmp/execution-context; else cat /tmp/execution-result; fi',
        ], text=True)
        return json.loads(output)

    def test_branch_tag_commit_and_gitops_resolution(self):
        for version, annotation, expected in (
            ("refs/heads/main", self.first, self.head),
            ("v1.0.0", self.head, self.first),
            (self.first, "", self.first),
            ("latest", self.first, self.first),
            ("latest", "", self.head),
        ):
            with self.subTest(version=version, annotation=annotation):
                self.assertEqual(self.resolve(version=version, annotation=annotation)["revision"], expected)

    def test_external_source_uses_its_own_head(self):
        self.profile["gitOpsRepository"] = "https://example.invalid/gitops.git"
        self.assertEqual(self.resolve(annotation="invalid-external-sha")["revision"], self.head)

    def test_pin_and_destination_binding(self):
        resolved = self.resolve(annotation=self.first)
        pinned = self.resolve(annotation=self.head, pinnedReference=self.first, binding=resolved["binding"])
        self.assertEqual(pinned["revision"], self.first)
        self.assertEqual(pinned["binding"], resolved["binding"])

        self.profile["backend"]["path"] = "/state/different.tfstate"
        self.assertEqual(self.resolve(binding=resolved["binding"])["reason"], "BindingChanged")

    def test_rejects_unapproved_context_and_gitops_annotation(self):
        self.assertEqual(self.resolve(context={"environment": "other"})["reason"], "ContextNotAllowed")
        self.assertEqual(self.resolve(annotation="HEAD")["reason"], "InvalidGitOpsRevision")


if __name__ == "__main__":
    unittest.main()

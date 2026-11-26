"""Regressões do reconciliador; sem cluster, Terraform ou AWS."""
from copy import deepcopy
from datetime import timedelta

import pytest
from kubernetes.client.exceptions import ApiException

from conftest import controller, NOW, SHA


def test_intent_pins_revisions_and_catalog(cluster):
    stack = cluster.add()
    data = controller.intent(stack, cluster.catalog)
    assert data["revision"] == data["configuration"] == SHA
    assert data["serviceAccount"] == "terraformstack-executor"
    assert data["backend"]["path"] == "/state/terraform.tfstate"


@pytest.mark.parametrize("change", [
    {"target": {"entity": "entity-a", "environment": "prod"}},
    {"sourceRef": {"name": "foundation-stacks", "path": "../other"}},
    {"configuration": {"version": "main"}},
])
def test_invalid_intent_blocks_execution(cluster, change):
    cluster.add(**change)
    cluster.tick()
    assert cluster.condition()["reason"] == "InvalidIntent"
    assert not cluster.workflows


def test_persist_before_create_and_resume_idempotently(cluster):
    stack = cluster.add()
    cluster.tick()
    assert controller.FINALIZER in stack["metadata"]["finalizers"]
    cluster.tick()
    assert cluster.run()["sequence"] == 1
    assert not cluster.workflows
    cluster.tick(count=5)
    assert len(cluster.workflows) == 1
    assert cluster.run()["created"] is True
    assert cluster.api.create_namespaced_custom_object.call_count == 1


def test_stale_snapshot_cannot_overwrite_execution(cluster):
    stack = cluster.add()
    cluster.tick()
    old = deepcopy(stack)
    cluster.tick()
    with pytest.raises(ApiException) as exc:
        controller.reconcile(cluster.api, old, {old["metadata"]["name"]: old}, cluster.catalog)
    assert exc.value.status == 409
    assert cluster.run()["sequence"] == 1
    cluster.tick(count=3)
    assert len(cluster.workflows) == 1


def test_missing_active_workflow_fails_closed(cluster):
    cluster.add()
    cluster.start()
    cluster.workflows.clear()
    cluster.tick(count=3)
    assert cluster.condition()["reason"] == "WorkflowMissing"
    assert not cluster.workflows


def test_pending_approval_cancelled_on_spec_change(cluster):
    cluster.add()
    wf = cluster.start()
    wf["status"] = {"phase": "Running", "nodes": {"gate": {
        "type": "Suspend", "templateName": "approval", "phase": "Running"}}}
    cluster.tick()
    assert cluster.condition()["type"] == "ApprovalRequired"
    cluster.change(configuration={"version": "b" * 40})
    cluster.tick(count=2)
    assert cluster.run()["cancelling"]
    assert wf["spec"]["shutdown"] == "Terminate"
    cluster.outcome(phase="Failed")
    cluster.tick(count=3)
    assert cluster.run()["sequence"] == 2
    assert cluster.run()["intent"]["configuration"] == "b" * 40


def test_apply_in_progress_not_terminated_on_spec_change(cluster):
    stack = cluster.add()
    wf = cluster.start()
    cluster.run()["applying"] = True
    cluster.change(configuration={"version": "b" * 40})
    cluster.tick()
    assert "shutdown" not in wf["spec"]
    cluster.outcome(decision="AutoApply", outcome="Verified")
    cluster.tick()
    assert cluster.condition()["reason"] == "PreviousExecutionFinished"
    cluster.tick()
    assert cluster.run()["sequence"] == 2


@pytest.mark.parametrize("decision,request,expected", [
    ("AutoApply", "hash", True),
    ("RequireApproval", "hash", True),
    ("RequireApproval", "different", False),
    ("Deny", "hash", False),
    ("Error", "hash", False),
])
def test_authorization_requires_allowed_decision_and_matching_hash(cluster, decision, request, expected):
    cluster.add()
    cluster.start()
    cluster.outcome(phase="Running", decision=decision, request=request, hash="hash")
    cluster.tick()
    assert (cluster.run().get("authorization") == "hash") is expected


def test_dependency_requires_current_ready_generation(cluster):
    cluster.add()
    child = "simulator-dependent"
    cluster.add(child, dependsOn=[{"name": "simulator-network"}])
    cluster.tick(child, 3)
    assert cluster.condition(child)["reason"] == "DependencyBlocked"
    cluster.start()
    cluster.outcome()
    cluster.tick()
    cluster.tick(child, 3)
    assert cluster.run(child)["sequence"] == 1
    cluster.change(configuration={"version": "b" * 40})
    cluster.tick(child, 2)
    assert cluster.run(child)["cancelling"]


def test_destroy_waits_for_dependents(cluster):
    cluster.add(lifecycle={"destroy": True})
    cluster.add("simulator-dependent", dependsOn=[{"name": "simulator-network"}])
    cluster.tick(count=3)
    assert cluster.condition()["reason"] == "DependencyBlocked"
    assert not cluster.workflows


def test_drift_interval_and_policy_denial(cluster, monkeypatch):
    cluster.add(reconciliation={"driftCheckInterval": "45s"})
    cluster.start()
    cluster.outcome()
    cluster.tick()
    cluster.tick()
    assert cluster.run()["sequence"] == 1
    monkeypatch.setattr(controller, "now", lambda: NOW + timedelta(seconds=45))
    cluster.tick(count=2)
    assert cluster.run()["trigger"] == "DriftCheck"
    cluster.outcome(decision="Deny", drift=True)
    cluster.tick(count=3)
    assert cluster.condition()["reason"] == "PolicyDenied"
    assert cluster.run()["drift"] is True
    assert cluster.run()["sequence"] == 2


def test_retry_respects_interval_and_limit(cluster, monkeypatch):
    cluster.add(reconciliation={"retryPolicy": {"strategy": "Fixed", "interval": "10s", "maxRetries": 1}})
    cluster.start()
    cluster.outcome(phase="Failed", decision="AutoApply")
    cluster.tick(count=2)
    assert cluster.run()["attempt"] == 0
    monkeypatch.setattr(controller, "now", lambda: NOW + timedelta(seconds=10))
    cluster.tick(count=2)
    assert cluster.run()["attempt"] == 1
    assert cluster.run()["trigger"] == "Retry"
    cluster.outcome(phase="Failed", decision="AutoApply")
    cluster.tick()
    monkeypatch.setattr(controller, "now", lambda: NOW + timedelta(hours=1))
    cluster.tick(count=3)
    assert cluster.run()["sequence"] == 2


def test_exponential_retry_backoff():
    settings = {"retryPolicy": {"maxRetries": 5, "interval": "10s", "maxInterval": "25s"}}
    assert not controller.retry_due(settings, {"attempt": 1}, 19)
    assert controller.retry_due(settings, {"attempt": 1}, 20)
    assert controller.retry_due(settings, {"attempt": 3}, 25)
    assert not controller.retry_due(settings, {"attempt": 5}, 1000)


def test_delete_retains_by_default(cluster):
    stack = cluster.add()
    cluster.tick()
    stack["metadata"]["deletionTimestamp"] = "2026-01-01T00:00:00Z"
    cluster.tick(count=2)
    assert controller.FINALIZER not in stack["metadata"]["finalizers"]
    assert not cluster.workflows


def test_destroy_finalizer_removed_only_after_verification(cluster):
    stack = cluster.add(lifecycle={"destroyResourcesOnDeletion": True})
    cluster.tick()
    stack["metadata"]["deletionTimestamp"] = "2026-01-01T00:00:00Z"
    cluster.tick(count=4)
    assert cluster.run()["intent"]["operation"] == "destroy"
    assert controller.FINALIZER in stack["metadata"]["finalizers"]
    cluster.outcome(decision="RequireApproval", outcome="Verified")
    cluster.tick(count=2)
    assert controller.FINALIZER not in stack["metadata"]["finalizers"]


def test_failed_destroy_retains_finalizer(cluster):
    stack = cluster.add(lifecycle={"destroyResourcesOnDeletion": True})
    cluster.tick()
    stack["metadata"]["deletionTimestamp"] = "2026-01-01T00:00:00Z"
    cluster.tick(count=4)
    cluster.outcome(phase="Failed", decision="RequireApproval")
    cluster.tick(count=3)
    assert controller.FINALIZER in stack["metadata"]["finalizers"]
    assert cluster.condition()["type"] == "Failed"


def test_suspended_stack_does_not_schedule(cluster):
    cluster.add(suspend=True)
    cluster.tick(count=4)
    assert cluster.condition()["reason"] == "Suspended"
    assert not cluster.workflows


def test_opa_error_overrides_decision(cluster):
    cluster.add()
    cluster.start()
    wf = cluster.outcome(phase="Failed", decision="AutoApply")
    wf["status"]["nodes"] = {"opa": {"templateName": "opa", "phase": "Error"}}
    cluster.tick()
    assert cluster.condition()["reason"] == "PolicyError"

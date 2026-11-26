package terraformstack.local

import rego.v1

snapshot(cidr) := {
    "metadata": [{"name": "lab06b-network", "namespace": "terraformstack-lab06b"}],
    "data": {"network.json": json.marshal({"vpc": {"cidr": cidr}})},
}

change(actions, before, after) := {
    "address": "kubernetes_config_map_v1.network",
    "change": {"actions": actions, "before": before, "after": after},
}

scenario(operation, drift, changes) := {
    "operation": operation, "slot": "lab06b-network", "drift": drift, "changes": changes,
}

test_auto_apply_baseline if {
    entry := change(["create"], null, snapshot("10.81.0.0/16"))
    result := decision with input as scenario("apply", false, [entry])
    result.decision == "AutoApply"
}

test_approval_for_declared_change if {
    entry := change(["update"], snapshot("10.81.0.0/16"), snapshot("10.82.0.0/16"))
    result := decision with input as scenario("apply", false, [entry])
    result.decision == "RequireApproval"
}

test_approval_for_declared_return_to_baseline if {
    entry := change(["update"], snapshot("10.82.0.0/16"), snapshot("10.81.0.0/16"))
    result := decision with input as scenario("apply", false, [entry])
    result.decision == "RequireApproval"
}

test_auto_apply_drift if {
    entry := change(["update"], snapshot("10.83.0.0/16"), snapshot("10.81.0.0/16"))
    result := decision with input as scenario("apply", true, [entry])
    result.decision == "AutoApply"
}

test_deny_public_cidr if {
    entry := change(["update"], snapshot("10.81.0.0/16"), snapshot("0.0.0.0/0"))
    result := decision with input as scenario("apply", false, [entry])
    result.decision == "Deny"
}

test_deny_foreign_resource if {
    entry := object.union(change(["create"], null, snapshot("10.81.0.0/16")), {"address": "other"})
    result := decision with input as scenario("apply", false, [entry])
    result.decision == "Deny"
}

test_approval_for_destroy if {
    entry := change(["delete"], snapshot("10.81.0.0/16"), null)
    result := decision with input as scenario("destroy", false, [entry])
    result.decision == "RequireApproval"
}

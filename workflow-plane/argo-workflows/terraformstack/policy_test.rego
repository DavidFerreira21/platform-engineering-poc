package terraformstack

import rego.v1

snapshot(cidr) := {
    "metadata": [{"name": "simulator-network", "namespace": "terraformstack-simulator"}],
    "data": {"network.json": json.marshal({"vpc": {"cidr": cidr}})},
}

change(actions, before, after) := {
    "address": "kubernetes_config_map_v1.network",
    "change": {"actions": actions, "before": before, "after": after},
}

resource_drift(false, changes) := []

resource_drift(true, changes) := changes

scenario(operation, drift, changes) := {
    "document": json.marshal({"resource_changes": changes, "resource_drift": resource_drift(drift, changes)}),
    "context": json.marshal({
        "operation": operation,
        "variables": {"slot": "simulator-network", "namespace": "terraformstack-simulator"},
        "exitCode": 2,
    }),
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


test_no_changes if {
    fixture := {
        "document": json.marshal({"resource_changes": []}),
        "context": json.marshal({"operation":"apply", "exitCode":0}),
    }
    result := decision with input as fixture
    result.decision == "NoChanges"
    result.documentHash == crypto.sha256(fixture.document)
    result.contextHash == crypto.sha256(fixture.context)
}

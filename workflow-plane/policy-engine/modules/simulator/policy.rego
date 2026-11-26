package policy_modules.simulator

import rego.v1

network_address := "kubernetes_config_map_v1.network"

plan := json.unmarshal(input.document)
context := json.unmarshal(input.context)

changes := [change |
    some change in object.get(plan, "resource_changes", [])
    change.change.actions != ["no-op"]
]

valid_snapshot(null) := true

valid_snapshot(item) if {
    count(item.metadata) == 1
    item.metadata[0].namespace == context.variables.namespace
    item.metadata[0].name == context.variables.slot
    count(item.data) == 1
    is_object(json.unmarshal(item.data["network.json"]))
}

has_snapshot(change) if change.before != null
has_snapshot(change) if change.after != null

safe if {
    every change in changes {
        change.address == network_address
        has_snapshot(change.change)
        valid_snapshot(change.change.before)
        valid_snapshot(change.change.after)
    }
}

changed_data(change) if change.before == null
changed_data(change) if change.after == null
changed_data(change) if change.before.data != change.after.data

default drift := false

drift if {
    some resource in object.get(plan, "resource_drift", [])
    resource.address == network_address
    changed_data(resource.change)
}

desired_cidr := cidr if {
    safe
    context.operation == "apply"
    count(changes) == 1
    change := changes[0]
    change.change.actions in [["create"], ["update"]]
    cidr := json.unmarshal(change.change.after.data["network.json"]).vpc.cidr
    net.cidr_contains("10.0.0.0/8", cidr)
}

baseline_automatic if {
    desired_cidr == "10.81.0.0/16"
    changes[0].change.actions == ["create"]
}

baseline_automatic if {
    desired_cidr == "10.81.0.0/16"
    drift
}

default gate := {"decision": "Deny", "reason": "Mudança fora do contrato do Simulator"}

gate := {"decision": "NotApplicable", "reason": "Plan sem mudanças"} if {
    context.operation == "apply"
    context.exitCode == 0
    count(changes) == 0
}

gate := {"decision": "RequireApproval", "reason": "Destroy exige aprovação"} if {
    safe
    context.operation == "destroy"
    every change in changes {
        change.change.actions == ["delete"]
    }
}

gate := {"decision": "AutoApply", "reason": "ConfigMap com CIDR baseline"} if baseline_automatic

gate := {"decision": "RequireApproval", "reason": "Alteração declarada no ConfigMap"} if {
    is_string(desired_cidr)
    not baseline_automatic
}

cidrs := {phase: cidr |
    some change in changes
    change.address == network_address
    some phase in ["before", "after"]
    cidr := json.unmarshal(change.change[phase].data["network.json"]).vpc.cidr
    net.cidr_is_valid(cidr)
}

# Cobertura explícita: somente endereços reconhecidos pelo módulo.
result := object.union(gate, {
    "covered": [change.address | some change in changes; change.address == network_address],
    "drift": drift,
    "cidrs": cidrs,
})

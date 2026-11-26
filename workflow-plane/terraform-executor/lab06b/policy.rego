package terraformstack.local

import rego.v1

default decision := {"decision": "Deny", "reason": "Mudança fora do contrato local"}

valid_change(change) if {
    change.address == "kubernetes_config_map_v1.network"
    some snapshot in [change.change.before, change.change.after]
    snapshot != null
    every item in [change.change.before, change.change.after] {
        valid_snapshot(item)
    }
}

valid_snapshot(null) := true

valid_snapshot(item) if {
    count(item.metadata) == 1
    item.metadata[0].namespace == "terraformstack-lab06b"
    item.metadata[0].name == input.slot
    count(item.data) == 1
    is_object(json.unmarshal(item.data["network.json"]))
}

safe if {
    input.slot in {"lab06b-network", "lab06b-dependent"}
    every change in input.changes {
        valid_change(change)
    }
}

destroy if {
    safe
    input.operation == "destroy"
    every change in input.changes {
        change.change.actions == ["delete"]
    }
}

desired_cidr := cidr if {
    safe
    input.operation == "apply"
    count(input.changes) == 1
    change := input.changes[0]
    change.change.actions in [["create"], ["update"]]
    cidr := json.unmarshal(change.change.after.data["network.json"]).vpc.cidr
    net.cidr_contains("10.0.0.0/8", cidr)
}

decision := {"decision": "RequireApproval", "reason": "Destroy local exige aprovação"} if destroy

baseline_automatic if {
    desired_cidr == "10.81.0.0/16"
    input.changes[0].change.actions == ["create"]
}

baseline_automatic if {
    desired_cidr == "10.81.0.0/16"
    input.drift == true
}

decision := {"decision": "AutoApply", "reason": "ConfigMap local com CIDR baseline"} if {
    baseline_automatic
}

decision := {"decision": "RequireApproval", "reason": "Alteração declarada no ConfigMap local"} if {
    is_string(desired_cidr)
    not baseline_automatic
}

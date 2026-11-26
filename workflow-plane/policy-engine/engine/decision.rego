package terraformstack

import rego.v1

# O contexto é produzido pelo runner a partir do perfil de execução confiável.
plan := json.unmarshal(input.document)
context := json.unmarshal(input.context)
profile := object.get(context, "policyProfile", "")
profiles := data.profiles
selected := object.get(profiles, profile, [])
changes := [change |
    some change in object.get(plan, "resource_changes", [])
    change.change.actions != ["no-op"]
]

# Contrato obrigatório de cada módulo. Falha ou resultado ausente não vira allow.
valid_module(result) if {
    result.decision in {"AutoApply", "RequireApproval", "Deny", "NotApplicable", "Error"}
    is_string(result.reason)
    result.reason != ""
    is_array(result.covered)
    every address in result.covered {
        is_string(address)
    }
    # Um módulo que não se aplica não pode autorizar cobertura.
    not invalid_coverage(result)
}

invalid_coverage(result) if {
    result.decision == "NotApplicable"
    count(result.covered) > 0
}

module_result(name) := result if {
    result := data.policy_modules[name].result
    valid_module(result)
} else := {"decision": "Error", "reason": "Módulo ausente ou inválido", "covered": []}

results := [object.union(module_result(name), {"name": name}) | some name in selected]

valid_profile if {
    is_array(selected)
    count(selected) > 1
    "common" in selected
    count(selected) == count({name | some name in selected})
    every name in selected {
        is_string(name)
    }
}

has_decision(value) if {
    some result in results
    result.decision == value
}

covered(address) if {
    some result in results
    result.decision in {"AutoApply", "RequireApproval"}
    address in result.covered
}

uncovered := {change.address | some change in changes; not covered(change.address)}

# Fail closed e precedência: Error > Deny > RequireApproval > AutoApply.
# NoChanges é uma decisão exclusiva do engine, após validar o perfil.
gate := {"decision": "Error", "reason": "Perfil ausente ou inválido"} if {
    not valid_profile
} else := {"decision": "Error", "reason": "Falha de avaliação de módulo"} if {
    has_decision("Error")
} else := {"decision": "Deny", "reason": "Mudança bloqueada por módulo"} if {
    has_decision("Deny")
} else := {"decision": "NoChanges", "reason": "Plan sem mudanças"} if {
    context.operation == "apply"
    context.exitCode == 0
    count(changes) == 0
} else := {"decision": "Deny", "reason": "Mudança sem cobertura de política"} if {
    count(uncovered) > 0
} else := {"decision": "RequireApproval", "reason": "Aprovação exigida por módulo"} if {
    has_decision("RequireApproval")
} else := {"decision": "AutoApply", "reason": "Mudanças cobertas e permitidas"} if {
    context.operation == "apply"
    has_decision("AutoApply")
} else := {"decision": "Deny", "reason": "Sem autorização explícita"}

default drift := false
drift if {
    some result in results
    object.get(result, "drift", false) == true
}

# Preserva o resumo usado no Lab 06; módulos podem expor seus próprios detalhes.
cidrs := object.get(object.get({result.name: result | some result in results}, "simulator", {}), "cidrs", {})

decision := object.union(gate, {
    "profile": profile,
    "modules": results,
    "documentHash": crypto.sha256(input.document),
    "contextHash": crypto.sha256(input.context),
    "drift": drift,
    "actions": [change.change.actions | some change in changes],
    "cidrs": cidrs,
})

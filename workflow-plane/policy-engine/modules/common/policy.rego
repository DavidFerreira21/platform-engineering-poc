package policy_modules.common

import rego.v1

context := json.unmarshal(input.context)

# Regra transversal: destroy nunca é automático. Este módulo não concede
# cobertura a recursos; os módulos especializados precisam fazê-lo.
default result := {"decision": "NotApplicable", "reason": "Sem restrição adicional", "covered": []}

result := {"decision": "RequireApproval", "reason": "Destroy exige aprovação", "covered": []} if {
    context.operation == "destroy"
}

locals {
  # A hierarquia segue a profundidade do diretório: global < Entity < Environment.
  # Apenas arquivos opcionais ausentes herdam valores. YAML inválido falha no plan.
  paths = [
    "${var.config_root}/${var.stack}.yaml",
    "${var.config_root}/${var.entity}/${var.stack}.yaml",
    "${var.config_root}/${var.entity}/${var.environment}/${var.stack}.yaml",
  ]

  global_values = fileexists(local.paths[0]) ? yamldecode(file(local.paths[0])) : {}
  entity_values = fileexists(local.paths[1]) ? yamldecode(file(local.paths[1])) : {}
  env_values    = fileexists(local.paths[2]) ? yamldecode(file(local.paths[2])) : {}

  # O contrato define três camadas de mesclagem em mapas. Para cada chave de
  # nível superior, submapas são mesclados até o terceiro nível. Listas e
  # valores escalares são substituídos pelo último nível declarado.
  top_keys = toset(concat(keys(local.global_values), keys(local.entity_values), keys(local.env_values)))
  merged = {
    for top in local.top_keys : top => merge(
      try(local.global_values[top], {}),
      try(local.entity_values[top], {}),
      try(local.env_values[top], {}),
      {
        for nested in toset(concat(
          keys(try(local.global_values[top], {})),
          keys(try(local.entity_values[top], {})),
          keys(try(local.env_values[top], {}))
        )) : nested => (
          can(tomap(try(local.global_values[top][nested], {}))) &&
          can(tomap(try(local.entity_values[top][nested], {}))) &&
          can(tomap(try(local.env_values[top][nested], {})))
        ) ? merge(
          try(local.global_values[top][nested], {}),
          try(local.entity_values[top][nested], {}),
          try(local.env_values[top][nested], {})
        ) : coalesce(
          try(local.env_values[top][nested], null),
          try(local.entity_values[top][nested], null),
          try(local.global_values[top][nested], null)
        )
      }
    )
  }
}

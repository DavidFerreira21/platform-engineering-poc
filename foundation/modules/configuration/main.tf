locals {
  # Precedência: global < Entity < Environment. YAML inválido falha explicitamente.
  paths = [
    "${var.config_root}/${var.stack}.yaml",
    "${var.config_root}/${var.entity}/${var.stack}.yaml",
    "${var.config_root}/${var.entity}/${var.environment}/${var.stack}.yaml",
  ]

  global_values = fileexists(local.paths[0]) ? yamldecode(file(local.paths[0])) : {}
  entity_values = fileexists(local.paths[1]) ? yamldecode(file(local.paths[1])) : {}
  env_values    = fileexists(local.paths[2]) ? yamldecode(file(local.paths[2])) : {}

  # Typed nested merges para os contratos conhecidos. Evitam que coerção de
  # map/ternary transforme flags booleanas em strings (true -> "true").
  network = merge(
    try(local.global_values.network, {}),
    try(local.entity_values.network, {}),
    try(local.env_values.network, {}),
    {
      vpc = merge(
        try(local.global_values.network.vpc, {}),
        try(local.entity_values.network.vpc, {}),
        try(local.env_values.network.vpc, {})
      )
      nat = merge(
        try(local.global_values.network.nat, {}),
        try(local.entity_values.network.nat, {}),
        try(local.env_values.network.nat, {})
      )
      endpoints = merge(
        try(local.global_values.network.endpoints, {}),
        try(local.entity_values.network.endpoints, {}),
        try(local.env_values.network.endpoints, {}),
        {
          s3 = merge(
            try(local.global_values.network.endpoints.s3, {}),
            try(local.entity_values.network.endpoints.s3, {}),
            try(local.env_values.network.endpoints.s3, {})
          )
          ecr = merge(
            try(local.global_values.network.endpoints.ecr, {}),
            try(local.entity_values.network.endpoints.ecr, {}),
            try(local.env_values.network.endpoints.ecr, {})
          )
        }
      )
    }
  )

  eks = merge(
    try(local.global_values.eks, {}),
    try(local.entity_values.eks, {}),
    try(local.env_values.eks, {}),
    {
      clusters = merge(
        try(local.global_values.eks.clusters, {}),
        try(local.entity_values.eks.clusters, {}),
        try(local.env_values.eks.clusters, {})
      )
      addons = merge(
        try(local.global_values.eks.addons, {}),
        try(local.entity_values.eks.addons, {}),
        try(local.env_values.eks.addons, {})
      )
      networking = merge(
        try(local.global_values.eks.networking, {}),
        try(local.entity_values.eks.networking, {}),
        try(local.env_values.eks.networking, {})
      )
    }
  )

  # Outras stacks podem introduzir seus próprios contratos de composição sem
  # modificar TerraformStack; o fallback é merge superficial, explícito.
  configurations = {
    network = { network = local.network }
    eks     = { eks = local.eks }
  }
  merged = try(local.configurations[var.stack], merge(local.global_values, local.entity_values, local.env_values))
}

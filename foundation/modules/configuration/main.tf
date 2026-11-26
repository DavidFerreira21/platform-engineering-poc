terraform {
  required_version = ">= 1.8.0"

  required_providers {
    deepmerge = {
      source  = "isometry/deepmerge"
      version = "1.3.1"
    }
  }
}

locals {
  paths = [
    "${var.config_root}/${var.stack}.yaml",
    "${var.config_root}/${var.entity}/${var.stack}.yaml",
    "${var.config_root}/${var.entity}/${var.environment}/${var.stack}.yaml",
  ]

  source_paths = [for path in local.paths : path if fileexists(path)]

  # Cada documento fornece um objeto. A decodificação preserva os tipos do YAML.
  layers = [for path in local.source_paths : merge({}, yamldecode(file(path)))]

  # Objetos são mesclados recursivamente. Os demais valores são substituídos.
  merged = provider::deepmerge::mergo({}, local.layers...)
}

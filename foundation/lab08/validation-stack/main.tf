variable "entity" {
  type = string
}
variable "environment" {
  type = string
}
variable "stack" {
  type = string
  validation {
    condition     = contains(["network", "eks"], var.stack)
    error_message = "Este laboratório valida apenas network ou eks."
  }
}

module "configuration" {
  source      = "../../modules/configuration"
  config_root = abspath("${path.module}/../../config")
  entity      = var.entity
  environment = var.environment
  stack       = var.stack
}

# Somente locals e outputs. Nenhum recurso cloud é provisionado nesta stack.
output "loaded_files" {
  value = [for p in module.configuration.source_paths : basename(p)]
}

output "configuration_loaded" {
  value = length(keys(module.configuration.values)) > 0
}

output "network_cidr" {
  value = try(module.configuration.values.network.vpc.cidr, null)
}

output "network_dns_hostnames" {
  value = try(module.configuration.values.network.vpc.dnsHostnames, null)
}

output "eks_management_enabled" {
  value = try(module.configuration.values.eks.clusters.management, null)
}

output "eks_pci_enabled" {
  value = try(module.configuration.values.eks.clusters.pci, null)
}

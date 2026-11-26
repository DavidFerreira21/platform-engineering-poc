terraform {
  required_version = "= 1.13.5"
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "= 2.38.0"
    }
  }
  backend "local" {}
}

variable "slot" {
  type = string
}

variable "namespace" {
  type = string
}

variable "config_root" {
  type = string
}

provider "kubernetes" {}

module "configuration" {
  source      = "../../modules/configuration"
  config_root = var.config_root
  stack       = "network"
  entity      = "entity-a"
  environment = "uat"
}

resource "kubernetes_config_map_v1" "network" {
  metadata {
    name      = var.slot
    namespace = var.namespace
  }
  data = {
    "network.json" = jsonencode(module.configuration.values.network)
  }
}

output "network" {
  value = module.configuration.values.network
}

terraform {
  required_version = ">= 1.10"
  required_providers {
    aws = { source = "hashicorp/aws", version = ">= 5.83, < 6" }
  }
  backend "s3" {}
}
variable "entity" { type = string }
variable "environment" { type = string }
variable "region" { type = string }
variable "account_id" { type = string }
provider "aws" {
  region              = var.region
  allowed_account_ids = [var.account_id]
}
module "configuration" {
  source      = "../../modules/configuration"
  config_root = abspath("${path.module}/../../config")
  entity      = var.entity
  environment = var.environment
  stack       = "eks"
}
locals {
  config = module.configuration.values.eks
  name   = "${var.entity}-${var.environment}"
}
variable "vpc_id" { type = string }
variable "subnets" { type = map(list(string)) }
variable "kubernetes_version" { type = string }
variable "api_client_cidrs" {
  description = "CIDRs privados autorizados a acessar a API via HTTPS a partir da rede de gestão."
  type        = list(string)
}
variable "admin_role_arn" { type = string }
module "eks" {
  source             = "../../modules/eks"
  name               = local.name
  api_client_cidrs   = var.api_client_cidrs
  clusters           = local.config.clusters
  private_endpoint   = local.config.networking.privateEndpoint
  kubernetes_version = var.kubernetes_version
  admin_role_arn     = var.admin_role_arn
  vpc_id             = var.vpc_id
  subnets            = var.subnets
}
output "clusters" { value = module.eks.clusters }
output "bridge_parameters" {
  value = local.config.publishParameterStore ? merge({}, [for profile, cluster in module.eks.clusters : {
    "eks/${profile}/cluster-name"     = cluster.name
    "eks/${profile}/cluster-endpoint" = cluster.endpoint
  }]...) : {}
}

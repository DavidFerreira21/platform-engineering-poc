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
  stack       = "network"
}
locals {
  config = module.configuration.values.network
  name   = "${var.entity}-${var.environment}"
}
variable "azs" { type = list(string) }
module "network" {
  source = "../../modules/network"
  name   = local.name
  azs    = var.azs
  config = local.config
}
output "vpc_id" { value = module.network.vpc_id }
output "subnets" { value = module.network.subnets }
output "bridge_parameters" {
  value = local.config.publishParameterStore ? merge(
    { "network/vpc-id" = module.network.vpc_id },
    { for group, ids in module.network.subnets : "network/${group}-subnet-ids" => join(",", ids) }
  ) : {}
}

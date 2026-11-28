provider "aws" {
  profile             = "platform-hub"
  region              = var.region
  allowed_account_ids = var.account_id == null ? null : [var.account_id]

  default_tags {
    tags = {
      Project     = "platform-engineering-poc"
      ManagedBy   = "terraform"
      Entity      = "platform"
      Environment = "hub"
    }
  }
}

locals {
  kubernetes_version = var.kubernetes_version != null ? var.kubernetes_version : format("1.%d", max([
    for version in data.aws_eks_cluster_versions.standard[0].cluster_versions :
    tonumber(split(".", version.cluster_version)[1])
  ]...))
  admin_role_arn = var.admin_role_arn != null ? var.admin_role_arn : data.aws_iam_session_context.current[0].issuer_arn
}

module "network" {
  source = "../../modules/network"
  name   = "platform-hub"
  azs    = var.azs

  config = {
    layout = "hub"
    vpc = {
      enabled      = true
      cidr         = var.vpc_cidr
      dnsHostnames = true
      dnsSupport   = true
    }
    nat = {
      enabled = true
      mode    = "regional"
    }
    endpoints = {
      s3  = { enabled = true }
      ecr = { enabled = false }
    }
  }
}

module "eks" {
  source = "../../modules/eks"
  name   = "platform"

  clusters = {
    hub        = true
    management = false
    nonPci     = false
    pci        = false
  }

  vpc_id                  = module.network.vpc_id
  subnets                 = module.network.subnets
  private_endpoint        = true
  kubernetes_version      = local.kubernetes_version
  api_client_cidrs        = var.api_client_cidrs
  api_public_access_cidrs = var.api_public_access_cidrs
  admin_role_arn          = local.admin_role_arn
}

module "argocd" {
  count  = var.identity_center == null ? 0 : 1
  source = "../../modules/argocd"

  cluster_name    = module.eks.clusters["hub"].name
  identity_center = var.identity_center
}

moved {
  from = module.argocd
  to   = module.argocd[0]
}

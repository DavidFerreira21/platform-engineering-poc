locals {
  enabled = { for k, v in { hub = var.clusters.hub, management = var.clusters.management, nonpci = var.clusters.nonPci, pci = var.clusters.pci } : k => v if v }
  subnet_group = {
    hub        = contains(keys(var.subnets), "app") ? "app" : "web-nonpci"
    management = "web-nonpci"
    nonpci     = "web-nonpci"
    pci        = "web-pci"
  }
}
module "cluster" {
  for_each                                 = local.enabled
  source                                   = "terraform-aws-modules/eks/aws"
  version                                  = "21.1.0"
  name                                     = "${var.name}-${each.key}"
  kubernetes_version                       = var.kubernetes_version
  vpc_id                                   = var.vpc_id
  subnet_ids                               = var.subnets[local.subnet_group[each.key]]
  endpoint_private_access                  = var.private_endpoint
  endpoint_public_access                   = length(var.api_public_access_cidrs) > 0
  endpoint_public_access_cidrs             = length(var.api_public_access_cidrs) > 0 ? var.api_public_access_cidrs : null
  compute_config                           = { enabled = true, node_pools = ["system", "general-purpose"] }
  authentication_mode                      = "API"
  enable_cluster_creator_admin_permissions = false
  enabled_log_types = each.key == "hub" ? [
    "api", "audit", "authenticator", "controllerManager", "scheduler"
  ] : ["audit", "api", "authenticator"]
  security_group_additional_rules = length(var.api_client_cidrs) == 0 ? {} : {
    management = {
      description = "HTTPS a partir da rede de gestao"
      protocol    = "tcp"
      from_port   = 443
      to_port     = 443
      type        = "ingress"
      cidr_blocks = var.api_client_cidrs
    }
  }
  access_entries = {
    platform = {
      principal_arn = var.admin_role_arn
      policy_associations = {
        admin = {
          policy_arn   = "arn:${data.aws_partition.current.partition}:eks::aws:cluster-access-policy/AmazonEKSClusterAdminPolicy"
          access_scope = { type = "cluster" }
        }
      }
    }
  }
}

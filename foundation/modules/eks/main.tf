variable "name" { type = string }
variable "vpc_id" { type = string }
variable "subnets" { type = map(list(string)) }
variable "kubernetes_version" { type = string }
variable "clusters" {
  type = object({ hub = optional(bool, false), management = bool, nonPci = bool, pci = bool })
  validation {
    condition     = !(var.clusters.hub && (var.clusters.management || var.clusters.nonPci || var.clusters.pci))
    error_message = "Use o perfil Hub ou os perfis da spoke em cada state."
  }
}
variable "private_endpoint" {
  type = bool
  validation {
    condition     = var.private_endpoint
    error_message = "O perfil de laboratório requer endpoint EKS privado."
  }
}
variable "api_client_cidrs" {
  description = "CIDRs privados autorizados a acessar a API via HTTPS a partir da rede de gestão."
  type        = list(string)
}
variable "admin_role_arn" { type = string }
data "aws_partition" "current" {}
locals {
  enabled = { for k, v in { hub = var.clusters.hub, management = var.clusters.management, nonpci = var.clusters.nonPci, pci = var.clusters.pci } : k => v if v }
}
module "cluster" {
  for_each                        = local.enabled
  source                          = "terraform-aws-modules/eks/aws"
  version                         = "20.37.2"
  cluster_name                    = "${var.name}-${each.key}"
  cluster_version                 = var.kubernetes_version
  vpc_id                          = var.vpc_id
  subnet_ids                      = var.subnets[each.key == "pci" ? "web-pci" : "web-nonpci"]
  cluster_endpoint_private_access = var.private_endpoint
  cluster_endpoint_public_access  = false
  cluster_compute_config          = { enabled = true, node_pools = ["system", "general-purpose"] }
  authentication_mode             = "API"
  cluster_security_group_additional_rules = length(var.api_client_cidrs) == 0 ? {} : {
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
output "clusters" {
  value = { for k, cluster in module.cluster : k => { name = cluster.cluster_name, endpoint = cluster.cluster_endpoint } }
}

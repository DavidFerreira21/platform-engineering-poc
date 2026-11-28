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

variable "api_public_access_cidrs" {
  description = "CIDRs autorizados no endpoint público do EKS. Uma lista vazia mantém o endpoint público desabilitado."
  type        = list(string)
  default     = []

  validation {
    condition = alltrue([
      for cidr in var.api_public_access_cidrs :
      try(cidrhost(cidr, 0) != "" && tonumber(split("/", cidr)[1]) > 0, false)
    ])
    error_message = "Informe CIDRs válidos e específicos para o endpoint público do EKS."
  }
}

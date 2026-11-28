variable "cluster_name" {
  description = "Nome do cluster EKS que recebe a capability."
  type        = string
}

variable "identity_center" {
  description = "Instância IAM Identity Center existente e grupos autorizados para o Argo CD."
  type = object({
    instance_arn = string
    region       = string
    groups       = map(set(string))
  })
  validation {
    condition = (
      length(lookup(var.identity_center.groups, "ADMIN", toset([]))) > 0 &&
      alltrue([for role, ids in var.identity_center.groups : contains(["ADMIN", "EDITOR", "VIEWER"], role) && length(ids) > 0])
    )
    error_message = "Configure ao menos um grupo ADMIN. Os papéis aceitos são ADMIN, EDITOR e VIEWER, com grupos não vazios."
  }
}

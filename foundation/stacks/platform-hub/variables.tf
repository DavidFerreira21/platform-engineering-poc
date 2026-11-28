variable "account_id" {
  description = "ID esperado da conta de plataforma para limitar o provider. Se omitido, usa a identidade ativa."
  type        = string
  default     = null
}

variable "region" {
  description = "Região do Hub. Se omitida, usa o ambiente ou o perfil AWS."
  type        = string
  default     = null
}

variable "azs" {
  description = "Três AZs disponíveis na conta e região. Preserve a ordem após o primeiro apply."
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b", "us-east-1c"]
}

variable "vpc_cidr" {
  description = "CIDR IPv4 da VPC do Hub. Conferir com o inventário de rede da POC."
  type        = string
  default     = "10.80.0.0/16"
}

variable "kubernetes_version" {
  description = "Versão Kubernetes em suporte padrão na região do Hub."
  type        = string
  default     = null
}

variable "api_client_cidrs" {
  description = "CIDRs da rede de gestão autorizada a chegar ao endpoint privado do EKS."
  type        = list(string)
  default     = []
}

variable "admin_role_arn" {
  description = "ARN IAM da role administrativa do EKS na conta de plataforma."
  type        = string
  default     = null
}

variable "identity_center" {
  description = "Instância IAM Identity Center e grupos autorizados para o Argo CD."
  type = object({
    instance_arn = string
    region       = string
    groups       = map(set(string))
  })
  default = null
}

variable "executor_state_bucket_name" {
  description = "Nome globalmente único do bucket de state das execuções autorizadas do TerraformStack."
  type        = string
  default     = null
}

variable "api_public_access_cidrs" {
  description = "CIDRs autorizados a acessar a API do Hub pela internet. Uma lista vazia desabilita o endpoint público."
  type        = list(string)
  default     = []
}

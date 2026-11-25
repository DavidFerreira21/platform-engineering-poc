variable "config_root" {
  description = "Caminho absoluto do diretório que contém configurações globais, por Entity e por Environment."
  type        = string
}

variable "stack" {
  description = "Nome da stack que seleciona o arquivo YAML, ex.: network ou eks."
  type        = string

  validation {
    condition     = can(regex("^[a-z][a-z0-9_-]*$", var.stack))
    error_message = "stack deve ser um nome simples, sem diretórios."
  }
}

variable "entity" {
  description = "Entity autorizada, ex.: entity-a."
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9][a-z0-9-]*$", var.entity))
    error_message = "entity não pode conter caminhos relativos ou caracteres especiais."
  }
}

variable "environment" {
  description = "Environment autorizado, ex.: uat."
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9][a-z0-9-]*$", var.environment))
    error_message = "environment não pode conter caminhos relativos ou caracteres especiais."
  }
}

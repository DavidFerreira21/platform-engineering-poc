variable "config_root" {
  description = "Caminho absoluto do diretório que contém configurações globais, por Entity e por Environment."
  type        = string
}

variable "stack" {
  description = "Nome base do arquivo YAML, com a mesma capitalização do arquivo."
  type        = string

  validation {
    condition     = can(regex("^[A-Za-z][A-Za-z0-9_-]*$", var.stack))
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

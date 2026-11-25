output "values" {
  description = "Configuração composta da stack; não usar para dados secretos."
  value       = local.merged
}

output "source_paths" {
  description = "Arquivos de configuração encontrados, na ordem de precedência."
  value       = [for path in local.paths : path if fileexists(path)]
}

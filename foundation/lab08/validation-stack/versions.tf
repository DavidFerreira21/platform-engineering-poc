terraform {
  required_version = ">= 1.10.0"
  # O executor injeta o backend S3 em tempo de execução; testes locais ficam sem AWS.
}

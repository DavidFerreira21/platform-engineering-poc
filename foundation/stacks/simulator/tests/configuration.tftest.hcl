mock_provider "kubernetes" {}

variables {
  slot        = "simulator-network"
  namespace   = "terraformstack-simulator"
  config_root = "../../config"
}

run "composes_network_configuration" {
  command = plan
  assert {
    condition     = output.network.vpc.cidr == "10.81.0.0/16" && output.network.vpc.dnsHostnames && output.network.vpc.dnsSupport
    error_message = "A configuração do Simulator deve combinar CIDR de UAT com o DNS do baseline."
  }
  assert {
    condition     = output.network.nat.enabled && output.network.endpoints.s3.enabled && !output.network.endpoints.ecr.enabled
    error_message = "As flags herdadas de NAT e endpoints devem preservar seus valores booleanos."
  }
}

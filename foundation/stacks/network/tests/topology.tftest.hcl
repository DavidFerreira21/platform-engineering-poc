mock_provider "aws" {
  mock_data "aws_region" {
    defaults = { name = "us-east-1" }
  }
}
variables {
  entity      = "entity-a"
  environment = "uat"
  region      = "us-east-1"
  account_id  = "000000000000"
  azs         = ["us-east-1a", "us-east-1b", "us-east-1c"]
}
run "uat_topology" {
  command = plan
  assert {
    condition     = length(output.subnets) == 4 && alltrue([for ids in values(output.subnets) : length(ids) == 3])
    error_message = "Esperadas 12 subnets privadas, em quatro grupos de três."
  }
  assert {
    condition     = length(keys(output.bridge_parameters)) == 5
    error_message = "UAT deve exportar os cinco parâmetros Network."
  }
}
run "hub_bridge_disabled" {
  command = plan
  variables {
    entity      = "platform"
    environment = "hub"
  }
  assert {
    condition     = length(keys(output.bridge_parameters)) == 0
    error_message = "Hub deve manter a exportação desligada."
  }
}

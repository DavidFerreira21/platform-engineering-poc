variables {
  config_root = "tests/fixtures"
  stack       = "EC2"
  entity      = "entity-a"
  environment = "uat"
}

run "merges_any_yaml_by_precedence" {
  command = plan

  assert {
    condition = (
      output.values.ec2.instance_type == "t3.small" &&
      output.values.ec2.network.endpoints.private.port == 9443 &&
      output.values.ec2.network.endpoints.private.dns == true &&
      output.values.ec2.network.ipv6 == true
    )
    error_message = "O valor mais específico deve prevalecer e preservar os demais campos aninhados."
  }

  assert {
    condition = (
      output.values.ec2.enabled == false &&
      output.values.ec2.replicas == 0 &&
      output.values.ec2.note == null &&
      length(output.values.ec2.empty_list) == 0 &&
      output.values.ec2.network.cidr_selection == ["10.82.0.0/16"] &&
      length(output.values.ec2.network.cidrs) == 1 &&
      output.values.ec2.network.cidrs[0] == "10.82.0.0/16" &&
      length(output.values.ec2.network.routes) == 1 &&
      output.values.ec2.network.routes[0].destination == "10.82.0.0/16" &&
      output.values.ec2.network.routes[0].enabled == false
    )
    error_message = "Booleanos, números e null mantêm seus tipos, e listas são substituídas por inteiro."
  }

  assert {
    condition = output.source_paths == [
      "tests/fixtures/EC2.yaml",
      "tests/fixtures/entity-a/EC2.yaml",
      "tests/fixtures/entity-a/uat/EC2.yaml",
    ]
    error_message = "Os arquivos encontrados devem aparecer em ordem de precedência."
  }
}

run "skips_missing_environment_file" {
  command = plan

  variables {
    environment = "prod"
  }

  assert {
    condition = (
      length(output.source_paths) == 2 &&
      output.values.ec2.network.endpoints.private.port == 8443 &&
      length(output.values.ec2.network.cidrs) == 2
    )
    error_message = "A ausência de override do ambiente deve preservar os outros níveis."
  }
}

run "returns_empty_object_without_files" {
  command = plan

  variables {
    stack = "missing"
  }

  assert {
    condition     = length(keys(output.values)) == 0 && length(output.source_paths) == 0
    error_message = "Sem arquivos, o módulo deve retornar um objeto vazio."
  }
}

output "aws_target" {
  value = {
    account_id = data.aws_caller_identity.current.account_id
    region     = data.aws_region.current.region
  }
}

output "vpc_id" {
  value = module.network.vpc_id
}

output "subnets" {
  value = module.network.subnets
}

output "cluster" {
  value = merge(module.eks.clusters["hub"], { kubernetes_version = local.kubernetes_version })
}

output "argocd" {
  value = var.identity_center == null ? null : {
    capability_arn      = module.argocd[0].capability_arn
    capability_role_arn = module.argocd[0].capability_role_arn
    server_url          = module.argocd[0].server_url
  }
}

output "executor_backend" {
  description = "Backend S3 disponível para o cadastro de destinos do TerraformStack."
  value = {
    bucket       = aws_s3_bucket.state.id
    region       = data.aws_region.current.region
    use_lockfile = true
  }
}

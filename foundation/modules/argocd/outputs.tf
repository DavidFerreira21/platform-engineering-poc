output "capability_arn" {
  value = aws_eks_capability.argocd.arn
}

output "capability_role_arn" {
  value = aws_iam_role.capability.arn
}

output "server_url" {
  value = aws_eks_capability.argocd.configuration[0].argo_cd[0].server_url
}

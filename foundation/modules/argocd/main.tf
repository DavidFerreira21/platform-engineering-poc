resource "aws_iam_role" "capability" {
  name_prefix        = "platform-hub-argocd-"
  assume_role_policy = data.aws_iam_policy_document.capability_trust.json
}

resource "aws_eks_capability" "argocd" {
  cluster_name              = var.cluster_name
  capability_name           = "argocd"
  type                      = "ARGOCD"
  role_arn                  = aws_iam_role.capability.arn
  delete_propagation_policy = "RETAIN"

  configuration {
    argo_cd {
      namespace = "argocd"

      aws_idc {
        idc_instance_arn = var.identity_center.instance_arn
        idc_region       = var.identity_center.region
      }

      dynamic "rbac_role_mapping" {
        for_each = var.identity_center.groups
        content {
          role = rbac_role_mapping.key

          dynamic "identity" {
            for_each = rbac_role_mapping.value
            content {
              id   = identity.value
              type = "SSO_GROUP"
            }
          }
        }
      }
    }
  }
}

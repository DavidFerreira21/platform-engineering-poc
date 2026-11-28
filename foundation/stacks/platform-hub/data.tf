data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

data "aws_iam_session_context" "current" {
  count = var.admin_role_arn == null ? 1 : 0
  arn   = data.aws_caller_identity.current.arn
}

data "aws_eks_cluster_versions" "standard" {
  count          = var.kubernetes_version == null ? 1 : 0
  cluster_type   = "eks"
  version_status = "STANDARD_SUPPORT"
}

data "aws_iam_policy_document" "state" {
  statement {
    sid     = "DenyInsecureTransport"
    effect  = "Deny"
    actions = ["s3:*"]

    resources = [
      aws_s3_bucket.state.arn,
      "${aws_s3_bucket.state.arn}/*",
    ]

    principals {
      type        = "*"
      identifiers = ["*"]
    }

    condition {
      test     = "Bool"
      variable = "aws:SecureTransport"
      values   = ["false"]
    }
  }
}

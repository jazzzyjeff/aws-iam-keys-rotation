data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

data "aws_iam_policy_document" "role" {
  version = "2012-10-17"

  statement {
    effect = "Allow"
    actions = [
      "ssm:GetParameter",
      "iam:ListAccessKeys",
      "iam:CreateAccessKey",
      "iam:DeleteAccessKey",
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:ssm:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:parameter${var.ado.user}",
      "arn:${data.aws_partition.current.partition}:ssm:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:parameter${var.ado.token}",
      "arn:${data.aws_partition.current.partition}:ssm:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:parameter${var.discord_webhook_url}",
      "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:user/${var.iam_user_name}"
    ]
  }
}

data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

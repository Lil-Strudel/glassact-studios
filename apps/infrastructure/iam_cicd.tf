resource "aws_iam_openid_connect_provider" "github" {
  url            = "https://token.actions.githubusercontent.com"
  client_id_list = ["sts.amazonaws.com"]
  thumbprint_list = [
    "6938fd4d98bab03faadb97b34396831e3780aea1",
    "1c58a3a8518e8759bf075b76b750d4f2df264fcd"
  ]
}

resource "aws_iam_role" "cicd" {
  name = "glassact-cicd"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Federated = aws_iam_openid_connect_provider.github.arn }
      Action    = "sts:AssumeRoleWithWebIdentity"
      Condition = {
        StringEquals = {
          "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
        }
        StringLike = {
          "token.actions.githubusercontent.com:sub" = [
            "repo:${var.github_repo}:ref:refs/heads/main",
            "repo:${var.github_repo}:environment:production",
          ]
        }
      }
    }]
  })
}

data "aws_iam_policy_document" "cicd" {
  statement {
    effect  = "Allow"
    actions = ["s3:PutObject", "s3:DeleteObject", "s3:GetObject", "s3:ListBucket"]
    resources = [
      aws_s3_bucket.landing.arn,
      "${aws_s3_bucket.landing.arn}/*",
      aws_s3_bucket.webapp.arn,
      "${aws_s3_bucket.webapp.arn}/*",
    ]
  }
  statement {
    effect  = "Allow"
    actions = ["cloudfront:CreateInvalidation"]
    resources = [
      aws_cloudfront_distribution.landing.arn,
      aws_cloudfront_distribution.webapp.arn,
    ]
  }
  statement {
    effect    = "Allow"
    actions   = ["ecr:GetAuthorizationToken"]
    resources = ["*"]
  }
  statement {
    effect = "Allow"
    actions = [
      "ecr:BatchCheckLayerAvailability",
      "ecr:PutImage",
      "ecr:InitiateLayerUpload",
      "ecr:UploadLayerPart",
      "ecr:CompleteLayerUpload",
      "ecr:BatchGetImage",
    ]
    resources = [aws_ecr_repository.api.arn, aws_ecr_repository.migrate.arn]
  }
  statement {
    effect  = "Allow"
    actions = ["ssm:SendCommand"]
    resources = [
      aws_instance.api.arn,
      "arn:aws:ssm:us-west-2::document/AWS-RunShellScript",
    ]
  }
  statement {
    effect    = "Allow"
    actions   = ["ssm:GetCommandInvocation", "ssm:ListCommandInvocations"]
    resources = ["*"] # these calls don't support resource-level scoping
  }
  statement {
    effect  = "Allow"
    actions = ["s3:PutObject", "s3:GetObject", "s3:ListBucket"]
    resources = [
      aws_s3_bucket.backups.arn,
      "${aws_s3_bucket.backups.arn}/deploy/*",
    ]
  }
}

resource "aws_iam_role_policy" "cicd" {
  name   = "cicd-deploy"
  role   = aws_iam_role.cicd.id
  policy = data.aws_iam_policy_document.cicd.json
}

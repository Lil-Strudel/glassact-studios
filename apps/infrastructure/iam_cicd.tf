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
          "token.actions.githubusercontent.com:sub" = "repo:${var.github_repo}:ref:refs/heads/main"
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
    actions   = ["lambda:UpdateFunctionCode"]
    resources = [aws_lambda_function.api.arn]
  }
}

resource "aws_iam_role_policy" "cicd" {
  name   = "cicd-deploy"
  role   = aws_iam_role.cicd.id
  policy = data.aws_iam_policy_document.cicd.json
}

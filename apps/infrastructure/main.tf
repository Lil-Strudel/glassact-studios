terraform {
  backend "s3" {
    bucket       = "tf-state-x3cn68j"
    key          = "terraform.tfstate"
    region       = "us-west-2"
    use_lockfile = true
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.27.0"
    }
  }
}

provider "aws" {
  region = "us-west-2"
}

resource "aws_s3_bucket" "state_bucket" {
  bucket = "tf-state-x3cn68j"
}

resource "aws_s3_bucket" "file_bucket" {
  bucket = "glassact-nkm52j"
}

resource "aws_iam_user" "dev" {
  name = "strudel-dev"
}

data "aws_iam_policy_document" "dev_s3" {
  statement {
    effect  = "Allow"
    actions = ["s3:*"]
    resources = [
      aws_s3_bucket.file_bucket.arn,
      "${aws_s3_bucket.file_bucket.arn}/*"
    ]
  }
}

resource "aws_iam_user_policy" "dev_s3" {
  name   = "dev-s3"
  user   = aws_iam_user.dev.name
  policy = data.aws_iam_policy_document.dev_s3.json
}

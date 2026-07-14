data "aws_caller_identity" "current" {}

# ---- Default VPC (no dedicated networking resources) ----
data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

data "aws_subnet" "selected" {
  id = data.aws_subnets.default.ids[0]
}

data "aws_ami" "al2023_arm64" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-*-kernel-*-arm64"]
  }
  filter {
    name   = "architecture"
    values = ["arm64"]
  }
  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

data "aws_ec2_managed_prefix_list" "cloudfront_origin_facing" {
  name = "com.amazonaws.global.cloudfront.origin-facing"
}

# ---- Security group: no SSH, CloudFront-only ingress on the API port ----
resource "aws_security_group" "api_ec2" {
  name        = "glassact-api-ec2"
  description = "GlassAct API host - CloudFront origin-facing ingress only, no SSH"
  vpc_id      = data.aws_vpc.default.id

  ingress {
    description     = "CloudFront origin-facing to API"
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    prefix_list_ids = [data.aws_ec2_managed_prefix_list.cloudfront_origin_facing.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# ---- IAM role for the instance (SSM + ECR pull + S3 + SSM Parameter Store) ----
resource "aws_iam_role" "ec2_api" {
  name = "glassact-ec2-api"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "ec2_ssm" {
  role       = aws_iam_role.ec2_api.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

data "aws_iam_policy_document" "ec2_api" {
  statement {
    effect    = "Allow"
    actions   = ["ecr:GetAuthorizationToken"]
    resources = ["*"]
  }
  statement {
    effect = "Allow"
    actions = [
      "ecr:BatchGetImage",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchCheckLayerAvailability",
    ]
    resources = [
      aws_ecr_repository.api.arn,
      aws_ecr_repository.migrate.arn,
    ]
  }
  statement {
    effect  = "Allow"
    actions = ["s3:GetObject", "s3:PutObject", "s3:DeleteObject", "s3:ListBucket"]
    resources = [
      aws_s3_bucket.file_bucket.arn,
      "${aws_s3_bucket.file_bucket.arn}/*",
    ]
  }
  statement {
    effect  = "Allow"
    actions = ["s3:GetObject", "s3:ListBucket"]
    resources = [
      aws_s3_bucket.backups.arn,
      "${aws_s3_bucket.backups.arn}/deploy/*",
    ]
  }
  statement {
    effect    = "Allow"
    actions   = ["s3:PutObject"]
    resources = ["${aws_s3_bucket.backups.arn}/postgres/*"]
  }
  statement {
    effect  = "Allow"
    actions = ["ssm:GetParameter", "ssm:GetParameters", "ssm:GetParametersByPath"]
    resources = [
      "arn:aws:ssm:us-west-2:${data.aws_caller_identity.current.account_id}:parameter/glassact/api/*",
    ]
  }
  statement {
    effect    = "Allow"
    actions   = ["kms:Decrypt"]
    resources = ["arn:aws:kms:us-west-2:${data.aws_caller_identity.current.account_id}:alias/aws/ssm"]
  }
}

resource "aws_iam_role_policy" "ec2_api" {
  name   = "glassact-ec2-api"
  role   = aws_iam_role.ec2_api.id
  policy = data.aws_iam_policy_document.ec2_api.json
}

resource "aws_iam_instance_profile" "ec2_api" {
  name = "glassact-ec2-api"
  role = aws_iam_role.ec2_api.name
}

# ---- EBS volume for Postgres data (survives instance replacement) ----
resource "aws_ebs_volume" "postgres_data" {
  availability_zone = data.aws_subnet.selected.availability_zone
  size              = 20
  type              = "gp3"
  encrypted         = true

  tags = { Name = "glassact-postgres-data" }

  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_instance" "api" {
  ami                         = data.aws_ami.al2023_arm64.id
  instance_type               = "t4g.small"
  subnet_id                   = data.aws_subnet.selected.id
  vpc_security_group_ids      = [aws_security_group.api_ec2.id]
  iam_instance_profile        = aws_iam_instance_profile.ec2_api.name
  associate_public_ip_address = true

  root_block_device {
    volume_size = 30
    volume_type = "gp3"
  }

  user_data = templatefile("${path.module}/ec2/user_data.sh.tftpl", {
    backup_bucket       = aws_s3_bucket.backups.bucket
    backup_service_unit = file("${path.module}/ec2/glassact-backup.service")
    backup_timer_unit   = file("${path.module}/ec2/glassact-backup.timer")
  })

  tags = { Name = "glassact-api" }

  lifecycle {
    ignore_changes = [ami] # roll the AMI deliberately, not on every plan
  }
}

resource "aws_volume_attachment" "postgres_data" {
  device_name = "/dev/sdf"
  volume_id   = aws_ebs_volume.postgres_data.id
  instance_id = aws_instance.api.id
}

resource "aws_eip" "api" {
  instance = aws_instance.api.id
  domain   = "vpc"
}

# ---- ECR repositories ----
resource "aws_ecr_repository" "api" {
  name                 = "glassact-api"
  image_tag_mutability = "IMMUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_repository" "migrate" {
  name                 = "glassact-migrate"
  image_tag_mutability = "IMMUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_lifecycle_policy" "api" {
  repository = aws_ecr_repository.api.name
  policy = jsonencode({
    rules = [{
      rulePriority = 1
      description  = "keep last 10 images"
      selection    = { tagStatus = "any", countType = "imageCountMoreThan", countNumber = 10 }
      action       = { type = "expire" }
    }]
  })
}

resource "aws_ecr_lifecycle_policy" "migrate" {
  repository = aws_ecr_repository.migrate.name
  policy = jsonencode({
    rules = [{
      rulePriority = 1
      description  = "keep last 10 images"
      selection    = { tagStatus = "any", countType = "imageCountMoreThan", countNumber = 10 }
      action       = { type = "expire" }
    }]
  })
}

# ---- Backups bucket (nightly pg_dump + CI deploy bundle) ----
resource "random_id" "backups_bucket_suffix" {
  byte_length = 3
}

resource "aws_s3_bucket" "backups" {
  bucket = "glassact-backups-${random_id.backups_bucket_suffix.hex}"
}

resource "aws_s3_bucket_public_access_block" "backups" {
  bucket                  = aws_s3_bucket.backups.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_lifecycle_configuration" "backups" {
  bucket = aws_s3_bucket.backups.id

  rule {
    id     = "expire-postgres-backups-30-days"
    status = "Enabled"
    filter { prefix = "postgres/" }
    expiration { days = 30 }
  }
}

# Config/secrets for the API, delivered to the EC2 instance via SSM Parameter
# Store. Real values are set once out-of-band (CLI/console) after `apply` -
# `ignore_changes` mirrors the pattern previously used for the Lambda
# `environment` block, so Terraform never manages plaintext secrets.
locals {
  api_ssm_params = {
    "ENV"                     = { type = "String", value = "production" }
    "PORT"                    = { type = "String", value = "8080" }
    "BASE_URL"                = { type = "String", value = "https://${var.webapp_domain}" }
    "S3_BUCKET_NAME"          = { type = "String", value = aws_s3_bucket.file_bucket.bucket }
    "AWS_REGION"              = { type = "String", value = "us-west-2" }
    "DATABASE_DSN"            = { type = "SecureString", value = "changeme" }
    "AUTH_SECRET"             = { type = "SecureString", value = "changeme" }
    "GOOGLE_CLIENT_ID"        = { type = "SecureString", value = "changeme" }
    "GOOGLE_CLIENT_SECRET"    = { type = "SecureString", value = "changeme" }
    "GOOGLE_REDIRECT_URL"     = { type = "String", value = "changeme" }
    "MICROSOFT_CLIENT_ID"     = { type = "SecureString", value = "changeme" }
    "MICROSOFT_CLIENT_SECRET" = { type = "SecureString", value = "changeme" }
    "MICROSOFT_REDIRECT_URL"  = { type = "String", value = "changeme" }
    "SMTP_HOST"               = { type = "String", value = "changeme" }
    "SMTP_PORT"               = { type = "String", value = "changeme" }
    "SMTP_USERNAME"           = { type = "SecureString", value = "changeme" }
    "SMTP_PASSWORD"           = { type = "SecureString", value = "changeme" }
    # Container init only (docker-compose.yml), not read by the Go app.
    "POSTGRES_PASSWORD" = { type = "SecureString", value = "changeme" }
  }
}

resource "aws_ssm_parameter" "api" {
  for_each = local.api_ssm_params

  name  = "/glassact/api/${each.key}"
  type  = each.value.type
  value = each.value.value

  lifecycle {
    ignore_changes = [value]
  }
}

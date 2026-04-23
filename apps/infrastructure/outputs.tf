output "landing_bucket_name" {
  value = aws_s3_bucket.landing.bucket
}

output "webapp_bucket_name" {
  value = aws_s3_bucket.webapp.bucket
}

output "landing_cloudfront_distribution_id" {
  value = aws_cloudfront_distribution.landing.id
}

output "webapp_cloudfront_distribution_id" {
  value = aws_cloudfront_distribution.webapp.id
}

output "lambda_function_name" {
  value = aws_lambda_function.api.function_name
}

output "cicd_role_arn" {
  value = aws_iam_role.cicd.arn
}

output "landing_cert_validation_records" {
  description = "Add these CNAMEs at your DNS registrar to validate the landing ACM cert"
  value = {
    for dvo in aws_acm_certificate.landing.domain_validation_options : dvo.domain_name => {
      name  = dvo.resource_record_name
      type  = dvo.resource_record_type
      value = dvo.resource_record_value
    }
  }
}

output "webapp_cert_validation_records" {
  description = "Add these CNAMEs at your DNS registrar to validate the webapp ACM cert"
  value = {
    for dvo in aws_acm_certificate.webapp.domain_validation_options : dvo.domain_name => {
      name  = dvo.resource_record_name
      type  = dvo.resource_record_type
      value = dvo.resource_record_value
    }
  }
}

output "landing_cloudfront_domain" {
  description = "Point glassactstudios.com CNAME to this value"
  value       = aws_cloudfront_distribution.landing.domain_name
}

output "webapp_cloudfront_domain" {
  description = "Point app.glassactstudios.com CNAME to this value"
  value       = aws_cloudfront_distribution.webapp.domain_name
}

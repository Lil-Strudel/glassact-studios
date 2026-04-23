resource "aws_acm_certificate" "landing" {
  provider          = aws.us_east_1
  domain_name       = var.landing_domain
  validation_method = "DNS"
  lifecycle { create_before_destroy = true }
}

resource "aws_acm_certificate" "webapp" {
  provider          = aws.us_east_1
  domain_name       = var.webapp_domain
  validation_method = "DNS"
  lifecycle { create_before_destroy = true }
}

resource "aws_acm_certificate_validation" "landing" {
  provider        = aws.us_east_1
  certificate_arn = aws_acm_certificate.landing.arn
}

resource "aws_acm_certificate_validation" "webapp" {
  provider        = aws.us_east_1
  certificate_arn = aws_acm_certificate.webapp.arn
}

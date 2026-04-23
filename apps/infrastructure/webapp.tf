resource "aws_s3_bucket" "webapp" {
  bucket = "glassact-webapp-${random_id.webapp_bucket_suffix.hex}"
}

resource "aws_s3_bucket_public_access_block" "webapp" {
  bucket                  = aws_s3_bucket.webapp.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_cloudfront_origin_access_control" "webapp" {
  name                              = "glassact-webapp-oac"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

resource "aws_s3_bucket_policy" "webapp" {
  bucket = aws_s3_bucket.webapp.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid    = "AllowCloudFrontOAC"
      Effect = "Allow"
      Principal = {
        Service = "cloudfront.amazonaws.com"
      }
      Action   = "s3:GetObject"
      Resource = "${aws_s3_bucket.webapp.arn}/*"
      Condition = {
        StringEquals = {
          "AWS:SourceArn" = aws_cloudfront_distribution.webapp.arn
        }
      }
    }]
  })
}

resource "aws_lambda_function" "api" {
  function_name = "glassact-api"
  role          = aws_iam_role.lambda_exec.arn
  runtime       = "provided.al2023"
  architectures = ["arm64"]
  handler       = "bootstrap"
  filename      = "${path.module}/lambda_placeholder.zip"

  environment {
    variables = {
      ENV            = "production"
      BASE_URL       = "https://${var.webapp_domain}"
      S3_BUCKET_NAME = aws_s3_bucket.file_bucket.bucket
    }
  }

  lifecycle {
    ignore_changes = [filename, source_code_hash, environment]
  }
}

resource "aws_lambda_function_url" "api" {
  function_name      = aws_lambda_function.api.function_name
  authorization_type = "NONE"
}

resource "aws_cloudfront_origin_access_control" "lambda" {
  name                              = "glassact-lambda-oac"
  origin_access_control_origin_type = "lambda"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

resource "aws_cloudfront_response_headers_policy" "no_cache" {
  name = "glassact-no-cache"

  custom_headers_config {
    items {
      header   = "Cache-Control"
      value    = "no-store"
      override = true
    }
  }
}

resource "aws_cloudfront_distribution" "webapp" {
  enabled             = true
  aliases             = [var.webapp_domain]
  default_root_object = "index.html"

  origin {
    domain_name              = aws_s3_bucket.webapp.bucket_regional_domain_name
    origin_id                = "webapp-s3"
    origin_access_control_id = aws_cloudfront_origin_access_control.webapp.id
  }

  origin {
    domain_name              = replace(replace(aws_lambda_function_url.api.function_url, "https://", ""), "/", "")
    origin_id                = "api-lambda"
    origin_access_control_id = aws_cloudfront_origin_access_control.lambda.id

    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  default_cache_behavior {
    target_origin_id       = "webapp-s3"
    viewer_protocol_policy = "redirect-to-https"
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    compress               = true
    default_ttl            = 300
    max_ttl                = 900

    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }
  }

  ordered_cache_behavior {
    path_pattern           = "/index.html"
    target_origin_id       = "webapp-s3"
    viewer_protocol_policy = "redirect-to-https"
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    compress               = true
    min_ttl                = 0
    default_ttl            = 0
    max_ttl                = 0

    response_headers_policy_id = aws_cloudfront_response_headers_policy.no_cache.id

    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }
  }

  ordered_cache_behavior {
    path_pattern           = "/api/*"
    target_origin_id       = "api-lambda"
    viewer_protocol_policy = "redirect-to-https"
    allowed_methods        = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods         = ["GET", "HEAD"]
    compress               = true

    cache_policy_id          = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad"
    origin_request_policy_id = "b689b0a8-53d0-40ab-baf2-68738e2966ac"
  }

  ordered_cache_behavior {
    path_pattern           = "/file/*"
    target_origin_id       = "api-lambda"
    viewer_protocol_policy = "redirect-to-https"
    allowed_methods        = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods         = ["GET", "HEAD"]
    compress               = true

    cache_policy_id          = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad"
    origin_request_policy_id = "b689b0a8-53d0-40ab-baf2-68738e2966ac"
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn      = aws_acm_certificate_validation.webapp.certificate_arn
    ssl_support_method       = "sni-only"
    minimum_protocol_version = "TLSv1.2_2021"
  }
}

resource "aws_lambda_permission" "cloudfront" {
  statement_id  = "AllowCloudFrontInvoke"
  action        = "lambda:InvokeFunctionUrl"
  function_name = aws_lambda_function.api.function_name
  principal     = "cloudfront.amazonaws.com"
  source_arn    = aws_cloudfront_distribution.webapp.arn
}

resource "random_id" "immich_backups_bucket_suffix" {
  byte_length = 3
}

resource "aws_s3_bucket" "immich_backups" {
  bucket = "immich-backups-${random_id.immich_backups_bucket_suffix.hex}"
}

resource "aws_s3_bucket_public_access_block" "immich_backups" {
  bucket                  = aws_s3_bucket.immich_backups.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_versioning" "immich_backups" {
  bucket = aws_s3_bucket.immich_backups.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "immich_backups" {
  bucket = aws_s3_bucket.immich_backups.id

  depends_on = [aws_s3_bucket_versioning.immich_backups]

  rule {
    id     = "whole-bucket-cleanup"
    status = "Enabled"
    filter {}

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
    noncurrent_version_expiration {
      noncurrent_days = 180
    }
    expiration {
      expired_object_delete_marker = true
    }
  }

  rule {
    id     = "expire-db-dumps-180-days"
    status = "Enabled"
    filter { prefix = "backups/" }

    expiration {
      days = 180
    }
    # Deep Archive bills 180 days minimum; a dump is already that old when it turns noncurrent.
    noncurrent_version_expiration {
      noncurrent_days = 1
    }
  }
}

resource "aws_iam_user" "immich_backup" {
  name = "immich-backup"
}

# No DeleteObject: a compromised Immich server must not be able to wipe its backups.
data "aws_iam_policy_document" "immich_backup" {
  statement {
    effect    = "Allow"
    actions   = ["s3:ListBucket", "s3:ListBucketMultipartUploads"]
    resources = [aws_s3_bucket.immich_backups.arn]
  }
  statement {
    effect = "Allow"
    actions = [
      "s3:PutObject",
      "s3:GetObject",
      "s3:GetObjectVersion",
      "s3:RestoreObject",
      "s3:AbortMultipartUpload",
      "s3:ListMultipartUploadParts",
    ]
    resources = ["${aws_s3_bucket.immich_backups.arn}/*"]
  }
}

resource "aws_iam_user_policy" "immich_backup" {
  name   = "immich-backup-s3"
  user   = aws_iam_user.immich_backup.name
  policy = data.aws_iam_policy_document.immich_backup.json
}

resource "aws_s3_bucket" "nlb_logs" {
  bucket = var.bucket_name
}

resource "aws_s3_bucket_lifecycle_configuration" "bucket-config" {
  bucket = aws_s3_bucket.nlb_logs.id

  rule {
    id = "short_retention_policy"

    expiration {
      days = 4
    }
    status = "Enabled"
  }
}

resource "aws_s3_bucket_public_access_block" "nlb_logs" {
  bucket = aws_s3_bucket.nlb_logs.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}


resource "aws_s3_bucket_server_side_encryption_configuration" "nlb_logs_default_encryption" {
  bucket = aws_s3_bucket.nlb_logs.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
    bucket_key_enabled = true
  }
}


resource "aws_s3_bucket_policy" "nlb_logs_policy" {
  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Principal": {
            "Service": "delivery.logs.amazonaws.com"
        },
        "Action": [
            "s3:PutObject"
        ],
        "Resource": [
            "arn:aws:s3:::${aws_s3_bucket.nlb_logs.id}/${var.bucket_prefix}/AWSLogs/${var.aws_account_id}/*"
        ]
      }
    ]
}
POLICY
  bucket = aws_s3_bucket.nlb_logs.bucket
}

resource "aws_s3_bucket_notification" "nlb_logs_bucket_notification" {
  bucket = aws_s3_bucket.nlb_logs.id

  queue {
    queue_arn = aws_sqs_queue.nlb_logs_queue.arn
    events    = ["s3:ObjectCreated:*"]
  }
}

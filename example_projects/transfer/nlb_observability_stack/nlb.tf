module "nlb_module" {
  source = "../../../modules/lb/nlb/v1"
  name              = var.nlb_name
  listener_protocol = "TLS"
  target_protocol   = "TCP"
  is_grpc           = false
  fqdn              = var.fqdn

  vpc_id = var.vpc_id
  subnet_ids = var.subnet_ids
  tags = var.tags

  access_logs_bucket = aws_s3_bucket.nlb_logs.bucket
  access_logs_bucket_prefix = var.bucket_prefix
  depends_on = [
    aws_s3_bucket.nlb_logs,
    aws_s3_bucket_policy.nlb_logs_policy
  ]
}

resource "aws_lb" "nlb" {
  name                             = var.name
  internal                         = false
  load_balancer_type               = "network"
  ip_address_type                  = "dualstack"
  subnets                          = var.subnet_ids
  security_groups                  = var.security_group_ids
  enable_cross_zone_load_balancing = true
  tags                             = var.tags
  dynamic "access_logs" {
    for_each = var.access_logs_bucket != null ? { enabled = true } : {}

    content {
      enabled = true
      bucket  = var.access_logs_bucket
      prefix  = var.access_logs_bucket_prefix
    }
  }
}

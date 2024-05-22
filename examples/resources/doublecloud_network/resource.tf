resource "doublecloud_network" "example-network" {
  project_id      = var.project_id
  name            = "example-clickhouse"
  region_id       = "eu-central-1"
  cloud_type      = "aws"
  ipv4_cidr_block = "10.0.0.0/16"
}
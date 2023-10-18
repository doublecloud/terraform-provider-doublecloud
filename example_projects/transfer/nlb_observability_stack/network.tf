resource "doublecloud_network" "nlb-network" {
  project_id = var.project_id
  name = var.network_name
  region_id = "eu-central-1"
  cloud_type = "aws"
  ipv4_cidr_block = "10.0.0.0/16"
}

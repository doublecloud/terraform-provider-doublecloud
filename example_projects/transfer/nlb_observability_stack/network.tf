resource "doublecloud_network" "nlb-network" {
  project_id      = var.project_id
  name            = var.network_name
  region_id       = var.region
  cloud_type      = var.cloud_type
  ipv4_cidr_block = var.ipv4_cidr
}

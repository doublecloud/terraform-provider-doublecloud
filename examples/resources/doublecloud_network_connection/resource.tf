# Create DoubleCloud Network
resource "doublecloud_network" "example" {
  project_id      = var.project_id
  name            = "example"
  region_id       = "eu-central-1"
  cloud_type      = "aws"
  ipv4_cidr_block = "10.42.0.0/16"
}

# Create AWS VPC
resource "aws_vpc" "own" {
  cidr_block                       = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block = true
}

# Get AWS AccountID
data "aws_caller_identity" "self" {}

# Get AWS RegionID
data "aws_region" "current" {}

# Create VPC Peering from DoubleCloud Network to AWS VPC
resource "doublecloud_network_connection" "example" {
  network_id = doublecloud_network.example.id
  aws = {
    peering = {
      vpc_id          = aws_vpc.own.id
      account_id      = data.aws_caller_identity.self.account_id
      region_id       = data.aws_region.current.id
      ipv4_cidr_block = aws_vpc.own.cidr_block
      ipv6_cidr_block = aws_vpc.own.ipv6_cidr_block
    }
  }
}

# Accept Peering Request on AWS side
resource "aws_vpc_peering_connection_accepter" "own" {
  vpc_peering_connection_id = doublecloud_network_connection.example.aws.peering.peering_connection_id
  auto_accept               = true
}

# Confirm Peering creation
resource "doublecloud_network_connection_accepter" "accept" {
  id = doublecloud_network_connection.example.id
}

# Create routes to DoubleCloud Network
resource "aws_route" "ipv4" {
  route_table_id            = aws_vpc.own.main_route_table_id
  destination_cidr_block    = doublecloud_network_connection.example.aws.peering.managed_ipv4_cidr_block
  vpc_peering_connection_id = doublecloud_network_connection.example.aws.peering.peering_connection_id
}

resource "aws_route" "ipv6" {
  route_table_id              = aws_vpc.own.main_route_table_id
  destination_ipv6_cidr_block = doublecloud_network_connection.example.aws.peering.managed_ipv6_cidr_block
  vpc_peering_connection_id   = doublecloud_network_connection.example.aws.peering.peering_connection_id
}

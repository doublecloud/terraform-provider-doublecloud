data "aws_caller_identity" "self" {}

data "aws_region" "self" {}

# Prepare BYOC VPC and IAM Role
module "doublecloud_byoc" {
  source  = "doublecloud/doublecloud-byoc/aws"
  version = "1.0.2"
  providers = {
    aws = aws
  }
  // This is CIDR of newly created Data VPC (where Clickhosue and ELT jobs would be hosted)
  ipv4_cidr = var.dc_ipv4_cidr
}

# Get account ID to peer with
data "aws_caller_identity" "peered" {
  provider = aws
}

# Create DoubleCloud BYOC Network
resource "doublecloud_network" "aws" {
  project_id = var.dc_project_id
  name       = "beta-network"
  region_id  = module.doublecloud_byoc.region_id
  cloud_type = "aws"
  aws = {
    vpc_id                             = module.doublecloud_byoc.vpc_id
    account_id                         = module.doublecloud_byoc.account_id
    iam_role_arn                       = module.doublecloud_byoc.iam_role_arn
    iam_policy_permission_boundary_arn = module.doublecloud_byoc.iam_policy_permission_boundary_arn

    // For sake of simplicity we put it in public subnets
    // so we can access from laptops
    private_subnets = false
  }
}

# Create VPC Peering from DoubleCloud Network to AWS VPC
resource "doublecloud_network_connection" "example" {
  network_id = doublecloud_network.aws.id
  aws = {
    peering = {
      vpc_id     = aws_vpc.docdb_vpc.id
      account_id = data.aws_caller_identity.peered.account_id
      region_id  = var.aws_region
      // This is host VPC. VPC where exist infra located (RDS and EC2 instances).
      // We peer those VPC to gain connectivity between *exist* VPC and *data* VPC
      ipv4_cidr_block = aws_vpc.docdb_vpc.cidr_block
      ipv6_cidr_block = aws_vpc.docdb_vpc.ipv6_cidr_block
    }
  }
}

# Accept Peering Request on AWS side
resource "aws_vpc_peering_connection_accepter" "own" {
  provider                  = aws
  vpc_peering_connection_id = time_sleep.avoid_aws_race.triggers["peering_connection_id"]
  auto_accept               = true
}

# Confirm Peering creation
resource "doublecloud_network_connection_accepter" "accept" {
  id = doublecloud_network_connection.example.id

  depends_on = [
    aws_vpc_peering_connection_accepter.own,
  ]
}

# Create ipv4 routes to DoubleCloud Network
resource "aws_route" "user_to_dc_ipv4_private_route" {
  provider                  = aws
  route_table_id            = aws_route_table.docdb_private_rt.id
  destination_cidr_block    = doublecloud_network_connection.example.aws.peering.managed_ipv4_cidr_block
  vpc_peering_connection_id = time_sleep.avoid_aws_race.triggers["peering_connection_id"]
}
resource "aws_route" "user_to_dc_ipv4_public_route" {
  provider                  = aws
  route_table_id            = aws_route_table.docdb_public_rt.id
  destination_cidr_block    = doublecloud_network_connection.example.aws.peering.managed_ipv4_cidr_block
  vpc_peering_connection_id = time_sleep.avoid_aws_race.triggers["peering_connection_id"]
}

# Sleep to avoid AWS InvalidVpcPeeringConnectionID.NotFound error
resource "time_sleep" "avoid_aws_race" {
  create_duration = "30s"

  triggers = {
    peering_connection_id = doublecloud_network_connection.example.aws.peering.peering_connection_id
  }
}

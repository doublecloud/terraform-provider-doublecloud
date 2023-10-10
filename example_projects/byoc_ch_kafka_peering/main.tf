# Prepare BYOC VPC and IAM Role
module "doublecloud_byoc" {
  source  = "doublecloud/doublecloud-byoc/aws"
  version = "1.0.2"
  providers = {
    aws = aws.byoc
  }
  ipv4_cidr = var.ipv4_cidr
}

# Create VPC to peer with
resource "aws_vpc" "peered" {
  cidr_block                       = var.peered_ipv4_cidr_block
  provider                         = aws.peered
  assign_generated_ipv6_cidr_block = true
}

# Get account ID to peer with
data "aws_caller_identity" "peered" {
  provider = aws.peered
}

# Create DoubleCloud BYOC Network
resource "doublecloud_network" "aws" {
  project_id = var.project_id
  name       = "alpha-network"
  region_id  = module.doublecloud_byoc.region_id
  cloud_type = "aws"
  aws = {
    vpc_id          = module.doublecloud_byoc.vpc_id
    account_id      = module.doublecloud_byoc.account_id
    iam_role_arn    = module.doublecloud_byoc.iam_role_arn
    private_subnets = true
  }
}

# Create VPC Peering from DoubleCloud Network to AWS VPC
resource "doublecloud_network_connection" "example" {
  network_id = doublecloud_network.aws.id
  aws = {
    peering = {
      vpc_id          = aws_vpc.peered.id
      account_id      = data.aws_caller_identity.peered.account_id
      region_id       = var.peered_region_id
      ipv4_cidr_block = aws_vpc.peered.cidr_block
      ipv6_cidr_block = aws_vpc.peered.ipv6_cidr_block
    }
  }
}

# Accept Peering Request on AWS side
resource "aws_vpc_peering_connection_accepter" "own" {
  provider                  = aws.peered
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
resource "aws_route" "ipv4" {
  provider                  = aws.peered
  route_table_id            = aws_vpc.peered.main_route_table_id
  destination_cidr_block    = doublecloud_network_connection.example.aws.peering.managed_ipv4_cidr_block
  vpc_peering_connection_id = time_sleep.avoid_aws_race.triggers["peering_connection_id"]
}

# Create ipv6 routes to DoubleCloud Network
resource "aws_route" "ipv6" {
  provider                    = aws.peered
  route_table_id              = aws_vpc.peered.main_route_table_id
  destination_ipv6_cidr_block = doublecloud_network_connection.example.aws.peering.managed_ipv6_cidr_block
  vpc_peering_connection_id   = time_sleep.avoid_aws_race.triggers["peering_connection_id"]
}

resource "doublecloud_clickhouse_cluster" "alpha-clickhouse" {
  project_id = var.project_id
  name       = "alpha-clickhouse"
  region_id  = var.region_id
  cloud_type = "aws"
  network_id = doublecloud_network.aws.id

  resources {
    clickhouse {
      resource_preset_id = "s1-c2-m4"
      disk_size          = 34359738368
      replica_count      = 1
    }
  }

  config {
    log_level       = "LOG_LEVEL_TRACE"
    max_connections = 120

    kafka {
      security_protocol = "SASL_SSL"
      sasl_mechanism    = "SCRAM_SHA_512"
      sasl_username     = doublecloud_kafka_cluster.alpha-kafka.connection_info.user
      sasl_password     = doublecloud_kafka_cluster.alpha-kafka.connection_info.password
    }
  }

  access {
    ipv4_cidr_blocks = [
      {
        value       = doublecloud_network.aws.ipv4_cidr_block
        description = "DC Network interconnection"
      },
      {
        value       = aws_vpc.peered.cidr_block
        description = "Peered VPC"
      }
    ]
    ipv6_cidr_blocks = [
      {
        value       = doublecloud_network.aws.ipv6_cidr_block
        description = "DC Network interconnection"
      },
      {
        value       = aws_vpc.peered.ipv6_cidr_block
        description = "Peered VPC"
      }
    ]
  }
}

resource "doublecloud_kafka_cluster" "alpha-kafka" {
  project_id = var.project_id
  name       = "alpha-kafka"
  region_id  = var.region_id
  cloud_type = "aws"
  network_id = doublecloud_network.aws.id

  resources {
    kafka {
      resource_preset_id = "s1-c2-m4"
      disk_size          = 34359738368
      broker_count       = 1
      zone_count         = 1
    }
  }

  schema_registry {
    enabled = false
  }

  access {
    ipv4_cidr_blocks = [
      {
        value       = doublecloud_network.aws.ipv4_cidr_block
        description = "DC Network interconnection"
      },
      {
        value       = aws_vpc.peered.cidr_block
        description = "Peered VPC"
      }
    ]
    ipv6_cidr_blocks = [
      {
        value       = doublecloud_network.aws.ipv6_cidr_block
        description = "DC Network interconnection"
      },
      {
        value       = aws_vpc.peered.ipv6_cidr_block
        description = "Peered VPC"
      }
    ]
  }
}

# Sleep to avoid AWS InvalidVpcPeeringConnectionID.NotFound error
resource "time_sleep" "avoid_aws_race" {
  create_duration = "30s"

  triggers = {
    peering_connection_id = doublecloud_network_connection.example.aws.peering.peering_connection_id
  }
}

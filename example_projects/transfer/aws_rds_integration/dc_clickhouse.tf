## Actual Clickhouse DB
resource "doublecloud_clickhouse_cluster" "alpha-clickhouse" {
  project_id = var.dc_project_id
  name       = "alpha-clickhouse"
  region_id  = var.aws_region
  cloud_type = "aws"
  network_id = doublecloud_network.aws.id

  resources {
    clickhouse {
      resource_preset_id = "s2-c2-m4"
      disk_size          = 34359738368 // 32 gb, but in bytes.
      replica_count      = 1
    }
  }

  config {
    log_level       = "LOG_LEVEL_TRACE"
    max_connections = 120
  }

  access {
    // this will add allow for Visualization
    data_services = ["visualization"]
    ipv4_cidr_blocks = [
      {
        // Connectivity within DC VPC
        value       = doublecloud_network.aws.ipv4_cidr_block
        description = "DC Network interconnection"
      },
      {
        // Connectivity with AWS VPC
        value       = aws_vpc.tutorial_vpc.cidr_block
        description = "Peered VPC"
      },
      {
        // My Local IP-v4 address
        value       = "${var.my_ip}/32"
        description = "My IP"
      }
    ]
    ipv6_cidr_blocks = [
      {
        // My Local IP-v6 address
        value       = "${var.my_ipv6}/128"
        description = "My IPv6"
      }
    ]
  }
}

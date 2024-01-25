resource "doublecloud_kafka_cluster" "input-kafka" {
  project_id = var.project_id
  name       = "clickstream-kafka"
  region_id  = var.region_id
  cloud_type = "aws"
  network_id = doublecloud_network.main-network.id

  resources {
    kafka {
      resource_preset_id = "s1-c2-m4"
      disk_size          = 32 * 1024 * 1024 * 1024 // 32 gb of Storage
      broker_count       = 1
      zone_count         = 1
    }
  }

  config {}

  schema_registry {
    enabled = false
  }

  access {
    ipv4_cidr_blocks = [
      {
        value       = "${var.my_ip}/32"
        description = "My IP4 for local access"
      }
    ]
    ipv6_cidr_blocks = [
      {
        value       = "${var.my_ipv6}/128"
        description = "My IP6 for local access"
      }
    ]
  }
}

// Grab newly create Clickhouse data
data "doublecloud_kafka" "input-kafka" {
  name       = doublecloud_kafka_cluster.input-kafka.name
  project_id = var.project_id
}

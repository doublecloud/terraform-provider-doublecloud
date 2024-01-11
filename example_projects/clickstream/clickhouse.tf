resource "doublecloud_clickhouse_cluster" "target-clickhouse" {
  project_id = var.project_id
  name       = "clickstream-clickhouse"
  region_id  = var.region_id
  cloud_type = "aws"
  network_id = doublecloud_network.main-network.id

  resources {
    clickhouse {
      resource_preset_id = "s1-c2-m4"
      disk_size          = 32 * 1024 * 1024 * 1024 // 32 gb of Storage
      replica_count      = 1
    }
  }

  config {
    log_level       = "LOG_LEVEL_TRACE"
    max_connections = 120

    kafka {
      security_protocol = "SASL_SSL"
      sasl_mechanism    = "SCRAM_SHA_512"
      sasl_username = resource.doublecloud_kafka_cluster.input-kafka.connection_info.user
      sasl_password = resource.doublecloud_kafka_cluster.input-kafka.connection_info.password
    }
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

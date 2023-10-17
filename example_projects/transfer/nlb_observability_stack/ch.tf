resource "doublecloud_clickhouse_cluster" "nlb-logs-clickhouse-cluster" {
  project_id = var.project_id
  name       = "nlb-logs-clickhouse-cluster"
  region_id  = "eu-central-1"
  cloud_type = "aws"
  network_id = resource.doublecloud_network.nlb-network.id

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
  }

  access {
    data_services = ["transfer"]
    ipv4_cidr_blocks = [
      {
        value       = "10.0.0.0/8"
        description = "Office in Berlin"
      }
    ]
  }
}


data "doublecloud_clickhouse" "nlb-logs-clickhouse" {
  project_id = var.project_id
  id         = doublecloud_clickhouse_cluster.nlb-logs-clickhouse-cluster.id
  depends_on = [
    resource.doublecloud_clickhouse_cluster.nlb-logs-clickhouse-cluster,
  ]
}



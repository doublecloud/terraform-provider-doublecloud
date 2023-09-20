resource "doublecloud_clickhouse_cluster" "example-clickhouse" {
  project_id = var.project_id
  name = "example-clickhouse"
  region_id = "eu-central-1"
  cloud_type = "aws"
  network_id = data.doublecloud_network.default.id

  resources {
    clickhouse {
      resource_preset_id = "s1-c2-m4"
      disk_size = 34359738368
      replica_count = 1
    }
  }

  config {
    log_level = "LOG_LEVEL_TRACE"
    max_connections = 120
  }
}
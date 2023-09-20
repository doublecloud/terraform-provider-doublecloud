resource "doublecloud_clickhouse_kafka" "example-kafka" {
  project_id = var.project_id
  name = "example-clickhouse"
  region_id = "eu-central-1"
  cloud_type = "aws"
  network_id = data.doublecloud_network.default.id

  resources {
    kafka {
      resource_preset_id = "s1-c2-m4"
      disk_size = 34359738368
      broker_count = 1
      zone_count = 1
    }
  }

  schema_registry {
    enabled = false
  }
}
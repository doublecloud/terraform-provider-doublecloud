data "doublecloud_network" "network" {
  project_id      = var.project_id
  name            = "default in eu-central-1" // network for clickhouse
}

resource "doublecloud_clickhouse_cluster" "backend" {
  project_id = var.project_id
  name       = "beta-clickhouse"
  region_id  = var.region_id
  cloud_type = "aws"
  network_id = data.doublecloud_network.network.id

  resources {
    clickhouse {
      resource_preset_id = var.clickhouse_cluster_resource_preset
      disk_size          = 34359738368
      replica_count      = 1
    }
  }

  config {
    log_level       = "LOG_LEVEL_INFORMATION"
    max_connections = 120
  }

  access {
    data_services = ["transfer"]
    ipv4_cidr_blocks = [
      {
        value       = var.ipv4_cidr
        description = "VPC CIDR"
      }
    ]
  }
}

data "doublecloud_clickhouse" "backend" {
  project_id = var.project_id
  id         = doublecloud_clickhouse_cluster.backend.id
}



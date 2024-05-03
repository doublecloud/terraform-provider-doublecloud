resource "doublecloud_kafka_cluster" "example-kafka" {
  project_id = var.project_id
  name = "example-kafka"
  region_id = "eu-central-1"
  cloud_type = "aws"
  network_id = data.doublecloud_network.default.id

  resources {
    kafka {
      resource_preset_id = "s1-c2-m4"
      disk_size = 34359738368 # 32GB
      broker_count = 1
      zone_count = 1
    }
  }

  schema_registry {
    enabled = false
  }

  access {
    data_services = ["transfer"]
    ipv4_cidr_blocks = [
      {
        value = "10.0.0.0/8"
        description = "Office in Berlin"
      }
	  ]
  }
}

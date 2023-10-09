resource "doublecloud_network" "aws" {
  project_id = var.project_id
  name       = "alpha-network"
  region_id  = module.doublecloud-byoc.region_id
  cloud_type = "aws"
  aws = {
    vpc_id       = module.doublecloud-byoc.vpc_id
    account_id   = module.doublecloud-byoc.account_id
    iam_role_arn = module.doublecloud-byoc.iam_role_arn
  }
}

resource "doublecloud_clickhouse_cluster" "alpha-clickhouse" {
  project_id = var.project_id
  name       = "alpha-clickhouse"
  region_id  = var.region_id
  cloud_type = "aws"
  network_id = resource.doublecloud_network.aws.id

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
      sasl_username = resource.doublecloud_kafka_cluster.alpha-kafka.connection_info.user
      sasl_password = resource.doublecloud_kafka_cluster.alpha-kafka.connection_info.password
    }
  }

  access {
    ipv4_cidr_blocks = [
      {
        value       = var.ipv4_cidr
        description = "Office in Berlin"
      }
    ]
  }

  depends_on = [
    resource.doublecloud_kafka_cluster.alpha-kafka,
    data.doublecloud_kafka.alpha-kafka
  ]
}

resource "doublecloud_kafka_cluster" "alpha-kafka" {
  project_id = var.project_id
  name       = "alpha-kafka"
  region_id  = var.region_id
  cloud_type = "aws"
  network_id = resource.doublecloud_network.aws.id

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
        value       = var.ipv4_cidr
        description = "Office in Berlin"
      },
      {
        value       = "192.168.44.0/24"
        description = "Office in Miami"
      }
    ]
  }
}

data "doublecloud_kafka" "alpha-kafka" {
  id         = resource.doublecloud_kafka_cluster.alpha-kafka.id
  project_id = var.project_id
}

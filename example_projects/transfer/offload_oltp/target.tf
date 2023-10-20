module "byoc" {
  source  = "doublecloud/doublecloud-byoc/aws"
  version = "1.0.3"

  doublecloud_controlplane_account_id = data.aws_caller_identity.self.account_id
  ipv4_cidr                           = "10.10.0.0/16"
}

data "aws_vpc" "infra" {
  id = var.vpc_id
}

resource "doublecloud_network" "network" {
  project_id = var.project_id
  name       = "k8s-logs-store-network"
  region_id  = module.byoc.region_id
  cloud_type = "aws"
  aws = {
    vpc_id       = module.byoc.vpc_id
    account_id   = module.byoc.account_id
    iam_role_arn = module.byoc.iam_role_arn
  }
}

resource "doublecloud_clickhouse_cluster" "dwh" {
  project_id  = var.project_id
  name        = "dwg"
  region_id   = "eu-central-1"
  cloud_type  = "aws"
  network_id  = doublecloud_network.network.id
  description = "Main DWH Cluster"

  resources {
    clickhouse {
      resource_preset_id = "s1-c2-m4"
      disk_size          = 51539607552
      replica_count      = var.is_prod ? 3 : 1 # for prod it's better to be more then 1 replica
      shard_count        = 1
    }

  }

  config {
    log_level      = "LOG_LEVEL_INFO"
    text_log_level = "LOG_LEVEL_INFO"
  }

  access {
    data_services = ["transfer", "visualization"]
    ipv4_cidr_blocks = [{
      value       = data.aws_vpc.infra.cidr_block
      description = "peered-net"
    }]
  }
}

data "doublecloud_clickhouse" "dwh" {
  name       = doublecloud_clickhouse_cluster.dwh.name
  project_id = var.project_id
}

resource "doublecloud_transfer_endpoint" "dwh-target" {
  name = "dwh-target"
  project_id = var.project_id
  settings {
    clickhouse_target {
      connection {
        address {
          cluster_id = doublecloud_clickhouse_cluster.dwh.id
        }
        database = "default"
        user     = data.doublecloud_clickhouse.dwh.connection_info.user
        password = data.doublecloud_clickhouse.dwh.connection_info.password
      }
    }
  }
}

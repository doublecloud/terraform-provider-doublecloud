### Off-load exist OLTP load into Clickhouse for analytics

Let’s say we have a typical web application that lives inside terraform and consist main storages (let it be Postgres).

And what we need to add some analytical capabilities here:

1. Offload oltp-db analytical to different storage
2. Aggregate all data in same place
3. Join it with data outside our scope.

![offload.drawio.svg](offload.drawio.svg)

First of all let’s take a look how to organize code between stages, I do prefer a module + several root here, so we can tweak it a bit easier.

Let’s start with a `main.tf`, usually it contains providers definition, nothing more:

```hcl
provider "doublecloud" {
  authorized_key = file(var.dc-token)
}
provider "aws" {
  profile = var.profile
}
```

This will just enable you usage for certain envs, like here it’s AWS and  [Double.Cloud](https://double.cloud/).

First thing we need to do - is create our storage, for this example we will use [Clickhouse](https://clickhouse.com/).

To enable clickhouse we need to create a network were to put it. For this example I choose [BYOA](https://double.cloud/blog/posts/2022/12/bring-your-own-account/) network, so it can be easily peered with exist infra.

```hcl
module "byoc" {
  source  = "doublecloud/doublecloud-byoc/aws"
  version = "1.0.3"

  ipv4_cidr = "10.10.0.0/16"
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

```

Once we have clickhouse we can start designing our data pipes.

First let`s make postgres-to-clickhouse:

```hcl
resource "doublecloud_transfer_endpoint" "pg-source" {
  name = "sample-pg2ch-source"
  project_id = var.project_id
  settings {
    postgres_source {
      connection {
        on_premise {
          hosts = [
            var.postgres_host
          ]
          port = 5432
        }
      }
      database = var.postgres_database
      user = var.postgres_user
      password = var.postgres_password
    }
  }
}

resource "doublecloud_transfer" "pg2ch" {
  name = "pg2ch"
  project_id = var.project_id
  source = doublecloud_transfer_endpoint.pg-source.id
  target = doublecloud_transfer_endpoint.dwh-target.id
  type = "SNAPSHOT_AND_INCREMENT"
  activated = false
}
```

This creates a simple replication pipeline between your exist postgres and newly created DWH clickhouse cluster.

As you can see a lot of stuff here actually comes as variables, so it’s quite easy to prepare different stages, simple add `stage_name.tfvars` and run `terraform apply` with it:

```hcl
variable "dc-token" {
  description = "Auth token for double cloud, see: https://github.com/doublecloud/terraform-provider-doublecloud"
}
variable "profile" {
  description = "Name of AWS profile"
}
variable "vpc_id" {
  description = "VPC ID of exist infra to peer with"
}
variable "is_prod" {
  description = "Is environment production"
}
variable "project_id" {
  description = "Double.Cloud project ID"
}
variable "postgres_host" {
  description = "Source host"
}
variable "postgres_database" {
  description = "Source database"
}
variable "postgres_user" {
  description = "Source user"
}
variable "postgres_password" {
  description = "Source Password"
}
```

That’s it. Your data stack is ready to consume. As next steps just setup your own visualization connection:

```hcl
resource "doublecloud_workbook" "k8s-logs-viewer" {
  project_id = var.project_id
  title      = "dwh"

  config = jsonencode({
    "datasets" : [],
    "charts" : [],
    "dashboards" : []
  })

  connect {
    name = "main"
    config = jsonencode({
      kind          = "clickhouse"
      cache_ttl_sec = 600
      host          = data.doublecloud_clickhouse.dwh.connection_info.host
      port          = 8443
      username      = data.doublecloud_clickhouse.dwh.connection_info.user
      secure        = true
      raw_sql_level = "off"
    })
    secret = data.doublecloud_clickhouse.dwh.connection_info.password
  }
}
```

Congratulations! You are awesome. Such configuration is really easy to deploy (just run terraform apply) and copy - run it with different variable-set like [this](https://registry.terraform.io/providers/terraform-redhat/rhcs/latest/docs/guides/terraform-vars).

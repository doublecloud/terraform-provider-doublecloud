// AWS RDS Source
resource "doublecloud_transfer_endpoint" "docdb-source" {
  name       = "docdb-source"
  project_id = var.dc_project_id
  settings {
    mongo_source {
      connection {
        on_premise {
          tls_mode {
            // by default we connect via VPC Peering and using TLS encryption, so we must specify ca-cert
            ca_certificate = file("../assets/global-bundle.pem")
          }
          hosts = [aws_docdb_cluster.service.endpoint]
          port  = 27017
        }
        user        = var.db_username
        password    = var.db_password
        auth_source = "admin"
      }
    }
  }
}

// Grab newly create Clickhouse data
data "doublecloud_clickhouse" "dwh" {
  name       = doublecloud_clickhouse_cluster.beta-clickhouse.name
  project_id = var.dc_project_id
}

// Endpoint for Clickhouse
resource "doublecloud_transfer_endpoint" "dwh-target" {
  name       = "beta-clickhouse-target"
  project_id = var.dc_project_id
  settings {
    clickhouse_target {
      connection {
        address {
          cluster_id = doublecloud_clickhouse_cluster.beta-clickhouse.id
        }
        // We use default database, user and password here, for sake of simplicity
        database = "default"
        user     = data.doublecloud_clickhouse.dwh.connection_info.user
        password = data.doublecloud_clickhouse.dwh.connection_info.password
      }
    }
  }
}

// Actual transfer, this will create transfer from RDS to Clickhouse.
resource "doublecloud_transfer" "pg2ch" {
  name       = "docdb-to-clickhouse-replication"
  project_id = var.dc_project_id
  source     = doublecloud_transfer_endpoint.docdb-source.id
  target     = doublecloud_transfer_endpoint.dwh-target.id
  type       = "SNAPSHOT_AND_INCREMENT"
  activated  = false
  runtime = {
    dedicated = {
      flavor = "TINY"
      vpc_id = doublecloud_network.aws.id
    }
  }
}

// AWS RDS Source
resource "doublecloud_transfer_endpoint" "pg-source" {
  name       = "chinook-pg-source"
  project_id = var.dc_project_id
  settings {
    postgres_source {
      connection {
        on_premise {
          tls_mode {
            // by default we connect via VPC Peering and using TLS encryption, so we must specify ca-cert
            ca_certificate = file("global-bundle.pem")
          }
          hosts = [
            // AWS RDS host-name
            aws_db_instance.tutorial_database.address
          ]
          port = 5432
        }
      }
      database = aws_db_instance.tutorial_database.db_name
      user     = aws_db_instance.tutorial_database.username
      password = var.db_password // here we pass it from variables, but can be value from AWS SecretManager
    }
  }
}

// Grab newly create Clickhouse data
data "doublecloud_clickhouse" "dwh" {
  name       = doublecloud_clickhouse_cluster.alpha-clickhouse.name
  project_id = var.dc_project_id
}

// Endpoint for Clickhouse
resource "doublecloud_transfer_endpoint" "dwh-target" {
  name       = "alpha-clickhouse-target"
  project_id = var.dc_project_id
  settings {
    clickhouse_target {
      connection {
        address {
          cluster_id = doublecloud_clickhouse_cluster.alpha-clickhouse.id
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
  name       = "postgres-to-clickhouse-snapshot"
  project_id = var.dc_project_id
  source     = doublecloud_transfer_endpoint.pg-source.id
  target     = doublecloud_transfer_endpoint.dwh-target.id
  type       = "SNAPSHOT_ONLY"
  activated  = false
}

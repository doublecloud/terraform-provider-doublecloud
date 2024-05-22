resource "doublecloud_transfer_endpoint" "sample-pg2ch-source" {
  name       = "sample-pg2ch-source"
  project_id = var.project_id
  settings {
    postgres_source {
      connection {
        on_premise {
          hosts = [
            "mypostgresql.host"
          ]
          port = 5432
        }
      }
      database = "postgres"
      user     = "postgres"
      password = var.postgresql_postgres_password
    }
  }
}

resource "doublecloud_transfer_endpoint" "sample-pg2ch-target" {
  name       = "sample-pg2ch-target"
  project_id = var.project_id
  settings {
    clickhouse_target {
      clickhouse_cleanup_policy = "DROP"
      connection {
        address {
          cluster_id = "chcexampleexampleexa"
        }
        database = "default"
        password = var.clickhouse_admin_password
        user     = "admin"
      }
    }
  }
}

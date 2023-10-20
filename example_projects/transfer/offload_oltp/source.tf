resource "doublecloud_transfer_endpoint" "pg-source" {
  name       = "sample-pg2ch-source"
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
      user     = var.postgres_user
      password = var.postgres_password
    }
  }
}

resource "doublecloud_transfer" "pg2ch" {
  name       = "pg2ch"
  project_id = var.project_id
  source     = doublecloud_transfer_endpoint.pg-source.id
  target     = doublecloud_transfer_endpoint.dwh-target.id
  type       = "SNAPSHOT_AND_INCREMENT"
  activated  = false
}

resource "doublecloud_transfer_endpoint" "clickstream-source" {
  count      = var.enable_transfer ? 1 : 0
  name       = "clickstream-source"
  project_id = var.project_id
  settings {
    kafka_source {
      connection {
        cluster_id = doublecloud_kafka_cluster.input-kafka.id
      }
      auth {
        sasl {
          user      = "admin"
          mechanism = "KAFKA_MECHANISM_SHA512"
        }
      }
      parser {
        json {
          schema {
            fields {
              field {
                name     = "user_ts"
                type     = "datetime"
                key      = false
                required = false
              }
              field {
                name     = "id"
                type     = "uint64"
                key      = false
                required = false
              }
              field {
                name     = "message"
                type     = "utf8"
                key      = false
                required = false
              }
            }
          }
          null_keys_allowed = false
          add_rest_column   = true
        }
      }
      topic_name = "clickhouse-events"
    }
  }
}

resource "doublecloud_transfer_endpoint" "clickstream-target" {
  count      = var.enable_transfer ? 1 : 0
  name       = "clickstream-target"
  project_id = var.project_id
  settings {
    clickhouse_target {
      clickhouse_cleanup_policy = "DROP"
      connection {
        address {
          cluster_id = doublecloud_clickhouse_cluster.target-clickhouse.id
        }
        database = "default"
        user     = "admin"
      }
    }
  }
}

resource "doublecloud_transfer" "clickstream-transfer" {
  count      = var.enable_transfer ? 1 : 0
  name       = "clickstream-transfer"
  project_id = var.project_id
  source     = doublecloud_transfer_endpoint.clickstream-source[count.index].id
  target     = doublecloud_transfer_endpoint.clickstream-target[count.index].id
  type       = "INCREMENT_ONLY"
  activated  = true
}

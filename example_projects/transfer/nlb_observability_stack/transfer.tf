resource "doublecloud_transfer_endpoint" "nlb-s3-s32ch-source" {
  name       = var.transfer_source_name
  project_id = var.project_id
  settings {
    object_storage_source {
      provider {
        bucket                = var.bucket_name
        path_prefix           = var.bucket_prefix
        aws_access_key_id     = var.aws_access_key_id
        aws_secret_access_key = var.aws_access_key_secret
        region                = var.region
        endpoint              = var.endpoint
        use_ssl               = true
        verify_ssl_cert       = false
      }
      format {
        csv {
          delimiter = " " // space as delimiter
          advanced_options {
          }
          additional_options {
          }
        }
      }
      event_source {
        sqs {
          queue_name = var.sqs_name
        }
      }
      result_table {
        add_system_cols = true
        table_name      = var.transfer_source_table_name
        table_namespace = var.transfer_source_table_namespace
      }
      result_schema {
        data_schema {
          fields {
            field {
              name     = "type"
              type     = "string"
              required = false
              key      = false
              path     = "0"
            }
            field {
              name     = "version"
              type     = "string"
              required = false
              key      = false
              path     = "1"
            }
            field {
              name     = "time"
              type     = "datetime"
              required = false
              key      = false
              path     = "2"
            }
            field {
              name     = "elb"
              type     = "string"
              required = false
              key      = false
              path     = "3"
            }
            field {
              name     = "listener"
              type     = "string"
              required = false
              key      = false
              path     = "4"
            }
            field {
              name     = "client_port"
              type     = "string"
              required = false
              key      = false
              path     = "5"
            }
            field {
              name     = "destination_port"
              type     = "string"
              required = false
              key      = false
              path     = "6"
            }
            field {
              name     = "connection_time"
              type     = "uint64"
              required = false
              key      = false
              path     = "7"
            }
            field {
              name     = "tls_handshake_time"
              type     = "string"
              required = false
              key      = false
              path     = "8"
            }
            field {
              name     = "received_bytes"
              type     = "uint64"
              required = false
              key      = false
              path     = "9"
            }
            field {
              name     = "sent_bytes"
              type     = "uint64"
              required = false
              key      = false
              path     = "10"
            }
            field {
              name     = "incoming_tls_alert"
              type     = "string"
              required = false
              key      = false
              path     = "11"
            }
            field {
              name     = "chosen_cert_arn"
              type     = "string"
              required = false
              key      = false
              path     = "12"
            }
            field {
              name     = "chosen_cert_serial"
              type     = "string"
              required = false
              key      = false
              path     = "13"
            }
            field {
              name     = "tls_cipher"
              type     = "string"
              required = false
              key      = false
              path     = "14"
            }
            field {
              name     = "tls_protocol_version"
              type     = "string"
              required = false
              key      = false
              path     = "15"
            }
            field {
              name     = "tls_named_group"
              type     = "string"
              required = false
              key      = false
              path     = "16"
            }
            field {
              name     = "domain_name"
              type     = "string"
              required = false
              key      = false
              path     = "17"
            }
            field {
              name     = "alpn_fe_protocol"
              type     = "string"
              required = false
              key      = false
              path     = "18"
            }
            field {
              name     = "alpn_be_protocol"
              type     = "string"
              required = false
              key      = false
              path     = "19"
            }
            field {
              name     = "alpn_client_preference_list"
              type     = "string"
              required = false
              key      = false
              path     = "20"
            }
            field {
              name     = "tls_connection_creation_time"
              type     = "datetime"
              required = false
              key      = false
              path     = "21"
            }
          }
        }
      }
    }
  }
  depends_on = [
    resource.doublecloud_clickhouse_cluster.nlb-logs-clickhouse-cluster,
    data.doublecloud_clickhouse.nlb-logs-clickhouse
  ]
}

resource "doublecloud_transfer_endpoint" "nlb-ch-s32ch-target" {
  name       = var.transfer_target_name
  project_id = var.project_id
  settings {
    clickhouse_target {
      clickhouse_cleanup_policy = "DROP"
      connection {
        address {
          cluster_id = doublecloud_clickhouse_cluster.nlb-logs-clickhouse-cluster.id
        }
        database = "default"
        password = data.doublecloud_clickhouse.nlb-logs-clickhouse.connection_info.password
        user     = data.doublecloud_clickhouse.nlb-logs-clickhouse.connection_info.user
      }
    }
  }
  depends_on = [
    resource.doublecloud_clickhouse_cluster.nlb-logs-clickhouse-cluster,
    data.doublecloud_clickhouse.nlb-logs-clickhouse
  ]
}

resource "doublecloud_transfer" "nlb-logs-s32ch" {
  name       = var.transfer_name
  project_id = var.project_id
  source     = doublecloud_transfer_endpoint.nlb-s3-s32ch-source.id
  target     = doublecloud_transfer_endpoint.nlb-ch-s32ch-target.id
  type       = "INCREMENT_ONLY"
  activated  = false
  depends_on = [
    resource.doublecloud_clickhouse_cluster.nlb-logs-clickhouse-cluster,
    data.doublecloud_clickhouse.nlb-logs-clickhouse
  ]
}

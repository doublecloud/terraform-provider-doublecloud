// Grab newly create Clickhouse data
data "doublecloud_clickhouse" "target-clickhouse" {
  name       = doublecloud_clickhouse_cluster.target-clickhouse.name
  project_id = var.project_id
}

// This will output the database endpoint
output "clikchouse_connection" {
  description = "Clickhouse Connection profile"
  value       = data.doublecloud_clickhouse.target-clickhouse.connection_info
}

// This will output the database port
output "kafka_connection" {
  description = "Kafka Connection profile"
  value       = data.doublecloud_kafka.input-kafka.connection_info
}

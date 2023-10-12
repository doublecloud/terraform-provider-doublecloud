variable "project_id" {
  type        = string
  description = "ID of the DoubleCloud project in which to create resources"
}

variable "postgresql_postgres_password" {
  type        = string
  default     = "password"
  description = "Password of the user 'postgres' in the source PostgreSQL cluster"
}

variable "clickhouse_admin_password" {
  type        = string
  default     = "password"
  sensitive   = true
  description = "Password of the user 'admin' in the target ClickHouse cluster"
}

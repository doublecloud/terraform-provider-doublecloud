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

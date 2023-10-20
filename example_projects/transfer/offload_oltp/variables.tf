variable "dc-token" {
  default     = ""
  description = "Auth token for double cloud, see: https://github.com/doublecloud/terraform-provider-doublecloud"
}
variable "profile" {
  default     = ""
  description = "Name of AWS profile"
}

variable "vpc_id" {
  default     = ""
  description = "VPC ID of exist infra to peer with"
}
variable "is_prod" {
  default     = ""
  description = "Is environment production"
}
variable "project_id" {
  default     = ""
  description = "Double.Cloud project ID"
}
variable "postgres_host" {
  default     = ""
  description = "Source host"
}
variable "postgres_database" {
  default     = ""
  description = "Source database"
}
variable "postgres_user" {
  default     = ""
  description = "Source user"
}
variable "postgres_password" {
  default     = ""
  description = "Source Password"
}

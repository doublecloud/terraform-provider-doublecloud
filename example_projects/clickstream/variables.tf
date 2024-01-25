// This variable contains your IP address. This
// is used when setting up the Access Rule on the
// clusters security group
variable "my_ip" {
  description = "Your IP address"
  type        = string
  sensitive   = true
}

// This variable contains your IP6 address. This
// is used when setting up the Access Rule on the
// clusters security group
variable "my_ipv6" {
  description = "Your IPv6 address"
  type        = string
  sensitive   = true
}

// This example host infra on top of AWS cloud provider, so we must choose AWS region
variable "region_id" {
  type        = string
  description = "ID of the AWS region in which to create resources"
  default     = "eu-central-1"
}

// You can see this project_id on https://app.double.cloud/project-settings page
variable "project_id" {
  type        = string
  description = "ID of the DoubleCloud project in which to create resources"
}

// Authorization in Double.Cloud work with key.json files, so we must specify were it located
variable "dc_key_path" {
  type        = string
  default     = "~/.config/auth_key.json"
  description = "Path to DC key"
}

// Will create a transfer between kafka and clickhouse
variable "enable_transfer" {
  type = bool
  default = false
  description = "Create delivery from kafka to clickhouse via DoubleCloud.Transfer"
}

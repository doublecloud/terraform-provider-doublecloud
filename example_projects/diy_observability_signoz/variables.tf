variable "project_id" {
  type        = string
  description = "ID of the DoubleCloud project in which to create resources"
}

// This variable contains your IP address. This
// is used when setting up the SSH rule on the
// web security group
variable "my_ip" {
  description = "Your IP address"
  type        = string
  sensitive   = true
}
// This variable contains your IP address. This
// is used when setting up the SSH rule on the
// web security group
variable "my_ipv6" {
  description = "Your IPv6 address"
  type        = string
  sensitive   = true
}

variable "ipv4_cidr" {
  type        = string
  description = "CIDR of used vpc"
  default     = "10.0.0.0/16"
}

variable "region_id" {
  type        = string
  description = "ID of the region in which to create resources"
  default     = "eu-central-1"
}

variable "aws_profile" {
  type        = string
  description = "AWS Profile Name"
  default     = "default"
}

variable "cluster_name" {
  type        = string
  description = "Name of K8S cluster which will host app"
  default     = "main"
}

variable "clickhouse_cluster_resource_preset" {
  type        = string
  default     = "s1-c2-m4"
  description = "Specs for the managed ClickHouse cluster created in DoubleCloud"
}

variable "signoz_namespace" {
  type        = string
  default     = "signoz"
  description = "Namespace in k8s cluster for signoz helm chart installation"
}

variable "path_to_dc_key" {
  type        = string
  default     = "~/.config/auth_key.json"
  description = "Path to DC key"
}

variable "project_id" {
  type        = string
  description = "ID of the DoubleCloud project in which to create resources"
}

variable "ipv4_cidr" {
  type        = string
  description = "CIDR of a used vpc"
  default     = "10.0.0.0/16"
}

variable "region_id" {
  type        = string
  description = "ID of the region in which to create resources"
  default     = "us-east-2"
}

#Variables for VPC Peering
variable "peered_region_id" {
  type        = string
  description = "ID of region in the user's AWS Account with which should be establish peering"
  default     = "eu-west-1"
}

variable "peered_ipv4_cidr_block" {
  type        = string
  description = "CIDR of VPC in account with which should be establish peering"
  default     = "172.31.0.0/16"
}

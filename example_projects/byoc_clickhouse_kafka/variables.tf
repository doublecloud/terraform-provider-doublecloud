variable "project_id" {
  type        = string
  description = "ID of the DoubleCloud project in which to create resources"
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

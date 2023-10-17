variable "name" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "access_logs_bucket" {
  type = string
  default = null
}

variable "access_logs_bucket_prefix" {
  type = string
  default = null
}

variable "subnet_ids" {
  type = list(string)
}

variable "tags" {
  type = map(string)
}

variable "security_group_ids" {
  type = list(string)

  default = null
}

variable "target_protocol" {
  type = string

  validation {
    condition     = var.target_protocol == "TCP" || var.target_protocol == "UDP" || var.target_protocol == "TLS"
    error_message = "NLB module supports, you guessed it, NLBs (L4 - TCP/UDP/TLS)! Use ALB if you want L7."
  }
}

variable "listener_protocol" {
  type = string

  validation {
    condition     = var.listener_protocol == "TCP" || var.listener_protocol == "UDP" || var.listener_protocol == "TLS"
    error_message = "NLB module supports, you guessed it, NLBs (L4 - TCP/UDP/TLS)! Use ALB if you want L7."
  }
}

variable "is_grpc" {
  type = bool
}

variable "fqdn" {
  type = string
}

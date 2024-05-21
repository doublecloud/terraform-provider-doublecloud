terraform {
  required_providers {
    doublecloud = {
      source = "registry.terraform.io/doublecloud/doublecloud"
    }
  }
}

data "doublecloud_kafka" "alpha-kafka" {
  id         = var.kafka_id
  project_id = var.project_id
}

provider "doublecloud" {
  federation_id = var.federation_id
}


variable "federation_id" {
  type = string
}

variable "project_id" {
  type = string
}

variable "kafka_id" {
  type = string
}

terraform {
  required_version = ">= 1.5.5"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.51.0"
    }
    time = {
      source = "hashicorp/time"
    }
    doublecloud = {
      source  = "registry.terraform.io/doublecloud/doublecloud"
      version = ">= 0.1.6"
    }
  }
}

provider "aws" {
  region = var.region_id
}


provider "doublecloud" {
  # See https://double.cloud/docs/en/public-api/tutorials/transfer-api-quickstart on how to obtain this file
  authorized_key = file("authorized_key.json")
}

module "doublecloud-byoc" {
  source  = "doublecloud/doublecloud-byoc/aws"
  version = "1.0.2"

  ipv4_cidr = var.ipv4_cidr
}
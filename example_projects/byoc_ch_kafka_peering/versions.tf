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
  alias  = "peered"
  region = var.peered_region_id
}

provider "aws" {
  alias  = "byoc"
  region = var.region_id
}

provider "doublecloud" {
  # See https://double.cloud/docs/en/public-api/tutorials/transfer-api-quickstart on how to obtain this file
  authorized_key = file("authorized_key.json")
}

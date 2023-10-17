terraform {
  required_providers {
    doublecloud = {
      source = "registry.terraform.io/doublecloud/doublecloud"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "= 5.19.0"
    }
  }
}

provider "doublecloud" {
  authorized_key = file("authorized_key.json")
  endpoint       = var.service_account_endpoint
  token_url      = var.service_account_token_endpoint
}

provider "aws" {
  profile             = var.aws_profile
  region              = var.region
  allowed_account_ids = [var.aws_account_id]
  default_tags {
    tags = var.tags
  }
}

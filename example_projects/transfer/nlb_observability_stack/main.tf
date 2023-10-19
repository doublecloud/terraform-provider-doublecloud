terraform {
  required_providers {
    doublecloud = {
      source  = "registry.terraform.io/doublecloud/doublecloud"
      version = "= 0.1.11"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "= 5.19.0"
    }
  }
}

provider "doublecloud" {
  authorized_key = file("authorized_key.json")
}

provider "aws" {
  profile             = var.aws_profile
  region              = var.region
  allowed_account_ids = [var.aws_account_id]
  default_tags {
    tags = var.tags
  }
}

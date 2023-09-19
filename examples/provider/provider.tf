terraform {
  required_providers {
    doublecloud = {
      source = "registry.terraform.io/doublecloud/doublecloud"
    }
  }
}

provider "doublecloud" {
  authorized_key = file("authorized_key.json") # See https://double.cloud/docs/en/public-api/tutorials/transfer-api-quickstart on how to obtain this file
}

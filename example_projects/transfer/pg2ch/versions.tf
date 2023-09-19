terraform {
  required_providers {
    doublecloud = {
      source = "registry.terraform.io/doublecloud/doublecloud"
    }
  }
}

provider "doublecloud" {
  authorized_key = file("authorized_key.json")
}

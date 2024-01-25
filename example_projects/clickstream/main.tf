// Here is where we are defining
// our Terraform settings
terraform {
  required_providers {
    doublecloud = {
      source  = "registry.terraform.io/doublecloud/doublecloud"
      version = ">= 0.1.6"
    }
  }

  // This is the required version of Terraform
  required_version = "1.5.6"
}

provider "doublecloud" {
  # See https://double.cloud/docs/en/public-api/tutorials/transfer-api-quickstart on how to obtain this file
  authorized_key = file(var.dc_key_path)
}

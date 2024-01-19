// Here is where we are defining
// our Terraform settings
terraform {
  required_providers {
    // The only required provider we need
    // is aws, and we want version 4.0.0
    aws = {
      source  = "hashicorp/aws"
      version = "4.51.0"
    }
    time = {
      source = "hashicorp/time"
    }
    doublecloud = {
      source  = "registry.terraform.io/doublecloud/doublecloud"
      version = ">= 0.1.6"
    }
  }

  // This is the required version of Terraform
  required_version = "1.5.6"
}

// Here we are configuring our aws provider.
// We are setting the region to the region of
// our variable "aws_region"
provider "aws" {
  region = var.aws_region
  profile = var.aws_profile
}

provider "doublecloud" {
  # See https://double.cloud/docs/en/public-api/tutorials/transfer-api-quickstart on how to obtain this file
  authorized_key = file(var.dc_key_path)
}

// This data object is going to be
// holding all the available availability
// zones in our defined region
data "aws_availability_zones" "available" {
  state = "available"
}

// Create a data object called "ubuntu" that holds the latest
// Ubuntu 20.04 server AMI
data "aws_ami" "ubuntu" {
  // We want the most recent AMI
  most_recent = "true"

  // We are filtering through the names of the AMIs. We want the
  // Ubuntu 20.04 server
  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }

  // We are filtering through the virtualization type to make sure
  // we only find AMIs with a virtualization type of hvm
  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  // This is the ID of the publisher that created the AMI.
  // The publisher of Ubuntu 20.04 LTS Focal is Canonical
  // and their ID is 099720109477
  owners = ["099720109477"]
}

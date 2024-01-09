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
    helm = {
      source  = "hashicorp/helm"
      version = "= 2.5.1"
    }
    doublecloud = {
      source  = "registry.terraform.io/doublecloud/doublecloud"
      version = ">= 0.1.6"
    }
  }

  // This is the required version of Terraform
  required_version = "1.5.6"
}

provider "aws" {
  profile = var.aws_profile
  region  = var.region_id
}

data "aws_eks_cluster" "k8s" {
  name = var.cluster_name
}

provider "helm" {
  kubernetes {
    host                   = data.aws_eks_cluster.k8s.endpoint
    cluster_ca_certificate = base64decode(data.aws_eks_cluster.k8s.certificate_authority[0].data)
    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      args        = ["eks", "get-token", "--profile", var.aws_profile, "--cluster-name", var.cluster_name]
      command     = "aws"
    }
  }
  debug = true
}

provider "doublecloud" {
  # See https://double.cloud/docs/en/public-api/tutorials/transfer-api-quickstart on how to obtain this file
  authorized_key = file(var.path_to_dc_key)
}

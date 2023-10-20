provider "doublecloud" {
  endpoint       = "api.double.cloud:443"
  authorized_key = file(var.dc-token)
}
provider "aws" {
  profile = var.profile
}

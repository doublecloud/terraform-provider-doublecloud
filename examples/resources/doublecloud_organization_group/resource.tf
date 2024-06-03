resource "doublecloud_organization_group" "sample-group" {
  organizaton_id = var.organization_id
  name           = "test-federation"
  description    = "test federation"
}

resource "doublecloud_saml_federation" "sample-federation" {
  organizaton_id               = var.organization_id
  name                         = "test-federation"
  description                  = "test federation"
  cookie_max_age               = "9h"
  auto_create_account_on_login = true
  issuer                       = "test-issuer"
  sso_binding                  = "POST"
  sso_url                      = "https://myawesome.federation.com/"
  case_insensitive_name_ids    = true
}

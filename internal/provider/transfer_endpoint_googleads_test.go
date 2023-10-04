package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTransferEndpointGoogleAdsSource(t *testing.T) {
	t.Parallel()

	testEndpointName := fmt.Sprintf("%s-googleads-source", testPrefix)
	testEndpointID := fmt.Sprintf("doublecloud_transfer_endpoint.%s", testEndpointName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: fmt.Sprintf(
					`resource "doublecloud_transfer_endpoint" %[1]q {
						project_id = %[2]q
						name = %[1]q
						settings {
							googleads_source {
								credentials {
									developer_token = "mysecretdevtoken"
									client_id = "mycliid"
									client_secret = "myclisecret"
									access_token = "mysecretacctoken"
									refresh_token = "mysecretreftoken"
								}
								customer_id = "customerid"
								start_date = "2020-02-13"
								end_date = "2020-03-12"
								custom_queries = [
									{
										query = "SELECT something WHERE something_else"
										table_name = "users"
									}
								]
								login_customer_id = "customerlogin"
								conversion_window_days = 7
							}
						}
					}`,
					testEndpointName,
					testProjectId,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEndpointID, "name", testEndpointName),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.credentials.developer_token", "mysecretdevtoken"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.credentials.client_id", "mycliid"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.credentials.client_secret", "myclisecret"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.credentials.access_token", "mysecretacctoken"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.credentials.refresh_token", "mysecretreftoken"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.customer_id", "customerid"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.start_date", "2020-02-13"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.end_date", "2020-03-12"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.custom_queries.0.query", "SELECT something WHERE something_else"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.custom_queries.0.table_name", "users"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.login_customer_id", "customerlogin"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.conversion_window_days", "7"),
				),
			},
			// Update and Read testing
			{
				Config: fmt.Sprintf(
					`resource "doublecloud_transfer_endpoint" %[1]q {
						project_id = %[2]q
						name = %[1]q
						settings {
							googleads_source {
								credentials {
									developer_token = "mysecretdevtoken2"
									client_id = "mycliid"
									client_secret = "myclisecret"
									access_token = "mysecretacctoken"
									refresh_token = "mysecretreftoken"
								}
								customer_id = "customerid2"
								start_date = "2020-03-13"
								end_date = "2020-04-12"
								custom_queries = [
									{
										query = "SELECT something WHERE something_else 2"
										table_name = "users"
									}
								]
								login_customer_id = "customerlogin"
								conversion_window_days = 7
							}
						}
					}`,
					testEndpointName,
					testProjectId,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEndpointID, "name", testEndpointName),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.credentials.developer_token", "mysecretdevtoken2"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.credentials.client_id", "mycliid"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.credentials.client_secret", "myclisecret"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.credentials.access_token", "mysecretacctoken"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.credentials.refresh_token", "mysecretreftoken"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.customer_id", "customerid2"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.start_date", "2020-03-13"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.end_date", "2020-04-12"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.custom_queries.0.query", "SELECT something WHERE something_else 2"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.custom_queries.0.table_name", "users"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.login_customer_id", "customerlogin"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.googleads_source.conversion_window_days", "7"),
				),
			},
			// Delete occurs automatically in TestCase
		},
	})
}

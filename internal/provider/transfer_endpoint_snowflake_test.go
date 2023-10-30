package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTransferEndpointSnowflakeSource(t *testing.T) {
	t.Parallel()

	testEndpointSnowflakeSourceName := fmt.Sprintf("%s-snowflake-source", testPrefix)
	testEndpointSnowflakeSourceID := fmt.Sprintf("doublecloud_transfer_endpoint.%s", testEndpointSnowflakeSourceName)

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
							snowflake_source {
								host = "test-1"
								role = "TEST"
								warehouse = "test-2"
								database = "test-3"
								schema = "test-4"
								jdbc_url_params = "test-5"

								credentials {
									oauth {
										client_id = "cliid"
										client_secret = "clisecret"
										access_token = "cliaccesstoken"
										refresh_token = "clirefreshtoken"
									}
								}
							}
						}
					}`,
					testEndpointSnowflakeSourceName,
					testProjectId,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEndpointSnowflakeSourceID, "name", testEndpointSnowflakeSourceName),
					resource.TestCheckResourceAttr(testEndpointSnowflakeSourceID, "settings.snowflake_source.host", "test-1"),
					resource.TestCheckResourceAttr(testEndpointSnowflakeSourceID, "settings.snowflake_source.role", "TEST"),
					resource.TestCheckResourceAttr(testEndpointSnowflakeSourceID, "settings.snowflake_source.warehouse", "test-2"),
					resource.TestCheckResourceAttr(testEndpointSnowflakeSourceID, "settings.snowflake_source.database", "test-3"),
					resource.TestCheckResourceAttr(testEndpointSnowflakeSourceID, "settings.snowflake_source.schema", "test-4"),
					resource.TestCheckResourceAttr(testEndpointSnowflakeSourceID, "settings.snowflake_source.jdbc_url_params", "test-5"),
				),
			},
			// Update and Read testing

			{
				Config: fmt.Sprintf(
					`resource "doublecloud_transfer_endpoint" %[1]q {
						project_id = %[2]q
						name = %[1]q
						settings {
							snowflake_source {
								host = "test-1"
								role = "TEST-2"
								warehouse = "test-2"
								database = "test-3"
								schema = "new-schema"
								jdbc_url_params = "test-5"

								credentials {
									oauth {
										client_id = "cliid"
										client_secret = "clisecret"
										access_token = "cliaccesstoken"
										refresh_token = "clirefreshtoken"
									}
								}
							}
						}
					}`,
					testEndpointSnowflakeSourceName,
					testProjectId,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEndpointSnowflakeSourceID, "name", testEndpointSnowflakeSourceName),
					resource.TestCheckResourceAttr(testEndpointSnowflakeSourceID, "settings.snowflake_source.role", "TEST-2"),
					resource.TestCheckResourceAttr(testEndpointSnowflakeSourceID, "settings.snowflake_source.schema", "new-schema"),
				),
			},
			// Update of Credentials does not work, because the values are not returned by API
			// Delete occurs automatically in TestCase
		},
	})
}

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTransferEndpointLinkedinAdsSource(t *testing.T) {
	t.Parallel()

	testEndpointLinkedinAdsSourceName := fmt.Sprintf("%s-linkedinads-source", testPrefix)
	testEndpointLinkedinAdsSourceID := fmt.Sprintf("doublecloud_transfer_endpoint.%s", testEndpointLinkedinAdsSourceName)

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
							linkedinads_source {
								start_date = "2021-05-17"
								credentials {
									oauth {
										client_id = "cliid"
										client_secret = "clisecret"
										refresh_token = "clirefreshtoken"
									}
								}
							}
						}
					}`,
					testEndpointLinkedinAdsSourceName,
					testProjectId,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEndpointLinkedinAdsSourceID, "name", testEndpointLinkedinAdsSourceName),
					resource.TestCheckResourceAttr(testEndpointLinkedinAdsSourceID, "settings.linkedinads_source.start_date", "2021-05-17"),
				),
			},
			// Update and Read testing
			{
				Config: fmt.Sprintf(
					`resource "doublecloud_transfer_endpoint" %[1]q {
						project_id = %[2]q
						name = %[1]q
						settings {
							linkedinads_source {
								start_date = "2022-06-18"
								account_ids = [1, 2, 3]
								credentials {
									oauth {
										client_id = "cliid"
										client_secret = "clisecret"
										refresh_token = "clirefreshtoken"
									}
								}
							}
						}
					}`,
					testEndpointLinkedinAdsSourceName,
					testProjectId,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEndpointLinkedinAdsSourceID, "name", testEndpointLinkedinAdsSourceName),
					resource.TestCheckResourceAttr(testEndpointLinkedinAdsSourceID, "settings.linkedinads_source.start_date", "2022-06-18"),
					resource.TestCheckResourceAttr(testEndpointLinkedinAdsSourceID, "settings.linkedinads_source.account_ids.0", "1"),
					resource.TestCheckResourceAttr(testEndpointLinkedinAdsSourceID, "settings.linkedinads_source.account_ids.1", "2"),
					resource.TestCheckResourceAttr(testEndpointLinkedinAdsSourceID, "settings.linkedinads_source.account_ids.2", "3"),
				),
			},
			// Update of Credentials does not work, because the values are not returned by API
			// Delete occurs automatically in TestCase
		},
	})
}

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTransferEndpointFacebookMarketingSource(t *testing.T) {
	t.Parallel()

	testEndpointName := fmt.Sprintf("%s-facebookmarketing-source", testPrefix)
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
							facebookmarketing_source {
								start_date = "2017-01-25T00:00:00Z"
								account_id = "111111111111111"
								end_date = "2017-01-25T23:59:59Z"
								access_token = "EAACEdExxx"
								custom_insights = [
									{
										name = "MyInsight1"
										fields = ["account_id", "account_currency"]
										breakdowns = ["app_id"]
										action_breakdowns = ["device"]
									},
									{
										name = "MyInsight2"
										fields = ["account_currency"]
										breakdowns = ["image_asset", "gender", "ad_format_asset"]
										action_breakdowns = ["type"]
									}
								]
							}
						}
					}`,
					testEndpointName,
					testProjectId,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEndpointID, "name", testEndpointName),
					resource.TestCheckResourceAttr(testEndpointID, "settings.facebookmarketing_source.start_date", "2017-01-25T00:00:00Z"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.facebookmarketing_source.account_id", "111111111111111"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.facebookmarketing_source.end_date", "2017-01-25T23:59:59Z"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.facebookmarketing_source.access_token", "EAACEdExxx"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.facebookmarketing_source.custom_insights.0.name", "MyInsight1"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.facebookmarketing_source.custom_insights.0.fields.0", "account_id"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.facebookmarketing_source.custom_insights.0.fields.1", "account_currency"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.facebookmarketing_source.custom_insights.0.breakdowns.0", "app_id"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.facebookmarketing_source.custom_insights.0.action_breakdowns.0", "device"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.facebookmarketing_source.custom_insights.1.name", "MyInsight2"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.facebookmarketing_source.custom_insights.1.fields.0", "account_currency"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.facebookmarketing_source.custom_insights.1.breakdowns.0", "image_asset"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.facebookmarketing_source.custom_insights.1.breakdowns.1", "gender"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.facebookmarketing_source.custom_insights.1.breakdowns.2", "ad_format_asset"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.facebookmarketing_source.custom_insights.1.action_breakdowns.0", "type"),
				),
			},
			// Update and Read testing
			{
				Config: fmt.Sprintf(
					`resource "doublecloud_transfer_endpoint" %[1]q {
						project_id = %[2]q
						name = %[1]q
						settings {
							facebookmarketing_source {
								start_date = "2017-01-26T00:00:00Z"
								account_id = "111111111111111"
								end_date = "2017-01-26T23:59:59Z"
								access_token = "EAACEdExxx"
							}
						}
					}`,
					testEndpointName,
					testProjectId,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEndpointID, "name", testEndpointName),
					resource.TestCheckResourceAttr(testEndpointID, "settings.facebookmarketing_source.start_date", "2017-01-26T00:00:00Z"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.facebookmarketing_source.account_id", "111111111111111"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.facebookmarketing_source.end_date", "2017-01-26T23:59:59Z"),
					resource.TestCheckResourceAttr(testEndpointID, "settings.facebookmarketing_source.access_token", "EAACEdExxx"),
				),
			},
			// Delete occurs automatically in TestCase
		},
	})
}

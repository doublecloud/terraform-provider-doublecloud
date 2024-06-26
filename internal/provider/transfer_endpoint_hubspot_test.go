package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

var (
	testEHubspotSourceName string = fmt.Sprintf("%v-hubspot-source", testPrefix)
	testEHubspotSourceId   string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testEHubspotSourceName)
)

func TestAccTransferEndpointHubspotResource(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTransferEndpointResourceHubspotConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEHubspotSourceId, "name", testEHubspotSourceName),
					resource.TestCheckResourceAttr(testEHubspotSourceId, "settings.hubspot_source.start_date", "2024-06-24"),
					resource.TestCheckResourceAttr(testEHubspotSourceId, "settings.hubspot_source.enable_experimental_streams", "true"),
					resource.TestCheckResourceAttr(testEHubspotSourceId, "settings.hubspot_source.credentials.private_app.access_token", "my_access_token"),
				),
			},

			// Update and Read testing
			{
				Config: testAccTransferEndpointResourceHubspotModifiedConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEHubspotSourceId, "name", testEHubspotSourceName),
					resource.TestCheckResourceAttr(testEHubspotSourceId, "settings.hubspot_source.start_date", "2024-06-25"),
					resource.TestCheckResourceAttr(testEHubspotSourceId, "settings.hubspot_source.enable_experimental_streams", "false"),
					resource.TestCheckResourceAttr(testEHubspotSourceId, "settings.hubspot_source.credentials.private_app.access_token", "my_new_access_token"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccTransferEndpointResourceHubspotConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[2]q
	name = %[1]q
	settings {
		hubspot_source {
			start_date = "2024-06-24"
			enable_experimental_streams = "true"
			credentials {
				private_app {
					access_token = "my_access_token"
				}
			}
		}
	}
}
`, testEHubspotSourceName, testProjectId)
}

func testAccTransferEndpointResourceHubspotModifiedConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[2]q
	name = %[1]q
	settings {
		hubspot_source {
			start_date = "2024-06-25"
			enable_experimental_streams = "false"
			credentials {
				private_app {
					access_token = "my_new_access_token"
				}
			}
		}
	}
}
`, testEHubspotSourceName, testProjectId)
}

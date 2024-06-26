package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

var (
	testEInstagramSourceName string = fmt.Sprintf("%v-instagram-source", testPrefix)

	testEInstagramSourceId string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testEInstagramSourceName)
)

func TestAccTransferEndpointInstagramResource(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTransferEndpointResourceInstagramConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEInstagramSourceId, "name", testEInstagramSourceName),
					resource.TestCheckResourceAttr(testEInstagramSourceId, "settings.instagram_source.start_date", "2024-06-24"),
					resource.TestCheckResourceAttr(testEInstagramSourceId, "settings.instagram_source.access_token", "test_access_token"),
				),
			},

			// Update and Read testing
			{
				Config: testAccTransferEndpointResourceInstagramModifiedConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEInstagramSourceId, "name", testEInstagramSourceName),
					resource.TestCheckResourceAttr(testEInstagramSourceId, "settings.instagram_source.start_date", "2024-06-25"),
					resource.TestCheckResourceAttr(testEInstagramSourceId, "settings.instagram_source.access_token", "new_test_access_token"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccTransferEndpointResourceInstagramConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[2]q
	name = %[1]q
	settings {
		instagram_source {
			start_date = "2024-06-24"
			access_token = "test_access_token"
		}
	}
}
`, testEInstagramSourceName, testProjectId)
}

func testAccTransferEndpointResourceInstagramModifiedConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[2]q
	name = %[1]q
	settings {
		instagram_source {
			start_date = "2024-06-25"
			access_token = "new_test_access_token"
		}
	}
}
`, testEInstagramSourceName, testProjectId)
}

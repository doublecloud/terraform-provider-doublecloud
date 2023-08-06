package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetworkDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccExampleDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.doublecloud_network.test", "name", testNetworkName),
					resource.TestCheckResourceAttr("data.doublecloud_network.test", "region_id", "eu-central-1"),
					resource.TestCheckResourceAttr("data.doublecloud_network.test", "ipv4_cidr_block", "172.42.0.0/16"),
					resource.TestCheckResourceAttr("data.doublecloud_network.test", "cloud_type", "aws"),
				),
			},
		},
	})
}

func testAccExampleDataSourceConfig() string {
	return fmt.Sprintf(`
data "doublecloud_network" "test" {
	name = "%v"
	project_id = "%v"
}`, testNetworkName, testProjectId)
}

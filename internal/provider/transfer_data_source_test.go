package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTransferDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccTransferDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.doublecloud_transfer.test", "name", testDSTransferName),
					resource.TestCheckResourceAttr("data.doublecloud_transfer.test", "type", "INCREMENT_ONLY"),
					resource.TestCheckResourceAttr("data.doublecloud_transfer.test", "status", "RUNNING"),
				),
			},
		},
	})
}

func testAccTransferDataSourceConfig() string {
	return fmt.Sprintf(`
data "doublecloud_transfer" "test" {
	name = "%v"
	project_id = "%v"
}`, testDSTransferName, testProjectId)
}

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccKafkaDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccKafkaDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.doublecloud_kafka.test", "region_id", "eu-central-1"),
					resource.TestCheckResourceAttr("data.doublecloud_kafka.test", "cloud_type", "aws"),
					resource.TestCheckResourceAttr("data.doublecloud_kafka.test", "version", "3.5"),
					resource.TestCheckResourceAttr("data.doublecloud_kafka.test", "connection_info.user", "admin"),
					resource.TestCheckResourceAttr("data.doublecloud_kafka.test", "private_connection_info.user", "admin"),
				),
			},
		},
	})
}

func testAccKafkaDataSourceConfig() string {
	return fmt.Sprintf(`
data "doublecloud_kafka" "test" {
	name = "%v"
	project_id = "%v"
}`, testKafkaName, testProjectId)
}

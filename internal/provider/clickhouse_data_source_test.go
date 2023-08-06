package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccClickhouseDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccClickhouseDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.doublecloud_clickhouse.test", "region_id", "eu-central-1"),
					resource.TestCheckResourceAttr("data.doublecloud_clickhouse.test", "cloud_type", "aws"),
					resource.TestCheckResourceAttr("data.doublecloud_clickhouse.test", "connection_info.user", "admin"),
					resource.TestCheckResourceAttr("data.doublecloud_clickhouse.test", "connection_info.https_port", "8443"),
					resource.TestCheckResourceAttr("data.doublecloud_clickhouse.test", "connection_info.tcp_port_secure", "9440"),
					resource.TestCheckResourceAttr("data.doublecloud_clickhouse.test", "private_connection_info.user", "admin"),
					resource.TestCheckResourceAttr("data.doublecloud_clickhouse.test", "private_connection_info.https_port", "8443"),
					resource.TestCheckResourceAttr("data.doublecloud_clickhouse.test", "private_connection_info.tcp_port_secure", "9440"),
				),
			},
		},
	})
}

func testAccClickhouseDataSourceConfig() string {
	return fmt.Sprintf(`
data "doublecloud_clickhouse" "test" {
	name = "%v"
	project_id = "%v"
}`, testClickhouseName, testProjectId)
}

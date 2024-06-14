package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

var (
	testERedshiftSourceName string = fmt.Sprintf("%v-redshift-source", testPrefix)
	testERedshiftSourceID   string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testERedshiftSourceName)
)

func TestAccTransferEndpointRedshiftSource(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read Testing
			{
				Config: testAccTransferEndpointRedshiftConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testERedshiftSourceID, "name", testERedshiftSourceName),
					resource.TestCheckResourceAttr(testERedshiftSourceID, "settings.redshift_source.host", "leader-0.company.tech"),
					resource.TestCheckResourceAttr(testERedshiftSourceID, "settings.redshift_source.port", "5439"),
					resource.TestCheckResourceAttr(testERedshiftSourceID, "settings.redshift_source.database", "production"),
					resource.TestCheckResourceAttr(testERedshiftSourceID, "settings.redshift_source.username", "dc-transfer-endpoint"),
					resource.TestCheckResourceAttr(testERedshiftSourceID, "settings.redshift_source.password", "test"),
					resource.TestCheckResourceAttr(testERedshiftSourceID, "settings.redshift_source.schemas.#", "1"),
					resource.TestCheckResourceAttr(testERedshiftSourceID, "settings.redshift_source.schemas.0", "public"),
				),
			},

			// Update and Read Testing
			{
				Config: testAccTransferEndpointRedshiftModifiedConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testERedshiftSourceID, "name", testERedshiftSourceName),
					resource.TestCheckResourceAttr(testERedshiftSourceID, "settings.redshift_source.host", "leader-1.company.tech"),
					resource.TestCheckResourceAttr(testERedshiftSourceID, "settings.redshift_source.port", "5439"),
					resource.TestCheckResourceAttr(testERedshiftSourceID, "settings.redshift_source.database", "production"),
					resource.TestCheckResourceAttr(testERedshiftSourceID, "settings.redshift_source.username", "dc-transfer"),
					resource.TestCheckResourceAttr(testERedshiftSourceID, "settings.redshift_source.password", "test2"),
					resource.TestCheckResourceAttr(testERedshiftSourceID, "settings.redshift_source.schemas.#", "2"),
					resource.TestCheckResourceAttr(testERedshiftSourceID, "settings.redshift_source.schemas.0", "public"),
					resource.TestCheckResourceAttr(testERedshiftSourceID, "settings.redshift_source.schemas.1", "private"),
				),
			},

			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccTransferEndpointRedshiftConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[2]q
	name       = %[1]q
	settings {
		redshift_source {
			host = "leader-0.company.tech"
			port = 5439
			database = "production"
			username = "dc-transfer-endpoint"
			password = "test"
			schemas = ["public"]
		}
	}
}
`, testERedshiftSourceName, testProjectId)
}

func testAccTransferEndpointRedshiftModifiedConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[2]q
	name       = %[1]q
	settings {
		redshift_source {
			host = "leader-1.company.tech"
			port = 5439
			database = "production"
			username = "dc-transfer"
			password = "test2"
			schemas = ["public", "private"]
		}
	}
}
`, testERedshiftSourceName, testProjectId)
}

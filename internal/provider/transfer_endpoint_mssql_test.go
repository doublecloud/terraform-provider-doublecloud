package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

var (
	testEMssqlSourceName string = fmt.Sprintf("%v-mssql-source", testPrefix)
	testEMssqlSourceId   string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testEMssqlSourceName)
)

func TestAccTransferEndpointMssqlResource(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTransferEndpointResourceMssqlConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEMssqlSourceId, "name", testEMssqlSourceName),
					resource.TestCheckResourceAttr(testEMssqlSourceId, "settings.mssql_source.host", "localhost"),
					resource.TestCheckResourceAttr(testEMssqlSourceId, "settings.mssql_source.port", "1433"),
					resource.TestCheckResourceAttr(testEMssqlSourceId, "settings.mssql_source.database", "testdb"),
					resource.TestCheckResourceAttr(testEMssqlSourceId, "settings.mssql_source.username", "testuser"),
					resource.TestCheckResourceAttr(testEMssqlSourceId, "settings.mssql_source.password", "testpass"),
					resource.TestCheckResourceAttr(testEMssqlSourceId, "settings.mssql_source.replication_method", "STANDARD"),
					resource.TestCheckResourceAttr(testEMssqlSourceId, "settings.mssql_source.ssl_method.unencrypted.%", "0"),
				),
			},

			// Update and Read testing
			{
				Config: testAccTransferEndpointResourceMssqlModifiedConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEMssqlSourceId, "name", testEMssqlSourceName),
					resource.TestCheckResourceAttr(testEMssqlSourceId, "settings.mssql_source.host", "localhost"),
					resource.TestCheckResourceAttr(testEMssqlSourceId, "settings.mssql_source.port", "1433"),
					resource.TestCheckResourceAttr(testEMssqlSourceId, "settings.mssql_source.database", "newdb"),
					resource.TestCheckResourceAttr(testEMssqlSourceId, "settings.mssql_source.username", "newuser"),
					resource.TestCheckResourceAttr(testEMssqlSourceId, "settings.mssql_source.password", "newpass"),
					resource.TestCheckResourceAttr(testEMssqlSourceId, "settings.mssql_source.replication_method", "CDC"),
					resource.TestCheckResourceAttr(testEMssqlSourceId, "settings.mssql_source.ssl_method.encrypted_verify_cert.host_name_in_certificate", "my_certificate"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccTransferEndpointResourceMssqlConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[2]q
	name = %[1]q
	settings {
		mssql_source {
			host = "localhost"
			port = 1433
			database = "testdb"
			username = "testuser"
			password = "testpass"
			replication_method = "STANDARD"
			ssl_method {
				unencrypted {}
			}
		}
	}
}
`, testEMssqlSourceName, testProjectId)
}

func testAccTransferEndpointResourceMssqlModifiedConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[2]q
	name = %[1]q
	settings {
		mssql_source {
			host = "localhost"
			port = 1433
			database = "newdb"
			username = "newuser"
			password = "newpass"
			replication_method = "CDC"
			ssl_method {
				encrypted_verify_cert {
					host_name_in_certificate = "my_certificate"
				}
			}
		}
	}
}
`, testEMssqlSourceName, testProjectId)
}

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	testEMysqlSourceName string = fmt.Sprintf("%v-mysql-source", testPrefix)
	testEMysqlTargetName string = fmt.Sprintf("%v-mysql-target", testPrefix)

	testEMysqlSourceId string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testEMysqlSourceName)
	testEMysqlTargetId string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testEMysqlTargetName)
)

func TestAccTransferEndpointMysqlResource(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTransferEndpointResourceMysqlConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEMysqlSourceId, "name", testEMysqlSourceName),
					resource.TestCheckResourceAttr(testEMysqlSourceId, "settings.mysql_source.connection.on_premise.port", "3306"),
					resource.TestCheckResourceAttr(testEMysqlSourceId, "settings.mysql_source.database", "production"),
					resource.TestCheckResourceAttr(testEMysqlSourceId, "settings.mysql_source.user", "dc-transfer"),
					resource.TestCheckResourceAttr(testEMysqlSourceId, "settings.mysql_source.password", "foobar123"),
					resource.TestCheckResourceAttr(testEMysqlSourceId, "settings.mysql_source.object_transfer_settings.tables", "BEFORE_DATA"),

					resource.TestCheckResourceAttr(testEMysqlTargetId, "name", testEMysqlTargetName),
					resource.TestCheckResourceAttr(testEMysqlTargetId, "settings.mysql_target.connection.on_premise.port", "3306"),
					resource.TestCheckResourceAttr(testEMysqlTargetId, "settings.mysql_target.database", "production"),
					resource.TestCheckResourceAttr(testEMysqlTargetId, "settings.mysql_target.user", "dc-transfer"),
					resource.TestCheckResourceAttr(testEMysqlTargetId, "settings.mysql_target.password", "foobar123"),
					resource.TestCheckNoResourceAttr(testEMysqlTargetId, "settings.mysql_target.object_transfer_settings"),
				),
			},
			// Update and Read testing
			{
				Config: testAccTransferEndpointResourceMysqlModifiedConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEMysqlSourceId, "name", testEMysqlSourceName),
					resource.TestCheckResourceAttr(testEMysqlSourceId, "settings.mysql_source.connection.on_premise.port", "3307"),
					resource.TestCheckResourceAttr(testEMysqlSourceId, "settings.mysql_source.database", "production"),
					resource.TestCheckResourceAttr(testEMysqlSourceId, "settings.mysql_source.user", "dc-transfer"),
					resource.TestCheckResourceAttr(testEMysqlSourceId, "settings.mysql_source.password", "foobar124"),
					resource.TestCheckResourceAttr(testEMysqlSourceId, "settings.mysql_source.timezone", "Africa/Johannesburg"),
					resource.TestCheckResourceAttr(testEMysqlSourceId, "settings.mysql_source.object_transfer_settings.tables", "AFTER_DATA"),
					resource.TestCheckResourceAttr(testEMysqlSourceId, "settings.mysql_source.object_transfer_settings.view", "NEVER"),

					resource.TestCheckResourceAttr(testEMysqlTargetId, "name", testEMysqlTargetName),
					resource.TestCheckResourceAttr(testEMysqlTargetId, "settings.mysql_target.connection.on_premise.port", "3307"),
					resource.TestCheckResourceAttr(testEMysqlTargetId, "settings.mysql_target.database", "production"),
					resource.TestCheckResourceAttr(testEMysqlTargetId, "settings.mysql_target.user", "dc-transfer"),
					resource.TestCheckResourceAttr(testEMysqlTargetId, "settings.mysql_target.password", "foobar124"),
					resource.TestCheckResourceAttr(testEMysqlTargetId, "settings.mysql_target.cleanup_policy", "DROP"),
					resource.TestCheckResourceAttr(testEMysqlTargetId, "settings.mysql_target.timezone", "Europe/Zurich"),
					resource.TestCheckNoResourceAttr(testEMysqlTargetId, "settings.mysql_target.object_transfer_settings.tables"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccTransferEndpointResourceMysqlConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[3]q
	name = %[1]q
	settings {
		mysql_source {
			connection {
				on_premise {
					hosts = ["leader-0.company.tech"]
					port = 3306
				}
			}
			database = "production"
			user = "dc-transfer"
			password = "foobar123"

			object_transfer_settings {
				tables = "BEFORE_DATA"
			}
		}
	}
}

resource "doublecloud_transfer_endpoint" %[2]q {
	project_id = %[3]q
	name = %[2]q
	settings {
		mysql_target {
			connection {
				on_premise {
					hosts = ["leader-0.company.tech"]
					port = 3306
				}
			}
			database = "production"
			user = "dc-transfer"
			password = "foobar123"
		}
	}
}
`, testEMysqlSourceName, testEMysqlTargetName, testProjectId)
}

func testAccTransferEndpointResourceMysqlModifiedConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[3]q
	name = %[1]q
	settings {
		mysql_source {
			connection {
				on_premise {
					hosts = ["follower-0.company.tech"]
					port = 3307
				}
			}
			database = "production"
			user = "dc-transfer"
			password = "foobar124"
			include_tables_regex = ["prod.users"]
			timezone = "Africa/Johannesburg"

			object_transfer_settings {
				tables = "AFTER_DATA"
				view = "NEVER"
			}
		}
	}
}

resource "doublecloud_transfer_endpoint" %[2]q {
	project_id = %[3]q
	name = %[2]q
	settings {
		mysql_target {
			connection {
				on_premise {
					hosts = ["leader-0.company.tech"]
					port = 3307
				}
			}
			database = "production"
			user = "dc-transfer"
			password = "foobar124"
			cleanup_policy = "DROP"
			timezone = "Europe/Zurich"
		}
	}
}
`, testEMysqlSourceName, testEMysqlTargetName, testProjectId)
}

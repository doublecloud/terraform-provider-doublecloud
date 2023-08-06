package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	testEPgSourceName string = fmt.Sprintf("%v-postgres-source", testPrefix)
	testEPgTargetName string = fmt.Sprintf("%v-postgres-target", testPrefix)

	testEPgSourceId string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testEPgSourceName)
	testEPgTargetId string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testEPgTargetName)
)

func TestAccTransferEndpointPostgresResource(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTransferEndpointResourcePostgresConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEPgSourceId, "name", testEPgSourceName),
					resource.TestCheckResourceAttr(testEPgSourceId, "settings.postgres_source.connection.on_premise.port", "5432"),
					resource.TestCheckResourceAttr(testEPgSourceId, "settings.postgres_source.database", "production"),
					resource.TestCheckResourceAttr(testEPgSourceId, "settings.postgres_source.user", "dc-transfer"),
					resource.TestCheckResourceAttr(testEPgSourceId, "settings.postgres_source.password", "foobar123"),

					resource.TestCheckResourceAttr(testEPgTargetId, "name", testEPgTargetName),
					resource.TestCheckResourceAttr(testEPgTargetId, "settings.postgres_target.connection.on_premise.port", "5432"),
					resource.TestCheckResourceAttr(testEPgTargetId, "settings.postgres_target.database", "production"),
					resource.TestCheckResourceAttr(testEPgTargetId, "settings.postgres_target.user", "dc-transfer"),
					resource.TestCheckResourceAttr(testEPgTargetId, "settings.postgres_target.password", "foobar123"),
				),
			},
			// Update and Read testing
			{
				Config: testAccTransferEndpointResourcePostgresModifiedConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEPgSourceId, "name", testEPgSourceName),
					resource.TestCheckResourceAttr(testEPgSourceId, "settings.postgres_source.connection.on_premise.port", "6432"),
					resource.TestCheckResourceAttr(testEPgSourceId, "settings.postgres_source.database", "production"),
					resource.TestCheckResourceAttr(testEPgSourceId, "settings.postgres_source.user", "dc-transfer"),
					resource.TestCheckResourceAttr(testEPgSourceId, "settings.postgres_source.password", "foobar124"),
					resource.TestCheckResourceAttr(testEPgSourceId, "settings.postgres_source.slot_byte_lag_limit", "8388608"),
					resource.TestCheckResourceAttr(testEPgSourceId, "settings.postgres_source.service_schema", "prod"),
					resource.TestCheckResourceAttr(testEPgSourceId, "settings.postgres_source.object_transfer_settings.table", "AFTER_DATA"),

					resource.TestCheckResourceAttr(testEPgTargetId, "name", testEPgTargetName),
					resource.TestCheckResourceAttr(testEPgTargetId, "settings.postgres_target.connection.on_premise.port", "6432"),
					resource.TestCheckResourceAttr(testEPgTargetId, "settings.postgres_target.database", "production"),
					resource.TestCheckResourceAttr(testEPgTargetId, "settings.postgres_target.user", "dc-transfer"),
					resource.TestCheckResourceAttr(testEPgTargetId, "settings.postgres_target.password", "foobar124"),
					resource.TestCheckResourceAttr(testEPgTargetId, "settings.postgres_target.cleanup_policy", "DROP"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccTransferEndpointResourcePostgresConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[3]q
	name = %[1]q
	settings {
		postgres_source {
			connection {
				on_premise {
					hosts = ["leader-0.company.tech"]
					port = 5432
				}
			}
			database = "production"
			user = "dc-transfer"
			password = "foobar123"
		}
	}
}

resource "doublecloud_transfer_endpoint" %[2]q {
	project_id = %[3]q
	name = %[2]q
	settings {
		postgres_target {
			connection {
				on_premise {
					hosts = ["leader-0.company.tech"]
					port = 5432
				}
			}
			database = "production"
			user = "dc-transfer"
			password = "foobar123"
		}
	}
}
`, testEPgSourceName, testEPgTargetName, testProjectId)
}

func testAccTransferEndpointResourcePostgresModifiedConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[3]q
	name = %[1]q
	settings {
		postgres_source {
			connection {
				on_premise {
					hosts = ["follower-0.company.tech"]
					port = 6432
				}
			}
			database = "production"
			user = "dc-transfer"
			password = "foobar124"
			include_tables = ["prod.users"]
			slot_byte_lag_limit = 8388608
			service_schema = "prod"

			object_transfer_settings {
				table = "AFTER_DATA"
			}
		}
	}
}

resource "doublecloud_transfer_endpoint" %[2]q {
	project_id = %[3]q
	name = %[2]q
	settings {
		postgres_target {
			connection {
				on_premise {
					hosts = ["follower-0.company.tech"]
					port = 6432
				}
			}
			database = "production"
			user = "dc-transfer"
			password = "foobar124"
			cleanup_policy = "DROP"
		}
	}
}
`, testEPgSourceName, testEPgTargetName, testProjectId)
}

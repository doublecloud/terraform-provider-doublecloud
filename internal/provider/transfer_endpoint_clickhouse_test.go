package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	testEChSourceName string = fmt.Sprintf("%v-clickhouse-source", testPrefix)
	testEChTargetName string = fmt.Sprintf("%v-clickhouse-target", testPrefix)

	testEChSourceId string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testEChSourceName)
	testEChTargetId string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testEChTargetName)
)

func TestAccTransferEndpointClickhouseResource(t *testing.T) {
	m := TransferEndpointModel{
		ProjectID: types.StringValue(testProjectId),
	}
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTransferEndpointResourceConfig(&m),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEChSourceId, "name", testEChSourceName),
					resource.TestCheckResourceAttr(testEChSourceId, "settings.clickhouse_source.connection.user", "admin"),
					resource.TestCheckResourceAttr(testEChSourceId, "settings.clickhouse_source.connection.password", "foobar123"),
					resource.TestCheckResourceAttr(testEChSourceId, "settings.clickhouse_source.connection.address.cluster_id", "cluster-foo-id"),
					resource.TestCheckResourceAttr(testEChTargetId, "name", testEChTargetName),
					resource.TestCheckResourceAttr(testEChTargetId, "settings.clickhouse_target.clickhouse_cluster_name", "production"),
					resource.TestCheckResourceAttr(testEChTargetId, "settings.clickhouse_target.connection.address.on_premise.http_port", "8443"),
				),
			},
			// Update and Read testing
			{
				Config: testAccTransferEndpointResourceConfigModified(&m),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEChSourceId, "name", testEChSourceName),
					resource.TestCheckResourceAttr(testEChSourceId, "settings.clickhouse_source.connection.user", "admin"),
					resource.TestCheckResourceAttr(testEChSourceId, "settings.clickhouse_source.connection.password", "foobar124"),
					resource.TestCheckNoResourceAttr(testEChSourceId, "settings.clickhouse_source.connection.address.cluster_id"),
					resource.TestCheckResourceAttr(testEChSourceId, "settings.clickhouse_source.connection.address.on_premise.shard.0.name", "first"),
					resource.TestCheckResourceAttr(testEChSourceId, "settings.clickhouse_source.connection.address.on_premise.tls_mode.ca_certificate", "<place-your-pem-here>"),

					resource.TestCheckResourceAttr(testEChTargetId, "name", testEChTargetName),
					resource.TestCheckResourceAttr(testEChTargetId, "settings.clickhouse_target.connection.password", "foobar124"),
					resource.TestCheckResourceAttr(testEChTargetId, "settings.clickhouse_target.clickhouse_cluster_name", "production"),
					resource.TestCheckResourceAttr(testEChTargetId, "settings.clickhouse_target.connection.address.on_premise.http_port", "8443"),
					resource.TestCheckResourceAttr(testEChTargetId, "settings.clickhouse_target.connection.address.on_premise.native_port", "9443"),
				),
			},
			// Delete testing automatically occurs in TestCase

		},
	})
}

func testAccTransferEndpointResourceConfig(m *TransferEndpointModel) string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[3]q
	name = %[1]q
	settings {
		clickhouse_source {
			connection {
				address {
					cluster_id = "cluster-foo-id"
				}
				database = "default"
				user = "admin"
				password = "foobar123"	
			}
		}
	}
}

resource "doublecloud_transfer_endpoint" %[2]q {
  project_id = %[3]q
  name = %[2]q
  settings {
	clickhouse_target {
		clickhouse_cluster_name = "production"
		alt_name {
			from_name = "foo"
			to_name = "bar"
		}
		connection {
			address {
				on_premise {
					http_port = 8443
					shard {
						name = "first"
						hosts = ["127.0.0.1", "127.0.0.2"]
					}
				}
			}
			database = "default"
			user = "admin"
			password = "foobar123"	
		}
		clickhouse_cleanup_policy = "TRUNCATE"
		}
	}
}
`, testEChSourceName, testEChTargetName, m.ProjectID.ValueString())
}

func testAccTransferEndpointResourceConfigModified(m *TransferEndpointModel) string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[3]q
	name = %[1]q
	settings {
		clickhouse_source {
			connection {
				address {
					on_premise {
						http_port = 8443
						shard {
							name = "first"
							hosts = ["127.0.0.1", "127.0.0.2"]
						}
						tls_mode {
							ca_certificate = "<place-your-pem-here>"
						}
					}	
				}
				database = "default"
				user = "admin"
				password = "foobar124"	
			}
			include_tables = ["foo", "bar"]
			exclude_tables = ["bunny", "wolf"]
		}
	}
}

resource "doublecloud_transfer_endpoint" %[2]q {
  project_id = %[3]q
  name = %[2]q
  settings {
	clickhouse_target {
		clickhouse_cluster_name = "production"
		alt_name {
			from_name = "foo"
			to_name = "bar"
		}
		connection {
			address {
				on_premise {
					http_port = 8443
					native_port = 9443
					shard {
						name = "first"
						hosts = ["127.0.0.1", "127.0.0.2"]
					}
					shard {
						name = "second"
						hosts = ["1.1.1.1", "2.2.2.2"]
					}
				}
			}
			database = "default"
			user = "admin"
			password = "foobar124"
		}
		clickhouse_cleanup_policy = "DROP"
		}
	}
}
`, testEChSourceName, testEChTargetName, m.ProjectID.ValueString())
}

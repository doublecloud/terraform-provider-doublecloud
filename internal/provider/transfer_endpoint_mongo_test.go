package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	testEMgSourceName string = fmt.Sprintf("%v-mongo-source", testPrefix)
	testEMgTargetName string = fmt.Sprintf("%v-mongo-target", testPrefix)

	testEMgSourceId string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testEMgSourceName)
	testEMgTargetId string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testEMgTargetName)
)

func TestAccTransferEndpointMongoResource(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTransferEndpointResourceMongoConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEMgSourceId, "name", testEMgSourceName),
					resource.TestCheckResourceAttr(testEMgSourceId, "settings.mongo_source.connection.connection_type.on_premise.port", "27015"),
					resource.TestCheckResourceAttr(testEMgSourceId, "settings.mongo_source.connection.auth_source", "production"),
					resource.TestCheckResourceAttr(testEMgSourceId, "settings.mongo_source.connection.user", "dc-transfer"),
					resource.TestCheckResourceAttr(testEMgSourceId, "settings.mongo_source.connection.password", "foobar123"),

					resource.TestCheckResourceAttr(testEMgTargetId, "name", testEMgTargetName),
					resource.TestCheckResourceAttr(testEMgTargetId, "settings.mongo_target.connection.connection_type.on_premise.port", "27015"),
					resource.TestCheckResourceAttr(testEMgTargetId, "settings.mongo_target.connection.auth_source", "production"),
					resource.TestCheckResourceAttr(testEMgTargetId, "settings.mongo_target.connection.user", "dc-transfer"),
					resource.TestCheckResourceAttr(testEMgTargetId, "settings.mongo_target.connection.password", "foobar123"),
				),
			},

			// Update and Read testing
			{
				Config: testAccTransferEndpointResourceMongoModifiedConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEMgSourceId, "name", testEMgSourceName),
					resource.TestCheckResourceAttr(testEMgSourceId, "settings.mongo_source.connection.connection_type.on_premise.port", "27015"),
					resource.TestCheckResourceAttr(testEMgSourceId, "settings.mongo_source.connection.auth_source", "production"),
					resource.TestCheckResourceAttr(testEMgSourceId, "settings.mongo_source.connection.user", "dc-transfer"),
					resource.TestCheckResourceAttr(testEMgSourceId, "settings.mongo_source.connection.password", "foobar124"),
					resource.TestCheckResourceAttr(testEMgSourceId, "settings.mongo_source.collection.0.database_name", "production"),
					resource.TestCheckResourceAttr(testEMgSourceId, "settings.mongo_source.collection.0.collection_name", "*"),
					resource.TestCheckResourceAttr(testEMgSourceId, "settings.mongo_source.excluded_collection.0.database_name", "backyard"),
					resource.TestCheckResourceAttr(testEMgSourceId, "settings.mongo_source.excluded_collection.0.collection_name", "queue"),

					resource.TestCheckResourceAttr(testEMgTargetId, "settings.mongo_target.connection.connection_type.on_premise.port", "27015"),
					resource.TestCheckResourceAttr(testEMgTargetId, "settings.mongo_target.connection.auth_source", "production"),
					resource.TestCheckResourceAttr(testEMgTargetId, "settings.mongo_target.connection.user", "dc-transfer"),
					resource.TestCheckResourceAttr(testEMgTargetId, "settings.mongo_target.connection.password", "foobar124"),
					resource.TestCheckResourceAttr(testEMgTargetId, "settings.mongo_target.database", "sink"),
					resource.TestCheckResourceAttr(testEMgTargetId, "settings.mongo_target.cleanup_policy", "DROP"),
				),
			},

			// Update and Read testing
			{
				Config: testAccTransferEndpointResourceMongoFinalModifiedConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEMgSourceId, "name", testEMgSourceName),
					resource.TestCheckResourceAttr(testEMgSourceId, "settings.mongo_source.connection.connection_type.srv.hostname", "follower-1.company.tech"),
					resource.TestCheckResourceAttr(testEMgSourceId, "settings.mongo_source.connection.auth_source", "production"),
					resource.TestCheckResourceAttr(testEMgSourceId, "settings.mongo_source.connection.user", "dc-transfer"),
					resource.TestCheckResourceAttr(testEMgSourceId, "settings.mongo_source.connection.password", "foobar124"),

					resource.TestCheckResourceAttr(testEMgTargetId, "settings.mongo_target.connection.connection_type.srv.hostname", "leader-0.company.tech"),
					resource.TestCheckResourceAttr(testEMgTargetId, "settings.mongo_target.connection.auth_source", "production"),
					resource.TestCheckResourceAttr(testEMgTargetId, "settings.mongo_target.connection.user", "dc-transfer"),
					resource.TestCheckResourceAttr(testEMgTargetId, "settings.mongo_target.connection.password", "foobar124"),
					resource.TestCheckResourceAttr(testEMgTargetId, "settings.mongo_target.database", "sink"),
					resource.TestCheckResourceAttr(testEMgTargetId, "settings.mongo_target.cleanup_policy", "DROP"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccTransferEndpointResourceMongoConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[3]q
	name = %[1]q
	settings {
		mongo_source {
			connection {
				connection_type {
					on_premise {
						hosts = ["leader-0.company.tech"]
						port = 27015
					}
				}	
				user = "dc-transfer"
				password = "foobar123"
				auth_source = "production"
			}
		}
	}
}

resource "doublecloud_transfer_endpoint" %[2]q {
	project_id = %[3]q
	name = %[2]q
	settings {
		mongo_target {
			connection {
				connection_type{
					on_premise {
						hosts = ["leader-0.company.tech"]
						port = 27015
					}
				}
				user = "dc-transfer"
				password = "foobar123"
				auth_source = "production"
			}
		}
	}
}
`, testEMgSourceName, testEMgTargetName, testProjectId)
}

func testAccTransferEndpointResourceMongoModifiedConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[3]q
	name = %[1]q
	settings {
		mongo_source {
			connection {
				connection_type {
					on_premise {
						hosts = ["follower-1.company.tech", "follower-2.company.tech", "leader-0.company.tech"]
						port = 27015
					}
				}
				user = "dc-transfer"
				password = "foobar124"
				auth_source = "production"
			}
			collection {
				database_name = "production"
				collection_name = "*"
			}
			excluded_collection {
				database_name = "backyard"
				collection_name = "queue"
			}
			excluded_collection {
				database_name = "backyard"
				collection_name = "dead_letter_queue"
			}
		}
	}
}

resource "doublecloud_transfer_endpoint" %[2]q {
	project_id = %[3]q
	name = %[2]q
	settings {
		mongo_target {
			connection {
				connection_type {
					on_premise {
						hosts = ["leader-0.company.tech"]
						port = 27015
					}
				}
				user = "dc-transfer"
				password = "foobar124"
				auth_source = "production"
			}
			database = "sink"
			cleanup_policy = "DROP"
		}
	}
}
`, testEMgSourceName, testEMgTargetName, testProjectId)
}

func testAccTransferEndpointResourceMongoFinalModifiedConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[3]q
	name = %[1]q
	settings {
		mongo_source {
			connection {
				connection_type {
					srv {
						hostname = "follower-1.company.tech"
					}
				}
				user = "dc-transfer"
				password = "foobar124"
				auth_source = "production"
			}
		}
	}
}

resource "doublecloud_transfer_endpoint" %[2]q {
	project_id = %[3]q
	name = %[2]q
	settings {
		mongo_target {
			connection {
				connection_type {
					srv {
						hostname = "leader-0.company.tech"
					}
				}
				user = "dc-transfer"
				password = "foobar124"
				auth_source = "production"
			}
			database = "sink"
			cleanup_policy = "DROP"
		}
	}
}
`, testEMgSourceName, testEMgTargetName, testProjectId)
}

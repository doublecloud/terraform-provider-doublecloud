package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	testEKfSourceName string = fmt.Sprintf("%v-kafka-source", testPrefix)
	testEKfTargetName string = fmt.Sprintf("%v-kafka-target", testPrefix)

	testEKfSourceId string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testEKfSourceName)
	testEKfTargetId string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testEKfTargetName)
)

func TestAccTransferEndpointKafkaResource(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTransferEndpointResourceKafkaConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEKfSourceId, "name", testEKfSourceName),
					resource.TestCheckResourceAttr(testEKfSourceId, "settings.kafka_source.topic_name", "orders"),
					resource.TestCheckResourceAttr(testEKfSourceId, "settings.kafka_source.connection.cluster_id", "cluster-foo-id"),
					resource.TestCheckResourceAttr(testEKfTargetId, "settings.kafka_target.connection.cluster_id", "cluster-foo-id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccTransferEndpointResourceKafkaModifiedConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEKfSourceId, "name", testEKfSourceName),
					resource.TestCheckResourceAttr(testEKfSourceId, "settings.kafka_source.topic_name", "orders"),

					resource.TestCheckResourceAttr(testEKfSourceId, "settings.kafka_source.connection.on_premise.tls_mode.ca_certificate", "<place-your-pem-here>"),
					resource.TestCheckResourceAttr(testEKfSourceId, "settings.kafka_source.auth.sasl.user", "sink-user"),
					resource.TestCheckResourceAttr(testEKfSourceId, "settings.kafka_source.auth.sasl.password", "foobar123"),

					resource.TestCheckResourceAttr(testEKfTargetId, "settings.kafka_target.connection.cluster_id", "cluster-foo-id"),
					resource.TestCheckResourceAttr(testEKfTargetId, "settings.kafka_target.auth.sasl.user", "sink-user"),
					resource.TestCheckResourceAttr(testEKfTargetId, "settings.kafka_target.auth.sasl.password", "foobar123"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccTransferEndpointResourceKafkaConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[3]q
	name = %[1]q
	settings {
		kafka_source {
			connection {
				cluster_id = "cluster-foo-id"
			}
			auth {
				no_auth {}
			}
			topic_name = "orders"
		}
	}
}

resource "doublecloud_transfer_endpoint" %[2]q {
	project_id = %[3]q
	name = %[2]q
	settings {
		kafka_target {
			connection {
				cluster_id = "cluster-foo-id"
			}
			auth {
				no_auth {}
			}
			topic_settings {
				topic_prefix = "prefix"
			}
			serializer {
				auto {}
			}
		}
	}
}
`, testEKfSourceName, testEKfTargetName, testProjectId)
}

func testAccTransferEndpointResourceKafkaModifiedConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[3]q
	name = %[1]q
	settings {
		kafka_source {
			connection {
				on_premise {					
					broker_urls = ["host1-az1:9091", "host2-az2:9091"]
					tls_mode {
						ca_certificate = "<place-your-pem-here>"
					}
				}
			}
			auth {
				sasl {
					user = "sink-user"
					password = "foobar123"
				}
			}
			topic_name = "orders"
		}
	}
}

resource "doublecloud_transfer_endpoint" %[2]q {
	project_id = %[3]q
	name = %[2]q
	settings {
		kafka_target {
			connection {
				cluster_id = "cluster-foo-id"
			}
			auth {
				sasl {
					user = "sink-user"
					password = "foobar123"
				}
			}
			topic_settings {
				topic_prefix = "prefix"
			}
			serializer {
				debezium {}
			}
		}
	}
}
`, testEKfSourceName, testEKfTargetName, testProjectId)
}

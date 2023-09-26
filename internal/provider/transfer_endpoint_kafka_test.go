package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTransferEndpointKafkaSource(t *testing.T) {
	testEKfSourceName := fmt.Sprintf("%v-kafka-source", testPrefix)
	testEKfSourceId := fmt.Sprintf("doublecloud_transfer_endpoint.%v", testEKfSourceName)

	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: fmt.Sprintf(`
					resource "doublecloud_transfer_endpoint" %[1]q {
						project_id = %[2]q
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
								parser {
									tskv {
										null_keys_allowed = false
										add_rest_column = false
										schema {
											fields {
												field {
													name = "f1"
													type = "int64"
													key = false
													required = false
												}
											}
										}
									}
								}
							}
						}
					}
				`, testEKfSourceName, testProjectId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEKfSourceId, "name", testEKfSourceName),
					resource.TestCheckResourceAttr(testEKfSourceId, "project_id", testProjectId),
					resource.TestCheckResourceAttr(testEKfSourceId, "settings.kafka_source.connection.cluster_id", "cluster-foo-id"),
					resource.TestCheckResourceAttr(testEKfSourceId, "settings.kafka_source.topic_name", "orders"),
					resource.TestCheckResourceAttr(testEKfSourceId, "settings.kafka_source.parser.tskv.null_keys_allowed", "false"),
					resource.TestCheckResourceAttr(testEKfSourceId, "settings.kafka_source.parser.tskv.add_rest_column", "false"),
				),
			},
			// Update and Read testing
			{
				Config: fmt.Sprintf(`
					resource "doublecloud_transfer_endpoint" %[1]q {
						project_id = %[2]q
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
										mechanism = "KAFKA_MECHANISM_SHA512"
									}
								}
								topic_name = "orders"
								parser {
									json {
										null_keys_allowed = true
										add_rest_column = true
										schema {
											fields {
												field {
													name = "f1"
													type = "int64"
													key = false
													required = false
												}
												field {
													name = "f2"
													type = "string"
													key = false
													required = false
												}
											}
										}
									}
								}
							}
						}
					}
				`, testEKfSourceName, testProjectId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEKfSourceId, "name", testEKfSourceName),
					resource.TestCheckResourceAttr(testEKfSourceId, "project_id", testProjectId),
					resource.TestCheckResourceAttr(testEKfSourceId, "settings.kafka_source.topic_name", "orders"),
					resource.TestCheckResourceAttr(testEKfSourceId, "settings.kafka_source.connection.on_premise.tls_mode.ca_certificate", "<place-your-pem-here>"),
					resource.TestCheckResourceAttr(testEKfSourceId, "settings.kafka_source.auth.sasl.user", "sink-user"),
					resource.TestCheckResourceAttr(testEKfSourceId, "settings.kafka_source.auth.sasl.password", "foobar123"),
					resource.TestCheckResourceAttr(testEKfSourceId, "settings.kafka_source.topic_name", "orders"),
					resource.TestCheckResourceAttr(testEKfSourceId, "settings.kafka_source.parser.json.null_keys_allowed", "true"),
					resource.TestCheckResourceAttr(testEKfSourceId, "settings.kafka_source.parser.json.add_rest_column", "true"),
				),
			},
			// Delete occurs automatically
		},
	})
}

func TestAccTransferEndpointKafkaTarget(t *testing.T) {
	testEKfTargetName := fmt.Sprintf("%v-kafka-target", testPrefix)
	testEKfTargetId := fmt.Sprintf("doublecloud_transfer_endpoint.%v", testEKfTargetName)

	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: fmt.Sprintf(`
					resource "doublecloud_transfer_endpoint" %[1]q {
						project_id = %[2]q
						name = %[1]q
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
				`, testEKfTargetName, testProjectId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEKfTargetId, "name", testEKfTargetName),
					resource.TestCheckResourceAttr(testEKfTargetId, "project_id", testProjectId),
					resource.TestCheckResourceAttr(testEKfTargetId, "settings.kafka_target.connection.cluster_id", "cluster-foo-id"),
					resource.TestCheckResourceAttr(testEKfTargetId, "settings.kafka_target.topic_settings.topic_prefix", "prefix"),
				),
			},
			// Update and Read testing
			{
				Config: fmt.Sprintf(`
					resource "doublecloud_transfer_endpoint" %[1]q {
						project_id = %[2]q
						name = %[1]q
						settings {
							kafka_target {
								connection {
									cluster_id = "cluster-foo-id"
								}
								auth {
									sasl {
										user = "sink-user"
										password = "foobar123"
										mechanism = "KAFKA_MECHANISM_SHA512"
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
				`, testEKfTargetName, testProjectId),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEKfTargetId, "name", testEKfTargetName),
					resource.TestCheckResourceAttr(testEKfTargetId, "project_id", testProjectId),
					resource.TestCheckResourceAttr(testEKfTargetId, "settings.kafka_target.connection.cluster_id", "cluster-foo-id"),
					resource.TestCheckResourceAttr(testEKfTargetId, "settings.kafka_target.auth.sasl.user", "sink-user"),
					resource.TestCheckResourceAttr(testEKfTargetId, "settings.kafka_target.auth.sasl.password", "foobar123"),
					resource.TestCheckResourceAttr(testEKfTargetId, "settings.kafka_target.auth.sasl.mechanism", "KAFKA_MECHANISM_SHA512"),
					resource.TestCheckResourceAttr(testEKfTargetId, "settings.kafka_target.topic_settings.topic_prefix", "prefix"),
				),
			},
			// Delete occurs automatically
		},
	})
}

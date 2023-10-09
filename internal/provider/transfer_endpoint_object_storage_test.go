package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	testEObjectStorageSourceName string = fmt.Sprintf("%v-object-storage-source", testPrefix)
	testEObjectStorageTargetName string = fmt.Sprintf("%v-object-storage-target", testPrefix)

	testEObjectStorageSourceId string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testEObjectStorageSourceName)
	testEObjectStorageTargetId string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testEObjectStorageTargetName)
)

func TestAccTransferEndpointObjectStorageResource(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTransferEndpointResourceObjectStorageConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEObjectStorageSourceId, "name", testEObjectStorageSourceName),
					resource.TestCheckResourceAttr(testEObjectStorageSourceId, "settings.object_storage_source.format.csv.delimiter", ","),
					resource.TestCheckResourceAttr(testEObjectStorageSourceId, "settings.object_storage_source.format.csv.quote_char", "\""),
					resource.TestCheckResourceAttr(testEObjectStorageSourceId, "settings.object_storage_source.format.csv.escape_char", "\\"),
					resource.TestCheckResourceAttr(testEObjectStorageSourceId, "settings.object_storage_source.format.csv.double_quote", "false"),
					resource.TestCheckResourceAttr(testEObjectStorageSourceId, "settings.object_storage_source.format.csv.newlines_in_values", "false"),
					resource.TestCheckResourceAttr(testEObjectStorageSourceId, "settings.object_storage_source.format.csv.additional_options.null_values.0", "NULL"),
					resource.TestCheckResourceAttr(testEObjectStorageSourceId, "settings.object_storage_source.format.csv.additional_options.decimal_point", ","),
					resource.TestCheckResourceAttr(testEObjectStorageSourceId, "settings.object_storage_source.format.csv.advanced_options.skip_rows", "1"),
					resource.TestCheckResourceAttr(testEObjectStorageSourceId, "settings.object_storage_source.format.csv.advanced_options.column_names.0", "Test-column-1"),
					resource.TestCheckResourceAttr(testEObjectStorageSourceId, "settings.object_storage_source.provider.bucket", "test-bucket"),
					resource.TestCheckResourceAttr(testEObjectStorageSourceId, "settings.object_storage_source.provider.use_ssl", "true"),
					resource.TestCheckResourceAttr(testEObjectStorageSourceId, "settings.object_storage_source.result_table.table_name", "test-name"),

					resource.TestCheckResourceAttr(testEObjectStorageTargetId, "name", testEObjectStorageTargetName),
					resource.TestCheckResourceAttr(testEObjectStorageTargetId, "settings.object_storage_target.output_format", "OBJECT_STORAGE_SERIALIZATION_FORMAT_CSV"),
					resource.TestCheckResourceAttr(testEObjectStorageTargetId, "settings.object_storage_target.output_encoding", "GZIP"),
					resource.TestCheckResourceAttr(testEObjectStorageTargetId, "settings.object_storage_target.connection.region", "eu-central-1"),
					resource.TestCheckResourceAttr(testEObjectStorageTargetId, "settings.object_storage_target.serializer_config.any_as_string", "false"),
				),
			},
			// Update and Read testing
			{
				Config: testAccTransferEndpointResourceObjectStorageModifiedConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEObjectStorageSourceId, "name", testEObjectStorageSourceName),
					resource.TestCheckResourceAttr(testEObjectStorageSourceId, "settings.object_storage_source.format.jsonl.newlines_in_values", "false"),
					resource.TestCheckResourceAttr(testEObjectStorageSourceId, "settings.object_storage_source.format.jsonl.unexpected_field_behavior", "UNEXPECTED_FIELD_BEHAVIOR_INFER"),
					resource.TestCheckResourceAttr(testEObjectStorageSourceId, "settings.object_storage_source.format.jsonl.block_size", "1000"),
					resource.TestCheckResourceAttr(testEObjectStorageSourceId, "settings.object_storage_source.event_source.sqs.queue_name", "test-queue"),
					resource.TestCheckResourceAttr(testEObjectStorageSourceId, "settings.object_storage_source.event_source.sqs.owner_id", "test-id"),

					resource.TestCheckResourceAttr(testEObjectStorageTargetId, "name", testEObjectStorageTargetName),
					resource.TestCheckResourceAttr(testEObjectStorageTargetId, "settings.object_storage_target.bucket", "test-bucket"),
					resource.TestCheckResourceAttr(testEObjectStorageTargetId, "settings.object_storage_target.output_format", "OBJECT_STORAGE_SERIALIZATION_FORMAT_JSON"),
					resource.TestCheckResourceAttr(testEObjectStorageTargetId, "settings.object_storage_target.output_encoding", "UNCOMPRESSED"),
					resource.TestCheckResourceAttr(testEObjectStorageTargetId, "settings.object_storage_target.connection.region", "eu-central-1"),
					resource.TestCheckResourceAttr(testEObjectStorageTargetId, "settings.object_storage_target.serializer_config.any_as_string", "false"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccTransferEndpointResourceObjectStorageConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[3]q
	name = %[1]q
	settings {
		object_storage_source {
			path_pattern = "*"
			format {
				csv {
					delimiter = ","
					quote_char = "\""
					escape_char = "\\"
					encoding = "UTF-8"
					double_quote = false
					newlines_in_values = false
					block_size = 1000
					additional_options {
						null_values = ["NULL", "0"]
						true_values = ["TRUE", "true", "t"]
						false_values = ["FALSE", "false", "f"]
						decimal_point = ","
						strings_can_be_null = true
						quoted_strings_can_be_null = true
						include_missing_columns = true
					}
					advanced_options {
						skip_rows = 1
						skip_rows_after_names = 0
						autogenerate_column_names = false
						column_names = ["Test-column-1", "Test-column-2"]
					}
				}
			}
			provider {
				bucket = "test-bucket"
				aws_access_key_id = "test-key-id"
				aws_secret_access_key = "test secret"
				path_prefix = "test-bucket-subfolder"
				endpoint = ""
				region = "eu-central-1"
				use_ssl = true
				verify_ssl_cert = false
			}
			result_table {
				table_namespace = "test-namespace"
				table_name = "test-name"
				add_system_cols = true
			}
			result_schema {
				infer {}
			}
		}
	}
}

resource "doublecloud_transfer_endpoint" %[2]q {
	project_id = %[3]q
	name = %[2]q
	settings {
		object_storage_target {
			bucket = "test-bucket"
			service_account_id = "test-id"
			output_format = "OBJECT_STORAGE_SERIALIZATION_FORMAT_CSV"
			bucket_layout = "test-layout"
			bucket_layout_timezone = "test-timezone"
			buffer_interval = "20s"
			output_encoding = "GZIP"
			connection {
				aws_access_key_id = "test-key-id"
				aws_secret_access_key = "test secret"
				endpoint = ""
				region = "eu-central-1"
				use_ssl = true
				verify_ssl_cert = false
			}
			serializer_config {
				any_as_string = false
			}
		}
	}
}
`, testEObjectStorageSourceName, testEObjectStorageTargetName, testProjectId)
}

func testAccTransferEndpointResourceObjectStorageModifiedConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[3]q
	name = %[1]q
	settings {
		object_storage_source {
			path_pattern = "*"
			format {
				jsonl {
					newlines_in_values = false
					unexpected_field_behavior = "UNEXPECTED_FIELD_BEHAVIOR_INFER"
					block_size = 1000
				}
			}
			event_source {
				sqs {
					queue_name = "test-queue"
					owner_id = "test-id"
				}
			}
			provider {
				bucket = "test-bucket"
				aws_access_key_id = "test-key-id"
				aws_secret_access_key = "test secret"
				path_prefix = "test-bucket-subfolder"
				endpoint = ""
				region = "eu-central-1"
				use_ssl = true
				verify_ssl_cert = false
			}
			result_table {
				table_namespace = "test-namespace"
				table_name = "test-name"
				add_system_cols = true
			}
			result_schema {
				data_schema {
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

resource "doublecloud_transfer_endpoint" %[2]q {
	project_id = %[3]q
	name = %[2]q
	settings {
		object_storage_target {
			bucket = "test-bucket"
			service_account_id = "test-id"
			output_format = "OBJECT_STORAGE_SERIALIZATION_FORMAT_JSON"
			bucket_layout = "test-layout"
			bucket_layout_timezone = "test-timezone"
			buffer_interval = "20s"
			output_encoding = "UNCOMPRESSED"
			connection {
				aws_access_key_id = "test-key-id"
				aws_secret_access_key = "test secret"
				endpoint = ""
				region = "eu-central-1"
				use_ssl = true
				verify_ssl_cert = false
			}
			serializer_config {
				any_as_string = false
			}
		}
	}
}
`, testEObjectStorageSourceName, testEObjectStorageTargetName, testProjectId)
}

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var (
	testES3CsvSourceName string = fmt.Sprintf("%v-s3-source-csv", testPrefix)
	testES3CsvSourceId   string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testES3CsvSourceName)

	testES3ParquetSourceName string = fmt.Sprintf("%v-s3-source-parquet", testPrefix)
	testES3ParquetSourceId   string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testES3ParquetSourceName)

	testES3AvroSourceName string = fmt.Sprintf("%v-s3-source-avro", testPrefix)
	testES3AvroSourceId   string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testES3AvroSourceName)

	testES3JsonlSourceName string = fmt.Sprintf("%v-s3-source-jsonl", testPrefix)
	testES3JsonlSourceId   string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testES3JsonlSourceName)
)

func TestAccTransferEndpointS3SourceResource(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTransferEndpointResourceS3CsvConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testES3CsvSourceId, "name", testES3CsvSourceName),
					resource.TestCheckResourceAttr(testES3CsvSourceId, "settings.s3_source.dataset", "events"),
					resource.TestCheckResourceAttr(testES3CsvSourceId, "settings.s3_source.path_pattern", "/events/"),
					resource.TestCheckResourceAttr(testES3CsvSourceId, "settings.s3_source.format.csv.delimiter", ","),
					resource.TestCheckResourceAttr(testES3CsvSourceId, "settings.s3_source.format.csv.double_quote", "true"),
				),
			},
			// Update and Read testing
			{
				Config: testAccTransferEndpointResourceS3CsvConfigModified(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testES3CsvSourceId, "name", testES3CsvSourceName),
					resource.TestCheckResourceAttr(testES3CsvSourceId, "settings.s3_source.dataset", "events"),
					resource.TestCheckResourceAttr(testES3CsvSourceId, "settings.s3_source.path_pattern", "/events/"),
					resource.TestCheckResourceAttr(testES3CsvSourceId, "settings.s3_source.format.csv.delimiter", ";"),
					resource.TestCheckResourceAttr(testES3CsvSourceId, "settings.s3_source.format.csv.double_quote", "false"),
					resource.TestCheckResourceAttr(testES3CsvSourceId, "settings.s3_source.format.csv.block_size", "1024"),
					resource.TestCheckResourceAttr(testES3CsvSourceId, "settings.s3_source.provider.endpoint", "s3.company.tech"),
					resource.TestCheckResourceAttr(testES3CsvSourceId, "settings.s3_source.provider.bucket", "cdc-production"),
					resource.TestCheckResourceAttr(testES3CsvSourceId, "settings.s3_source.provider.use_ssl", "false"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTransferEndpointResourceS3ParquetConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testES3ParquetSourceId, "name", testES3ParquetSourceName),
					resource.TestCheckResourceAttr(testES3ParquetSourceId, "settings.s3_source.dataset", "events"),
					resource.TestCheckResourceAttr(testES3ParquetSourceId, "settings.s3_source.path_pattern", "/events/"),
				),
			},
			// Update and Read testing
			{
				Config: testAccTransferEndpointResourceS3ParquetConfigModified(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testES3ParquetSourceId, "name", testES3ParquetSourceName),
					resource.TestCheckResourceAttr(testES3ParquetSourceId, "settings.s3_source.format.parquet.buffer_size", "1024"),
					resource.TestCheckResourceAttr(testES3ParquetSourceId, "settings.s3_source.format.parquet.batch_size", "100"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTransferEndpointResourceS3AvroConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testES3AvroSourceId, "name", testES3AvroSourceName),
					resource.TestCheckResourceAttr(testES3AvroSourceId, "settings.s3_source.dataset", "events"),
					resource.TestCheckResourceAttr(testES3AvroSourceId, "settings.s3_source.path_pattern", "/events/"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTransferEndpointResourceS3JsonlConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testES3JsonlSourceId, "name", testES3JsonlSourceName),
					resource.TestCheckResourceAttr(testES3JsonlSourceId, "settings.s3_source.dataset", "events"),
					resource.TestCheckResourceAttr(testES3JsonlSourceId, "settings.s3_source.path_pattern", "/events/"),
					resource.TestCheckResourceAttr(testES3JsonlSourceId, "settings.s3_source.format.jsonl.newlines_in_values", "false"),
					resource.TestCheckResourceAttr(testES3JsonlSourceId, "settings.s3_source.format.jsonl.block_size", "65536"),
				),
			},
			// Update and Read testing
			{
				Config: testAccTransferEndpointResourceS3JsonlConfigModified(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testES3JsonlSourceId, "name", testES3JsonlSourceName),
					resource.TestCheckResourceAttr(testES3JsonlSourceId, "settings.s3_source.format.jsonl.newlines_in_values", "false"),
					resource.TestCheckResourceAttr(testES3JsonlSourceId, "settings.s3_source.format.jsonl.unexpected_field_behavior", "UNEXPECTED_FIELD_BEHAVIOR_IGNORE"),
					resource.TestCheckResourceAttr(testES3JsonlSourceId, "settings.s3_source.format.jsonl.block_size", "524288"),
				),
			},

			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccTransferEndpointResourceS3CsvConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[2]q
	name = %[1]q
	settings {
		s3_source {
			dataset = "events"
			path_pattern = "/events/"
			format {
				csv {
					delimiter = ","
					double_quote = true
				}
			}
			provider {
				bucket = "cdc-production"
			}
		}
	}
}
`, testES3CsvSourceName, testProjectId)
}

func testAccTransferEndpointResourceS3CsvConfigModified() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[2]q
	name = %[1]q
	settings {
		s3_source {
			dataset = "events"
			path_pattern = "/events/"
			format {
				csv {
					delimiter = ";"
					double_quote = false
					block_size = 1024
				}
			}
			provider {
				endpoint = "s3.company.tech"
				bucket = "cdc-production"
				use_ssl = false
			}
		}
	}
}
`, testES3CsvSourceName, testProjectId)
}

func testAccTransferEndpointResourceS3ParquetConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[2]q
	name = %[1]q
	settings {
		s3_source {
			dataset = "events"
			path_pattern = "/events/"
			format {
				parquet {
					columns = ["event_id", "event_type"]
				}
			}
			provider {
				bucket = "cdc-production"
			}
		}
	}
}
`, testES3ParquetSourceName, testProjectId)
}

func testAccTransferEndpointResourceS3ParquetConfigModified() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[2]q
	name = %[1]q
	settings {
		s3_source {
			dataset = "events"
			path_pattern = "/events/"
			format {
				parquet {
					buffer_size = 1024
					columns = ["event_id", "event_type", "event_started", "event_finished", "user_id"]
					batch_size = 100
				}
			}
			provider {
				bucket = "cdc-production"
			}
		}
	}
}
`, testES3ParquetSourceName, testProjectId)
}

func testAccTransferEndpointResourceS3AvroConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[2]q
	name = %[1]q
	settings {
		s3_source {
			dataset = "events"
			path_pattern = "/events/"
			format {
				avro {}
			}
			provider {
				bucket = "cdc-production"
			}
		}
	}
}
`, testES3AvroSourceName, testProjectId)
}

func testAccTransferEndpointResourceS3JsonlConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[2]q
	name = %[1]q
	settings {
		s3_source {
			dataset = "events"
			path_pattern = "/events/"
			format {
				jsonl {
					newlines_in_values = false
					block_size = 65536
				}
			}
			provider {
				bucket = "cdc-production"
			}
		}
	}
}
`, testES3JsonlSourceName, testProjectId)
}

func testAccTransferEndpointResourceS3JsonlConfigModified() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[2]q
	name = %[1]q
	settings {
		s3_source {
			dataset = "events"
			path_pattern = "/events/"
			format {
				jsonl {
					newlines_in_values = false
					unexpected_field_behavior = "UNEXPECTED_FIELD_BEHAVIOR_IGNORE"
					block_size = 524288
				}
			}
			provider {
				bucket = "cdc-production"
			}
		}
	}
}
`, testES3JsonlSourceName, testProjectId)
}

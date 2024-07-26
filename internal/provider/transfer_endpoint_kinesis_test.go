package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTransferEndpointKinesisSource(t *testing.T) {
	t.Parallel()

	testEndpointKinesisSourceName := fmt.Sprintf("%s-kinesis-source", testPrefix)
	testEndpointKinesis := fmt.Sprintf("doublecloud_transfer_endpoint.%s", testEndpointKinesisSourceName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: fmt.Sprintf(
					`resource "doublecloud_transfer_endpoint" %[1]q {
						project_id = %[2]q
						name = %[1]q
						settings {
							kinesis_source {
								stream_name = "my-test-stream"
								region = "eu-central-1"
								aws_access_key_id = "aws_access_key_id"
								aws_secret_access_key = "aws_secret_access_key"
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
											}
										}
									}
								}
							}
						}
					}`,
					testEndpointKinesisSourceName,
					testProjectId,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEndpointKinesis, "name", testEndpointKinesisSourceName),
					resource.TestCheckResourceAttr(testEndpointKinesis, "settings.kinesis_source.region", "eu-central-1"),
					resource.TestCheckResourceAttr(testEndpointKinesis, "settings.kinesis_source.stream_name", "my-test-stream"),
					resource.TestCheckResourceAttr(testEndpointKinesis, "settings.kinesis_source.parser.json.null_keys_allowed", "true"),
					resource.TestCheckResourceAttr(testEndpointKinesis, "settings.kinesis_source.parser.json.add_rest_column", "true"),
				),
			},
			// Update and Read testing
			{
				Config: fmt.Sprintf(
					`resource "doublecloud_transfer_endpoint" %[1]q {
						project_id = %[2]q
						name = %[1]q
						settings {
							kinesis_source {
								stream_name = "my-test-stream"
								region = "eu-central-1"
								aws_access_key_id = "aws_access_key_id"
								aws_secret_access_key = "aws_secret_access_key"
								parser {
									json {
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
					}`,
					testEndpointKinesisSourceName,
					testProjectId,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEndpointKinesis, "name", testEndpointKinesisSourceName),
					resource.TestCheckResourceAttr(testEndpointKinesis, "settings.kinesis_source.parser.json.null_keys_allowed", "false"),
					resource.TestCheckResourceAttr(testEndpointKinesis, "settings.kinesis_source.parser.json.add_rest_column", "false"),
				),
			},
			// Update of Credentials does not work, because the values are not returned by API
			// Delete occurs automatically in TestCase
		},
	})
}

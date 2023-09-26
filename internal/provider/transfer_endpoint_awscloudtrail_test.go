package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTransferEndpointAwsCloudTrailSource(t *testing.T) {
	t.Parallel()

	const testEndpointResource = "doublecloud_transfer_endpoint.tte-awscloudtrail-source"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: fmt.Sprintf(
					`resource "doublecloud_transfer_endpoint" "tte-awscloudtrail-source" {
						project_id = %[1]q
						name = "tte-awscloudtrail-source"
						settings {
							aws_cloudtrail_source {
								key_id = "mykeyid"
								secret_key = "mysecretkey"
								region_name = "us-west-1"
								start_date = "2021-01-25"
							}
						}
					}`,
					testProjectId,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEndpointResource, "name", "tte-awscloudtrail-source"),
					resource.TestCheckResourceAttr(testEndpointResource, "settings.aws_cloudtrail_source.region_name", "us-west-1"),
					resource.TestCheckResourceAttr(testEndpointResource, "settings.aws_cloudtrail_source.start_date", "2021-01-25"),
				),
			},
			// Update
			{
				Config: fmt.Sprintf(
					`resource "doublecloud_transfer_endpoint" "tte-awscloudtrail-source" {
						project_id = %[1]q
						name = "tte-awscloudtrail-source"
						settings {
							aws_cloudtrail_source {
								key_id = "mykeyid"
								secret_key = "mysecretkey"
								region_name = "us-east-2"
								start_date = "2022-02-26"
							}
						}
					}`,
					testProjectId,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEndpointResource, "settings.aws_cloudtrail_source.region_name", "us-east-2"),
					resource.TestCheckResourceAttr(testEndpointResource, "settings.aws_cloudtrail_source.start_date", "2022-02-26"),
				),
			},
			// Update credentials
			{
				Config: fmt.Sprintf(
					`resource "doublecloud_transfer_endpoint" "tte-awscloudtrail-source" {
						project_id = %[1]q
						name = "tte-awscloudtrail-source"
						settings {
							aws_cloudtrail_source {
								key_id = "my_new_keyid"
								secret_key = "my_new_secretkey"
								region_name = "us-east-2"
								start_date = "2022-02-26"
							}
						}
					}`,
					testProjectId,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEndpointResource, "settings.aws_cloudtrail_source.key_id", "my_new_keyid"),
					resource.TestCheckResourceAttr(testEndpointResource, "settings.aws_cloudtrail_source.secret_key", "my_new_secretkey"),
				),
			},
			// Delete occurs automatically
		},
	})
}

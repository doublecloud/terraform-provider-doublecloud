package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

var (
	testEBigquerySourceName string = fmt.Sprintf("%v-bigquery-source", testPrefix)
	testEBigqueryTargetName string = fmt.Sprintf("%v-bigquery-target", testPrefix)

	testEBigquerySourceId string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testEBigquerySourceName)
	testEBigqueryTargetId string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testEBigqueryTargetName)
)

func TestAccTransferEndpointBigqueryResource(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTransferEndpointResourceBigqueryConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEBigquerySourceId, "name", testEBigquerySourceName),
					resource.TestCheckResourceAttr(testEBigquerySourceId, "settings.bigquery_source.project_id", "project_123"),
					resource.TestCheckResourceAttr(testEBigquerySourceId, "settings.bigquery_source.dataset_id", "dataset_231"),
					resource.TestCheckResourceAttr(testEBigquerySourceId, "settings.bigquery_source.credentials_json", "my_credentials_json"),

					resource.TestCheckResourceAttr(testEBigqueryTargetId, "name", testEBigqueryTargetName),
					resource.TestCheckResourceAttr(testEBigqueryTargetId, "settings.bigquery_target.project_id", "project_123"),
					resource.TestCheckResourceAttr(testEBigqueryTargetId, "settings.bigquery_target.dataset_id", "dataset_231"),
					resource.TestCheckResourceAttr(testEBigqueryTargetId, "settings.bigquery_target.credentials_json", "my_credentials_json"),
				),
			},

			// Update and Read testing
			{
				Config: testAccTransferEndpointResourceBigqueryModifiedConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEBigquerySourceId, "name", testEBigquerySourceName),
					resource.TestCheckResourceAttr(testEBigquerySourceId, "settings.bigquery_source.project_id", "project_12"),
					resource.TestCheckResourceAttr(testEBigquerySourceId, "settings.bigquery_source.dataset_id", "dataset_23"),
					resource.TestCheckResourceAttr(testEBigquerySourceId, "settings.bigquery_source.credentials_json", "my_new_credentials_json"),

					resource.TestCheckResourceAttr(testEBigqueryTargetId, "name", testEBigqueryTargetName),
					resource.TestCheckResourceAttr(testEBigqueryTargetId, "settings.bigquery_target.project_id", "project_12"),
					resource.TestCheckResourceAttr(testEBigqueryTargetId, "settings.bigquery_target.dataset_id", "dataset_23"),
					resource.TestCheckResourceAttr(testEBigqueryTargetId, "settings.bigquery_target.credentials_json", "my_new_credentials_json"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccTransferEndpointResourceBigqueryConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[3]q
	name = %[1]q
	settings {
		bigquery_source {
			project_id = "project_123"
			dataset_id = "dataset_231"
			credentials_json = "my_credentials_json"
		}
	}
}

resource "doublecloud_transfer_endpoint" %[2]q {
	project_id = %[3]q
	name = %[2]q
	settings {
		bigquery_target {
			project_id = "project_123"
			dataset_id = "dataset_231"
			credentials_json = "my_credentials_json"
		}
	}
}
`, testEBigquerySourceName, testEBigqueryTargetName, testProjectId)
}

func testAccTransferEndpointResourceBigqueryModifiedConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[3]q
	name = %[1]q
	settings {
		bigquery_source {
			project_id = "project_12"
			dataset_id = "dataset_23"
			credentials_json = "my_new_credentials_json"
		}
	}
}

resource "doublecloud_transfer_endpoint" %[2]q {
	project_id = %[3]q
	name = %[2]q
	settings {
		bigquery_target {
			project_id = "project_12"
			dataset_id = "dataset_23"
			credentials_json = "my_new_credentials_json"
		}
	}
}
`, testEBigquerySourceName, testEBigqueryTargetName, testProjectId)
}

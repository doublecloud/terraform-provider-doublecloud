package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

var (
	testEMetrikaSourceName string = fmt.Sprintf("%v-metrika-source", testPrefix)
	testEMetrikaSourceID   string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testEMetrikaSourceName)
)

func TestAccTransferEndpointMetrikaSource(t *testing.T) {
	//t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read Testing
			{
				Config: testAccTransferEndpointMetrikaConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEMetrikaSourceID, "name", testEMetrikaSourceName),
					resource.TestCheckResourceAttr(testEMetrikaSourceID, "settings.metrika_source.counter_ids.#", "2"),
					resource.TestCheckResourceAttr(testEMetrikaSourceID, "settings.metrika_source.counter_ids.0", "1"),
					resource.TestCheckResourceAttr(testEMetrikaSourceID, "settings.metrika_source.counter_ids.1", "2"),
					resource.TestCheckResourceAttr(testEMetrikaSourceID, "settings.metrika_source.token", "randomToken"),
					resource.TestCheckResourceAttr(testEMetrikaSourceID, "settings.metrika_source.streams.#", "1"),
					resource.TestCheckResourceAttr(testEMetrikaSourceID, "settings.metrika_source.streams.0.stream_type", "METRIKA_STREAM_TYPE_HITS_V2"),
					resource.TestCheckResourceAttr(testEMetrikaSourceID, "settings.metrika_source.streams.0.columns.#", "2"),
					resource.TestCheckResourceAttr(testEMetrikaSourceID, "settings.metrika_source.streams.0.columns.0", "column1"),
					resource.TestCheckResourceAttr(testEMetrikaSourceID, "settings.metrika_source.streams.0.columns.1", "column2"),
				),
			},
			// Update and Read Testing
			{
				Config: testAccTransferEndpointMetrikaModifiedConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEMetrikaSourceID, "name", testEMetrikaSourceName),
					resource.TestCheckResourceAttr(testEMetrikaSourceID, "settings.metrika_source.counter_ids.#", "2"),
					resource.TestCheckResourceAttr(testEMetrikaSourceID, "settings.metrika_source.counter_ids.0", "3"),
					resource.TestCheckResourceAttr(testEMetrikaSourceID, "settings.metrika_source.counter_ids.1", "4"),
					resource.TestCheckResourceAttr(testEMetrikaSourceID, "settings.metrika_source.token", "modifiedToken"),
					resource.TestCheckResourceAttr(testEMetrikaSourceID, "settings.metrika_source.streams.#", "1"),
					resource.TestCheckResourceAttr(testEMetrikaSourceID, "settings.metrika_source.streams.0.stream_type", "METRIKA_STREAM_TYPE_VISITS"),
					resource.TestCheckResourceAttr(testEMetrikaSourceID, "settings.metrika_source.streams.0.columns.#", "2"),
					resource.TestCheckResourceAttr(testEMetrikaSourceID, "settings.metrika_source.streams.0.columns.0", "column1"),
					resource.TestCheckResourceAttr(testEMetrikaSourceID, "settings.metrika_source.streams.0.columns.1", "column2"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})

}

func testAccTransferEndpointMetrikaConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[2]q
	name       = %[1]q
	settings {
		metrika_source {
			counter_ids = [1, 2]
			token       = "randomToken"
			metrika_stream {
				stream_type = "METRIKA_STREAM_TYPE_HITS_V2"
				columns     = ["column1", "column2"]
			}
		}
	}
}
`, testEMetrikaSourceName, testProjectId)
}

func testAccTransferEndpointMetrikaModifiedConfig() string {
	return fmt.Sprintf(`
resource "doublecloud_transfer_endpoint" %[1]q {
	project_id = %[2]q
	name       = %[1]q
	settings {
		metrika_source {
			counter_ids = [3, 4]
			token       = "modifiedToken"
			metrika_stream {
				stream_type = "METRIKA_STREAM_TYPE_VISITS"
				columns     = ["column1", "column2"]
			}
		}
	}
}
`, testEMetrikaSourceName, testProjectId)
}

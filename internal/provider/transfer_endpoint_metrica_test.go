package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

var (
	testEMetricaSourceName string = fmt.Sprintf("%v-metrica-source", testPrefix)
	testEMetricaSourceID   string = fmt.Sprintf("doublecloud_transfer_endpoint.%v", testEMetricaSourceName)
)

func TestAccTransferEndpointMetrikaSource(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read Testing
			{
				Config: testAccTransferEndpointMetrikaConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testEMetricaSourceID, "name", testEMetricaSourceName),
					resource.TestCheckResourceAttr(testEMetricaSourceID, "settings.metrica_source.counter_ids.#", "2"),
					resource.TestCheckResourceAttr(testEMetricaSourceID, "settings.metrica_source.counter_ids.0", "1"),
					resource.TestCheckResourceAttr(testEMetricaSourceID, "settings.metrica_source.counter_ids.1", "2"),
					resource.TestCheckResourceAttr(testEMetricaSourceID, "settings.metrica_source.token", "randomToken"),
					resource.TestCheckResourceAttr(testEMetricaSourceID, "settings.metrica_source.metrica_stream.#", "1"),
					resource.TestCheckResourceAttr(testEMetricaSourceID, "settings.metrica_source.metrica_stream.0.stream_type", "METRICA_STREAM_TYPE_HITS_V2"),
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
		metrica_source {
			counter_ids = [1, 2]
			token       = "randomToken"
			metrica_stream {
				stream_type = "METRICA_STREAM_TYPE_HITS_V2"
			}
		}
	}
}
`, testEMetricaSourceName, testProjectId)
}

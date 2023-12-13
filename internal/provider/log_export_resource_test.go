package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLogExportResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccLogExportResourceConfig("e2e-test-export", "123", "DATADOG_HOST_DATADOGHQ"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("doublecloud_log_export.data-dog", "project_id", testProjectId),
					resource.TestCheckResourceAttr("doublecloud_log_export.data-dog", "name", "e2e-test-export"),
					resource.TestCheckResourceAttr("doublecloud_log_export.data-dog", "datadog.api_key", "123"),
					resource.TestCheckResourceAttr("doublecloud_log_export.data-dog", "datadog.datadog_host", "DATADOG_HOST_DATADOGHQ"),
				),
			},
		},
	})
}

func testAccLogExportResourceConfig(name, apiKey, host string) string {
	return fmt.Sprintf(`
data "doublecloud_clickhouse" "ch_for_logs" {
	name = %[5]q
	project_id = %[1]q
}

resource "doublecloud_log_export" "data-dog" {
  project_id = %[1]q
  name = %[2]q

  sources = [{
    type="LOG_SOURCE_TYPE_CLICKHOUSE"
    id=data.doublecloud_clickhouse.ch_for_logs.id
  }]

  datadog = {
    api_key = %[3]q
    datadog_host =  %[4]q
  }
}`, testProjectId, name, apiKey, host, testClickhouseName)
}

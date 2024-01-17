package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

//nolint:govet
func TestAccWorkbookResource(t *testing.T) {
	return
	t.Parallel()
	m := WorkbookResourceModel{
		ProjectID: types.StringValue(testProjectId),
		Title:     types.StringValue("tf-acc-test"),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccWorkbookResourceConfig(&m),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("doublecloud_workbook.test", "title", m.Title.ValueString()),
				),
			},
			// Update and Read testing
			{
				Config: testAccWorkbookResourceConfigModified(&m),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("doublecloud_workbook.test", "title", m.Title.ValueString()),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccWorkbookResourceConfig(m *WorkbookResourceModel) string {
	return fmt.Sprintf(`
resource "doublecloud_workbook" "test" {
  project_id = %[1]q
  title = %[2]q

  config = %[3]q

  connect {
	name = "conn_ch_1"
	config = jsonencode({
		kind = "clickhouse"
		cache_ttl_sec = 600
		host = "rw.chcpbbeap8lpuv24hhh4.at.double.cloud"
		port = 8443
		username = "examples_user"
		secure = true
		raw_sql_level = "off"
	})
	secret = "yahj@ah5foth"
  }

}
`, m.ProjectID.ValueString(),
		m.Title.ValueString(),
		exampleWorkbook())
}

func testAccWorkbookResourceConfigModified(m *WorkbookResourceModel) string {
	return fmt.Sprintf(`
resource "doublecloud_workbook" "test" {
  project_id = %[1]q
  title = %[2]q

  config = %[3]q

  connect {
	name = "conn_ch_1"
	config = jsonencode({
		kind = "clickhouse"
		cache_ttl_sec = 900
		host = "rw.chcpbbeap8lpuv24hhh4.at.double.cloud"
		port = 8443
		username = "examples_user"
		secure = true
		raw_sql_level = "off"
	})
	secret = "yahj@ah5foth"
  }

  connect {
	name = "conn_ch_2"
	config = jsonencode({
		kind = "clickhouse"
		cache_ttl_sec = 420
		host = "rw.chcpbbeap8lpuv24hhh4.at.double.cloud"
		port = 8443
		username = "examples_user"
		secure = true
		raw_sql_level = "off"
	})
	secret = "yahj@ah5foth"
  }
}
`, m.ProjectID.ValueString(),
		m.Title.ValueString(),
		exampleWorkbook())

}

func exampleWorkbook() string {
	return `
    {
        "datasets": [
        {
            "name": "ds_hits_sample",
            "dataset": {
                "fields": [
                {
                    "description": null,
                    "id": "hit_id",
                    "cast": "integer",
                    "calc_spec": {
                        "kind": "direct",
                        "avatar_id": "production_marts",
                        "field_name": "Hit_ID"
                    },
                    "aggregation": "none",
                    "title": "Hit_ID",
                    "hidden": false
                },
                {
                    "description": null,
                    "id": "date",
                    "cast": "date",
                    "title": "Date",
                    "aggregation": "none",
                    "calc_spec": {
                        "kind": "direct",
                        "avatar_id": "production_marts",
                        "field_name": "Date"
                    },
                    "hidden": false
                },
                {
                    "description": null,
                    "id": "time_spent",
                    "cast": "float",
                    "calc_spec": {
                        "kind": "direct",
                        "avatar_id": "production_marts",
                        "field_name": "Time_Spent"
                    },
                    "hidden": false,
                    "aggregation": "none",
                    "title": "Time_Spent"
                },
                {
                    "description": null,
                    "id": "cookie_enabled",
                    "cast": "integer",
                    "calc_spec": {
                        "kind": "direct",
                        "avatar_id": "production_marts",
                        "field_name": "Cookie_Enabled"
                    },
                    "hidden": false,
                    "title": "Cookie_Enabled",
                    "aggregation": "none"
                },
                {
                    "description": null,
                    "id": "region_id",
                    "cast": "integer",
                    "title": "Region_ID",
                    "hidden": false,
                    "calc_spec": {
                        "kind": "direct",
                        "avatar_id": "production_marts",
                        "field_name": "Region_ID"
                    },
                    "aggregation": "none"
                },
                {
                    "description": null,
                    "id": "gender",
                    "cast": "string",
                    "calc_spec": {
                        "kind": "direct",
                        "avatar_id": "production_marts",
                        "field_name": "Gender"
                    },
                    "hidden": false,
                    "title": "Gender",
                    "aggregation": "none"
                },
                {
                    "description": null,
                    "id": "browser",
                    "cast": "string",
                    "calc_spec": {
                        "kind": "direct",
                        "avatar_id": "production_marts",
                        "field_name": "Browser"
                    },
                    "aggregation": "none",
                    "hidden": false,
                    "title": "Browser"
                },
                {
                    "description": null,
                    "id": "traffic_source",
                    "cast": "string",
                    "calc_spec": {
                        "kind": "direct",
                        "avatar_id": "production_marts",
                        "field_name": "Traffic_Source"
                    },
                    "aggregation": "none",
                    "hidden": false,
                    "title": "Traffic_Source"
                },
                {
                    "description": null,
                    "id": "technology",
                    "cast": "string",
                    "calc_spec": {
                        "kind": "direct",
                        "avatar_id": "production_marts",
                        "field_name": "Technology"
                    },
                    "hidden": false,
                    "aggregation": "none",
                    "title": "Technology"
                },
                {
                    "description": null,
                    "id": "hits_count",
                    "cast": "integer",
                    "calc_spec": {
                        "kind": "direct",
                        "field_name": "Hit_ID"
                    },
                    "aggregation": "countunique",
                    "hidden": false,
                    "title": "Total hits"
                }],
            "avatars": null,
            "sources": [
                {
                    "title": "hits_sample",
                    "connection_ref": "conn_ch_1",
                    "id": "production_marts",
                    "spec": {
                        "db_name": "examples",
                        "kind": "sql_table",
                        "table_name": "hits"
                    }
                }
                ]
            }
            }
        ],
        "charts": [],
        "dashboards": []
    }`
	// an_indicator_chart(), a_donut_chart(), a_column_chart())
	// , a_column_chart())
}

//nolint:unused
func an_indicator_chart() string {
	// Example of an indicator chart
	// See https://double.cloud/docs/en/data-visualization/quickstart#create-an-indicator
	return `{
        "chart": {
            "ad_hoc_fields": [],
            "visualization": {
                "field": {
                    "source": {
                        "kind": "ref",
                        "id": "hits_count"
                    }
                },
                "kind": "indicator"
            },
            "datasets": ["ds_hits_sample"]
        },
        "name": "chart_total_hits"
    }`
}

//nolint:unused
func a_donut_chart() string {
	// Example of a donut chart
	// See https://double.cloud/docs/en/data-visualization/quickstart#create-a-donut-chart
	return `{
        "chart": {
            "ad_hoc_fields": [
                {
                    "field": {
                        "description": null,
                        "id": "browser_count_unique",
                        "cast": "string",
                        "title": "Browser Count Unique",
                        "aggregation": "countunique",
                        "calc_spec": {"kind": "direct", "avatar_id": "production_marts", "field_name": "Browser"},
                        "hidden": false
                    }
                }
            ],
            "visualization": {
                "kind": "donut_chart",
                "sort": [],
                "coloring": {
                    "mounts": [],
                    "palette_id": null,
                    "source": {"kind": "ref", "id": "technology"}
                },
                "measures": {
                    "source": {
                        "kind": "ref",
                        "id": "browser_count_unique"
                    }
                }
            },
            "datasets": ["ds_hits_sample"]
        },
        "name": "chart_user_share_by_platform"
    }`
}

//nolint:unused
func a_column_chart() string {
	// This is an example of a column chart
	// See https://double.cloud/docs/en/data-visualization/quickstart#create-a-column-chart
	// return ``
	return `{
        "chart": {
            "ad_hoc_fields": [
                {
                    "field": {
                        "description": null,
                        "id": "time_spent_sum",
                        "cast": "float",
                        "calc_spec": {
                            "kind": "id_formula",
                            "formula": "SUM([Time_Spent])"
                        },
                        "aggregation": "none",
                        "hidden": false,
                        "title": "Time Spent Sum"
                    },
                    "dataset_name": "ds_hits_sample"
                }
            ],
            "visualization": {
                "y": [
                    {
                        "source": {
                            "kind": "ref",
                            "id": "time_spent_sum"
                        }
                    }
                ],
                "kind": "column_chart",
                "sort": [],
                "coloring": {
                    "mounts": [],
                    "palette_id": null,
                    "kind": "dimension",
                    "source": {"kind": "ref", "id": "browser"}
                },
                "x": [{"source": {"kind": "ref", "id": "date"}}]
            },
            "datasets": ["ds_hits_sample"]
        },
        "name": "chart_time_spent_per_browser"
    }`
}

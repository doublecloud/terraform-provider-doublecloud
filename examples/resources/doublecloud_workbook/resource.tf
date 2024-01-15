resource "doublecloud_workbook" "example_workbook" {
  project_id = var.dc_project_id
  title      = "Example workbook"

  config = jsonencode({
    "datasets" : [],
    "charts" : [],
    "dashboards" : []
  })

  connect {
    name = "main"
    config = jsonencode({
      kind          = "clickhouse"
      cache_ttl_sec = 600
      host          = data.doublecloud_clickhouse.dwh.connection_info.host
      port          = 8443
      username      = data.doublecloud_clickhouse.dwh.connection_info.user
      secure        = true
      raw_sql_level = "off"
    })
    secret = data.doublecloud_clickhouse.dwh.connection_info.password
  }
}
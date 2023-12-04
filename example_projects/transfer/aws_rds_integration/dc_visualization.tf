// Add a connection for to Clickhouse Cluster
// With this connection we can later add datasets / charts / dashboard
resource "doublecloud_workbook" "dwh-viewer" {
  project_id = var.dc_project_id
  title      = "DWH Viewer"

  config = jsonencode({ // Empty for now
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

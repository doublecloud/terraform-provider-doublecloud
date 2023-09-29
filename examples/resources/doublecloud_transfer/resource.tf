resource "doublecloud_transfer" "sample-pg2ch" {
  name = "sample-pg2ch"
  project_id = var.project_id
  source = doublecloud_transfer_endpoint.sample-pg2ch-source.id
  target = doublecloud_transfer_endpoint.sample-pg2ch-target.id
  type = "SNAPSHOT_ONLY"
  activated = false
  transformation = {
    transformers = [
      {
        dbt = {
          git_repository_link = "https://github.com/doublecloud/tests-clickhouse-dbt.git"
          profile_name = "my_clickhouse_profile"
          operation = "run"
        }
      },
    ]
  }
}

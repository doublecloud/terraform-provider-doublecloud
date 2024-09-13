resource "doublecloud_airflow_cluster" "example-airflow" {
  project_id = var.project_id
  name       = "example-airflow"
  region_id  = "eu-central-1"
  cloud_type = "aws"
  network_id = data.doublecloud_network.default.id

  resources {
    airflow {
      max_worker_count   = 1
      min_worker_count   = 1
      environment_flavor = "dev_test"
      worker_concurrency = 16
      worker_disk_size   = 10
      worker_preset      = "small"
    }
  }

  config {
    version_id = "2.9.0"
    sync_config {
      repo_url  = "https://github.com/apache/airflow"
      branch    = "main"
      dags_path = "airflow/example_dags"
    }
  }
}

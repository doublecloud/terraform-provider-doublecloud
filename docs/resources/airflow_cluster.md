---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "doublecloud_airflow_cluster Resource - terraform-provider-doublecloud"
subcategory: ""
description: |-
  Airflow Cluster resource
---

# doublecloud_airflow_cluster (Resource)

Airflow Cluster resource

## Example Usage

```terraform
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

  access {
    data_services = ["transfer"]
    ipv4_cidr_blocks = [
      {
        value       = "10.0.0.0/8"
        description = "Office in Berlin"
      }
    ]
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cloud_type` (String) Cloud provider (`aws`)
- `name` (String) Cluster name
- `network_id` (String) Cluster network ID
- `project_id` (String) Project ID
- `region_id` (String) Region where the cluster is located

### Optional

- `access` (Block, Optional) Access control configuration (see [below for nested schema](#nestedblock--access))
- `config` (Block, Optional) Cluster configuration (see [below for nested schema](#nestedblock--config))
- `description` (String) Cluster description
- `resources` (Block, Optional) Cluster resources (see [below for nested schema](#nestedblock--resources))

### Read-Only

- `connection_info` (Attributes) Public connection info (see [below for nested schema](#nestedatt--connection_info))
- `cr_connection_info` (Attributes) Remote connection info (see [below for nested schema](#nestedatt--cr_connection_info))
- `id` (String) Cluster ID

<a id="nestedblock--access"></a>
### Nested Schema for `access`

Optional:

- `data_services` (List of String) List of allowed services
- `ipv4_cidr_blocks` (Attributes List) IPv4 CIDR blocks (see [below for nested schema](#nestedatt--access--ipv4_cidr_blocks))
- `ipv6_cidr_blocks` (Attributes List) IPv6 CIDR blocks (see [below for nested schema](#nestedatt--access--ipv6_cidr_blocks))

<a id="nestedatt--access--ipv4_cidr_blocks"></a>
### Nested Schema for `access.ipv4_cidr_blocks`

Required:

- `value` (String) CIDR block

Optional:

- `description` (String) CIDR block description


<a id="nestedatt--access--ipv6_cidr_blocks"></a>
### Nested Schema for `access.ipv6_cidr_blocks`

Required:

- `value` (String) CIDR block

Optional:

- `description` (String) CIDR block description



<a id="nestedblock--config"></a>
### Nested Schema for `config`

Required:

- `version_id` (String) Airflow cluster version ID

Optional:

- `airflow_env_variable` (Block List) Environment variables (see [below for nested schema](#nestedblock--config--airflow_env_variable))
- `custom_image_digest` (String) Custom Airflow image digest
- `managed_requirements_txt` (String) Path to the managed `requirements.txt` file
- `sync_config` (Block, Optional) DAG repository configuration (see [below for nested schema](#nestedblock--config--sync_config))
- `user_service_account` (String) Service account for the Airflow cluster

<a id="nestedblock--config--airflow_env_variable"></a>
### Nested Schema for `config.airflow_env_variable`

Optional:

- `name` (String) Environment variable name
- `value` (String) Environment variable value


<a id="nestedblock--config--sync_config"></a>
### Nested Schema for `config.sync_config`

Required:

- `branch` (String) DAG repository branch name
- `dags_path` (String) Path to the directory with DAGs
- `repo_url` (String) DAG repository URL

Optional:

- `credentials` (Block, Optional) DAG repository credentials (see [below for nested schema](#nestedblock--config--sync_config--credentials))
- `revision` (String) DAG repository revision

<a id="nestedblock--config--sync_config--credentials"></a>
### Nested Schema for `config.sync_config.credentials`

Optional:

- `api_credentials` (Block, Optional) API credentials for accessing the DAG repository (see [below for nested schema](#nestedblock--config--sync_config--credentials--api_credentials))

<a id="nestedblock--config--sync_config--credentials--api_credentials"></a>
### Nested Schema for `config.sync_config.credentials.api_credentials`

Optional:

- `password` (String, Sensitive) Password
- `username` (String) Username





<a id="nestedblock--resources"></a>
### Nested Schema for `resources`

Optional:

- `airflow` (Block, Optional) (see [below for nested schema](#nestedblock--resources--airflow))

<a id="nestedblock--resources--airflow"></a>
### Nested Schema for `resources.airflow`

Required:

- `environment_flavor` (String) Environment configuration
- `max_worker_count` (Number) Maximum number of workers
- `min_worker_count` (Number) Minimum number of workers
- `worker_concurrency` (Number) Worker concurrency
- `worker_disk_size` (Number) Worker disk size
- `worker_preset` (String) Worker resource preset



<a id="nestedatt--connection_info"></a>
### Nested Schema for `connection_info`

Read-Only:

- `host` (String) Webserver URL
- `password` (String, Sensitive) Password for the Airflow user
- `user` (String) Airflow user


<a id="nestedatt--cr_connection_info"></a>
### Nested Schema for `cr_connection_info`

Read-Only:

- `host` (String) host to use in clients
- `password` (String, Sensitive) Password for the Airflow user
- `user` (String) Airflow user
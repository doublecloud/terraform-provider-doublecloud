# Example terraform to create a transfer from PostgreSQL to ClickHouse

The example in this directory illustrates how a transfer from PostgreSQL to ClickHouse can be created using DoubleCloud public Terraform.

Before running `terraform apply`, replace the variables and values in the Terraform files with the appropriate ones to enable access to databases in your environment.

## Contents

### `versions.tf`

Contains definition of the DoubleCloud public Terraform provider and the following important block:

```ini
provider "doublecloud" {
  authorized_key = file("authorized_key.json")
}
```

The `authorized_key.json` is an authentication file required to use a service account. See [DoubleCloud documentation](https://double.cloud/docs/en/public-api/tutorials/transfer-api-quickstart) for the instructions on how to obtain this file.

### `variables.tf`

Contains several Terraform variables. See their descriptions and usage for details.

### `transfer.tf`

Contains definitions of three resources:

* `"doublecloud_transfer_endpoint" "sample-pg2ch-source"` is the source endpoint
    * `hosts` and other parameters should be replaced with the actual values enabling access to a PostgreSQL instance. The instance itself should be created separately, for example, using [AWS RDS provider](https://registry.terraform.io/modules/terraform-aws-modules/rds/aws/latest), or manually.
* `"doublecloud_transfer_endpoint" "sample-pg2ch-target"` is the target endpoint
    * `address` and other connection parameters should be replaced with the values enabling access to a ClickHouse instance. The example uses a managed ClickHouse instance. Such an instance can be created using a `doublecloud_clickhouse_cluster` resource or manually. This is better be made separately in advance, because ClickHouse cluster creation takes long time, and the credentials to access a newly-created cluster should be obtained manually (they are not displayed in Terraform for security reasons).
* `"doublecloud_transfer" "sample-pg2ch"` is the transfer itself
    * Set `activated` to `true` to activate a transfer automatically on creation.

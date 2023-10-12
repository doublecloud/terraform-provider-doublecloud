# Example terraform to create a ClickHouse and Apache Kafka in BYOC

The example in this directory illustrates how to create ClickHouse and Apache Kafka with BYOC in user's AWS account with configured KafkaEngine in Clickhouse.

Before running `terraform apply`, replace the variables and values in the Terraform files with the appropriate ones to enable access to databases in your environment.

For download all providers and modules run `terraform init`.


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

### `main.tf`

Contains definitions of three resources:

* `"doublecloud_network" "aws"` is the doublecloud network which is connected to BYOC AWS network
* `"doublecloud_clickhouse_cluster" "alpha-clickhouse"` is the main ClickHouse cluster which located in your AWS profile 
* `"doublecloud_kafka_cluster" "alpha-kafka"` is the Kafka which connected to ClickHouse through `config.kafka` block
